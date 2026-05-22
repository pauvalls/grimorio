package update

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v2"
)

// updateAssets downloads and updates skills or agents from the latest release.
func updateAssets(assetType string) error {
	goos, goarch, err := detectPlatform()
	if err != nil {
		return fmt.Errorf("detecting platform: %w", err)
	}

	fmt.Printf("Updating %s (platform: %s/%s)...\n", assetType, goos, goarch)

	// Fetch latest release
	release, err := fetchLatestRelease("pauvalls", "grimorio", nil, "")
	if err != nil {
		return fmt.Errorf("fetching latest release: %w", err)
	}

	// Find the correct asset
	downloadURL, _, err := findAsset(release, goos, goarch)
	if err != nil {
		return fmt.Errorf("finding asset: %w", err)
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("grimorio-update-%s-*", assetType))
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer cleanupOnError([]string{tmpDir})

	archivePath := filepath.Join(tmpDir, archiveName(goos, goarch))

	// Download archive
	fmt.Printf("Downloading %s...\n", downloadURL)
	client := &http.Client{Timeout: 5 * time.Minute}
	if err := downloadFile(downloadURL, archivePath, client); err != nil {
		return fmt.Errorf("downloading archive: %w", err)
	}

	// Extract
	extractDir := filepath.Join(tmpDir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return fmt.Errorf("creating extract dir: %w", err)
	}

	fmt.Printf("Extracting archive...\n")
	if err := extractArchive(archivePath, extractDir); err != nil {
		return fmt.Errorf("extracting archive: %w", err)
	}

	// Find the actual extracted contents (GoReleaser wraps in a directory)
	entries, err := os.ReadDir(extractDir)
	if err != nil {
		return fmt.Errorf("reading extract dir: %w", err)
	}

	var sourceDir string
	for _, entry := range entries {
		if entry.IsDir() {
			sourceDir = filepath.Join(extractDir, entry.Name())
			break
		}
	}
	if sourceDir == "" {
		sourceDir = extractDir
	}

	// Determine target directories
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("finding home directory: %w", err)
	}

	var sourceSubdir, pluginTarget, globalTarget string
	switch assetType {
	case "skills":
		sourceSubdir = filepath.Join(sourceDir, "skills")
		pluginTarget = filepath.Join(home, ".config", "opencode", "plugins", "grimorio", "skills")
		globalTarget = "" // Skills only go to plugin dir
	case "agents":
		sourceSubdir = filepath.Join(sourceDir, "agents")
		pluginTarget = filepath.Join(home, ".config", "opencode", "plugins", "grimorio", "agents")
		globalTarget = filepath.Join(home, ".config", "opencode", "agents")
	default:
		return fmt.Errorf("unknown asset type: %s", assetType)
	}

	// Verify source exists
	if _, err := os.Stat(sourceSubdir); os.IsNotExist(err) {
		return fmt.Errorf("%s/ not found in release archive — the release may not include %s", assetType, assetType)
	}

	// Copy to plugin directory
	fmt.Printf("Copying %s to plugin directory...\n", assetType)
	if err := copyDir(sourceSubdir, pluginTarget); err != nil {
		return fmt.Errorf("copying %s to plugin dir: %w", assetType, err)
	}
	fmt.Printf("%s updated in plugin directory: %s\n", assetType, pluginTarget)

	// For agents, also copy to global agents directory
	if assetType == "agents" && globalTarget != "" {
		fmt.Printf("Copying agents to global agents directory...\n")
		if err := copyDir(sourceSubdir, globalTarget); err != nil {
			return fmt.Errorf("copying agents to global dir: %w", err)
		}
		fmt.Printf("Agents updated in global directory: %s\n", globalTarget)
	}

	fmt.Printf("Successfully updated %s from %s!\n", assetType, release.TagName)
	return nil
}

// copyDir recursively copies a directory tree.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Compute relative path
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		// Copy file
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		return copyFile(path, target)
	})
}

// NewUpdateSkillsCommand returns the "update skills" CLI command.
func NewUpdateSkillsCommand() *cli.Command {
	return &cli.Command{
		Name:  "skills",
		Usage: "Update Grimorio skills to the latest version",
		Action: func(cCtx *cli.Context) error {
			return updateAssets("skills")
		},
	}
}

// NewUpdateAgentsCommand returns the "update agents" CLI command.
func NewUpdateAgentsCommand() *cli.Command {
	return &cli.Command{
		Name:  "agents",
		Usage: "Update Grimorio agents to the latest version",
		Action: func(cCtx *cli.Context) error {
			return updateAssets("agents")
		},
	}
}
