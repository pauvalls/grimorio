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
	_ = os.WriteFile(filepath.Join(tmpDir, "introduction.md"), []byte("# Introduction\n\nWelcome to the campaign.\n\n## Overview\n\nThis is the overview."), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "lore.md"), []byte("# Lore\n\nAncient history."), 0644)

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
	defer func() { _ = zr.Close() }()

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

// TestEPUBExporter_LanguageIsEnglish is a regression test for the
// i18n-english-default change: EPUB metadata MUST declare `dc:language` of
// `en` (English default) and MUST NOT declare `es` (Spanish is opt-in).
func TestEPUBExporter_LanguageIsEnglish(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "intro.md"), []byte("# Welcome\n\nWelcome, traveler."), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	exporter := NewEPUBExporter()
	path, err := exporter.Export(context.Background(), tmpDir, "The Sword of Vlad")
	if err != nil {
		t.Fatalf("Export() error: %v", err)
	}

	zr, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("open epub: %v", err)
	}
	defer func() { _ = zr.Close() }()

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

	if !strings.Contains(opfContent, "<dc:language>en</dc:language>") {
		t.Errorf("content.opf must declare English locale (i18n-english-default); got:\n%s", opfContent)
	}
	if strings.Contains(opfContent, "<dc:language>es</dc:language>") {
		t.Errorf("content.opf must NOT declare Spanish locale; got:\n%s", opfContent)
	}
}

// TestLocaleDefaultsToEnglish greps the entire `internal/` tree for
// `lang="es"` and `<dc:language>es` to assert the i18n-english-default
// change has not regressed: no Spanish locale metadata may leak into
// source code after the flip. This is the global guard required by the
// spec's locale-metadata-en capability.
func TestLocaleDefaultsToEnglish(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	// The compiler package is rooted at the parent of the test file, which
	// is the `internal/compiler` directory. We walk up to find the
	// module root by looking for `go.mod`.
	moduleRoot := root
	for {
		if _, err := os.Stat(filepath.Join(moduleRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(moduleRoot)
		if parent == moduleRoot {
			t.Fatal("could not locate go.mod from " + root)
		}
		moduleRoot = parent
	}
	internalDir := filepath.Join(moduleRoot, "internal")

	bannedPatterns := []string{
		`lang="es"`,
		`<dc:language>es</dc:language>`,
		`<dc:language>es>`,
	}

	err = filepath.Walk(internalDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		// Test files legitimately contain the Spanish locale string as
		// a negative assertion (e.g. `must NOT declare lang="es"`). The
		// guard targets production code only.
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(data)
		for _, pattern := range bannedPatterns {
			if strings.Contains(content, pattern) {
				t.Errorf("file %s still contains banned Spanish locale pattern %q — i18n-english-default regression", path, pattern)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}
