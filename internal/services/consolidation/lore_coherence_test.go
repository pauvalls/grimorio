package consolidation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoreCoherence_Analyze_ContradictoryDate(t *testing.T) {
	files := []CampaignFile{
		{RelPath: "lore/treaties.md", Content: "The Treaty of Ashford was signed in 1247."},
		{RelPath: "acts/act2.md", Content: "The Treaty of Ashford happened in 1251."},
	}

	l := NewLoreCoherence()
	res, err := l.Analyze(context.Background(), files)
	require.NoError(t, err)
	assert.False(t, res.Passed)
	assert.Equal(t, "timeline_consistency", res.Rule)
	require.Len(t, res.Issues, 1)
	assert.Contains(t, res.Issues[0].Message, "Treaty of Ashford")
	assert.Contains(t, res.Issues[0].Message, "1247")
	assert.Contains(t, res.Issues[0].Message, "1251")
}

func TestLoreCoherence_Analyze_NoContradiction(t *testing.T) {
	files := []CampaignFile{
		{RelPath: "lore/treaties.md", Content: "The Treaty of Ashford was signed in 1247."},
		{RelPath: "acts/act2.md", Content: "The Treaty of Ashford was signed in 1247."},
	}

	l := NewLoreCoherence()
	res, err := l.Analyze(context.Background(), files)
	require.NoError(t, err)
	assert.True(t, res.Passed)
}

func TestLoreCoherence_Analyze_PrimordialEntityMultipleContexts(t *testing.T) {
	files := []CampaignFile{
		{RelPath: "lore/old.md", Content: "# The Primordial Flame\nAn old one."},
		{RelPath: "acts/act4.md", Content: "# The Ancient Flame\nThe ancient Flame destroyed the city."},
	}

	l := NewLoreCoherence()
	res, err := l.Analyze(context.Background(), files)
	require.NoError(t, err)
	assert.False(t, res.Passed)
	assert.NotEmpty(t, res.Issues)
}
