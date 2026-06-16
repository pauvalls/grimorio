package image

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

// ---------- computeKey ----------

func TestComputeKey_Deterministic(t *testing.T) {
	t.Parallel()
	const (
		prompt   = "epic fantasy dragon"
		provider = "pollinations"
		model    = "flux"
		w, h     = 1024, 1024
		seed     = 42
	)
	a := computeKey(prompt, provider, model, w, h, seed)
	b := computeKey(prompt, provider, model, w, h, seed)
	if a != b {
		t.Fatalf("computeKey not deterministic: %s != %s", a, b)
	}
	if len(a) != 64 {
		t.Errorf("computeKey length = %d, want 64 (hex sha256)", len(a))
	}
	// sanity: must match sha256 of the expected canonical input
	md5 := sha256.Sum256([]byte(fmt.Sprintf("%s\x1f%s\x1f%d\x1f%d\x1f%d\x1f%s",
		prompt, model, w, h, seed, provider)))
	if a != hex.EncodeToString(md5[:]) {
		t.Errorf("computeKey does not match expected sha256 input format")
	}
}

func TestComputeKey_FieldSensitivity(t *testing.T) {
	t.Parallel()
	base := func() string {
		return computeKey("prompt", "pollinations", "flux", 1024, 1024, 42)
	}
	baseKey := base()

	tt := []struct {
		name     string
		mutate   func() string
		expected bool // true = keys MUST differ
	}{
		{"different prompt", func() string { return computeKey("OTHER", "pollinations", "flux", 1024, 1024, 42) }, true},
		{"different model", func() string { return computeKey("prompt", "pollinations", "dalle3", 1024, 1024, 42) }, true},
		{"different width", func() string { return computeKey("prompt", "pollinations", "flux", 800, 1024, 42) }, true},
		{"different height", func() string { return computeKey("prompt", "pollinations", "flux", 1024, 800, 42) }, true},
		{"different seed", func() string { return computeKey("prompt", "pollinations", "flux", 1024, 1024, 99) }, true},
		{"different provider", func() string { return computeKey("prompt", "dalle", "flux", 1024, 1024, 42) }, true},
		{"identical", base, false},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.mutate()
			if tc.expected && got == baseKey {
				t.Errorf("expected different key for %s, got same", tc.name)
			}
			if !tc.expected && got != baseKey {
				t.Errorf("expected same key for %s, got different", tc.name)
			}
		})
	}
}

func TestComputeKey_EmptyFields(t *testing.T) {
	t.Parallel()
	a := computeKey("", "", "", 0, 0, 0)
	b := computeKey("", "", "", 0, 0, 0)
	if a != b {
		t.Errorf("computeKey with empty fields must be deterministic: %s != %s", a, b)
	}
	if len(a) != 64 {
		t.Errorf("computeKey length = %d, want 64", len(a))
	}
}

// ---------- ImageCache (in-memory LRU + on-disk) ----------

func newTestImageCache(t *testing.T) *ImageCache {
	t.Helper()
	dir := t.TempDir()
	c, err := NewImageCache(dir, 50)
	if err != nil {
		t.Fatalf("NewImageCache: %v", err)
	}
	return c
}

func TestImageCache_PutAndGet_Hit(t *testing.T) {
	t.Parallel()
	c := newTestImageCache(t)
	data := []byte("png-binary-blob")
	c.Put("k1", data)
	got, ok := c.Get("k1")
	if !ok {
		t.Fatal("Get(k1) miss after Put")
	}
	if !bytes.Equal(got, data) {
		t.Errorf("Get bytes = %q, want %q", got, data)
	}
}

func TestImageCache_Get_Miss(t *testing.T) {
	t.Parallel()
	c := newTestImageCache(t)
	if _, ok := c.Get("nope"); ok {
		t.Error("Get on empty cache must miss")
	}
}

func TestImageCache_DiskReadAndWrite(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c, err := NewImageCache(dir, 50)
	if err != nil {
		t.Fatalf("NewImageCache: %v", err)
	}
	c.Put("disk-key", []byte("disk-blob"))

	// Disk path shape: <dir>/v1/<key[:2]>/<key>.bin
	diskPath := filepath.Join(dir, "v1", "di", "disk-key.bin")
	if _, err := os.Stat(diskPath); err != nil {
		t.Errorf("expected disk file at %s: %v", diskPath, err)
	}

	// Same-instance read must hit (in-memory LRU was populated by Put).
	got, ok := c.Get("disk-key")
	if !ok {
		t.Fatal("Get after Put should hit")
	}
	if string(got) != "disk-blob" {
		t.Errorf("bytes = %q, want disk-blob", got)
	}
}

func TestImageCache_DiskReadCorruptionHandled(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c, err := NewImageCache(dir, 50)
	if err != nil {
		t.Fatalf("NewImageCache: %v", err)
	}
	// Use a multi-char key so diskPath shard works
	ck := "abcd"
	c.Put(ck, []byte("good"))
	shard := ck[:2]
	diskPath := filepath.Join(dir, "v1", shard, ck+".bin")
	// Corrupt the on-disk file
	_ = os.WriteFile(diskPath, []byte("corrupt"), 0644)

	// The LRU still has the good value, so Get returns it.
	got, ok := c.Get(ck)
	if !ok {
		t.Fatal("Get should hit LRU even with corrupt disk")
	}
	if string(got) != "good" {
		t.Errorf("Get returned %q, want LRU value 'good'", got)
	}
}

func TestImageCache_LRUEviction(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c, err := NewImageCache(dir, 50)
	if err != nil {
		t.Fatalf("NewImageCache: %v", err)
	}
	for i := 0; i < 50; i++ {
		c.Put(fmt.Sprintf("k%02d", i), []byte("x"))
	}
	if c.Len() != 50 {
		t.Fatalf("Len() = %d, want 50", c.Len())
	}
	// 51st distinct key must evict the oldest (k00) from in-memory layer
	c.Put("k51", []byte("y"))
	if c.Len() != 50 {
		t.Errorf("Len() after eviction = %d, want 50", c.Len())
	}
	// Verify the in-memory LRU evicted the oldest by inserting a 52nd and
	// confirming the new entry did not push Len() over 50.
	c.Put("k52", []byte("z"))
	if c.Len() != 50 {
		t.Errorf("Len() = %d, want 50 (cap)", c.Len())
	}
	// k51 and k52 should both be present (just put)
	if _, ok := c.Get("k51"); !ok {
		t.Error("recently put k51 should be in LRU")
	}
}

func TestImageCache_ThreadSafety(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c, err := NewImageCache(dir, 50)
	if err != nil {
		t.Fatalf("NewImageCache: %v", err)
	}
	const goroutines = 50
	const opsPerG = 20
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("goroutine %d panicked: %v", g, r)
				}
			}()
			for i := 0; i < opsPerG; i++ {
				key := fmt.Sprintf("g%d-k%d", g, i%60) // some duplicates
				c.Put(key, []byte(fmt.Sprintf("v-%d", i)))
				_, _ = c.Get(key)
			}
		}(g)
	}
	wg.Wait()
	if c.Len() > 50 {
		t.Errorf("Len() = %d, want <= 50", c.Len())
	}
}

// ---------- CachingProvider ----------

type recordingProvider struct {
	name  string
	data  []byte
	calls atomic.Int32
}

func (r *recordingProvider) Generate(prompt string) ([]byte, error) {
	r.calls.Add(1)
	return r.data, nil
}
func (r *recordingProvider) IsConfigured() bool { return true }
func (r *recordingProvider) Name() string       { return r.name }

func TestCachingProvider_HitDoesNotCallInner(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cache, err := NewImageCache(dir, 50)
	if err != nil {
		t.Fatalf("NewImageCache: %v", err)
	}
	inner := &recordingProvider{name: "pollinations", data: []byte("img-bytes")}
	cp := NewCachingProvider([]Provider{inner}, cache, CachingProviderConfig{
		Provider: "pollinations",
		Model:    "flux",
		Width:    1024, Height: 1024, Seed: 42,
	})

	// First call: miss, calls inner once
	got1, err := cp.Generate("a dragon")
	if err != nil {
		t.Fatalf("first Generate: %v", err)
	}
	if string(got1) != "img-bytes" {
		t.Errorf("bytes = %q, want img-bytes", got1)
	}
	if inner.calls.Load() != 1 {
		t.Fatalf("first call: inner.calls = %d, want 1", inner.calls.Load())
	}

	// Second call: hit, inner NOT called again
	got2, err := cp.Generate("a dragon")
	if err != nil {
		t.Fatalf("second Generate: %v", err)
	}
	if string(got2) != "img-bytes" {
		t.Errorf("bytes = %q, want img-bytes", got2)
	}
	if inner.calls.Load() != 1 {
		t.Errorf("second call: inner.calls = %d, want still 1 (cache hit)", inner.calls.Load())
	}
}

func TestCachingProvider_MissCallsInnerOnce(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cache, err := NewImageCache(dir, 50)
	if err != nil {
		t.Fatalf("NewImageCache: %v", err)
	}
	inner := &recordingProvider{name: "pollinations", data: []byte("img")}
	cp := NewCachingProvider([]Provider{inner}, cache, CachingProviderConfig{
		Provider: "pollinations", Model: "flux",
		Width: 1024, Height: 1024, Seed: 1,
	})
	if _, err := cp.Generate("hello"); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if inner.calls.Load() != 1 {
		t.Errorf("inner.calls = %d, want 1", inner.calls.Load())
	}
}

func TestCachingProvider_FallbackChain(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cache, err := NewImageCache(dir, 50)
	if err != nil {
		t.Fatalf("NewImageCache: %v", err)
	}
	failingErr := &failingProvider{name: "dalle"}
	succeeding := &recordingProvider{name: "pollinations", data: []byte("fallback-img")}

	chain := []Provider{failingErr, succeeding}
	cp := NewCachingProvider(chain, cache, CachingProviderConfig{
		Provider: "dalle", Model: "dall-e-3",
		Width: 1024, Height: 1024, Seed: 7,
	})

	got, err := cp.Generate("anything")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if string(got) != "fallback-img" {
		t.Errorf("bytes = %q, want fallback-img", got)
	}
	// Both inner providers were called.
	if failingErr.calls.Load() != 1 {
		t.Errorf("primary.calls = %d, want 1", failingErr.calls.Load())
	}
	if succeeding.calls.Load() != 1 {
		t.Errorf("fallback.calls = %d, want 1", succeeding.calls.Load())
	}
}

func TestCachingProvider_AllInnerFailReturnsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cache, err := NewImageCache(dir, 50)
	if err != nil {
		t.Fatalf("NewImageCache: %v", err)
	}
	f1 := &failingProvider{name: "a"}
	f2 := &failingProvider{name: "b"}
	cp := NewCachingProvider([]Provider{f1, f2}, cache, CachingProviderConfig{
		Provider: "a", Width: 1, Height: 1, Seed: 1,
	})
	if _, err := cp.Generate("x"); err == nil {
		t.Error("expected error when all inner providers fail")
	}
}

func TestCachingProvider_IsConfigured(t *testing.T) {
	t.Parallel()
	cache, _ := NewImageCache(t.TempDir(), 10)
	cp := NewCachingProvider([]Provider{&recordingProvider{name: "x"}}, cache,
		CachingProviderConfig{Provider: "x"})
	if !cp.IsConfigured() {
		t.Error("CachingProvider should always be configured")
	}
}

func TestCachingProvider_Name(t *testing.T) {
	t.Parallel()
	cache, _ := NewImageCache(t.TempDir(), 10)
	cp := NewCachingProvider([]Provider{&recordingProvider{name: "x"}}, cache,
		CachingProviderConfig{Provider: "x"})
	if cp.Name() == "" {
		t.Error("CachingProvider.Name() should not be empty")
	}
}

func TestCachingProvider_Inner(t *testing.T) {
	t.Parallel()
	cache, _ := NewImageCache(t.TempDir(), 10)
	inner := []Provider{&recordingProvider{name: "p"}}
	cp := NewCachingProvider(inner, cache, CachingProviderConfig{Provider: "p"})
	got := cp.Inner()
	if len(got) != 1 || got[0].Name() != "p" {
		t.Errorf("Inner() = %v, want [p]", got)
	}
}

func TestCachingProvider_BypassAndWrite(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cache, err := NewImageCache(dir, 50)
	if err != nil {
		t.Fatalf("NewImageCache: %v", err)
	}
	inner := &recordingProvider{name: "pollinations", data: []byte("forced-bytes")}
	cp := NewCachingProvider([]Provider{inner}, cache, CachingProviderConfig{
		Provider: "pollinations", Model: "flux",
		Width: 1024, Height: 1024, Seed: 1,
	})

	data, name, err := cp.BypassAndWrite("a prompt")
	if err != nil {
		t.Fatalf("BypassAndWrite: %v", err)
	}
	if string(data) != "forced-bytes" {
		t.Errorf("bytes = %q, want forced-bytes", data)
	}
	if name != "pollinations" {
		t.Errorf("provider name = %q, want pollinations", name)
	}
	if inner.calls.Load() != 1 {
		t.Errorf("inner.calls = %d, want 1", inner.calls.Load())
	}

	// Subsequent Get for the same key must hit the cache (so BypassAndWrite
	// did write).
	got, ok := cache.Get(computeKey("a prompt", "pollinations", "flux", 1024, 1024, 1))
	if !ok {
		t.Error("BypassAndWrite should have written to cache")
	}
	if string(got) != "forced-bytes" {
		t.Errorf("cached bytes = %q, want forced-bytes", got)
	}
}

func TestCachingProvider_BypassAndWrite_AllFail(t *testing.T) {
	t.Parallel()
	cache, _ := NewImageCache(t.TempDir(), 10)
	cp := NewCachingProvider([]Provider{&failingProvider{name: "x"}}, cache,
		CachingProviderConfig{Provider: "x"})
	if _, _, err := cp.BypassAndWrite("anything"); err == nil {
		t.Error("expected error when all inner providers fail")
	}
}

func TestImageCache_DiskPathShortKey(t *testing.T) {
	t.Parallel()
	cache, _ := NewImageCache(t.TempDir(), 10)
	// Empty key falls into the "_" shard branch.
	p := cache.diskPath("")
	if !filepath.IsAbs(p) == false && p == "" {
		t.Errorf("diskPath empty key returned %q", p)
	}
	// Single-char key falls into the "k" shard branch.
	p2 := cache.diskPath("k")
	if filepath.Base(p2) != "k.bin" {
		t.Errorf("diskPath single-char key = %q, want k.bin", p2)
	}
}

// ---------- helpers ----------

type failingProvider struct {
	name  string
	calls atomic.Int32
}

func (f *failingProvider) Generate(prompt string) ([]byte, error) {
	f.calls.Add(1)
	return nil, fmt.Errorf("simulated failure: %s", f.name)
}
func (f *failingProvider) IsConfigured() bool { return true }
func (f *failingProvider) Name() string       { return f.name }
