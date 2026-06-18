package consolidation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatBlockResolver_Analyze_ConflictingCR(t *testing.T) {
	files := []CampaignFile{
		{RelPath: "bestiary/bosses.md", Content: "# Araxos\n*Large dragon*\nCR 7"},
		{RelPath: "acts/act3.md", Content: "# Araxos\nCR 9"},
	}

	s := NewStatBlockResolver()
	res, err := s.Analyze(context.Background(), files)
	require.NoError(t, err)
	assert.False(t, res.Passed)
	assert.Equal(t, "stat_block_consistency", res.Rule)
	require.Len(t, res.Issues, 1)
	assert.Contains(t, res.Issues[0].Message, "bestiary CR 7")
}

func TestStatBlockResolver_Analyze_NoConflict(t *testing.T) {
	files := []CampaignFile{
		{RelPath: "bestiary/bosses.md", Content: "# Araxos\nCR 7"},
		{RelPath: "acts/act3.md", Content: "# Araxos\nCR 7"},
	}

	s := NewStatBlockResolver()
	res, err := s.Analyze(context.Background(), files)
	require.NoError(t, err)
	assert.True(t, res.Passed)
}
