package consolidation

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapReferenceChecker_Analyze_MissingMap(t *testing.T) {
	tmp := t.TempDir()
	files := []CampaignFile{
		{RelPath: "areas/cellar.md", Content: "![cellar](assets/maps/cellar.svg)"},
	}

	m := NewMapReferenceChecker(tmp)
	res, err := m.Analyze(context.Background(), files)
	require.NoError(t, err)
	assert.False(t, res.Passed)
	require.Len(t, res.Issues, 1)
	assert.Contains(t, res.Issues[0].Message, "cellar.svg")
}

func TestMapReferenceChecker_Analyze_ExistingMap(t *testing.T) {
	tmp := t.TempDir()
	mapPath := filepath.Join(tmp, "assets", "maps", "cellar.svg")
	require.NoError(t, os.MkdirAll(filepath.Dir(mapPath), 0755))
	require.NoError(t, os.WriteFile(mapPath, []byte("<svg/>"), 0644))

	files := []CampaignFile{
		{RelPath: "areas/cellar.md", Content: "![cellar](assets/maps/cellar.svg)"},
	}

	m := NewMapReferenceChecker(tmp)
	res, err := m.Analyze(context.Background(), files)
	require.NoError(t, err)
	assert.True(t, res.Passed)
}
