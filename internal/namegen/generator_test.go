package namegen

import (
	"fmt"
	"strings"
	"testing"
)

func TestGenerate_AllCategoryStylePairs(t *testing.T) {
	categories := []Category{
		CategoryCharacter, CategoryNPC, CategoryCity,
		CategoryTavern, CategoryMonster, CategoryFaction, CategoryItem,
	}
	styles := []Style{
		StyleGenericFantasy, StyleElven, StyleDwarven, StyleOrcish, StyleHumanMedieval,
	}

	for _, cat := range categories {
		for _, style := range styles {
			cat := cat
			style := style
			t.Run(fmt.Sprintf("%s/%s", cat, style), func(t *testing.T) {
				t.Parallel()
				g := NewWithSeed(42)
				names, err := g.Generate(cat, style, 5)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(names) != 5 {
					t.Fatalf("expected 5 names, got %d", len(names))
				}
				for _, n := range names {
					if n == "" {
						t.Error("got empty name")
					}
					// Every name must start with an uppercase letter or "The " for taverns
					if cat == CategoryTavern {
						if !strings.HasPrefix(n, "The ") {
							t.Errorf("tavern name %q missing 'The ' prefix", n)
						}
					} else {
						first := n[0]
						if first < 'A' || first > 'Z' {
							t.Errorf("name %q does not start with uppercase", n)
						}
					}
					if len(n) < 3 || len(n) > 50 {
						t.Errorf("name %q length %d out of reasonable bounds", n, len(n))
					}
				}
			})
		}
	}
}

func TestGenerate_DeterministicWithSeed(t *testing.T) {
	g1 := NewWithSeed(123)
	g2 := NewWithSeed(123)

	names1, err := g1.Generate(CategoryCharacter, StyleElven, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	names2, err := g2.Generate(CategoryCharacter, StyleElven, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := range names1 {
		if names1[i] != names2[i] {
			t.Errorf("determinism failed at index %d: %q vs %q", i, names1[i], names2[i])
		}
	}
}

func TestGenerate_InvalidCategory(t *testing.T) {
	g := NewWithSeed(1)
	_, err := g.Generate("spaceship", StyleGenericFantasy, 5)
	if err == nil {
		t.Fatal("expected error for invalid category")
	}
}

func TestGenerate_InvalidStyle(t *testing.T) {
	g := NewWithSeed(1)
	_, err := g.Generate(CategoryNPC, "cyberpunk", 5)
	if err == nil {
		t.Fatal("expected error for invalid style")
	}
}

func TestGenerate_InvalidCount(t *testing.T) {
	g := NewWithSeed(1)
	cases := []int{0, -1, 51, 100}
	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("count=%d", c), func(t *testing.T) {
			t.Parallel()
			_, err := g.Generate(CategoryNPC, StyleGenericFantasy, c)
			if err == nil {
				t.Fatalf("expected error for count=%d", c)
			}
		})
	}
}

func TestGenerate_Uniqueness(t *testing.T) {
	g := NewWithSeed(99)
	names, err := g.Generate(CategoryNPC, StyleGenericFantasy, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	seen := make(map[string]struct{}, len(names))
	for _, n := range names {
		if _, exists := seen[n]; exists {
			t.Errorf("duplicate name: %q", n)
		}
		seen[n] = struct{}{}
	}

	if len(seen) != 50 {
		t.Fatalf("expected 50 unique names, got %d", len(seen))
	}
}

func TestGenerate_Pronounceability(t *testing.T) {
	g := NewWithSeed(77)
	// Generate many names to stress-test the pronounceability filter.
	names, err := g.Generate(CategoryMonster, StyleOrcish, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, n := range names {
		if len(n) < 3 || len(n) > 15 {
			t.Errorf("name %q length %d out of bounds [3,15]", n, len(n))
		}

		// Check for triple consonants.
		lower := strings.ToLower(n)
		for i := 0; i < len(lower)-2; i++ {
			if isConsonantRune(rune(lower[i])) &&
				isConsonantRune(rune(lower[i+1])) &&
				isConsonantRune(rune(lower[i+2])) {
				t.Errorf("name %q contains triple consonant at pos %d", n, i)
			}
		}

		// Check for triple identical vowels.
		for i := 0; i < len(lower)-2; i++ {
			if isVowelRune(rune(lower[i])) &&
				lower[i] == lower[i+1] &&
				lower[i] == lower[i+2] {
				t.Errorf("name %q contains triple identical vowel at pos %d", n, i)
			}
		}
	}
}

func TestGenerate_BoundaryCounts(t *testing.T) {
	t.Run("count=1", func(t *testing.T) {
		t.Parallel()
		g := NewWithSeed(42)
		names, err := g.Generate(CategoryCity, StyleDwarven, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(names) != 1 {
			t.Fatalf("expected 1 name, got %d", len(names))
		}
	})

	t.Run("count=50", func(t *testing.T) {
		t.Parallel()
		g := NewWithSeed(42)
		names, err := g.Generate(CategoryItem, StyleHumanMedieval, 50)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(names) != 50 {
			t.Fatalf("expected 50 names, got %d", len(names))
		}
	})
}

// Helper functions for tests (duplicated to avoid export).
func isConsonantRune(r rune) bool {
	switch r {
	case 'b', 'c', 'd', 'f', 'g', 'h', 'j', 'k', 'l', 'm',
		'n', 'p', 'q', 'r', 's', 't', 'v', 'w', 'x', 'y', 'z':
		return true
	}
	return false
}

func isVowelRune(r rune) bool {
	switch r {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	}
	return false
}
