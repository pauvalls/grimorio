package compiler

import (
	"archive/zip"
	"context"
	"io"
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

// TestEPUBExporter_LanguageIsSpanish is a regression test for Fase 4 i18n:
// Spanish-language campaigns must be tagged as `es` in the EPUB metadata
// and MUST NOT regress to the English `en` tag.
func TestEPUBExporter_LanguageIsSpanish(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "intro.md"), []byte("# Introducción\n\nBienvenidos."), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	exporter := NewEPUBExporter()
	path, err := exporter.Export(context.Background(), tmpDir, "La Hoja de Vlad")
	if err != nil {
		t.Fatalf("Export() error: %v", err)
	}

	zr, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("open epub: %v", err)
	}
	defer zr.Close()

	var opfContent string
	for _, f := range zr.File {
		if f.Name == "OEBPS/content.opf" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("open opf: %v", err)
			}
			buf, err := io.ReadAll(rc)
			_ = rc.Close()
			if err != nil {
				t.Fatalf("read opf: %v", err)
			}
			opfContent = string(buf)
		}
	}

	if opfContent == "" {
		t.Fatal("OEBPS/content.opf missing from EPUB")
	}

	if !strings.Contains(opfContent, "<dc:language>es</dc:language>") {
		t.Errorf("content.opf must declare Spanish locale; got:\n%s", opfContent)
	}
	if strings.Contains(opfContent, "<dc:language>en</dc:language>") {
		t.Errorf("content.opf must NOT declare English locale; got:\n%s", opfContent)
	}
}
