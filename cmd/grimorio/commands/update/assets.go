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

	// Build grimorio MCP entry
	grimorioMCP := map[string]interface{}{
		"command": []string{exePath},
		"type":    "local",
		"enabled": true,
	}

	// Build grimorio command entry
	grimorioCmd := map[string]interface{}{
		"description": "Generate a complete D&D 5e campaign or one-shot from an idea (executes in main thread)",
		"subtask":     false,
		"template": `Generate a D&D 5e campaign or one-shot from the user's idea.

## IMPORTANT: Use the grimorio-architect agent. It handles everything end-to-end.

## Workflow (followed by grimorio-architect)

### Phase 1: Gather Requirements
Ask the user these questions (one at a time, interactively):
1. What's the campaign name? (kebab-case, e.g. "sunken-city")
2. One-shot or full campaign?
3. Campaign idea / brief description? (What story do you want to tell? 2-3 sentences)
4. Player level? (1-3, 4-6, 7-10, 11-15, 16-20)
5. Desired tone? (heroic, dark, humorous, political intrigue)
6. Duration? (one-shot, 3-5 sessions, long campaign)

### Phase 2: Create Campaign Structure
Use the grimorio MCP tool create_campaign to create the structure.

### Phase 3-13: End-to-End Orchestration (sequential batches)
The architect follows strict batch ordering — each batch waits for the previous:

- **Batch 1** (parallel): lore, NPCs, bestiary, maps, setting guide, introduction
  → Narrative validation (narrative-custodian)
  → WotC validation
- **Batch 2** (parallel): quests, encounters, characters, appendices
  → Narrative validation (narrative-custodian)
  → WotC validation
  → Update Narrative State
- **Batch 3** (sequential, 1 area at a time to avoid timeout):
  - Chapter 1 areas → Narrative validation → WotC validation
  - Chapter 2 areas → Narrative validation → WotC validation
  - Chapter 3 areas → Narrative validation → WotC validation
- **Phase 6**: Art generation — cover, NPC portraits, monster illustrations, scene artwork (batch-spec then generate)
- **Phase 7**: Update ALL markdown references with generated images
- **Phase 8**: Living World tools (factions, random tables, handouts, consequences) → Consistency Gate
- **Phase 9**: Final validations — integration check, narrative coherence, macguffin consistency, random tables, character sheets
- **Phase 10**: DM Experience tools (session prep, flowchart)
- **Phase 11**: Compile PDF (embeds all images + flowchart + CSS styling)
- **Phase 12**: Final report

The architect reports progress to the user after each phase.

### Final: Report
After completion, report to the user:
- Where the PDF was saved
- What content was generated
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
