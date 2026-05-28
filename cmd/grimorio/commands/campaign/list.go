package campaign

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/pauvalls/grimorio/internal/config"
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
	fsrepo "github.com/pauvalls/grimorio/internal/repository/fs"
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
	canonRepo := repository.NewFilesystemCanonRepository(cfg.OutputDir)
	fsMonsterRepo := fsrepo.NewFilesystemMonsterRepository(cfg.OutputDir)

	svc := services.NewCampaignService(
		campaignRepo, actRepo, charRepo, npcRepo, questRepo,
		canonRepo,
		&monsterRepoWrapper{fs: fsMonsterRepo},
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
_, _ = fmt.Fprintln(w, "NAME\tTITLE\tSTATUS\tACTS\tNPCS\tCHARS\tUPDATED")
	for _, c := range campaigns {
		updated := c.UpdatedAt.Format(time.RFC3339)
		if c.UpdatedAt.IsZero() {
			updated = "-"
		}
_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%d\t%s\n",
			c.Name, c.Title, c.Status, c.Acts, c.NPCs, c.Characters, updated)
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush output: %w", err)
	}
	return nil
}

// monsterRepoWrapper adapts fs.FilesystemMonsterRepository to repository.MonsterRepository
type monsterRepoWrapper struct {
	fs *fsrepo.FilesystemMonsterRepository
}

func (w *monsterRepoWrapper) Save(monster *domain.Monster) error {
	return w.fs.Save(context.Background(), monster.CampaignID, monster)
}

func (w *monsterRepoWrapper) Read(campaignID, name string) (*domain.Monster, error) {
	return w.fs.Read(context.Background(), campaignID, name)
}

func (w *monsterRepoWrapper) List(campaignID string) ([]domain.Monster, error) {
	ptrs, err := w.fs.List(context.Background(), campaignID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Monster, len(ptrs))
	for i, p := range ptrs {
		result[i] = *p
	}
	return result, nil
}

func (w *monsterRepoWrapper) Delete(ctx context.Context, campaignID, name string) error {
	return w.fs.Delete(ctx, campaignID, name)
}
