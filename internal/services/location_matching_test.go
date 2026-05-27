package services

import (
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestExtractLocationKeywords(t *testing.T) {
	tests := []struct {
		name     string
		hint     string
		wantKeys []string
	}{
		{
			name:     "simple forest",
			hint:     "forest",
			wantKeys: []string{"forest"},
		},
		{
			name:     "multi-word location",
			hint:     "dark forest clearing",
			wantKeys: []string{"dark", "forest", "clearing"},
		},
		{
			name:     "with stop words filtered",
			hint:     "the ancient forest in the mountain",
			wantKeys: []string{"ancient", "forest", "mountain"},
		},
		{
			name:     "with punctuation",
			hint:     "forest, swamp, and mountain.",
			wantKeys: []string{"forest", "swamp", "mountain"},
		},
		{
			name:     "empty hint",
			hint:     "",
			wantKeys: nil,
		},
		{
			name:     "short words filtered",
			hint:     "a big red barn",
			wantKeys: []string{"big", "red", "barn"},
		},
		{
			name:     "mixed case",
			hint:     "Dark FOREST Clearing",
			wantKeys: []string{"dark", "forest", "clearing"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractLocationKeywords(tt.hint)
			if len(got) != len(tt.wantKeys) {
				t.Fatalf("extractLocationKeywords(%q) returned %d keywords, want %d: got=%v, want=%v",
					tt.hint, len(got), len(tt.wantKeys), got, tt.wantKeys)
			}
			for i, key := range got {
				if i < len(tt.wantKeys) && key != tt.wantKeys[i] {
					t.Errorf("keyword[%d] = %q, want %q", i, key, tt.wantKeys[i])
				}
			}
		})
	}
}

func TestMatchesLocation(t *testing.T) {
	tests := []struct {
		name     string
		fact     domain.CanonFact
		keywords []string
		want     bool
	}{
		{
			name: "category exact match",
			fact: domain.CanonFact{
				Category:  "forest",
				Statement: "Patrol in the woods",
			},
			keywords: []string{"forest"},
			want:     true,
		},
		{
			name: "statement keyword match",
			fact: domain.CanonFact{
				Category:  "creature",
				Statement: "A wolf prowls through the dark forest",
			},
			keywords: []string{"forest"},
			want:     true,
		},
		{
			name: "case insensitive match",
			fact: domain.CanonFact{
				Category:  "encounter",
				Statement: "DENSE FOREST patrol",
			},
			keywords: []string{"forest"},
			want:     true,
		},
		{
			name: "no match",
			fact: domain.CanonFact{
				Category:  "urban",
				Statement: "Marketplace bustle",
			},
			keywords: []string{"forest"},
			want:     false,
		},
		{
			name: "empty keywords matches all",
			fact: domain.CanonFact{
				Category:  "urban",
				Statement: "Marketplace",
			},
			keywords: []string{},
			want:     true,
		},
		{
			name: "multiple keywords any match",
			fact: domain.CanonFact{
				Category:  "encounter",
				Statement: "Mountain pass ambush",
			},
			keywords: []string{"forest", "mountain", "swamp"},
			want:     true,
		},
		{
			name: "partial word match",
			fact: domain.CanonFact{
				Category:  "encounter",
				Statement: "Forested area patrol",
			},
			keywords: []string{"forest"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesLocation(tt.fact, tt.keywords)
			if got != tt.want {
				t.Errorf("matchesLocation(fact, %v) = %v, want %v", tt.keywords, got, tt.want)
			}
		})
	}
}
