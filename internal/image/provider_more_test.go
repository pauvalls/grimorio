package image

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.png")

	provider := &mockProvider{name: "test", fail: false}
	err := GenerateAndSave(provider, "test prompt", outputPath)
	if err != nil {
		t.Fatalf("GenerateAndSave() error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); err != nil {
		t.Errorf("GenerateAndSave() did not create file: %v", err)
	}
}

func TestGenerateAndSave_ProviderError(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.png")

	provider := &mockProvider{name: "test", fail: true}
	err := GenerateAndSave(provider, "test prompt", outputPath)
	if err == nil {
		t.Error("GenerateAndSave() should error when provider fails")
	}
}

func TestGenerateAndSaveWithFallback(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.png")

	p1 := &mockProvider{name: "primary", fail: true}
	p2 := &mockProvider{name: "fallback", fail: false}

	providerName, err := GenerateAndSaveWithFallback([]Provider{p1, p2}, "test prompt", outputPath)
	if err != nil {
		t.Fatalf("GenerateAndSaveWithFallback() error: %v", err)
	}
	if providerName != "fallback" {
		t.Errorf("Expected provider 'fallback', got '%s'", providerName)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); err != nil {
		t.Errorf("GenerateAndSaveWithFallback() did not create file: %v", err)
	}
}

func TestGenerateAndSaveWithFallback_AllFail(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.png")

	p1 := &mockProvider{name: "p1", fail: true}
	p2 := &mockProvider{name: "p2", fail: true}

	_, err := GenerateAndSaveWithFallback([]Provider{p1, p2}, "test prompt", outputPath)
	if err == nil {
		t.Error("GenerateAndSaveWithFallback() should error when all providers fail")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Provider != "pollinations" {
		t.Errorf("Expected default provider 'pollinations', got '%s'", cfg.Provider)
	}
	if cfg.DalleModel != "dall-e-3" {
		t.Errorf("Expected default DalleModel 'dall-e-3', got '%s'", cfg.DalleModel)
	}
	if cfg.Width != 1024 {
		t.Errorf("Expected default Width 1024, got %d", cfg.Width)
	}
	if cfg.Height != 1024 {
		t.Errorf("Expected default Height 1024, got %d", cfg.Height)
	}
	if cfg.Seed != -1 {
		t.Errorf("Expected default Seed -1, got %d", cfg.Seed)
	}
}
