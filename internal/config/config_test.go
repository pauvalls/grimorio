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
	if cfg.PDFEngine != "" {
		t.Errorf("expected PDFEngine empty string (auto-detect), got '%s'", cfg.PDFEngine)
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
	if cfg.CompilerVersion != 2 {
		t.Errorf("expected default CompilerVersion 2, got %d", cfg.CompilerVersion)
	}
	if cfg.ImageCacheDir == "" {
		t.Error("expected default ImageCacheDir to be set")
	}
	if cfg.ImageCacheSize != 50 {
		t.Errorf("expected default ImageCacheSize 50, got %d", cfg.ImageCacheSize)
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
	if cfg.PDFEngine != "" {
		t.Errorf("expected default PDFEngine empty (auto-detect), got '%s'", cfg.PDFEngine)
	}
}

func TestLoadConfig_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	data := `{"output_dir":"/tmp/test","pdf_engine":"weasyprint","image_provider":"raphael","compiler_version":1}`
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
	if cfg.CompilerVersion != 1 {
		t.Errorf("expected CompilerVersion 1, got %d", cfg.CompilerVersion)
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
	if cfg.PDFEngine != "" {
		t.Errorf("expected default PDFEngine empty (auto-detect), got '%s'", cfg.PDFEngine)
	}
	if cfg.Provider != "pollinations" {
		t.Errorf("expected default Provider, got '%s'", cfg.Provider)
	}
}

func TestLoadConfig_DalleKeyFromEnv(t *testing.T) {
	_ = os.Setenv("OPENAI_API_KEY", "test-key")
	defer func() { _ = os.Unsetenv("OPENAI_API_KEY") }()

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

func TestDefaultTTSConfig(t *testing.T) {
	cfg := DefaultTTSConfig()
	if cfg.Enabled != false {
		t.Errorf("expected TTS.Enabled false, got %v", cfg.Enabled)
	}
	if cfg.Piper.Port != 5000 {
		t.Errorf("expected Piper.Port 5000, got %d", cfg.Piper.Port)
	}
	if cfg.Piper.Host != "127.0.0.1" {
		t.Errorf("expected Piper.Host '127.0.0.1', got '%s'", cfg.Piper.Host)
	}
	if cfg.Piper.LengthScale != 1.0 {
		t.Errorf("expected Piper.LengthScale 1.0, got %f", cfg.Piper.LengthScale)
	}
	if cfg.Piper.Volume != 0.8 {
		t.Errorf("expected Piper.Volume 0.8, got %f", cfg.Piper.Volume)
	}
	if cfg.Piper.MaxRestarts != 3 {
		t.Errorf("expected Piper.MaxRestarts 3, got %d", cfg.Piper.MaxRestarts)
	}
	if cfg.Chunker.MaxChunkSize != 150 {
		t.Errorf("expected Chunker.MaxChunkSize 150, got %d", cfg.Chunker.MaxChunkSize)
	}
	if cfg.Audio.Player != "auto" {
		t.Errorf("expected Audio.Player 'auto', got '%s'", cfg.Audio.Player)
	}
	if cfg.Audio.PreloadBuffer != 1 {
		t.Errorf("expected Audio.PreloadBuffer 1, got %d", cfg.Audio.PreloadBuffer)
	}
}

func TestDefaultConfig_IncludesTTSDefaults(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.TTS.Piper.Port != 5000 {
		t.Errorf("expected default TTS.Piper.Port 5000, got %d", cfg.TTS.Piper.Port)
	}
	if cfg.TTS.Chunker.MaxChunkSize != 150 {
		t.Errorf("expected default TTS.Chunker.MaxChunkSize 150, got %d", cfg.TTS.Chunker.MaxChunkSize)
	}
}

func TestLoadConfig_TTSFields(t *testing.T) {
	// Isolate from environment variables that would override config file values
	envVars := []string{
		"GRIMORIO_TTS_ENABLED",
		"PIPER_MODEL_PATH", "PIPER_CONFIG_PATH", "PIPER_PORT", "PIPER_HOST",
		"PIPER_LENGTH_SCALE", "PIPER_VOLUME", "PIPER_CACHE_DIR", "PIPER_MAX_RESTARTS",
		"CHUNKER_MAX_SIZE", "AUDIO_PLAYER", "AUDIO_DEVICE", "AUDIO_PRELOAD_BUFFER",
	}
	oldVals := make(map[string]string)
	for _, k := range envVars {
		oldVals[k] = os.Getenv(k)
		_ = os.Unsetenv(k)
	}
	t.Cleanup(func() {
		for k, v := range oldVals {
			if v != "" {
				_ = os.Setenv(k, v)
			} else {
				_ = os.Unsetenv(k)
			}
		}
	})

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	data := `{
		"output_dir": "/tmp/test",
		"tts": {
			"enabled": true,
			"piper": {
				"model_path": "/models/es.onnx",
				"port": 8080,
				"length_scale": 1.2,
				"volume": 0.5
			},
			"chunker": {
				"max_chunk_size": 200
			},
			"audio": {
				"player": "aplay",
				"preload_buffer": 2
			}
		}
	}`
	if err := os.WriteFile(configPath, []byte(data), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() returned error: %v", err)
	}
	if !cfg.TTS.Enabled {
		t.Error("expected TTS.Enabled true")
	}
	if cfg.TTS.Piper.ModelPath != "/models/es.onnx" {
		t.Errorf("expected ModelPath '/models/es.onnx', got '%s'", cfg.TTS.Piper.ModelPath)
	}
	if cfg.TTS.Piper.Port != 8080 {
		t.Errorf("expected Port 8080, got %d", cfg.TTS.Piper.Port)
	}
	if cfg.TTS.Piper.LengthScale != 1.2 {
		t.Errorf("expected LengthScale 1.2, got %f", cfg.TTS.Piper.LengthScale)
	}
	if cfg.TTS.Piper.Volume != 0.5 {
		t.Errorf("expected Volume 0.5, got %f", cfg.TTS.Piper.Volume)
	}
	if cfg.TTS.Chunker.MaxChunkSize != 200 {
		t.Errorf("expected MaxChunkSize 200, got %d", cfg.TTS.Chunker.MaxChunkSize)
	}
	if cfg.TTS.Audio.Player != "aplay" {
		t.Errorf("expected Player 'aplay', got '%s'", cfg.TTS.Audio.Player)
	}
	if cfg.TTS.Audio.PreloadBuffer != 2 {
		t.Errorf("expected PreloadBuffer 2, got %d", cfg.TTS.Audio.PreloadBuffer)
	}
}

func TestLoadConfig_TTSEnvOverrides(t *testing.T) {
	_ = os.Setenv("GRIMORIO_TTS_ENABLED", "true")
	_ = os.Setenv("PIPER_MODEL_PATH", "/env/model.onnx")
	_ = os.Setenv("PIPER_PORT", "9000")
	_ = os.Setenv("PIPER_LENGTH_SCALE", "0.9")
	_ = os.Setenv("PIPER_VOLUME", "0.6")
	_ = os.Setenv("AUDIO_PLAYER", "ffplay")
	defer func() {
		_ = os.Unsetenv("GRIMORIO_TTS_ENABLED")
		_ = os.Unsetenv("PIPER_MODEL_PATH")
		_ = os.Unsetenv("PIPER_PORT")
		_ = os.Unsetenv("PIPER_LENGTH_SCALE")
		_ = os.Unsetenv("PIPER_VOLUME")
		_ = os.Unsetenv("AUDIO_PLAYER")
	}()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	_ = os.WriteFile(configPath, []byte(`{}`), 0644)

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() returned error: %v", err)
	}
	if !cfg.TTS.Enabled {
		t.Error("expected TTS.Enabled true from env")
	}
	if cfg.TTS.Piper.ModelPath != "/env/model.onnx" {
		t.Errorf("expected ModelPath from env, got '%s'", cfg.TTS.Piper.ModelPath)
	}
	if cfg.TTS.Piper.Port != 9000 {
		t.Errorf("expected Port 9000 from env, got %d", cfg.TTS.Piper.Port)
	}
	if cfg.TTS.Piper.LengthScale != 0.9 {
		t.Errorf("expected LengthScale 0.9 from env, got %f", cfg.TTS.Piper.LengthScale)
	}
	if cfg.TTS.Piper.Volume != 0.6 {
		t.Errorf("expected Volume 0.6 from env, got %f", cfg.TTS.Piper.Volume)
	}
	if cfg.TTS.Audio.Player != "ffplay" {
		t.Errorf("expected Player 'ffplay' from env, got '%s'", cfg.TTS.Audio.Player)
	}
}

func TestLoadConfig_TTSDefaultsApplied(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	// Provide a minimal TTS block with only some fields set.
	data := `{"tts": {"enabled": true, "piper": {"model_path": "/models/test.onnx"}}}`
	_ = os.WriteFile(configPath, []byte(data), 0644)

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() returned error: %v", err)
	}
	// Fields present in JSON should be preserved.
	if cfg.TTS.Piper.ModelPath != "/models/test.onnx" {
		t.Errorf("expected ModelPath '/models/test.onnx', got '%s'", cfg.TTS.Piper.ModelPath)
	}
	// Zero-valued fields should receive defaults.
	if cfg.TTS.Piper.Port != 5000 {
		t.Errorf("expected default Port 5000, got %d", cfg.TTS.Piper.Port)
	}
	if cfg.TTS.Piper.Host != "127.0.0.1" {
		t.Errorf("expected default Host '127.0.0.1', got '%s'", cfg.TTS.Piper.Host)
	}
	if cfg.TTS.Piper.LengthScale != 1.0 {
		t.Errorf("expected default LengthScale 1.0, got %f", cfg.TTS.Piper.LengthScale)
	}
	if cfg.TTS.Chunker.MaxChunkSize != 150 {
		t.Errorf("expected default MaxChunkSize 150, got %d", cfg.TTS.Chunker.MaxChunkSize)
	}
}
