package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// detectPlatform returns the current OS and architecture using runtime values.
func detectPlatform() (string, string, error) {
	return runtime.GOOS, runtime.GOARCH, nil
}

// mapGoArchToGoreleaser maps Go runtime OS/arch values to GoReleaser naming conventions.
// GoReleaser uses title-cased OS names and maps amd64 → x86_64.
func mapGoArchToGoreleaser(goos, goarch string) (string, string) {
	osName := goos
	if len(goos) > 0 {
		osName = strings.ToUpper(goos[:1]) + goos[1:]
	}
	archName := goarch
	if goarch == "amd64" {
		archName = "x86_64"
	}
	return osName, archName
}

// archiveName returns the GoReleaser archive filename for the given platform.
func archiveName(goos, goarch string) string {
	gos, arch := mapGoArchToGoreleaser(goos, goarch)
	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}
	return fmt.Sprintf("grimorio_%s_%s%s", gos, arch, ext)
}

// --- GitHub API ---

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

// fetchLatestRelease queries the GitHub API for the latest release.
// apiBaseURL is optional (for testing); defaults to https://api.github.com.
func fetchLatestRelease(owner, repo string, client *http.Client, apiBaseURL string) (*githubRelease, error) {
	if client == nil {
		client = http.DefaultClient
		client.Timeout = 30 * time.Second
	}
	if apiBaseURL == "" {
		apiBaseURL = "https://api.github.com"
	}

	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", apiBaseURL, owner, repo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decoding release JSON: %w", err)
	}

	return &release, nil
}

// findAsset locates the download URL for the platform archive and the checksums.txt URL.
func findAsset(release *githubRelease, goos, goarch string) (downloadURL, checksumURL string, err error) {
	expectedName := archiveName(goos, goarch)

	for _, asset := range release.Assets {
		if asset.Name == expectedName {
			downloadURL = asset.BrowserDownloadURL
		}
		if asset.Name == "checksums.txt" {
			checksumURL = asset.BrowserDownloadURL
		}
	}

	if downloadURL == "" {
		return "", "", fmt.Errorf("archive %s not found in release %s", expectedName, release.TagName)
	}

	return downloadURL, checksumURL, nil
}

// buildFallbackURL constructs a direct download URL using the static pattern.
func buildFallbackURL(owner, repo, tag, goos, goarch string) string {
	name := archiveName(goos, goarch)
	return fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", owner, repo, tag, name)
}
