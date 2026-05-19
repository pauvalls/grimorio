package campaign

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- isBinaryOrSkipped ---

func TestIsBinaryOrSkipped(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"file.md", false},
		{"file.json", false},
		{"file.txt", false},
		{"file.svg", true},
		{"file.pdf", true},
		{"file.png", true},
		{"file.jpg", true},
		{"file.jpeg", true},
		{"file.tar.gz", true},
		{"file.zip", true},
		{"nested/file.md", false},
		{"nested/image.svg", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBinaryOrSkipped(tt.name); got != tt.expected {
				t.Errorf("isBinaryOrSkipped(%q) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

// --- buildFileMap ---

func TestBuildFileMap(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some test files
	if err := os.MkdirAll(filepath.Join(tmpDir, "sub"), 0755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "a.md"), []byte("hello"), 0644); err != nil {
		t.Fatalf("write a.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "b.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("write b.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "sub", "c.md"), []byte("world"), 0644); err != nil {
		t.Fatalf("write c.md: %v", err)
	}

	files := buildFileMap(tmpDir)

	expected := map[string]bool{
		"a.md":     true,
		"b.json":   true,
		"sub/c.md": true,
	}

	for rel := range expected {
		if _, ok := files[rel]; !ok {
			t.Errorf("buildFileMap missing entry: %s", rel)
		}
	}

	if len(files) != len(expected) {
		t.Errorf("buildFileMap has %d entries, want %d", len(files), len(expected))
	}
}

func TestBuildFileMap_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	files := buildFileMap(tmpDir)
	if len(files) != 0 {
		t.Errorf("buildFileMap on empty dir should return empty map, got %d entries", len(files))
	}
}

// --- CreateTarGz and GetArchiveTopLevelDir ---

func TestCreateTarGzAndGetTopLevelDir(t *testing.T) {
	tmpDir := t.TempDir()
	campaignDir := filepath.Join(tmpDir, "my-campaign")
	_ = os.MkdirAll(filepath.Join(campaignDir, "maps"), 0755)
	_ = os.WriteFile(filepath.Join(campaignDir, "lore.md"), []byte("# Lore"), 0644)
	_ = os.WriteFile(filepath.Join(campaignDir, "maps", "dungeon.svg"), []byte("<svg/>"), 0644)

	archivePath := filepath.Join(tmpDir, "my-campaign.tar.gz")

	err := CreateTarGz(campaignDir, archivePath)
	if err != nil {
		t.Fatalf("CreateTarGz() error: %v", err)
	}

	// Verify archive exists
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Fatal("CreateTarGz() did not create archive file")
	}

	// Test GetArchiveTopLevelDir
	topDir, err := GetArchiveTopLevelDir(archivePath)
	if err != nil {
		t.Fatalf("GetArchiveTopLevelDir() error: %v", err)
	}
	if topDir != "my-campaign" {
		t.Errorf("GetArchiveTopLevelDir() = %q, want %q", topDir, "my-campaign")
	}

	// Verify archive contents by reading it
	file, err := os.Open(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = gzReader.Close() }()

	tarReader := tar.NewReader(gzReader)
	foundFiles := make(map[string]bool)
	for {
		header, err := tarReader.Next()
		if err != nil {
			break
		}
		foundFiles[header.Name] = true
	}

	expectedFiles := []string{"my-campaign/", "my-campaign/lore.md", "my-campaign/maps/", "my-campaign/maps/dungeon.svg"}
	for _, f := range expectedFiles {
		if !foundFiles[f] {
			t.Errorf("Archive missing entry: %s", f)
		}
	}
}

func TestCreateTarGz_EmptyOutputDir(t *testing.T) {
	tmpDir := t.TempDir()
	emptyCampaign := filepath.Join(tmpDir, "empty-campaign")
	_ = os.MkdirAll(emptyCampaign, 0755)

	archivePath := filepath.Join(tmpDir, "empty.tar.gz")
	err := CreateTarGz(emptyCampaign, archivePath)
	if err != nil {
		t.Fatalf("CreateTarGz() on empty dir error: %v", err)
	}

	// Should create valid gzip
	file, err := os.Open(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatalf("Created archive is not valid gzip: %v", err)
	}
	_ = gzReader.Close()
}

// --- GetArchiveTopLevelDir edge cases ---

func TestGetArchiveTopLevelDir_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "not-a-tar.gz")
	_ = os.WriteFile(invalidPath, []byte("not-gzip-data"), 0644)

	_, err := GetArchiveTopLevelDir(invalidPath)
	if err == nil {
		t.Error("GetArchiveTopLevelDir should error on invalid gzip")
	}
	if !strings.Contains(err.Error(), "not a valid gzip") {
		t.Errorf("Error should mention invalid gzip, got: %v", err)
	}
}

func TestGetArchiveTopLevelDir_NotFound(t *testing.T) {
	_, err := GetArchiveTopLevelDir("/nonexistent/file.tar.gz")
	if err == nil {
		t.Error("GetArchiveTopLevelDir should error on nonexistent file")
	}
}

// --- ExtractTarGz ---

func TestExtractTarGz(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a source tar.gz to extract
	sourceCampaign := filepath.Join(tmpDir, "src-campaign")
	_ = os.MkdirAll(sourceCampaign, 0755)
	_ = os.WriteFile(filepath.Join(sourceCampaign, "note.md"), []byte("# Note"), 0644)
	_ = os.MkdirAll(filepath.Join(sourceCampaign, "sub"), 0755)
	_ = os.WriteFile(filepath.Join(sourceCampaign, "sub", "deep.md"), []byte("deep"), 0644)

	archivePath := filepath.Join(tmpDir, "test.tar.gz")
	if err := CreateTarGz(sourceCampaign, archivePath); err != nil {
		t.Fatal(err)
	}

	// Extract to a different location
	extractDir := filepath.Join(tmpDir, "output")
	_ = os.MkdirAll(extractDir, 0755)

	if err := ExtractTarGz(archivePath, extractDir); err != nil {
		t.Fatalf("ExtractTarGz() error: %v", err)
	}

	// Verify extracted files
	extractedCampaign := filepath.Join(extractDir, "src-campaign")
	if _, err := os.Stat(filepath.Join(extractedCampaign, "note.md")); os.IsNotExist(err) {
		t.Error("ExtractTarGz did not extract note.md")
	}
	if _, err := os.Stat(filepath.Join(extractedCampaign, "sub", "deep.md")); os.IsNotExist(err) {
		t.Error("ExtractTarGz did not extract sub/deep.md")
	}
}

func TestExtractTarGz_DirectoryTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a malicious tar.gz with ../ paths
	archivePath := filepath.Join(tmpDir, "malicious.tar.gz")
	outFile, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	gzWriter := gzip.NewWriter(outFile)
	tarWriter := tar.NewWriter(gzWriter)

	// Add a malicious entry
	hdr := &tar.Header{
		Name:     "../../etc/passwd",
		Size:     4,
		Typeflag: tar.TypeReg,
		Mode:     0644,
	}
	if err := tarWriter.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write([]byte("root")); err != nil {
		t.Fatal(err)
	}
	_ = tarWriter.Close()
	_ = gzWriter.Close()
	_ = outFile.Close()

	// Try extracting — should fail with security error
	extractDir := filepath.Join(tmpDir, "safe-extract")
	_ = os.MkdirAll(extractDir, 0755)

	err = ExtractTarGz(archivePath, extractDir)
	if err == nil {
		t.Fatal("ExtractTarGz should reject directory traversal")
	}
	if !strings.Contains(err.Error(), "security error") {
		t.Errorf("Error should mention security error, got: %v", err)
	}
}

func TestExtractTarGz_InvalidGzip(t *testing.T) {
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "invalid.gz")
	_ = os.WriteFile(invalidPath, []byte("not-gzip"), 0644)

	err := ExtractTarGz(invalidPath, tmpDir)
	if err == nil {
		t.Error("ExtractTarGz should error on invalid gzip")
	}
	if !strings.Contains(err.Error(), "not a valid gzip") {
		t.Errorf("Error should mention invalid gzip, got: %v", err)
	}
}

// --- Export → Import Round Trip ---

func TestExportImportRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	campaignName := "roundtrip-campaign"
	sourceDir := filepath.Join(tmpDir, campaignName)

	// Create a campaign
	_ = os.MkdirAll(sourceDir, 0755)
	_ = os.WriteFile(filepath.Join(sourceDir, "lore.md"), []byte("# Round Trip Lore"), 0644)
	_ = os.MkdirAll(filepath.Join(sourceDir, "npcs"), 0755)
	_ = os.WriteFile(filepath.Join(sourceDir, "npcs", "gandalf.md"), []byte("## Gandalf\nA wizard"), 0644)

	// Export
	archivePath := filepath.Join(tmpDir, campaignName+".tar.gz")
	if err := CreateTarGz(sourceDir, archivePath); err != nil {
		t.Fatal(err)
	}

	// Remove original
	_ = os.RemoveAll(sourceDir)

	// Import
	importDir := filepath.Join(tmpDir, "imported")
	_ = os.MkdirAll(importDir, 0755)
	if err := ExtractTarGz(archivePath, importDir); err != nil {
		t.Fatal(err)
	}

	// Verify
	importedCampaign := filepath.Join(importDir, campaignName)
	data, err := os.ReadFile(filepath.Join(importedCampaign, "lore.md"))
	if err != nil {
		t.Fatalf("Failed to read imported lore.md: %v", err)
	}
	if string(data) != "# Round Trip Lore" {
		t.Errorf("Imported lore.md content mismatch: got %q", string(data))
	}

	data2, err := os.ReadFile(filepath.Join(importedCampaign, "npcs", "gandalf.md"))
	if err != nil {
		t.Fatalf("Failed to read imported npcs/gandalf.md: %v", err)
	}
	if !strings.Contains(string(data2), "A wizard") {
		t.Errorf("Imported gandalf.md content mismatch: got %q", string(data2))
	}
}
