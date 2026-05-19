package campaign

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/pauvalls/grimorio/internal/config"
	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/services"
	"github.com/urfave/cli/v2"
)

// NewListCommand returns the "campaign list" subcommand.
func NewListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all campaigns as a formatted table",
		Action: func(cCtx *cli.Context) error {
			return runList(cCtx)
		},
	}
}

// runList executes the list logic. Exported via closure above; testable separately.
func runList(cCtx *cli.Context) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	configPath := filepath.Join(home, ".config", "grimorio", "config.json")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	campaignRepo := repository.NewFilesystemCampaignRepository(cfg.OutputDir)
	actRepo := repository.NewFilesystemActRepository(cfg.OutputDir)
	charRepo := repository.NewFilesystemCharacterRepository(cfg.OutputDir)
	npcRepo := repository.NewFilesystemNPCRepository(cfg.OutputDir)
	questRepo := repository.NewFilesystemQuestRepository(cfg.OutputDir)

	svc := services.NewCampaignService(
		campaignRepo, actRepo, charRepo, npcRepo, questRepo,
		cfg.OutputDir, cfg.PDFEngine,
	)

	campaigns, err := svc.ListCampaigns()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing campaigns: %v\n", err)
		return nil
	}

	if len(campaigns) == 0 {
		fmt.Fprintln(os.Stderr, "No campaigns found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tTITLE\tSTATUS\tACTS\tNPCS\tCHARS\tUPDATED")
	for _, c := range campaigns {
		updated := c.UpdatedAt.Format(time.RFC3339)
		if c.UpdatedAt.IsZero() {
			updated = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%d\t%s\n",
			c.Name, c.Title, c.Status, c.Acts, c.NPCs, c.Characters, updated)
	}
	w.Flush()
	return nil
}
