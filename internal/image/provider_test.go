package image

import (
	"testing"
)

func TestNewProvider_Pollinations(t *testing.T) {
	cfg := Config{Provider: "pollinations", Width: 512, Height: 512}
	p, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider(pollinations) error: %v", err)
	}
	if p.Name() != "pollinations" {
		t.Errorf("Expected 'pollinations', got %s", p.Name())
	}
	if !p.IsConfigured() {
		t.Error("Pollinations should always be configured")
	}
}

func TestNewProvider_Raphael(t *testing.T) {
	cfg := Config{Provider: "raphael"}
	p, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider(raphael) error: %v", err)
	}
	if p.Name() != "raphael" {
		t.Errorf("Expected 'raphael', got %s", p.Name())
	}
	if !p.IsConfigured() {
		t.Error("Raphael should always be configured")
	}
}

func TestNewProvider_Default(t *testing.T) {
	cfg := Config{Provider: "unknown"}
	p, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider(unknown) error: %v", err)
	}
	if p.Name() != "pollinations" {
		t.Errorf("Expected default 'pollinations', got %s", p.Name())
	}
}

func TestNewProvider_Dalle_MissingKey(t *testing.T) {
	cfg := Config{Provider: "dalle", DalleKey: ""}
	_, err := NewProvider(cfg)
	if err == nil {
		t.Fatal("Expected error for DALL-E without API key")
	}
}

func TestNewProviderChain(t *testing.T) {
	cfg := DefaultConfig()
	chain := NewProviderChain(cfg)

	if len(chain) == 0 {
		t.Fatal("Provider chain should not be empty")
	}

	// Should have at least 2 providers (primary + fallback)
	if len(chain) < 2 {
		t.Errorf("Expected at least 2 providers in chain, got %d", len(chain))
	}

	// Should include raphael
	hasRaphael := false
	for _, p := range chain {
		if p.Name() == "raphael" {
			hasRaphael = true
			break
		}
	}
	if !hasRaphael {
		t.Error("Provider chain should include 'raphael'")
	}
}

// TestNewProviderChain_WithCache verifies the CachingProvider is at index 0
// of the chain when a cache is supplied, and that the inner chain still
// includes raphael + pollinations as fallbacks.
func TestNewProviderChain_WithCache(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewImageCache(dir, 50)
	if err != nil {
		t.Fatalf("NewImageCache: %v", err)
	}
	cfg := DefaultConfig()
	chain := NewProviderChainWithCache(cfg, cache)

	if len(chain) != 1 {
		t.Fatalf("chain length = %d, want 1 (the CachingProvider encapsulates the inner chain)", len(chain))
	}
	cp, ok := chain[0].(*CachingProvider)
	if !ok {
		t.Fatalf("chain[0] should be a *CachingProvider, got %T", chain[0])
	}
	// The CachingProvider's inner chain must include raphael and pollinations.
	hasRaphael, hasPollinations := false, false
	for _, p := range cp.inner {
		switch p.Name() {
		case "raphael":
			hasRaphael = true
		case "pollinations":
			hasPollinations = true
		}
	}
	if !hasRaphael {
		t.Error("CachingProvider.inner must include 'raphael'")
	}
	if !hasPollinations {
		t.Error("CachingProvider.inner must include 'pollinations'")
	}
}

func TestGenerateWithFallback_Success(t *testing.T) {
	p1 := &mockProvider{name: "primary", fail: true}
	p2 := &mockProvider{name: "fallback", fail: false}

	providers := []Provider{p1, p2}
	data, providerName, err := GenerateWithFallback(providers, "test prompt")

	if err != nil {
		t.Fatalf("GenerateWithFallback error: %v", err)
	}

	if providerName != "fallback" {
		t.Errorf("Expected 'fallback', got %s", providerName)
	}

	if string(data) != "mock-image-data" {
		t.Errorf("Unexpected data: %s", string(data))
	}

	// Primary should have been tried
	if p1.callCount != 1 {
		t.Errorf("Primary should have been called once, got %d", p1.callCount)
	}
}

func TestGenerateWithFallback_AllFail(t *testing.T) {
	p1 := &mockProvider{name: "p1", fail: true}
	p2 := &mockProvider{name: "p2", fail: true}

	providers := []Provider{p1, p2}
	_, _, err := GenerateWithFallback(providers, "test prompt")

	if err == nil {
		t.Fatal("Expected error when all providers fail")
	}

	if p1.callCount != 1 || p2.callCount != 1 {
		t.Error("All providers should have been tried")
	}
}

// mockProvider for testing
type mockProvider struct {
	name      string
	fail      bool
	callCount int
}

func (m *mockProvider) Generate(prompt string) ([]byte, error) {
	m.callCount++
	if m.fail {
		return nil, errProviderFailure
	}
	return []byte("mock-image-data"), nil
}

func (m *mockProvider) IsConfigured() bool { return true }
func (m *mockProvider) Name() string       { return m.name }

var errProviderFailure = &mockError{msg: "provider failure"}

type mockError struct {
	msg string
}

func (e *mockError) Error() string { return e.msg }
