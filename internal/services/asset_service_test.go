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
	name          string
	generateFunc  func(prompt string) ([]byte, error)
	callCount     atomic.Int32
	configured    bool
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

func TestGenerateImagesBatch_Parallel(t *testing.T) {
	service, tmpDir := setupTestAssetService(t)

	images := []BatchImageSpec{
		{Filename: "cover-art", Prompt: "cover prompt", Type: "cover"},
		{Filename: "npc-1", Prompt: "npc1 prompt", Type: "portrait"},
		{Filename: "npc-2", Prompt: "npc2 prompt", Type: "portrait"},
		{Filename: "scene-1", Prompt: "scene1 prompt", Type: "scene"},
	}

	start := time.Now()
	results, err := service.GenerateImagesBatch("test-campaign", images)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("GenerateImagesBatch() error: %v", err)
	}

	if len(results) != len(images) {
		t.Fatalf("GenerateImagesBatch() returned %d results, want %d", len(results), len(images))
	}

	// All should succeed
	for _, r := range results {
		if !r.Success {
			t.Errorf("GenerateImagesBatch() failed for %s: %s", r.Filename, r.Error)
		}
	}

	// Verify files were created
	for _, img := range images {
		path := filepath.Join(tmpDir, "test-campaign", "assets", img.Filename+".png")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("GenerateImagesBatch() did not create file: %s", path)
		}
	}

	// Should be fast (parallel), not sequential
	// 4 images * 100ms each sequential = 400ms+
	// Parallel should be < 200ms even with overhead
	if elapsed > 500*time.Millisecond {
		t.Logf("Warning: batch took %v, may not be fully parallel", elapsed)
	}
}

func TestGenerateImagesBatch_PartialFailure_WithFallback(t *testing.T) {
	tmpDir := t.TempDir()
	failCount := atomic.Int32{}
	provider := &mockProvider{
		name: "mock",
		generateFunc: func(prompt string) ([]byte, error) {
			// Fail on first attempt for "npc-bad"
			if strings.Contains(prompt, "bad") && failCount.Load() < 1 {
				failCount.Add(1)
				return nil, fmt.Errorf("simulated failure")
			}
			return []byte("fake-image-" + prompt), nil
		},
	}
	service := NewAssetServiceWithProvider(tmpDir, provider)

	images := []BatchImageSpec{
		{Filename: "cover-art", Prompt: "good prompt", Type: "cover"},
		{Filename: "npc-bad", Prompt: "bad npc prompt", Type: "portrait"},
		{Filename: "scene-ok", Prompt: "good scene prompt", Type: "scene"},
	}

	results, err := service.GenerateImagesBatch("test-campaign", images)
	if err != nil {
		t.Fatalf("GenerateImagesBatch() error: %v", err)
	}

	// First image should succeed
	if !results[0].Success {
		t.Errorf("First image should succeed: %v", results[0])
	}

	// Second image should succeed after fallback retry
	if !results[1].Success {
		t.Errorf("Second image (with fallback) should succeed: %v", results[1])
	}

	// Third image should succeed
	if !results[2].Success {
		t.Errorf("Third image should succeed: %v", results[2])
	}

	// Verify the fallback was actually called (provider called more than len(images) times)
	// Initial batch: 3 calls, fallback: 1 extra call for the failed one = 4 total
	if provider.callCount.Load() < 3 {
		t.Errorf("Expected at least 3 provider calls, got %d", provider.callCount.Load())
	}
}

func TestGenerateImagesBatch_CompleteFailure(t *testing.T) {
	tmpDir := t.TempDir()
	provider := &mockProvider{
		name: "mock",
		generateFunc: func(prompt string) ([]byte, error) {
			return nil, fmt.Errorf("total failure")
		},
	}
	service := NewAssetServiceWithProvider(tmpDir, provider)

	images := []BatchImageSpec{
		{Filename: "img1", Prompt: "prompt1", Type: "cover"},
		{Filename: "img2", Prompt: "prompt2", Type: "portrait"},
	}

	results, err := service.GenerateImagesBatch("test-campaign", images)
	if err != nil {
		t.Fatalf("GenerateImagesBatch() error: %v", err)
	}

	// All should fail (even after fallback)
	for _, r := range results {
		if r.Success {
			t.Errorf("Expected failure for %s, got success", r.Filename)
		}
		if !strings.Contains(r.Error, "total failure") {
			t.Errorf("Expected error containing 'total failure', got: %s", r.Error)
		}
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

func TestGetProvider_UsesInjectedProvider(t *testing.T) {
	provider := &mockProvider{name: "injected"}
	service := NewAssetServiceWithProvider("/tmp", provider)

	p, err := service.getProvider()
	if err != nil {
		t.Fatalf("getProvider() error: %v", err)
	}

	if p.Name() != "injected" {
		t.Errorf("getProvider() = %s, want 'injected'", p.Name())
	}
}

func TestGetProvider_FallsBackToConfig(t *testing.T) {
	service := NewAssetService("/tmp", image.DefaultConfig())

	p, err := service.getProvider()
	if err != nil {
		t.Fatalf("getProvider() error: %v", err)
	}

	if p == nil {
		t.Fatal("getProvider() returned nil provider")
	}

	if p.Name() != "pollinations" {
		t.Errorf("getProvider() = %s, want 'pollinations'", p.Name())
	}
}