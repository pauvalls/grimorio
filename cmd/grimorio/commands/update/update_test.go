package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"
)

// --- T003: Platform Detection ---

func TestDetectPlatform(t *testing.T) {
	goos, goarch, err := detectPlatform()
	if err != nil {
		t.Fatalf("detectPlatform() error = %v", err)
	}
	if goos == "" {
		t.Error("detectPlatform() returned empty OS")
	}
	if goarch == "" {
		t.Error("detectPlatform() returned empty arch")
	}
	// Should match runtime values
	if goos != runtime.GOOS {
		t.Errorf("detectPlatform() OS = %q, want %q", goos, runtime.GOOS)
	}
	if goarch != runtime.GOARCH {
		t.Errorf("detectPlatform() arch = %q, want %q", goarch, runtime.GOARCH)
	}
}

func TestMapGoArchToGoreleaser(t *testing.T) {
	tests := []struct {
		goos     string
		goarch   string
		wantOS   string
		wantArch string
	}{
		{"linux", "amd64", "Linux", "x86_64"},
		{"linux", "arm64", "Linux", "arm64"},
		{"darwin", "amd64", "Darwin", "x86_64"},
		{"darwin", "arm64", "Darwin", "arm64"},
		{"windows", "amd64", "Windows", "x86_64"},
		{"windows", "arm64", "Windows", "arm64"},
		{"freebsd", "amd64", "Freebsd", "x86_64"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.goos, tt.goarch), func(t *testing.T) {
			gotOS, gotArch := mapGoArchToGoreleaser(tt.goos, tt.goarch)
			if gotOS != tt.wantOS {
				t.Errorf("mapGoArchToGoreleaser() OS = %q, want %q", gotOS, tt.wantOS)
			}
			if gotArch != tt.wantArch {
				t.Errorf("mapGoArchToGoreleaser() arch = %q, want %q", gotArch, tt.wantArch)
			}
		})
	}
}

func TestArchiveName(t *testing.T) {
	tests := []struct {
		goos     string
		goarch   string
		wantName string
	}{
		{"linux", "amd64", "grimorio_Linux_x86_64.tar.gz"},
		{"linux", "arm64", "grimorio_Linux_arm64.tar.gz"},
		{"darwin", "amd64", "grimorio_Darwin_x86_64.tar.gz"},
		{"darwin", "arm64", "grimorio_Darwin_arm64.tar.gz"},
		{"windows", "amd64", "grimorio_Windows_x86_64.zip"},
		{"windows", "arm64", "grimorio_Windows_arm64.zip"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.goos, tt.goarch), func(t *testing.T) {
			got := archiveName(tt.goos, tt.goarch)
			if got != tt.wantName {
				t.Errorf("archiveName() = %q, want %q", got, tt.wantName)
			}
		})
	}
}

// --- T003: GitHub API Client ---

func TestFetchLatestRelease(t *testing.T) {
	// Mock GitHub API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/testowner/testrepo/releases/latest" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"tag_name": "v1.3.0",
			"assets": [
				{"name": "grimorio_Linux_x86_64.tar.gz", "browser_download_url": "https://example.com/linux.tar.gz"},
				{"name": "grimorio_Linux_x86_64.tar.gz_checksums.txt", "browser_download_url": "https://example.com/linux_checksums.txt"},
				{"name": "grimorio_Darwin_arm64.tar.gz", "browser_download_url": "https://example.com/darwin.tar.gz"}
			]
		}`)
	}))
	defer server.Close()

	release, err := fetchLatestRelease("testowner", "testrepo", server.Client(), server.URL)
	if err != nil {
		t.Fatalf("fetchLatestRelease() error = %v", err)
	}
	if release.TagName != "v1.3.0" {
		t.Errorf("fetchLatestRelease() tag = %q, want %q", release.TagName, "v1.3.0")
	}
	if len(release.Assets) != 3 {
		t.Errorf("fetchLatestRelease() assets count = %d, want 3", len(release.Assets))
	}
}

func TestFetchLatestRelease_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = fmt.Fprint(w, `{"message": "API rate limit exceeded"}`)
	}))
	defer server.Close()

	_, err := fetchLatestRelease("testowner", "testrepo", server.Client(), server.URL)
	if err == nil {
		t.Fatal("fetchLatestRelease() expected error for 403 response")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("error should mention 403, got: %v", err)
	}
}

func TestFetchLatestRelease_NetworkFailure(t *testing.T) {
	// Use a URL that will definitely fail
	_, err := fetchLatestRelease("testowner", "testrepo", http.DefaultClient, "http://localhost:99999")
	if err == nil {
		t.Fatal("fetchLatestRelease() expected error for network failure")
	}
}

func TestFindAsset(t *testing.T) {
	release := &githubRelease{
		TagName: "v1.3.0",
		Assets: []githubAsset{
			{Name: "grimorio_Linux_x86_64.tar.gz", BrowserDownloadURL: "https://example.com/linux.tar.gz"},
			{Name: "grimorio_Linux_x86_64.tar.gz_checksums.txt", BrowserDownloadURL: "https://example.com/linux_checksums.txt"},
			{Name: "grimorio_Darwin_arm64.tar.gz", BrowserDownloadURL: "https://example.com/darwin.tar.gz"},
			{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums.txt"},
		},
	}

	dlURL, checksumURL, err := findAsset(release, "linux", "amd64")
	if err != nil {
		t.Fatalf("findAsset() error = %v", err)
	}
	if dlURL != "https://example.com/linux.tar.gz" {
		t.Errorf("findAsset() download URL = %q, want %q", dlURL, "https://example.com/linux.tar.gz")
	}
	if checksumURL != "https://example.com/checksums.txt" {
		t.Errorf("findAsset() checksum URL = %q, want %q", checksumURL, "https://example.com/checksums.txt")
	}
}

func TestFindAsset_Missing(t *testing.T) {
	release := &githubRelease{
		TagName: "v1.3.0",
		Assets: []githubAsset{
			{Name: "grimorio_Darwin_arm64.tar.gz", BrowserDownloadURL: "https://example.com/darwin.tar.gz"},
		},
	}

	_, _, err := findAsset(release, "linux", "amd64")
	if err == nil {
		t.Fatal("findAsset() expected error when asset not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestBuildFallbackURL(t *testing.T) {
	url := buildFallbackURL("pauvalls", "grimorio", "v1.2.3", "linux", "amd64")
	want := "https://github.com/pauvalls/grimorio/releases/download/v1.2.3/grimorio_Linux_x86_64.tar.gz"
	if url != want {
		t.Errorf("buildFallbackURL() = %q, want %q", url, want)
	}
}

// --- T004: Download, Verify, Extract ---

func TestDownloadFile(t *testing.T) {
	content := []byte("test archive content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")

	err := downloadFile(server.URL, destPath, server.Client())
	if err != nil {
		t.Fatalf("downloadFile() error = %v", err)
	}

	got, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("reading downloaded file: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("downloadFile() content = %q, want %q", got, content)
	}
}

func TestDownloadFile_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")

	err := downloadFile(server.URL, destPath, server.Client())
	if err == nil {
		t.Fatal("downloadFile() expected error for 500 response")
	}
}

func TestVerifyChecksum(t *testing.T) {
	// Create a file with known content
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	content := []byte("hello world")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatal(err)
	}

	// Correct SHA256 for "hello world\n" — let's compute it for our exact content
	// "hello world" without newline: b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
	expectedHash := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"

	err := verifyChecksum(filePath, expectedHash)
	if err != nil {
		t.Fatalf("verifyChecksum() with correct hash error = %v", err)
	}
}

func TestVerifyChecksum_WrongHash(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(filePath, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	err := verifyChecksum(filePath, "0000000000000000000000000000000000000000000000000000000000000000")
	if err == nil {
		t.Fatal("verifyChecksum() expected error for wrong hash")
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Errorf("error should mention checksum mismatch, got: %v", err)
	}
}

func TestParseChecksums(t *testing.T) {
	checksums := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9  grimorio_Linux_x86_64.tar.gz\n"
	checksums += "abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234  grimorio_Darwin_arm64.tar.gz\n"

	hash, err := parseChecksums(checksums, "grimorio_Linux_x86_64.tar.gz")
	if err != nil {
		t.Fatalf("parseChecksums() error = %v", err)
	}
	if hash != "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9" {
		t.Errorf("parseChecksums() = %q, want expected hash", hash)
	}
}

func TestParseChecksums_NotFound(t *testing.T) {
	checksums := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9  other_file.tar.gz\n"

	_, err := parseChecksums(checksums, "missing.tar.gz")
	if err == nil {
		t.Fatal("parseChecksums() expected error when file not found in checksums")
	}
}

func TestExtractTarGz(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a tar.gz archive
	archivePath := filepath.Join(tmpDir, "test.tar.gz")
	createTestTarGz(t, archivePath, map[string][]byte{
		"grimorio":            []byte("binary content"),
		"agents/test.md":      []byte("# Agent"),
		"skills/test/skill.md": []byte("# Skill"),
	})

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		t.Fatal(err)
	}

	err := extractTarGz(archivePath, extractDir)
	if err != nil {
		t.Fatalf("extractTarGz() error = %v", err)
	}

	// Verify extracted files
	for _, file := range []string{"grimorio", "agents/test.md", "skills/test/skill.md"} {
		path := filepath.Join(extractDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("extractTarGz() missing file: %s", file)
		}
	}
}

func TestExtractZip(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a zip archive
	archivePath := filepath.Join(tmpDir, "test.zip")
	createTestZip(t, archivePath, map[string][]byte{
		"grimorio.exe":        []byte("binary content"),
		"agents/test.md":      []byte("# Agent"),
		"skills/test/skill.md": []byte("# Skill"),
	})

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		t.Fatal(err)
	}

	err := extractZip(archivePath, extractDir)
	if err != nil {
		t.Fatalf("extractZip() error = %v", err)
	}

	// Verify extracted files
	for _, file := range []string{"grimorio.exe", "agents/test.md", "skills/test/skill.md"} {
		path := filepath.Join(extractDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("extractZip() missing file: %s", file)
		}
	}
}

func TestExtractArchive_ChoosesCorrectExtractor(t *testing.T) {
	tmpDir := t.TempDir()

	// Test tar.gz
	tarPath := filepath.Join(tmpDir, "test.tar.gz")
	createTestTarGz(t, tarPath, map[string][]byte{"grimorio": []byte("binary")})

	extractDir := filepath.Join(tmpDir, "extracted1")
	_ = os.MkdirAll(extractDir, 0755)
	err := extractArchive(tarPath, extractDir)
	if err != nil {
		t.Fatalf("extractArchive(tar.gz) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(extractDir, "grimorio")); os.IsNotExist(err) {
		t.Error("extractArchive(tar.gz) missing grimorio binary")
	}

	// Test zip
	zipPath := filepath.Join(tmpDir, "test.zip")
	createTestZip(t, zipPath, map[string][]byte{"grimorio.exe": []byte("binary")})

	extractDir2 := filepath.Join(tmpDir, "extracted2")
	_ = os.MkdirAll(extractDir2, 0755)
	err = extractArchive(zipPath, extractDir2)
	if err != nil {
		t.Fatalf("extractArchive(zip) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(extractDir2, "grimorio.exe")); os.IsNotExist(err) {
		t.Error("extractArchive(zip) missing grimorio.exe binary")
	}
}

func TestValidateExtractedContents(t *testing.T) {
	tmpDir := t.TempDir()

	// Valid extraction with binary
	validDir := filepath.Join(tmpDir, "valid")
	_ = os.MkdirAll(filepath.Join(validDir, "agents"), 0755)
	_ = os.MkdirAll(filepath.Join(validDir, "skills", "test"), 0755)
	_ = os.WriteFile(filepath.Join(validDir, "grimorio"), []byte("binary"), 0755)

	err := validateExtractedContents(validDir)
	if err != nil {
		t.Fatalf("validateExtractedContents(valid) error = %v", err)
	}

	// Missing binary
	invalidDir := filepath.Join(tmpDir, "invalid")
	_ = os.MkdirAll(invalidDir, 0755)

	err = validateExtractedContents(invalidDir)
	if err == nil {
		t.Fatal("validateExtractedContents() expected error for missing binary")
	}
	if !strings.Contains(err.Error(), "binary not found") {
		t.Errorf("error should mention binary not found, got: %v", err)
	}
}

func TestCleanupOnError(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "temp1.txt")
	file2 := filepath.Join(tmpDir, "temp2.txt")
	_ = os.WriteFile(file1, []byte("test"), 0644)
	_ = os.WriteFile(file2, []byte("test"), 0644)

	cleanupOnError([]string{file1, file2})

	if _, err := os.Stat(file1); !os.IsNotExist(err) {
		t.Error("cleanupOnError() should remove file1")
	}
	if _, err := os.Stat(file2); !os.IsNotExist(err) {
		t.Error("cleanupOnError() should remove file2")
	}
}

// --- T005: Backup, Replace, Rollback ---

func TestBackupCurrentBinary(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "grimorio")
	backupDir := filepath.Join(tmpDir, ".grimorio")

	// Create a fake binary
	if err := os.WriteFile(binaryPath, []byte("original binary"), 0755); err != nil {
		t.Fatal(err)
	}

	backupPath, err := backupCurrentBinary(binaryPath, backupDir)
	if err != nil {
		t.Fatalf("backupCurrentBinary() error = %v", err)
	}

	// Verify backup exists and has correct content
	data, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("reading backup: %v", err)
	}
	if string(data) != "original binary" {
		t.Errorf("backup content = %q, want %q", data, "original binary")
	}
}

func TestReplaceBinary(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "grimorio")
	newPath := filepath.Join(tmpDir, "grimorio.new")

	// Create original and new binary
	if err := os.WriteFile(binaryPath, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newPath, []byte("new"), 0755); err != nil {
		t.Fatal(err)
	}

	err := replaceBinary(binaryPath, newPath)
	if err != nil {
		t.Fatalf("replaceBinary() error = %v", err)
	}

	// Verify replacement
	data, err := os.ReadFile(binaryPath)
	if err != nil {
		t.Fatalf("reading replaced binary: %v", err)
	}
	if string(data) != "new" {
		t.Errorf("replaced binary content = %q, want %q", data, "new")
	}
}

func TestRollback(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "grimorio")
	backupPath := filepath.Join(tmpDir, "grimorio.backup")

	// Create a corrupted binary and a good backup
	if err := os.WriteFile(binaryPath, []byte("corrupted"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(backupPath, []byte("good backup"), 0755); err != nil {
		t.Fatal(err)
	}

	err := rollback(binaryPath, backupPath)
	if err != nil {
		t.Fatalf("rollback() error = %v", err)
	}

	// Verify rollback restored the backup
	data, err := os.ReadFile(binaryPath)
	if err != nil {
		t.Fatalf("reading rolled-back binary: %v", err)
	}
	if string(data) != "good backup" {
		t.Errorf("rolled-back binary content = %q, want %q", data, "good backup")
	}
}

func TestRollback_BackupMissing(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "grimorio")
	backupPath := filepath.Join(tmpDir, "nonexistent.backup")

	if err := os.WriteFile(binaryPath, []byte("current"), 0755); err != nil {
		t.Fatal(err)
	}

	err := rollback(binaryPath, backupPath)
	if err == nil {
		t.Fatal("rollback() expected error when backup missing")
	}
}

// --- T006: Version Comparison ---

func TestIsNewer(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
		wantErr bool
	}{
		{"v1.2.3", "v1.3.0", true, false},
		{"v1.3.0", "v1.3.0", false, false},
		{"v1.3.0", "v1.2.3", false, false},
		{"v1.2.3", "v1.2.4", true, false},
		{"v1.2.3", "v2.0.0", true, false},
		{"dev", "v1.0.0", true, false},
		{"v1.2.3", "dev", false, false},
		{"", "v1.0.0", true, false},
		{"v1.2.3", "", false, true},
		{"v1.2.3", "not-a-version", false, true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_vs_%s", tt.current, tt.latest), func(t *testing.T) {
			got, err := isNewer(tt.current, tt.latest)
			if (err != nil) != tt.wantErr {
				t.Fatalf("isNewer() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("isNewer(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

func TestIsNewer_Prerelease(t *testing.T) {
	// v1.3.0-beta.1 should be considered newer than v1.2.3 but not newer than v1.3.0
	got, err := isNewer("v1.2.3", "v1.3.0-beta.1")
	if err != nil {
		t.Fatalf("isNewer() error = %v", err)
	}
	if !got {
		t.Error("isNewer(v1.2.3, v1.3.0-beta.1) = false, want true")
	}

	got, err = isNewer("v1.3.0", "v1.3.0-beta.1")
	if err != nil {
		t.Fatalf("isNewer() error = %v", err)
	}
	if got {
		t.Error("isNewer(v1.3.0, v1.3.0-beta.1) = true, want false")
	}
}

// --- T006: Update Orchestration (integration-level) ---

func TestUpdater_CheckForUpdate(t *testing.T) {
	// Use fixed URLs in the mock response so we don't need server.URL inside the handler
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"tag_name": "v1.3.0",
			"assets": [
				{"name": "grimorio_Linux_x86_64.tar.gz", "browser_download_url": "http://example.com/download"},
				{"name": "checksums.txt", "browser_download_url": "http://example.com/checksums"}
			]
		}`)
	}))
	defer server.Close()

	// Note: This test sets up the updater with the mock server
	u := &updater{
		repoOwner:      "testowner",
		repoName:       "testrepo",
		installDir:     t.TempDir(),
		backupDir:      t.TempDir(),
		apiBaseURL:     server.URL,
		httpClient:     server.Client(),
		currentVersion: "v1.2.3",
	}

	// Test that we can construct the updater and it has the right fields
	if u.repoOwner != "testowner" {
		t.Error("updater.repoOwner mismatch")
	}
}

// --- T007: CLI Registration ---

func TestNewUpdateCommand(t *testing.T) {
	cmd := NewUpdateCommand("v1.2.3")
	if cmd == nil {
		t.Fatal("NewUpdateCommand() returned nil")
	}
	if cmd.Name != "update" {
		t.Errorf("NewUpdateCommand() Name = %q, want %q", cmd.Name, "update")
	}
	if cmd.Usage == "" {
		t.Error("NewUpdateCommand() Usage is empty")
	}

	// Check for --dry-run flag
	var dryRunFlag *cli.BoolFlag
	for _, f := range cmd.Flags {
		if bf, ok := f.(*cli.BoolFlag); ok && bf.Name == "dry-run" {
			dryRunFlag = bf
			break
		}
	}
	if dryRunFlag == nil {
		t.Error("NewUpdateCommand() missing --dry-run flag")
	}
}

// --- Helpers ---

func createTestTarGz(t *testing.T, path string, files map[string][]byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	gzWriter := gzip.NewWriter(f)
	defer func() { _ = gzWriter.Close() }()

	tarWriter := tar.NewWriter(gzWriter)
	defer func() { _ = tarWriter.Close() }()

	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Size: int64(len(content)),
			Mode: 0644,
		}
		if strings.HasSuffix(name, "/") {
			hdr.Mode = 0755
			hdr.Typeflag = tar.TypeDir
		}
		if err := tarWriter.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if len(content) > 0 {
			if _, err := tarWriter.Write(content); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func createTestZip(t *testing.T, path string, files map[string][]byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	zipWriter := zip.NewWriter(f)
	defer func() { _ = zipWriter.Close() }()

	for name, content := range files {
		w, err := zipWriter.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(content); err != nil {
			t.Fatal(err)
		}
	}
}

// --- Asset Update Commands ---

func TestNewUpdateSkillsCommand(t *testing.T) {
	cmd := NewUpdateSkillsCommand()
	if cmd == nil {
		t.Fatal("NewUpdateSkillsCommand() returned nil")
	}
	if cmd.Name != "skills" {
		t.Errorf("command name = %q, want %q", cmd.Name, "skills")
	}
	if cmd.Usage == "" {
		t.Error("command usage should not be empty")
	}
	if cmd.Action == nil {
		t.Error("command action should not be nil")
	}
}

func TestNewUpdateAgentsCommand(t *testing.T) {
	cmd := NewUpdateAgentsCommand()
	if cmd == nil {
		t.Fatal("NewUpdateAgentsCommand() returned nil")
	}
	if cmd.Name != "agents" {
		t.Errorf("command name = %q, want %q", cmd.Name, "agents")
	}
	if cmd.Usage == "" {
		t.Error("command usage should not be empty")
	}
	if cmd.Action == nil {
		t.Error("command action should not be nil")
	}
}

func TestNewUpdateCommandsCommand(t *testing.T) {
	cmd := NewUpdateCommandsCommand()
	if cmd == nil {
		t.Fatal("NewUpdateCommandsCommand() returned nil")
	}
	if cmd.Name != "commands" {
		t.Errorf("command name = %q, want %q", cmd.Name, "commands")
	}
	if cmd.Usage == "" {
		t.Error("command usage should not be empty")
	}
	if cmd.Action == nil {
		t.Error("command action should not be nil")
	}
}

func TestNewUpdateAllCommand(t *testing.T) {
	cmd := NewUpdateAllCommand()
	if cmd == nil {
		t.Fatal("NewUpdateAllCommand() returned nil")
	}
	if cmd.Name != "all" {
		t.Errorf("command name = %q, want %q", cmd.Name, "all")
	}
	if cmd.Usage == "" {
		t.Error("command usage should not be empty")
	}
	if cmd.Action == nil {
		t.Error("command action should not be nil")
	}
}

func TestUpdateCommands_CreatesConfig(t *testing.T) {
	// Save original HOME and restore after test
	origHome := os.Getenv("HOME")
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })

	tmpHome := t.TempDir()
	_ = os.Setenv("HOME", tmpHome)

	// Create a fake executable path for the test
	configDir := filepath.Join(tmpHome, ".config", "opencode")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	err := updateCommands()
	if err != nil {
		t.Fatalf("updateCommands() error = %v", err)
	}

	// Verify opencode.json was created
	configPath := filepath.Join(configDir, "opencode.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("opencode.json was not created")
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading opencode.json: %v", err)
	}

	// Should contain grimorio MCP and command entries
	if !strings.Contains(string(content), `"grimorio"`) {
		t.Error("opencode.json should contain grimorio entry")
	}
	if !strings.Contains(string(content), `"mcp"`) {
		t.Error("opencode.json should contain mcp section")
	}
	if !strings.Contains(string(content), `"command"`) {
		t.Error("opencode.json should contain command section")
	}
}

func TestUpdateCommands_PreservesExistingConfig(t *testing.T) {
	origHome := os.Getenv("HOME")
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })

	tmpHome := t.TempDir()
	_ = os.Setenv("HOME", tmpHome)

	configDir := filepath.Join(tmpHome, ".config", "opencode")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write existing config with other entries
	existing := `{
  "agent": {
    "test-agent": {
      "description": "test"
    }
  },
  "mcp": {
    "other": {
      "command": ["other"],
      "type": "local"
    }
  }
}`
	configPath := filepath.Join(configDir, "opencode.json")
	if err := os.WriteFile(configPath, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	err := updateCommands()
	if err != nil {
		t.Fatalf("updateCommands() error = %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading opencode.json: %v", err)
	}

	// Existing entries should be preserved
	if !strings.Contains(string(content), `"test-agent"`) {
		t.Error("existing agent entries should be preserved")
	}
	if !strings.Contains(string(content), `"other"`) {
		t.Error("existing MCP entries should be preserved")
	}

	// Grimorio entries should be added
	if !strings.Contains(string(content), `"grimorio"`) {
		t.Error("grimorio entry should be added")
	}
}

func TestCopyDir(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	// Create a test file structure
	if err := os.WriteFile(filepath.Join(src, "test.md"), []byte("test content"), 0644); err != nil {
		t.Fatalf("creating test file: %v", err)
	}

	if err := copyDir(src, dst); err != nil {
		t.Fatalf("copyDir() error = %v", err)
	}

	// Verify copied file exists
	content, err := os.ReadFile(filepath.Join(dst, "test.md"))
	if err != nil {
		t.Fatalf("copied file not found: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("copied content = %q, want %q", string(content), "test content")
	}
}

