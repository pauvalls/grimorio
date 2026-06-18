package consolidation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityResolver_Analyze_DetectsSimilarNames(t *testing.T) {
	files := []CampaignFile{
		{RelPath: "npcs/npcs.md", Content: "# Velaplata\nA noble house."},
		{RelPath: "lore/history.md", Content: "# Velaplanta\nA misspelled variant."},
	}

	r := NewEntityResolver(0.85)
	res, err := r.Analyze(context.Background(), files)
	require.NoError(t, err)
	assert.False(t, res.Passed)
	assert.Equal(t, "entity_name_uniqueness", res.Rule)
	require.Len(t, res.Fixes, 1)
	assert.Equal(t, "Velaplanta", res.Fixes[0].After)
}

func TestEntityResolver_Analyze_AmbiguousSimilarity(t *testing.T) {
	files := []CampaignFile{
		{RelPath: "a.md", Content: "# Durmiente\nAn ancient being."},
		{RelPath: "b.md", Content: "# Corazón\nA different being."},
	}

	r := NewEntityResolver(0.85)
	res, err := r.Analyze(context.Background(), files)
	require.NoError(t, err)
	// These names are short and dissimilar; expect no collision.
	assert.True(t, res.Passed)
	assert.Empty(t, res.Questions)
}

func TestEntityResolver_CanonicalMap(t *testing.T) {
	files := []CampaignFile{
		{RelPath: "a.md", Content: "# Velaplata"},
		{RelPath: "b.md", Content: "# Velaplanta"},
	}

	r := NewEntityResolver(0.85)
	m := r.CanonicalMap(files)
	assert.Equal(t, "Velaplanta", m["Velaplata"])
}

func TestEntityResolver_NoFalsePositives(t *testing.T) {
	files := []CampaignFile{
		{RelPath: "a.md", Content: "# Aldric"},
		{RelPath: "b.md", Content: "# Beatrice"},
	}

	r := NewEntityResolver(0.85)
	res, err := r.Analyze(context.Background(), files)
	require.NoError(t, err)
	assert.True(t, res.Passed)
}
