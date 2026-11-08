package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

// openCodeDir returns the global OpenCode configuration directory.
func openCodeDir(home string) string {
	return filepath.Join(home, ".config", "opencode")
}

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

	// Find the actual extracted contents.
	// GoReleaser may wrap files in a directory (wrap_in_directory=true)
	// or place them directly at the archive root (wrap_in_directory=false).
	// We search for the asset subdir in extractDir first, then in any
	// top-level subdirectory.
	sourceDir, err := findAssetSourceDir(extractDir, assetType)
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("finding home directory: %w", err)
	}

	targets, err := assetTargets(home, assetType)
	if err != nil {
		return err
	}

	sourceSubdir := filepath.Join(sourceDir, assetType)
	if _, err := os.Stat(sourceSubdir); os.IsNotExist(err) {
		return fmt.Errorf("%s/ not found in release archive — the release may not include %s", assetType, assetType)
	}

	keep := assetFileFilter(assetType)
	for _, target := range targets {
		fmt.Printf("Copying %s to %s...\n", assetType, target)
		if err := copyDirFiltered(sourceSubdir, target, keep); err != nil {
			return fmt.Errorf("copying %s to %s: %w", assetType, target, err)
		}
	}

	fmt.Printf("Successfully updated %s from %s!\n", assetType, release.TagName)
	return nil
}

// assetTargets returns every directory where an asset type must be installed.
//
// OpenCode V2 discovers skills from ~/.config/opencode/skills and agents from
// ~/.config/opencode/agents. The legacy plugin directory is still populated so
// existing V1 installs keep working, but the global V2 directories are the ones
// that make the assets usable on V2.
func assetTargets(home, assetType string) ([]string, error) {
	configDir := openCodeDir(home)
	legacyDir := filepath.Join(configDir, "plugins", "grimorio")

	switch assetType {
	case "skills":
		return []string{
			filepath.Join(configDir, "skills"),
			filepath.Join(legacyDir, "skills"),
		}, nil
	case "agents":
		return []string{
			filepath.Join(configDir, "agents"),
			filepath.Join(legacyDir, "agents"),
		}, nil
	}

	return nil, fmt.Errorf("unknown asset type: %s", assetType)
}

// assetFileFilter returns a predicate that decides which files inside an asset
// directory should be installed. Agents are Markdown-only: the repository keeps
// Go test files next to the agent specs and those must never land in
// ~/.config/opencode/agents.
func assetFileFilter(assetType string) func(rel string) bool {
	if assetType == "agents" {
		return func(rel string) bool {
			return strings.HasSuffix(rel, ".md")
		}
	}
	return nil
}

// copyDir recursively copies a directory tree.
func copyDir(src, dst string) error {
	return copyDirFiltered(src, dst, nil)
}

// copyDirFiltered recursively copies a directory tree, skipping files for which
// keep returns false. A nil filter copies everything.
func copyDirFiltered(src, dst string, keep func(rel string) bool) error {
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

		if keep != nil && !keep(rel) {
			return nil
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

// grimorioCommandTemplate is the Markdown command installed as /grimorio.
// The prompt stays deliberately thin: the grimorio-architect agent owns the
// workflow, so it must not be duplicated here where it silently goes stale.
const grimorioCommandTemplate = `---
description: Generate a complete D&D 5e campaign or one-shot from an idea
agent: grimorio-architect
subagent: false
---

Start the Grimorio campaign generation workflow.

User request: $ARGUMENTS

Follow the mandatory intake (language first), then generate the campaign end-to-end and compile the PDF.
`

// buildMCPEnvironment returns the environment for the MCP server entry.
//
// It never hardcodes user-specific absolute paths: PATH comes from the current
// environment, the Piper voice is discovered under $HOME, and the remaining
// tuning values use generic defaults. Values already present in the user's
// configuration are preserved.
func buildMCPEnvironment(home string, existing map[string]interface{}) map[string]string {
	env := make(map[string]string, len(existing))
	for k, v := range existing {
		if s, ok := v.(string); ok {
			env[k] = s
		}
	}

	if env["PATH"] == "" {
		if path := os.Getenv("PATH"); path != "" {
			env["PATH"] = path
		} else {
			env["PATH"] = "/usr/local/bin:/usr/bin:/bin"
		}
	}

	// Discover the Piper voice only when the user has one installed.
	if env["PIPER_MODEL_PATH"] == "" {
		matches, _ := filepath.Glob(filepath.Join(home, ".local", "share", "piper", "*.onnx"))
		if len(matches) > 0 {
			env["PIPER_MODEL_PATH"] = matches[0]
			if env["PIPER_CONFIG_PATH"] == "" {
				if _, err := os.Stat(matches[0] + ".json"); err == nil {
					env["PIPER_CONFIG_PATH"] = matches[0] + ".json"
				}
			}
		}
	}

	defaults := map[string]string{
		"PIPER_PORT":           "5000",
		"PIPER_HOST":           "127.0.0.1",
		"PIPER_LENGTH_SCALE":   "1.0",
		"PIPER_VOLUME":         "0.8",
		"GRIMORIO_TTS_ENABLED": "true",
		"AUDIO_PLAYER":         "auto",
	}
	for k, v := range defaults {
		if env[k] == "" {
			env[k] = v
		}
	}

	return env
}

// updateCommands writes the OpenCode V2 integration: the MCP server entry under
// mcp.servers and the /grimorio command as a Markdown command file.
//
// It migrates away from the V1 shapes written by older installers
// (mcp.grimorio and command.grimorio) so a re-run leaves no duplicate or
// half-configured state, and it preserves every unrelated setting.
func updateCommands() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding executable: %w", err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("resolving executable path: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("finding home directory: %w", err)
	}

	configPath := filepath.Join(openCodeDir(home), "opencode.json")
	fmt.Printf("Updating OpenCode integration in %s...\n", configPath)

	data := map[string]interface{}{}
	if content, err := os.ReadFile(configPath); err == nil {
		if len(content) > 0 {
			if err := json.Unmarshal(content, &data); err != nil {
				return fmt.Errorf("parsing opencode.json: %w", err)
			}
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("reading opencode.json: %w", err)
	}

	if _, err := os.Stat(configPath); err == nil {
		backupPath := configPath + ".backup." + time.Now().Format("20060102150405")
		if err := copyFile(configPath, backupPath); err != nil {
			return fmt.Errorf("backing up opencode.json: %w", err)
		}
		fmt.Printf("Backup created: %s\n", backupPath)
	}

	mcpSection := ensureMap(data, "mcp")
	servers := ensureMap(mcpSection, "servers")

	existingEnv, _ := servers["grimorio"].(map[string]interface{})
	var previousEnv map[string]interface{}
	if existingEnv != nil {
		previousEnv, _ = existingEnv["environment"].(map[string]interface{})
	}
	if previousEnv == nil && existingEnv != nil {
		previousEnv, _ = existingEnv["env"].(map[string]interface{})
	}

	servers["grimorio"] = map[string]interface{}{
		"type":        "local",
		"command":     []string{exePath},
		"disabled":    false,
		"environment": buildMCPEnvironment(home, previousEnv),
	}

	// Migrate the V1 shape: a top-level mcp.grimorio entry would otherwise be
	// normalized next to the V2 entry and could shadow it.
	delete(mcpSection, "grimorio")

	// Migrate the V1 shape: command.grimorio is replaced by the Markdown
	// command file written below.
	if legacyCommands, ok := data["command"].(map[string]interface{}); ok {
		delete(legacyCommands, "grimorio")
		if len(legacyCommands) == 0 {
			delete(data, "command")
		}
	}

	if err := writeJSONIfChanged(configPath, data); err != nil {
		return err
	}

	commandPath := filepath.Join(openCodeDir(home), "commands", "grimorio.md")
	if err := writeFileIfChanged(commandPath, []byte(grimorioCommandTemplate), 0644); err != nil {
		return fmt.Errorf("writing command file: %w", err)
	}

	fmt.Printf("MCP server: mcp.servers.grimorio -> %s\n", exePath)
	fmt.Printf("Command:   %s\n", commandPath)
	fmt.Println("Successfully updated OpenCode integration")
	return nil
}

// ensureMap returns data[key] as a map, creating it when missing or malformed.
func ensureMap(data map[string]interface{}, key string) map[string]interface{} {
	if existing, ok := data[key].(map[string]interface{}); ok {
		return existing
	}
	created := make(map[string]interface{})
	data[key] = created
	return created
}

// writeJSONIfChanged marshals data and writes it only when the bytes differ.
func writeJSONIfChanged(path string, data map[string]interface{}) error {
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	out = append(out, '\n')
	if err := writeFileIfChanged(path, out, 0644); err != nil {
		return fmt.Errorf("writing opencode.json: %w", err)
	}
	return nil
}

// writeFileIfChanged writes content only when it differs from what is on disk.
// This keeps repeated updates from churning mtimes or triggering config reloads.
func writeFileIfChanged(path string, content []byte, perm os.FileMode) error {
	if current, err := os.ReadFile(path); err == nil && string(current) == string(content) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, content, perm)
}

// updateAll runs every update step. A failing step does not stop the others:
// a failed download must not leave the local OpenCode integration unfixed.
func updateAll(currentVersion string) error {
	fmt.Println("=== Grimorio Full Update ===")

	steps := []struct {
		name string
		run  func() error
	}{
		{"commands", updateCommands},
		{"skills", func() error { return updateAssets("skills") }},
		{"agents", func() error { return updateAssets("agents") }},
		{"binary", func() error { return runBinaryUpdate(currentVersion, false) }},
	}

	var failures []string
	for _, step := range steps {
		fmt.Printf("\n--- Updating %s ---\n", step.name)
		if err := step.run(); err != nil {
			fmt.Printf("ERROR updating %s: %v\n", step.name, err)
			failures = append(failures, fmt.Sprintf("%s: %v", step.name, err))
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("update completed with %d failure(s): %s", len(failures), strings.Join(failures, "; "))
	}

	fmt.Println("\n=== All updates completed successfully! ===")
	return nil
}

// runBinaryUpdate runs the binary self-update. Extracted so updateAll can reuse
// the same logic that powers `grimorio update`.
func runBinaryUpdate(currentVersion string, dryRun bool) error {
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

	// Detect ALL install dirs (CLI + MCP). The MCP binary lives at
	// $HOME/.grimorio/grimorio and is the one the OpenCode MCP server
	// runs (per mcp.servers.grimorio in opencode.json). Without this, the
	// MCP server stays on the old version after `grimorio update`.
	installDirs, err := discoverInstallDirs(exePath, home)
	if err != nil {
		return fmt.Errorf("discovering install dirs: %w", err)
	}

	u := &updater{
		repoOwner:      "pauvalls",
		repoName:       "grimorio",
		installDirs:    installDirs,
		backupDir:      filepath.Join(home, ".grimorio"),
		currentVersion: currentVersion,
		httpClient:     nil,
	}
	return u.runUpdate(dryRun)
}

// NewUpdateCommandsCommand returns the "update commands" CLI command.
func NewUpdateCommandsCommand() *cli.Command {
	return &cli.Command{
		Name:  "commands",
		Usage: "Write the Grimorio MCP server entry and /grimorio command into OpenCode",
		Action: func(cCtx *cli.Context) error {
			return updateCommands()
		},
	}
}

// findAssetSourceDir searches for the directory containing the requested
// asset subdirectory (e.g. "skills" or "agents"). It first checks the
// extract root, then any top-level subdirectory, to handle both
// wrap_in_directory=true and wrap_in_directory=false archive layouts.
func findAssetSourceDir(extractDir, assetType string) (string, error) {
	// Check root first
	rootPath := filepath.Join(extractDir, assetType)
	if _, err := os.Stat(rootPath); err == nil {
		return extractDir, nil
	}

	// Check top-level subdirectories
	entries, err := os.ReadDir(extractDir)
	if err != nil {
		return "", fmt.Errorf("reading extract dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subPath := filepath.Join(extractDir, entry.Name(), assetType)
			if _, err := os.Stat(subPath); err == nil {
				return filepath.Join(extractDir, entry.Name()), nil
			}
		}
	}

	return "", fmt.Errorf("%s/ not found in release archive — the release may not include %s", assetType, assetType)
}

// NewUpdateAllCommand returns the "update all" CLI command.
func NewUpdateAllCommand(currentVersion string) *cli.Command {
	return &cli.Command{
		Name:  "all",
		Usage: "Update all Grimorio assets: OpenCode integration, skills, agents, and binary",
		Action: func(cCtx *cli.Context) error {
			return updateAll(currentVersion)
		},
	}
}
