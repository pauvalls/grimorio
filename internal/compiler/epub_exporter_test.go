package compiler

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEPUBExporter_Export_ValidStructure(t *testing.T) {
	tmpDir := t.TempDir()
	// Create minimal campaign content
	os.WriteFile(filepath.Join(tmpDir, "introduction.md"), []byte("# Introduction\n\nWelcome to the campaign.\n\n## Overview\n\nThis is the overview."), 0644)
	os.WriteFile(filepath.Join(tmpDir, "lore.md"), []byte("# Lore\n\nAncient history."), 0644)

	exporter := NewEPUBExporter()
	path, err := exporter.Export(context.Background(), tmpDir, "Test Campaign")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify ZIP structure
	zr, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("failed to open epub zip: %v", err)
	}
	defer zr.Close()

	requiredFiles := map[string]bool{
		"mimetype":                 false,
		"META-INF/container.xml":   false,
		"OEBPS/content.opf":        false,
		"OEBPS/toc.ncx":            false,
	}

	for _, f := range zr.File {
		if _, ok := requiredFiles[f.Name]; ok {
			requiredFiles[f.Name] = true
		}
		if strings.HasPrefix(f.Name, "OEBPS/chapter") && strings.HasSuffix(f.Name, ".xhtml") {
			// At least one XHTML chapter
			requiredFiles["has_chapters"] = true
		}
	}

	for name, found := range requiredFiles {
		if !found {
			t.Errorf("missing required EPUB file: %s", name)
		}
	}
}

func TestEPUBExporter_Format(t *testing.T) {
	exporter := NewEPUBExporter()
	if exporter.Format() != "epub" {
		t.Errorf("expected format 'epub', got %q", exporter.Format())
	}
}
