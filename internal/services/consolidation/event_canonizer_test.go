package consolidation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventCanonizer_Analyze_DuplicatePlacement(t *testing.T) {
	files := []CampaignFile{
		{RelPath: "acts/act2.md", Content: "# Murder of Duke Aldric\nThe duke is found dead."},
		{RelPath: "acts/act4.md", Content: "# Murder of Duke Aldric\nThe true killer is revealed."},
	}

	e := NewEventCanonizer()
	res, err := e.Analyze(context.Background(), files)
	require.NoError(t, err)
	assert.False(t, res.Passed)
	assert.Equal(t, "event_canonical_location", res.Rule)
	require.Len(t, res.Questions, 1)
	assert.Contains(t, res.Questions[0].Question, "Duke Aldric")
}

func TestEventCanonizer_Analyze_SinglePlacement(t *testing.T) {
	files := []CampaignFile{
		{RelPath: "acts/act2.md", Content: "# Murder of Duke Aldric\nThe duke is found dead."},
	}

	e := NewEventCanonizer()
	res, err := e.Analyze(context.Background(), files)
	require.NoError(t, err)
	assert.True(t, res.Passed)
}
