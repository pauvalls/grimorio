package campaign

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pauvalls/grimorio/internal/config"
	fsrepo "github.com/pauvalls/grimorio/internal/repository/fs"
	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/services"
	"github.com/urfave/cli/v2"
)

// NewRegisterCommand returns the "campaign register" subcommand.
func NewRegisterCommand() *cli.Command {
	return &cli.Command{
		Name:      "register",
		Usage:     "Register NPCs and bestiary from markdown files into canon",
		ArgsUsage: "<campaign-name>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "npcs",
				Usage: "Register only NPCs",
			},
			&cli.BoolFlag{
				Name:  "bestiary",
				Usage: "Register only bestiary",
			},
		},
		Action: func(cCtx *cli.Context) error {
			return runRegister(cCtx)
		},
	}
}

// runRegister executes the register logic.
func runRegister(cCtx *cli.Context) error {
	if cCtx.NArg() < 1 {
		return fmt.Errorf("campaign name required")
	}

	campaignID := cCtx.Args().Get(0)
	registerNPCs := !cCtx.Bool("bestiary") // default true unless only bestiary
	registerBestiary := !cCtx.Bool("npcs") // default true unless only npcs

	if cCtx.Bool("npcs") && cCtx.Bool("bestiary") {
		return fmt.Errorf("cannot specify both --npcs and --bestiary exclusively")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	configPath := filepath.Join(home, ".config", "grimorio", "config.json")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	baseDir := cfg.OutputDir

	// Setup repositories
	campaignRepo := repository.NewFilesystemCampaignRepository(baseDir)
	actRepo := repository.NewFilesystemActRepository(baseDir)
	charRepo := repository.NewFilesystemCharacterRepository(baseDir)
	npcRepo := repository.NewFilesystemNPCRepository(baseDir)
	questRepo := repository.NewFilesystemQuestRepository(baseDir)
	canonRepo := repository.NewFilesystemCanonRepository(baseDir)
	fsMonsterRepo := fsrepo.NewFilesystemMonsterRepository(baseDir)

	service := services.NewCampaignService(
		campaignRepo, actRepo, charRepo, npcRepo, questRepo,
		canonRepo,
		&monsterRepoWrapper{fs: fsMonsterRepo},
		baseDir, cfg.PDFEngine,
	)

	// Register NPCs
	if registerNPCs {
		npcsPath := filepath.Join(baseDir, campaignID, "npcs", "npcs_and_factions.md")
		if _, err := os.Stat(npcsPath); os.IsNotExist(err) {
			fmt.Printf("⚠️  NPCs file not found: %s\n", npcsPath)
		} else {
			npcsContent, err := os.ReadFile(npcsPath)
			if err != nil {
				return fmt.Errorf("error reading NPCs: %w", err)
			}

			fmt.Printf("📖 Registering NPCs for %s...\n", campaignID)
			if err := service.SaveNPCs(campaignID, string(npcsContent)); err != nil {
				return fmt.Errorf("error saving NPCs: %w", err)
			}
			fmt.Println("✅ NPCs registered")
		}
	}

	// Register Bestiary
	if registerBestiary {
		bestiaryPath := filepath.Join(baseDir, campaignID, "bestiary", "bestiary.md")
		if _, err := os.Stat(bestiaryPath); os.IsNotExist(err) {
			fmt.Printf("⚠️  Bestiary file not found: %s\n", bestiaryPath)
		} else {
			bestiaryContent, err := os.ReadFile(bestiaryPath)
			if err != nil {
				return fmt.Errorf("error reading bestiary: %w", err)
			}

			fmt.Printf("⚔️  Registering bestiary for %s...\n", campaignID)
			if err := service.SaveBestiary(campaignID, string(bestiaryContent)); err != nil {
				return fmt.Errorf("error saving bestiary: %w", err)
			}
			fmt.Println("✅ Bestiary registered")
		}
	}

	fmt.Printf("\n✨ Campaign %s fully registered!\n", campaignID)
	fmt.Println("   You can now use grimorio_dm_session_context without warnings.")

	return nil
}
