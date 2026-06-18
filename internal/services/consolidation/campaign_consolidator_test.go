package consolidation

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sliceReader struct {
	files []CampaignFile
}

func (r *sliceReader) ReadCampaign(ctx context.Context, campaignID string) ([]CampaignFile, error) {
	return r.files, nil
}

func TestCampaignConsolidator_Detect_FindsDrift(t *testing.T) {
	files := []CampaignFile{
		{RelPath: "npcs/npcs.md", Content: "# Velaplata"},
		{RelPath: "lore/history.md", Content: "# Velaplanta"},
		{RelPath: "acts/act2.md", Content: "# Murder of Duke Aldric"},
		{RelPath: "acts/act4.md", Content: "# Murder of Duke Aldric"},
	}

	c := NewCampaignConsolidatorWithReader("/tmp", &sliceReader{files: files})
	report, err := c.Detect(context.Background(), "test")
	require.NoError(t, err)
	assert.Equal(t, "test", report.CampaignID)
	assert.NotEmpty(t, report.ChecksRun)
	assert.NotEmpty(t, report.RemainingIssues)
	assert.NotEmpty(t, report.NeedsHuman)
}

func TestCampaignConsolidator_Consolidate_AutoFix(t *testing.T) {
	tmp := t.TempDir()
	npcsPath := filepath.Join(tmp, "npcs.md")
	historyPath := filepath.Join(tmp, "history.md")
	require.NoError(t, os.WriteFile(npcsPath, []byte("# Velaplata\nVelaplata is a house."), 0644))
	require.NoError(t, os.WriteFile(historyPath, []byte("# Velaplanta\nVelaplanta is old."), 0644))

	files := []CampaignFile{
		{Path: npcsPath, RelPath: "npcs.md", Content: "# Velaplata\nVelaplata is a house."},
		{Path: historyPath, RelPath: "history.md", Content: "# Velaplanta\nVelaplanta is old."},
	}

	c := NewCampaignConsolidatorWithReader(tmp, &sliceReader{files: files})
	report, err := c.Consolidate(context.Background(), "test", domain.ConsolidationOptions{AutoFix: true})
	require.NoError(t, err)
	assert.NotEmpty(t, report.FixesApplied)

	content, err := os.ReadFile(historyPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Velaplanta")
	// With high-similarity canonical map, only variants below threshold remain unchanged.
}

func TestCampaignConsolidator_RegenerateIndex(t *testing.T) {
	tmp := t.TempDir()
	files := []CampaignFile{
		{RelPath: "lore/history.md", Content: "# History"},
	}

	c := NewCampaignConsolidatorWithReader(tmp, &sliceReader{files: files})
	require.NoError(t, c.RegenerateIndex(context.Background(), "test"))

	indexPath := filepath.Join(tmp, "test", "INDEX.md")
	content, err := os.ReadFile(indexPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Campaign Index")
	assert.Contains(t, string(content), "lore/history.md")
}

func TestCampaignConsolidator_VerifyFreshness(t *testing.T) {
	tmp := t.TempDir()
	campaignDir := filepath.Join(tmp, "test")
	require.NoError(t, os.MkdirAll(campaignDir, 0755))
	now := time.Now()
	require.NoError(t, os.WriteFile(filepath.Join(campaignDir, "campaign.md"), []byte("old"), 0644))
	require.NoError(t, os.Chtimes(filepath.Join(campaignDir, "campaign.md"), now.Add(-time.Hour), now.Add(-time.Hour)))

	files := []CampaignFile{
		{RelPath: "acts/act1.md", Content: "new", ModTime: now},
	}

	c := NewCampaignConsolidatorWithReader(tmp, &sliceReader{files: files})
	fresh, err := c.VerifyFreshness(context.Background(), "test")
	require.NoError(t, err)
	assert.True(t, fresh.CampaignMDStale)
	assert.Contains(t, fresh.Message, "stale")
}
