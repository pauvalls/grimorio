package update

import (
	"encoding/json"
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

	// Find the actual extracted contents.
	// GoReleaser may wrap files in a directory (wrap_in_directory=true)
	// or place them directly at the archive root (wrap_in_directory=false).
	// We search for the asset subdir in extractDir first, then in any
	// top-level subdirectory.
	sourceDir, err := findAssetSourceDir(extractDir, assetType)
	if err != nil {
		return err
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

// updateCommands updates the opencode.json configuration with grimorio MCP and command entries.
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

	configPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	fmt.Printf("Updating commands in %s...\n", configPath)

	// Read existing config
	data := map[string]interface{}{}
	if _, err := os.Stat(configPath); err == nil {
		content, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("reading opencode.json: %w", err)
		}
		if len(content) > 0 {
			if err := json.Unmarshal(content, &data); err != nil {
				return fmt.Errorf("parsing opencode.json: %w", err)
			}
		}
	}

	// Backup existing config
	backupPath := configPath + ".backup." + time.Now().Format("20060102150405")
	if _, err := os.Stat(configPath); err == nil {
		if err := copyFile(configPath, backupPath); err != nil {
			return fmt.Errorf("backing up opencode.json: %w", err)
		}
		fmt.Printf("Backup created: %s\n", backupPath)
	}

	// Build grimorio MCP entry with TTS environment variables
	grimorioMCP := map[string]interface{}{
		"command": []string{exePath},
		"type":    "local",
		"enabled": true,
		"env": map[string]string{
			"PATH":                       "/home/pau/.local/bin:/home/pau/.local/bin/piper:/usr/local/bin:/usr/bin:/bin",
			"PIPER_MODEL_PATH":            "/home/pau/.local/share/piper/es_ES-davefx-medium.onnx",
			"PIPER_CONFIG_PATH":           "/home/pau/.local/share/piper/es_ES-davefx-medium.onnx.json",
			"PIPER_PORT":                  "5000",
			"PIPER_HOST":                  "127.0.0.1",
			"PIPER_LENGTH_SCALE":          "1.0",
			"PIPER_VOLUME":                "0.8",
			"GRIMORIO_TTS_ENABLED":        "true",
			"AUDIO_PLAYER":                "auto",
		},
	}

	// Build grimorio command entry
	grimorioCmd := map[string]interface{}{
		"description": "Generate a complete D&D 5e campaign or one-shot from an idea (executes in main thread)",
		"subtask":     false,
		"template": `Generate a D&D 5e campaign or one-shot from the user's idea.

**Version:** 5.0.0 — Chapter-Sequential Pipeline

## IMPORTANT: Use the grimorio-architect agent. It handles everything end-to-end.

## 0. Language Intake (Mandatory, First Question)

Before any other question, ask the user:

> **¿En qué idioma prefieres jugar? / What language do you prefer to play in? [es/en]**

Default to "en" if the user skips. The architect stores this and propagates to all sub-agents via LANG: preamble on every delegate call.

## 1. Workflow (5 Macro-Phases, chapter-sequential)

The architect follows this sequence with BLOCKING gates at each macro-phase.

### Macro-Phase 1: Foundation
- create_campaign(name, setting, title)
- generate_adventure_bible(...) → canon.json
- generate_names (all 7 categories: npc, monster, character, city, faction, tavern, item)
- save_introduction
- save_setting_guide
- save_lore
- GATE: validate_canon → approved (BLOCKING)

### Macro-Phase 2: Chapters (sequential, 1 at a time)
For each chapter (typically 3):
- save_chapter(chapter_number, title, content) with WotC standards
- generate_map + generate_divider for this chapter
- GATE: narrative-custodian (BLOCKING)
- GATE: WotC format validation (BLOCKING)

**Why chapters first?** Chapters are the spatial and narrative skeleton. NPCs, bestiary, encounters, quests, and treasure all anchor to specific chapters and areas.

### Macro-Phase 3: Bestiary & Characters (parallel, anchored to chapters)
- save_npcs (anchored to chapter/area)
- save_bestiary (creatures tied to chapter habitats)
- save_encounters (per-chapter, with generate_treasure for hoards)
- save_quests (main + side + personal)
- save_characters (pre-gens + generate_character_hooks)
- save_appendices (consolidated reference)
- GATE: cross-reference validation (BLOCKING)

### Macro-Phase 4: Art & Living World (parallel)
- grimorio-artist → batch spec (cover + NPCs + scenes + monsters)
- generate_image (sequential, 3s delay, force_regenerate=false to use cache)
- Update markdown references with images
- generate_random_tables (encounters, rumors, weather, treasure)
- generate_handouts (summary, quest, lore, faction)
- generate_treasure (per hoard, SRD-compliant)
- update_faction_reputation (initial setup)
- process_consistency_gate (living world batch)
- GATE: campaign_health_dashboard (score ≥ 70 recommended)

### Macro-Phase 5: Export & Deliver
- grimorio validate {name} --scope=all (BLOCKING CLI gate)
- export_campaign --format=pdf (default) | --format=markdown | --format=epub
- generate_session_prep + generate_flowchart (optional)
- Final report

The architect reports progress to the user after each macro-phase.

## 2. Available MCP Tools (v5.0)

- Creation: create_campaign, generate_adventure_bible, generate_names
- Save: save_introduction, save_setting_guide, save_lore, save_chapter, save_npcs, save_bestiary, save_encounters, save_maps, save_quests, save_characters, save_appendices
- Assets: generate_image, generate_map, generate_divider, generate_flowchart, generate_random_tables, generate_handouts, generate_treasure, generate_session_prep
- Validation: validate_canon, check_consistency, process_consistency_gate, evaluate_consequences
- State: update_narrative_state, update_faction_reputation, update_quest_status
- Quality (v5.0): campaign_health_dashboard, export_campaign
- Compilation: compile_pdf

## Final: Report
After completion, report to the user:
- Where the PDF was saved
- What content was generated
- Campaign health score
- Any issues encountered

**DO NOT launch subagents from the command thread — the architect manages all delegation internally.**`,
	}

	// Ensure mcp section exists
	mcpSection, ok := data["mcp"].(map[string]interface{})
	if !ok {
		mcpSection = make(map[string]interface{})
		data["mcp"] = mcpSection
	}
	mcpSection["grimorio"] = grimorioMCP

	// Ensure command section exists
	cmdSection, ok := data["command"].(map[string]interface{})
	if !ok {
		cmdSection = make(map[string]interface{})
		data["command"] = cmdSection
	}
	cmdSection["grimorio"] = grimorioCmd

	// Write back
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(configPath, out, 0644); err != nil {
		return fmt.Errorf("writing opencode.json: %w", err)
	}

	fmt.Printf("Successfully updated commands in opencode.json\n")
	return nil
}

// updateAll runs all three asset updates sequentially: skills, agents, commands.
func updateAll() error {
	fmt.Println("=== Grimorio Full Update ===")

	fmt.Println("\n--- Updating skills ---")
	if err := updateAssets("skills"); err != nil {
		return fmt.Errorf("skills update failed: %w", err)
	}

	fmt.Println("\n--- Updating agents ---")
	if err := updateAssets("agents"); err != nil {
		return fmt.Errorf("agents update failed: %w", err)
	}

	fmt.Println("\n--- Updating commands ---")
	if err := updateCommands(); err != nil {
		return fmt.Errorf("commands update failed: %w", err)
	}

	fmt.Println("\n=== All updates completed successfully! ===")
	return nil
}

// NewUpdateCommandsCommand returns the "update commands" CLI command.
func NewUpdateCommandsCommand() *cli.Command {
	return &cli.Command{
		Name:  "commands",
		Usage: "Update Grimorio MCP and command entries in opencode.json",
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
func NewUpdateAllCommand() *cli.Command {
	return &cli.Command{
		Name:  "all",
		Usage: "Update all Grimorio assets: skills, agents, and commands",
		Action: func(cCtx *cli.Context) error {
			return updateAll()
		},
	}
}
