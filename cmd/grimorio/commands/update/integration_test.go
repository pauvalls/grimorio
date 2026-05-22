package update

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// T013: Integration test for archive extraction
// ---------------------------------------------------------------------------

func TestIntegration_DownloadAndExtractArchive(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create a mock archive with binary + agents + skills
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "grimorio_Linux_x86_64.tar.gz")

	files := map[string][]byte{
		"grimorio":              []byte("#!/bin/sh\necho 'grimorio v1.3.0'"),
		"agents/test.md":        []byte("# Test Agent"),
		"skills/test/skill.md":  []byte("# Test Skill"),
	}
	createTestTarGz(t, archivePath, files)

	// Compute SHA256 of the archive
	hash := computeFileSHA256(t, archivePath)

	// Mock HTTP server serving archive and checksums
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/grimorio_Linux_x86_64.tar.gz":
			http.ServeFile(w, r, archivePath)
		case "/checksums.txt":
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%s  grimorio_Linux_x86_64.tar.gz\n", hash)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Download the archive
	downloadPath := filepath.Join(tmpDir, "downloaded.tar.gz")
	err := downloadFile(server.URL+"/grimorio_Linux_x86_64.tar.gz", downloadPath, server.Client())
	if err != nil {
		t.Fatalf("downloadFile error = %v", err)
	}

	// Verify file was downloaded
	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
		t.Fatal("downloaded archive does not exist")
	}

	// Verify checksum
	err = verifyChecksum(downloadPath, hash)
	if err != nil {
		t.Fatalf("verifyChecksum error = %v", err)
	}

	// Extract
	extractDir := filepath.Join(tmpDir, "extracted")
	_ = os.MkdirAll(extractDir, 0755)
	err = extractArchive(downloadPath, extractDir)
	if err != nil {
		t.Fatalf("extractArchive error = %v", err)
	}

	// Verify extracted files: binary, agents/, skills/
	expectedFiles := []string{"grimorio", "agents/test.md", "skills/test/skill.md"}
	for _, file := range expectedFiles {
		path := filepath.Join(extractDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("missing extracted file: %s", file)
		}
	}

	// Verify binary content
	binaryPath := filepath.Join(extractDir, "grimorio")
	content, err := os.ReadFile(binaryPath)
	if err != nil {
		t.Fatalf("reading extracted binary: %v", err)
	}
	if !strings.Contains(string(content), "grimorio v1.3.0") {
		t.Errorf("binary content mismatch: %s", string(content))
	}
}

func TestIntegration_DownloadAndExtract_WithBadChecksum(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "grimorio_Linux_x86_64.tar.gz")
	files := map[string][]byte{
		"grimorio": []byte("binary"),
	}
	createTestTarGz(t, archivePath, files)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/grimorio_Linux_x86_64.tar.gz":
			http.ServeFile(w, r, archivePath)
		case "/checksums.txt":
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "0000000000000000000000000000000000000000000000000000000000000000  grimorio_Linux_x86_64.tar.gz\n")
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	downloadPath := filepath.Join(tmpDir, "downloaded.tar.gz")
	err := downloadFile(server.URL+"/grimorio_Linux_x86_64.tar.gz", downloadPath, server.Client())
	if err != nil {
		t.Fatalf("downloadFile error = %v", err)
	}

	err = verifyChecksum(downloadPath, "0000000000000000000000000000000000000000000000000000000000000000")
	if err == nil {
		t.Fatal("verifyChecksum expected error for bad hash")
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Errorf("error should mention checksum mismatch, got: %v", err)
	}
}

func computeFileSHA256(t *testing.T, path string) string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
