package consolidation

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileConsolidator_Analyze_DuplicateFiles(t *testing.T) {
	files := []CampaignFile{
		{RelPath: "a.md", Content: "same content"},
		{RelPath: "b.md", Content: "same content"},
	}

	f := NewFileConsolidator()
	res, err := f.Analyze(context.Background(), files)
	require.NoError(t, err)
	assert.False(t, res.Passed)
	require.Len(t, res.Issues, 1)
	assert.Equal(t, "duplicate_file", res.Issues[0].Rule)
}

func TestFileConsolidator_Analyze_StaleGeneratedFile(t *testing.T) {
	now := time.Now()
	files := []CampaignFile{
		{RelPath: "campaign.md", Content: "compiled", ModTime: now.Add(-time.Hour)},
		{RelPath: "acts/act1.md", Content: "source", ModTime: now},
	}

	f := NewFileConsolidator()
	res, err := f.Analyze(context.Background(), files)
	require.NoError(t, err)
	assert.False(t, res.Passed)
	assert.Contains(t, res.Issues[0].Message, "campaign.md")
}

func TestRemoveDuplicateFiles(t *testing.T) {
	tmp := t.TempDir()
	pathA := filepath.Join(tmp, "a.md")
	pathB := filepath.Join(tmp, "b.md")
	require.NoError(t, os.WriteFile(pathA, []byte("dup"), 0644))
	require.NoError(t, os.WriteFile(pathB, []byte("dup"), 0644))

	files := []CampaignFile{
		{Path: pathA, RelPath: "a.md", Content: "dup"},
		{Path: pathB, RelPath: "b.md", Content: "dup"},
	}

	removed, err := RemoveDuplicateFiles(files)
	require.NoError(t, err)
	require.Len(t, removed, 1)
	assert.Equal(t, "b.md", removed[0])
	assert.NoFileExists(t, pathB)
	assert.FileExists(t, pathA)
}
