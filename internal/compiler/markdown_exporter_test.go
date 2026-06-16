package compiler

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMarkdownExporter_Export_OrderAndSeparator(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files in arbitrary order
	_ = os.WriteFile(filepath.Join(tmpDir, "appendices.md"), []byte("# Appendices\n\nItems and monsters."), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "introduction.md"), []byte("# Introduction\n\nWelcome."), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "lore.md"), []byte("# Lore\n\nAncient history."), 0644)
	_ = os.MkdirAll(filepath.Join(tmpDir, "chapters"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "chapters", "chapter_02.md"), []byte("# Chapter 2\n\nThe dungeon."), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "chapters", "chapter_01.md"), []byte("# Chapter 1\n\nThe town."), 0644)

	exporter := NewMarkdownExporter()
	path, err := exporter.Export(context.Background(), tmpDir, "Test Campaign")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}

	// Should contain all sections
	if !strings.Contains(string(content), "# Introduction") {
		t.Error("missing Introduction section")
	}
	if !strings.Contains(string(content), "# Lore") {
		t.Error("missing Lore section")
	}
	if !strings.Contains(string(content), "# Chapter 1") {
		t.Error("missing Chapter 1 section")
	}
	if !strings.Contains(string(content), "# Chapter 2") {
		t.Error("missing Chapter 2 section")
	}
	if !strings.Contains(string(content), "# Appendices") {
		t.Error("missing Appendices section")
	}

	// Should be separated by ---
	if !strings.Contains(string(content), "\n---\n") {
		t.Error("missing separator between sections")
	}

	// Verify canonical order: introduction before lore before chapter_01 before chapter_02 before appendices
	introIdx := strings.Index(string(content), "# Introduction")
	loreIdx := strings.Index(string(content), "# Lore")
	ch1Idx := strings.Index(string(content), "# Chapter 1")
	ch2Idx := strings.Index(string(content), "# Chapter 2")
	appIdx := strings.Index(string(content), "# Appendices")

	if introIdx >= loreIdx || loreIdx >= ch1Idx || ch1Idx >= ch2Idx || ch2Idx >= appIdx {
		t.Error("sections are not in canonical order")
	}
}

func TestMarkdownExporter_Export_MissingSectionsSkipped(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "introduction.md"), []byte("# Introduction\n\nOnly this."), 0644)

	exporter := NewMarkdownExporter()
	path, err := exporter.Export(context.Background(), tmpDir, "Test Campaign")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}

	if strings.Contains(string(content), "---") {
		t.Error("should not have separator with only one section")
	}
}

func TestMarkdownExporter_Export_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	exporter := NewMarkdownExporter()
	_, err := exporter.Export(context.Background(), tmpDir, "Test Campaign")
	if err == nil {
		t.Error("expected error for empty campaign directory")
	}
}

func TestMarkdownExporter_Format(t *testing.T) {
	exporter := NewMarkdownExporter()
	if exporter.Format() != "markdown" {
		t.Errorf("expected format 'markdown', got %q", exporter.Format())
	}
}
