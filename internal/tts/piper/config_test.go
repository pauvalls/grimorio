package piper

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ModelPath != "" {
		t.Errorf("expected empty ModelPath, got '%s'", cfg.ModelPath)
	}
	if cfg.ConfigPath != "" {
		t.Errorf("expected empty ConfigPath, got '%s'", cfg.ConfigPath)
	}
	if cfg.Port != 5000 {
		t.Errorf("expected Port 5000, got %d", cfg.Port)
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("expected Host '127.0.0.1', got '%s'", cfg.Host)
	}
	if cfg.LengthScale != 1.0 {
		t.Errorf("expected LengthScale 1.0, got %f", cfg.LengthScale)
	}
	if cfg.Volume != 0.8 {
		t.Errorf("expected Volume 0.8, got %f", cfg.Volume)
	}
	if cfg.CacheDir != "" {
		t.Errorf("expected empty CacheDir, got '%s'", cfg.CacheDir)
	}
	if cfg.MaxRestarts != 3 {
		t.Errorf("expected MaxRestarts 3, got %d", cfg.MaxRestarts)
	}
	if cfg.HealthcheckTimeout != 2*time.Second {
		t.Errorf("expected HealthcheckTimeout 2s, got %v", cfg.HealthcheckTimeout)
	}
	if cfg.Player != "auto" {
		t.Errorf("expected Player 'auto', got '%s'", cfg.Player)
	}
	if cfg.Device != "" {
		t.Errorf("expected empty Device, got '%s'", cfg.Device)
	}
	if cfg.PreloadBuffer != 1 {
		t.Errorf("expected PreloadBuffer 1, got %d", cfg.PreloadBuffer)
	}
	if cfg.MaxChunkSize != 150 {
		t.Errorf("expected MaxChunkSize 150, got %d", cfg.MaxChunkSize)
	}
}
