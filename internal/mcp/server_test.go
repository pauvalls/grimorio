package mcp

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/pauvalls/grimorio/internal/config"
)

func TestNewServer(t *testing.T) {
	cfg := config.DefaultConfig()
	server, shutdown := NewServer(cfg)
	if server == nil {
		t.Fatal("NewServer() returned nil")
	}
	if shutdown == nil {
		t.Fatal("NewServer() returned nil shutdown")
	}
	_ = shutdown()
}

func TestNewServer_StructuredLogging_LegacyMode(t *testing.T) {
	// Capture stderr to verify slog output
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w

	// Create a local logger that writes to our pipe
	logger := slog.New(slog.NewTextHandler(w, nil))
	logger.Warn("CANON_LEGACY_MODE is enabled", "component", "canon")

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	if !strings.Contains(output, "CANON_LEGACY_MODE is enabled") {
		t.Errorf("expected structured log warning, got: %s", output)
	}
	if !strings.Contains(output, "component=canon") {
		t.Errorf("expected component field in structured log, got: %s", output)
	}
}

func TestNewServer_StructuredLogging_TTSNotInstalled(t *testing.T) {
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w

	logger := slog.New(slog.NewTextHandler(w, nil))
	logger.Info("Piper TTS not installed", "component", "tts")

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	if !strings.Contains(output, "Piper TTS not installed") {
		t.Errorf("expected structured log info, got: %s", output)
	}
	if !strings.Contains(output, "component=tts") {
		t.Errorf("expected component field in structured log, got: %s", output)
	}
}
