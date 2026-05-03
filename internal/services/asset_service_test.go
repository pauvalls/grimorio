package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pauvalls/grimorio/internal/image"
)

// mockProvider is a test double for image.Provider
type mockProvider struct {
	name         string
	generateFunc func(prompt string) ([]byte, error)
	callCount    atomic.Int32
	configured   bool
}

func (m *mockProvider) Generate(prompt string) ([]byte, error) {
	m.callCount.Add(1)
	if m.generateFunc != nil {
		return m.generateFunc(prompt)
	}
	return []byte("fake-image-data-" + prompt), nil
}

func (m *mockProvider) IsConfigured() bool { return m.configured }
func (m *mockProvider) Name() string       { return m.name }

func setupTestAssetService(t *testing.T) (*AssetService, string) {
	t.Helper()
	tmpDir := t.TempDir()
	provider := &mockProvider{name: "mock", configured: true}
	service := NewAssetServiceWithProvider(tmpDir, provider)
	return service, tmpDir
}

func TestGenerateImage_Success(t *testing.T) {
	service, tmpDir := setupTestAssetService(t)

	path, err := service.GenerateImage("test-campaign", "cover-art", "epic fantasy cover", "cover")
	if err != nil {
		t.Fatalf("GenerateImage() error: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "test-campaign", "assets", "cover-art.png")
	if path != expectedPath {
		t.Errorf("GenerateImage() path = %s, want %s", path, expectedPath)
	}

	// Verify file was created
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("GenerateImage() did not create file at %s", path)
	}
}

func TestGenerateImage_ProviderError(t *testing.T) {
	tmpDir := t.TempDir()
	provider := &mockProvider{
		name: "mock",
		generateFunc: func(prompt string) ([]byte, error) {
			return nil, fmt.Errorf("provider error: simulated failure")
		},
	}
	service := NewAssetServiceWithProvider(tmpDir, provider)

	_, err := service.GenerateImage("test-campaign", "fail-test", "prompt", "cover")
	if err == nil {
		t.Fatal("GenerateImage() expected error, got nil")
	}

	if !strings.Contains(err.Error(), "provider error") {
		t.Errorf("GenerateImage() error = %v, want containing 'provider error'", err)
	}
}

func TestGenerateImage_Sequential(t *testing.T) {
	service, tmpDir := setupTestAssetService(t)

	// Track call times to verify sequential execution
	var callTimes []time.Time
	provider := &mockProvider{
		name: "mock",
		generateFunc: func(prompt string) ([]byte, error) {
			callTimes = append(callTimes, time.Now())
			return []byte("fake-image-data-" + prompt), nil
		},
	}
	service = NewAssetServiceWithProvider(tmpDir, provider)

	// Generate multiple images
	start := time.Now()
	for i := 0; i < 3; i++ {
		_, err := service.GenerateImage("test-campaign", fmt.Sprintf("img-%d", i), "prompt", "cover")
		if err != nil {
			t.Fatalf("GenerateImage() error: %v", err)
		}
	}
	elapsed := time.Since(start)

	// Should take at least 6 seconds (3 images * 2s minimum between them)
	// Each image has 3s delay
	if elapsed < 6*time.Second {
		t.Errorf("Sequential generation took %v, expected at least 6s", elapsed)
	}

	// Verify images were created
	for i := 0; i < 3; i++ {
		path := filepath.Join(tmpDir, "test-campaign", "assets", fmt.Sprintf("img-%d.png", i))
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("GenerateImage() did not create file: %s", path)
		}
	}
}

func TestGenerateImage_WithProviderFallback(t *testing.T) {
	tmpDir := t.TempDir()

	// Primary provider always fails
	primary := &mockProvider{
		name: "primary",
		generateFunc: func(prompt string) ([]byte, error) {
			return nil, fmt.Errorf("primary always fails")
		},
	}

	// Fallback provider succeeds
	fallback := &mockProvider{
		name: "fallback",
		generateFunc: func(prompt string) ([]byte, error) {
			return []byte("fallback-image-" + prompt), nil
		},
	}

	providers := []image.Provider{primary, fallback}

	outputPath := filepath.Join(tmpDir, "test-campaign", "assets", "test-image.png")
	providerName, err := image.GenerateAndSaveWithFallback(providers, "test prompt", outputPath)

	if err != nil {
		t.Fatalf("GenerateAndSaveWithFallback() error: %v", err)
	}

	if providerName != "fallback" {
		t.Errorf("Expected fallback provider, got %s", providerName)
	}

	// Verify file was created by fallback
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Fallback did not create file at %s", outputPath)
	}

	// Primary should have been called
	if primary.callCount.Load() != 1 {
		t.Errorf("Primary provider should have been called once, got %d", primary.callCount.Load())
	}

	// Fallback should have been called
	if fallback.callCount.Load() != 1 {
		t.Errorf("Fallback provider should have been called once, got %d", fallback.callCount.Load())
	}
}

func TestGenerateImage_AllProvidersFail(t *testing.T) {
	tmpDir := t.TempDir()

	// All providers fail
	provider1 := &mockProvider{
		name: "provider1",
		generateFunc: func(prompt string) ([]byte, error) {
			return nil, fmt.Errorf("provider1 error")
		},
	}
	provider2 := &mockProvider{
		name: "provider2",
		generateFunc: func(prompt string) ([]byte, error) {
			return nil, fmt.Errorf("provider2 error")
		},
	}

	providers := []image.Provider{provider1, provider2}
	outputPath := filepath.Join(tmpDir, "test-campaign", "assets", "test-image.png")
	_, err := image.GenerateAndSaveWithFallback(providers, "test prompt", outputPath)

	if err == nil {
		t.Fatal("Expected error when all providers fail")
	}

	if !strings.Contains(err.Error(), "all providers failed") {
		t.Errorf("Expected 'all providers failed' error, got: %v", err)
	}

	// Both providers should have been tried
	if provider1.callCount.Load() != 1 {
		t.Errorf("Provider1 should have been called once, got %d", provider1.callCount.Load())
	}
	if provider2.callCount.Load() != 1 {
		t.Errorf("Provider2 should have been called once, got %d", provider2.callCount.Load())
	}
}

func TestGenerateMap_Success(t *testing.T) {
	service, _ := setupTestAssetService(t)

	svg, err := service.GenerateMap("test", "dungeon", "dungeon", "The Dark Cave", 5, []string{"Entrance", "Boss"})
	if err != nil {
		t.Fatalf("GenerateMap() error: %v", err)
	}

	if !strings.Contains(svg, "<svg") {
		t.Error("GenerateMap() output should contain SVG markup")
	}

	if !strings.Contains(svg, "The Dark Cave") {
		t.Error("GenerateMap() output should contain the title")
	}
}

func TestGenerateDivider_Success(t *testing.T) {
	service, _ := setupTestAssetService(t)

	svg, err := service.GenerateDivider("test", "divider", "ornate", 600)
	if err != nil {
		t.Fatalf("GenerateDivider() error: %v", err)
	}

	if !strings.Contains(svg, "<svg") {
		t.Error("GenerateDivider() output should contain SVG markup")
	}
}

func TestGetProviders_WithInjectedProvider(t *testing.T) {
	provider := &mockProvider{name: "injected"}
	service := NewAssetServiceWithProvider("/tmp", provider)

	providers := service.getProviders()
	if len(providers) != 1 {
		t.Fatalf("getProviders() returned %d providers, want 1", len(providers))
	}

	if providers[0].Name() != "injected" {
		t.Errorf("getProviders()[0] = %s, want 'injected'", providers[0].Name())
	}
}

func TestGetProviders_FallsBackToChain(t *testing.T) {
	service := NewAssetService("/tmp", image.DefaultConfig())

	providers := service.getProviders()
	if len(providers) == 0 {
		t.Fatal("getProviders() returned empty providers")
	}

	// Should have pollinations in the chain
	hasPollinations := false
	for _, p := range providers {
		if p.Name() == "pollinations" {
			hasPollinations = true
			break
		}
	}
	if !hasPollinations {
		t.Errorf("getProviders() should include 'pollinations' in fallback chain")
	}
}

func TestGenerateImage_DoubleExtension(t *testing.T) {
	service, tmpDir := setupTestAssetService(t)

	// filename already includes .png extension
	path, err := service.GenerateImage("test-campaign", "morgus-portrait.png", "A wizard portrait", "portrait")
	if err != nil {
		t.Fatalf("GenerateImage() error: %v", err)
	}

	// Should NOT create morgus-portrait.png.png
	badPath := filepath.Join(tmpDir, "test-campaign", "assets", "morgus-portrait.png.png")
	if _, err := os.Stat(badPath); !os.IsNotExist(err) {
		t.Errorf("GenerateImage() created double extension file: %s", badPath)
	}

	// Should create morgus-portrait.png
	goodPath := filepath.Join(tmpDir, "test-campaign", "assets", "morgus-portrait.png")
	if path != goodPath {
		t.Errorf("GenerateImage() path = %s, want %s", path, goodPath)
	}

	if _, err := os.Stat(goodPath); os.IsNotExist(err) {
		t.Errorf("GenerateImage() did not create file at %s", goodPath)
	}
}
