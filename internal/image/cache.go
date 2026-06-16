package image

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/pauvalls/grimorio/internal/cache"
)

// fieldSep is the US (Unit Separator, 0x1F) byte used to join cache-key
// fields. It can never appear in a normal text prompt or model name, so
// it safely delimits the fields inside the SHA-256 input.
const fieldSep = "\x1f"

// computeKey returns a deterministic 64-char hex SHA-256 hash of
// (prompt || model || width || height || seed || provider) joined with
// the US (0x1F) byte. Any change in any field produces a different key.
func computeKey(prompt, provider, model string, w, height, seed int) string {
	h := sha256.New()
	// Order matters and is part of the public contract: tests assert it.
	_, _ = h.Write([]byte(prompt))
	_, _ = h.Write([]byte(fieldSep))
	_, _ = h.Write([]byte(model))
	_, _ = h.Write([]byte(fieldSep))
	_, _ = fmt.Fprintf(h, "%d", w)
	_, _ = h.Write([]byte(fieldSep))
	_, _ = fmt.Fprintf(h, "%d", height)
	_, _ = h.Write([]byte(fieldSep))
	_, _ = fmt.Fprintf(h, "%d", seed)
	_, _ = h.Write([]byte(fieldSep))
	_, _ = h.Write([]byte(provider))
	return hex.EncodeToString(h.Sum(nil))
}

// ImageCache is a two-tier cache for generated image bytes:
//   - L0: thread-safe in-memory LRU (default capacity 50)
//   - L1: on-disk sharded mirror at <diskDir>/v1/<key[:2]>/<key>.bin
//
// Get() reads L0 first, then L1 (promoting the value into L0 on a hit).
// Put() writes to both layers. The on-disk file is a byte-preserving
// copy of the generated image, allowing the cache to survive process
// restarts. Concurrent same-key writes are last-writer-wins with
// identical bytes (the key fixes the inputs).
type ImageCache struct {
	lru     *cache.LRUCache[string, []byte]
	diskDir string

	// mu guards disk directory creation. Per-key writes are unique
	// paths so they can proceed concurrently; the LRU has its own lock.
	mu sync.RWMutex
}

// NewImageCache creates a new cache rooted at diskDir. memCap is the
// maximum number of in-memory entries; disk entries are unbounded.
func NewImageCache(diskDir string, memCap int) (*ImageCache, error) {
	if memCap <= 0 {
		memCap = 50
	}
	if err := os.MkdirAll(diskDir, 0o755); err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}
	return &ImageCache{
		lru:     cache.NewLRU[string, []byte](memCap),
		diskDir: diskDir,
	}, nil
}

// ComputeKey is the public re-export of the package-level computeKey so
// other packages (e.g. services) can compute the same key without
// duplicating the field-separator contract.
func ComputeKey(prompt, provider, model string, w, height, seed int) string {
	return computeKey(prompt, provider, model, w, height, seed)
}

// Get returns cached bytes for key (searching memory, then disk).
// On a disk hit the value is promoted into the in-memory LRU.
func (c *ImageCache) Get(key string) ([]byte, bool) {
	if v, ok := c.lru.Get(key); ok {
		return v, true
	}
	path := c.diskPath(key)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	// Promote to LRU
	c.lru.Put(key, data)
	return data, true
}

// Put stores bytes in both the in-memory LRU and on disk.
func (c *ImageCache) Put(key string, data []byte) {
	c.lru.Put(key, data)
	_ = c.writeDisk(key, data)
}

// Len returns the number of in-memory entries (LRU length).
func (c *ImageCache) Len() int { return c.lru.Len() }

// diskPath returns the sharded on-disk location for a given key.
func (c *ImageCache) diskPath(key string) string {
	var shard string
	if len(key) >= 2 {
		shard = key[:2]
	} else if len(key) == 1 {
		shard = key
	} else {
		shard = "_"
	}
	return filepath.Join(c.diskDir, "v1", shard, key+".bin")
}

func (c *ImageCache) writeDisk(key string, data []byte) error {
	path := c.diskPath(key)
	c.mu.RLock()
	dir := filepath.Dir(path)
	c.mu.RUnlock()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create shard dir: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write cache file: %w", err)
	}
	return nil
}

// CachingProviderConfig carries the cache-key fields for a CachingProvider.
// It mirrors the relevant image.Config fields plus a Provider hint used in
// the key. The Provider hint is the *primary* provider in the wrapped chain;
// changing it changes the key, so cache contents are never served for a
// different primary.
type CachingProviderConfig struct {
	Provider string
	Model    string
	Width    int
	Height   int
	Seed     int
}

// CachingProvider is a Provider that fronts a fallback chain and serves
// repeated requests from the cache. On miss it delegates to the inner
// chain via GenerateWithFallback and writes the result back to the cache.
type CachingProvider struct {
	inner    []Provider
	cache    *ImageCache
	key      CachingProviderConfig
	name     string
}

// NewCachingProvider wraps inner with a cache keyed by key. The returned
// provider satisfies the Provider interface and should be placed at index
// 0 of the provider chain so all traffic flows through the cache first.
func NewCachingProvider(inner []Provider, c *ImageCache, key CachingProviderConfig) *CachingProvider {
	name := "cached"
	if len(inner) > 0 && inner[0] != nil {
		if first := inner[0].Name(); first != "" {
			name = "cached-" + first
		}
	}
	return &CachingProvider{
		inner: inner,
		cache: c,
		key:   key,
		name:  name,
	}
}

// Name returns a stable identifier for this provider.
func (p *CachingProvider) Name() string { return p.name }

// IsConfigured is always true — the CachingProvider is always usable
// as long as at least one inner provider is configured (or the cache
// has a hit).
func (p *CachingProvider) IsConfigured() bool { return true }

// Generate returns image bytes for prompt, served from cache when
// possible. On a miss it delegates to the inner fallback chain and
// caches the result.
func (p *CachingProvider) Generate(prompt string) ([]byte, error) {
	key := computeKey(prompt, p.key.Provider, p.key.Model, p.key.Width, p.key.Height, p.key.Seed)
	if data, ok := p.cache.Get(key); ok {
		return data, nil
	}
	data, _, err := GenerateWithFallback(p.inner, prompt)
	if err != nil {
		return nil, err
	}
	p.cache.Put(key, data)
	return data, nil
}

// BypassAndWrite calls the inner chain directly (no cache read) and
// writes the result to the cache. It is used by the asset service when
// the caller passes force_regenerate=true. It returns the bytes and the
// name of the inner provider that produced them.
func (p *CachingProvider) BypassAndWrite(prompt string) ([]byte, string, error) {
	data, name, err := GenerateWithFallback(p.inner, prompt)
	if err != nil {
		return nil, "", err
	}
	key := computeKey(prompt, p.key.Provider, p.key.Model, p.key.Width, p.key.Height, p.key.Seed)
	p.cache.Put(key, data)
	return data, name, nil
}

// Inner returns the wrapped inner provider chain. Useful for tests and
// for callers that need to compute a cache key with the first inner
// provider's name.
func (p *CachingProvider) Inner() []Provider { return p.inner }
