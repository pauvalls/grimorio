package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pauvalls/grimorio/internal/image"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}
	if cfg.PDFEngine != "wkhtmltopdf" {
		t.Errorf("expected PDFEngine 'wkhtmltopdf', got '%s'", cfg.PDFEngine)
	}
	if cfg.Provider != "pollinations" {
		t.Errorf("expected Provider 'pollinations', got '%s'", cfg.Provider)
	}
	if cfg.DalleModel != "dall-e-3" {
		t.Errorf("expected DalleModel 'dall-e-3', got '%s'", cfg.DalleModel)
	}
	if cfg.OutputDir == "" {
		t.Error("expected OutputDir to be set")
	}
}

func TestLoadConfig_FileNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.json")

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() returned error for missing file: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadConfig() returned nil for missing file")
	}
	if cfg.PDFEngine != "wkhtmltopdf" {
		t.Errorf("expected default PDFEngine, got '%s'", cfg.PDFEngine)
	}
}

func TestLoadConfig_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	data := `{"output_dir":"/tmp/test","pdf_engine":"weasyprint","image_provider":"raphael"}`
	if err := os.WriteFile(configPath, []byte(data), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() returned error: %v", err)
	}
	if cfg.OutputDir != "/tmp/test" {
		t.Errorf("expected OutputDir '/tmp/test', got '%s'", cfg.OutputDir)
	}
	if cfg.PDFEngine != "weasyprint" {
		t.Errorf("expected PDFEngine 'weasyprint', got '%s'", cfg.PDFEngine)
	}
	if cfg.Provider != "raphael" {
		t.Errorf("expected Provider 'raphael', got '%s'", cfg.Provider)
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(configPath, []byte("not json"), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig() expected error for invalid JSON")
	}
}

func TestLoadConfig_EmptyFields(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	data := `{"output_dir":"","pdf_engine":"","provider":""}`
	if err := os.WriteFile(configPath, []byte(data), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() returned error: %v", err)
	}
	if cfg.OutputDir == "" {
		t.Error("expected OutputDir to be defaulted")
	}
	if cfg.PDFEngine != "wkhtmltopdf" {
		t.Errorf("expected default PDFEngine, got '%s'", cfg.PDFEngine)
	}
	if cfg.Provider != "pollinations" {
		t.Errorf("expected default Provider, got '%s'", cfg.Provider)
	}
}

func TestLoadConfig_DalleKeyFromEnv(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	data := `{}`
	if err := os.WriteFile(configPath, []byte(data), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() returned error: %v", err)
	}
	if cfg.DalleKey != "test-key" {
		t.Errorf("expected DalleKey from env, got '%s'", cfg.DalleKey)
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := &Config{
		OutputDir: "/tmp/test",
		PDFEngine: "weasyprint",
		Config:    image.Config{Provider: "raphael"},
	}

	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read saved config: %v", err)
	}
	if len(data) == 0 {
		t.Error("saved config file is empty")
	}
}
