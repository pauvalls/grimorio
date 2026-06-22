package update

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
)

// --- T005: Backup, Replace, Rollback ---

// backupCurrentBinary copies the current binary to a backup location.
func backupCurrentBinary(binaryPath, backupDir string) (string, error) {
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("creating backup dir: %w", err)
	}

	backupPath := filepath.Join(backupDir, "grimorio.backup")

	src, err := os.Open(binaryPath)
	if err != nil {
		return "", fmt.Errorf("opening current binary: %w", err)
	}
	defer func() { _ = src.Close() }()

	dst, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("creating backup file: %w", err)
	}
	defer func() { _ = dst.Close() }()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("copying binary to backup: %w", err)
	}

	// Preserve executable permission
	if info, err := src.Stat(); err == nil {
		_ = os.Chmod(backupPath, info.Mode()|0111)
	}

	return backupPath, nil
}

// replaceBinary atomically replaces the current binary with the new one.
// It uses a temp file + rename to avoid "text file busy" on Unix.
//
// Linux semantics: `os.Rename` succeeds even when the destination has an
// open file descriptor (the running process keeps the old inode via its fd,
// the new file at the same path is a fresh inode). This is the rolling
// update pattern that lets the MCP server keep running while the binary
// is replaced. The only failure mode is when something has the destination
// open for writing (mtime update on running), which is rare in practice.
func replaceBinary(binaryPath, newPath string) error {
	// Copy new binary to a unique temp location in the same directory
	tempFile, err := os.CreateTemp(filepath.Dir(binaryPath), "grimorio-update-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tempPath := tempFile.Name()
	_ = tempFile.Close()

	if err := copyFile(newPath, tempPath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("copying new binary to temp: %w", err)
	}

	// Make it executable
	if err := os.Chmod(tempPath, 0755); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("making new binary executable: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, binaryPath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("replacing binary: %w", err)
	}

	return nil
}

// rollback restores the backup binary.
func rollback(binaryPath, backupPath string) error {
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup not found at %s", backupPath)
	}

	if err := copyFile(backupPath, binaryPath); err != nil {
		return fmt.Errorf("restoring backup: %w", err)
	}

	if err := os.Chmod(binaryPath, 0755); err != nil {
		return fmt.Errorf("setting executable permission on restored binary: %w", err)
	}

	return nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source: %w", err)
	}
	defer func() { _ = source.Close() }()

	destination, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}
	defer func() { _ = destination.Close() }()

	if _, err := io.Copy(destination, source); err != nil {
		return fmt.Errorf("copying data: %w", err)
	}

	return nil
}

// --- T006: Version Comparison ---

// isNewer compares two semantic version strings.
// Returns true if latest is newer than current.
func isNewer(current, latest string) (bool, error) {
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	// "dev" or empty current version always needs update (if latest is valid)
	if current == "dev" || current == "" {
		return latest != "" && latest != "dev", nil
	}
	// "dev" as latest means no update available (development build)
	if latest == "dev" {
		return false, nil
	}
	// empty latest is invalid
	if latest == "" {
		return false, fmt.Errorf("invalid latest version: %q", latest)
	}

	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	// Compare major, minor, patch
	for i := 0; i < 3 && i < len(currentParts) && i < len(latestParts); i++ {
		cv, err := strconv.Atoi(strings.Split(currentParts[i], "-")[0])
		if err != nil {
			return false, fmt.Errorf("parsing current version %q: %w", current, err)
		}
		lv, err := strconv.Atoi(strings.Split(latestParts[i], "-")[0])
		if err != nil {
			return false, fmt.Errorf("parsing latest version %q: %w", latest, err)
		}

		if lv > cv {
			return true, nil
		}
		if lv < cv {
			return false, nil
		}
	}

	// If numeric parts are equal, check for prerelease
	// v1.3.0 is NOT newer than v1.3.0, but v1.3.0-beta is older than v1.3.0
	if len(latestParts) >= 3 && strings.Contains(latestParts[2], "-") {
		// latest is a prerelease (e.g., 1.3.0-beta), current is stable at same version
		// current = 1.3.0, latest = 1.3.0-beta → latest is NOT newer
		if !strings.Contains(currentParts[2], "-") {
			return false, nil
		}
	}

	return false, nil
}

// --- Update Orchestration ---

type updater struct {
	repoOwner      string
	repoName       string
	installDirs    []string // every binary that must be replaced (CLI + MCP)
	backupDir      string
	apiBaseURL     string
	httpClient     *http.Client
	currentVersion string
}

// discoverInstallDirs returns the list of binary paths that the updater
// should replace. It always includes the CLI binary (where the user invoked
// `grimorio update` from). It also includes the MCP server binary if it
// exists — the opencode MCP server runs grimorio from $HOME/.grimorio/grimorio
// (per `~/.config/opencode/plugins/grimorio/.mcp.json`), which is a separate
// path from the CLI binary in $HOME/.local/bin/grimorio.
//
// Returning the MCP path is non-fatal when the file does not exist: in that
// case, the user is probably running grimorio as a CLI only (no MCP server
// configured), and the update should still succeed for the CLI binary.
//
// The homeDir parameter is injected for testability (tests use t.TempDir()
// instead of touching the real $HOME).
func discoverInstallDirs(cliBinary, homeDir string) ([]string, error) {
	dirs := []string{cliBinary}

	mcpBinary := filepath.Join(homeDir, ".grimorio", "grimorio")
	if goos := runtime.GOOS; goos == "windows" {
		mcpBinary += ".exe"
	}

	if _, err := os.Stat(mcpBinary); err == nil {
		dirs = append(dirs, mcpBinary)
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("checking MCP binary at %s: %w", mcpBinary, err)
	}

	return dirs, nil
}

// runUpdate orchestrates the full update flow.
//
// Downloads the latest release once, then replaces the binary in every
// `installDirs` path. If any single replacement fails, the loop continues
// (logs the error) and the function returns an aggregated error at the end.
// This is intentional: a partial update (CLI updated, MCP still old) is
// better than no update at all, and the user can re-run `grimorio update`.
func (u *updater) runUpdate(dryRun bool) error {
	goos, goarch, err := detectPlatform()
	if err != nil {
		return fmt.Errorf("detecting platform: %w", err)
	}

	fmt.Printf("Checking for updates (current: %s, platform: %s/%s)...\n", u.currentVersion, goos, goarch)

	// Fetch latest release
	release, err := fetchLatestRelease(u.repoOwner, u.repoName, u.httpClient, u.apiBaseURL)
	if err != nil {
		return fmt.Errorf("fetching latest release: %w", err)
	}

	latestTag := release.TagName

	// Compare versions
	needsUpdate, err := isNewer(u.currentVersion, latestTag)
	if err != nil {
		return fmt.Errorf("comparing versions: %w", err)
	}

	if !needsUpdate {
		fmt.Println("Already up to date")
		return nil
	}

	fmt.Printf("Update available: %s → %s\n", u.currentVersion, latestTag)

	if dryRun {
		fmt.Println("Dry run: would download and install", latestTag)
		return nil
	}

	// Find the correct asset
	downloadURL, checksumURL, err := findAsset(release, goos, goarch)
	if err != nil {
		return fmt.Errorf("finding asset: %w", err)
	}

	// Create temp directory for download
	tmpDir, err := os.MkdirTemp("", "grimorio-update-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer cleanupOnError([]string{tmpDir})

	archivePath := filepath.Join(tmpDir, archiveName(goos, goarch))

	// Download archive
	fmt.Printf("Downloading %s...\n", downloadURL)
	if err := downloadFile(downloadURL, archivePath, u.httpClient); err != nil {
		return fmt.Errorf("downloading archive: %w", err)
	}

	// Download and verify checksum if available
	if checksumURL != "" {
		checksumsPath := filepath.Join(tmpDir, "checksums.txt")
		fmt.Printf("Downloading checksums...\n")
		if err := downloadFile(checksumURL, checksumsPath, u.httpClient); err != nil {
			fmt.Printf("Warning: could not download checksums: %v\n", err)
		} else {
			checksumsData, err := os.ReadFile(checksumsPath)
			if err != nil {
				return fmt.Errorf("reading checksums: %w", err)
			}
			expectedHash, err := parseChecksums(string(checksumsData), archiveName(goos, goarch))
			if err != nil {
				fmt.Printf("Warning: could not parse checksums: %v\n", err)
			} else {
				fmt.Printf("Verifying checksum...\n")
				if err := verifyChecksum(archivePath, expectedHash); err != nil {
					return fmt.Errorf("checksum verification failed: %w", err)
				}
				fmt.Println("Checksum verified ✓")
			}
		}
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

	// Validate
	if err := validateExtractedContents(extractDir); err != nil {
		return fmt.Errorf("validating extracted contents: %w", err)
	}

	// Find the actual binary path (GoReleaser wraps in a subdirectory)
	newBinaryPath, err := findBinaryInExtractedDir(extractDir)
	if err != nil {
		return fmt.Errorf("finding binary in extracted archive: %w", err)
	}

	// Replace the binary in EVERY install dir (CLI + MCP).
	// Failures are collected and reported at the end — a partial update is
	// better than no update at all.
	var errors []string
	for _, binaryPath := range u.installDirs {
		fmt.Printf("Installing new binary at %s...\n", binaryPath)

		// Backup the existing binary (skip if it doesn't exist — first install)
		if _, err := os.Stat(binaryPath); err == nil {
			if _, berr := backupCurrentBinary(binaryPath, u.backupDir); berr != nil {
				errors = append(errors, fmt.Sprintf("backing up %s: %v", binaryPath, berr))
				continue
			}
		}

		if rerr := replaceBinary(binaryPath, newBinaryPath); rerr != nil {
			errors = append(errors, fmt.Sprintf("replacing %s: %v", binaryPath, rerr))
			// Try to roll back from the backup we just made
			backupPath := filepath.Join(u.backupDir, "grimorio.backup")
			if _, serr := os.Stat(backupPath); serr == nil {
				if rberr := rollback(binaryPath, backupPath); rberr != nil {
					errors = append(errors, fmt.Sprintf("rollback for %s: %v", binaryPath, rberr))
				}
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("update completed with errors: %s", strings.Join(errors, "; "))
	}

	fmt.Printf("Successfully updated to %s!\n", latestTag)
	return nil
}

// --- T007: CLI Registration ---

// NewUpdateCommand returns the "update" CLI command.
func NewUpdateCommand(currentVersion string) *cli.Command {
	return &cli.Command{
		Name:  "update",
		Usage: "Update grimorio to the latest version",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Show what would be updated without making changes",
			},
		},
		Action: func(cCtx *cli.Context) error {
			// Get current binary path
			exePath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("finding current executable: %w", err)
			}
			exePath, err = filepath.EvalSymlinks(exePath)
			if err != nil {
				return fmt.Errorf("resolving executable path: %w", err)
			}

			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("finding home directory: %w", err)
			}
			backupDir := filepath.Join(home, ".grimorio")

			// Detect ALL install dirs (CLI + MCP). The MCP binary lives at
			// $HOME/.grimorio/grimorio and is the one the opencode MCP server
			// runs (per `~/.config/opencode/plugins/grimorio/.mcp.json`).
			installDirs, err := discoverInstallDirs(exePath, home)
			if err != nil {
				return fmt.Errorf("discovering install dirs: %w", err)
			}

			u := &updater{
				repoOwner:      "pauvalls",
				repoName:       "grimorio",
				installDirs:    installDirs,
				backupDir:      backupDir,
				currentVersion: currentVersion,
				httpClient:     nil,
			}

			return u.runUpdate(cCtx.Bool("dry-run"))
		},
	}
}
