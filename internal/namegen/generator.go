package namegen

import (
	"fmt"
	"math/rand"
	"strings"
	"unicode"
)

// NameGenerator produces deterministic fantasy names from syllable pools.
type NameGenerator struct {
	pools map[Category]map[Style]*SyllablePool
	rng   *rand.Rand
}

// New creates a NameGenerator with a default random source.
func New() *NameGenerator {
	return &NameGenerator{
		pools: Pools,
		rng:   rand.New(rand.NewSource(rand.Int63())),
	}
}

// NewWithSeed creates a NameGenerator with a seeded random source for reproducibility.
func NewWithSeed(seed int64) *NameGenerator {
	return &NameGenerator{
		pools: Pools,
		rng:   rand.New(rand.NewSource(seed)),
	}
}

// Generate returns 'count' unique names for the given category and style.
// Names are assembled using category-specific algorithms and filtered for
// pronounceability. Invalid combinations return an error.
func (g *NameGenerator) Generate(cat Category, style Style, count int) ([]string, error) {
	if err := Validate(cat, style); err != nil {
		return nil, err
	}
	if count < 1 || count > 50 {
		return nil, fmt.Errorf("count must be between 1 and 50, got %d", count)
	}

	pool, ok := g.pools[cat][style]
	if !ok {
		return nil, fmt.Errorf("no pool found for category %q and style %q", cat, style)
	}

	seen := make(map[string]struct{}, count)
	names := make([]string, 0, count)
	maxAttempts := count * 20 // safety valve to prevent infinite loops

	for attempts := 0; len(names) < count && attempts < maxAttempts; attempts++ {
		name := g.assembleName(cat, pool)
		if !g.isPronounceable(name) {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}

	if len(names) < count {
		return nil, fmt.Errorf("could only generate %d unique names out of %d requested", len(names), count)
	}

	return names, nil
}

// assembleName builds a single name according to the category algorithm.
func (g *NameGenerator) assembleName(cat Category, pool *SyllablePool) string {
	switch cat {
	case CategoryCharacter, CategoryNPC:
		return g.assembleSyllables(pool, false)
	case CategoryCity:
		return g.capitalize(pool.Prefixes[g.rng.Intn(len(pool.Prefixes))] +
			pool.Suffixes[g.rng.Intn(len(pool.Suffixes))])
	case CategoryTavern:
		return "The " + pool.Prefixes[g.rng.Intn(len(pool.Prefixes))] +
			" " + pool.Suffixes[g.rng.Intn(len(pool.Suffixes))]
	case CategoryMonster:
		return g.assembleSyllables(pool, true)
	case CategoryFaction:
		prefix := pool.Prefixes[g.rng.Intn(len(pool.Prefixes))]
		suffix := pool.Suffixes[g.rng.Intn(len(pool.Suffixes))]
		if g.rng.Intn(2) == 0 {
			return prefix + " of the " + suffix
		}
		return g.capitalize(prefix + " " + suffix)
	case CategoryItem:
		return pool.Prefixes[g.rng.Intn(len(pool.Prefixes))] +
			" " + pool.Suffixes[g.rng.Intn(len(pool.Suffixes))]
	default:
		return ""
	}
}

// assembleSyllables builds a name by stacking onset+vowel+[coda] syllables.
// When 'harsh' is true (monster), apostrophes are inserted at ~20% chance.
func (g *NameGenerator) assembleSyllables(pool *SyllablePool, harsh bool) string {
	syllableCount := pool.MinSyllables + g.rng.Intn(pool.MaxSyllables-pool.MinSyllables+1)
	var sb strings.Builder

	for i := 0; i < syllableCount; i++ {
		onset := pool.Onsets[g.rng.Intn(len(pool.Onsets))]
		vowel := pool.Vowels[g.rng.Intn(len(pool.Vowels))]
		var coda string
		if g.rng.Intn(2) == 0 { // 50% chance of coda
			coda = pool.Codas[g.rng.Intn(len(pool.Codas))]
		}

		sb.WriteString(onset)
		sb.WriteString(vowel)
		sb.WriteString(coda)

		if harsh && i < syllableCount-1 && g.rng.Intn(5) == 0 { // 20% chance of apostrophe
			sb.WriteString("'")
		}
	}

	return g.capitalize(sb.String())
}

// capitalize ensures the first rune is upper-case.
func (g *NameGenerator) capitalize(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// isPronounceable enforces the phonetic guard rules.
// Syllable-stacked names are limited to 3-15 chars; compound names can be longer.
func (g *NameGenerator) isPronounceable(name string) bool {
	if len(name) < 3 {
		return false
	}
	// Compound names (spaces or apostrophes in non-monster) may be longer.
	if strings.Contains(name, " ") || strings.Contains(name, "'") {
		if len(name) > 40 {
			return false
		}
	} else if len(name) > 15 {
		return false
	}

	runes := []rune(name)

	// Rule 1: no triple consonant clusters.
	consonantStreak := 0
	for _, r := range runes {
		if g.isConsonant(r) {
			consonantStreak++
			if consonantStreak >= 3 {
				return false
			}
		} else {
			consonantStreak = 0
		}
	}

	// Rule 2: no more than 2 consecutive identical vowels.
	for i := 0; i < len(runes)-2; i++ {
		if g.isVowel(runes[i]) && runes[i] == runes[i+1] && runes[i] == runes[i+2] {
			return false
		}
	}

	return true
}

// isConsonant reports whether r is an English consonant (case-insensitive).
// Apostrophes and spaces are treated as non-consonants.
func (g *NameGenerator) isConsonant(r rune) bool {
	switch unicode.ToLower(r) {
	case 'b', 'c', 'd', 'f', 'g', 'h', 'j', 'k', 'l', 'm',
		'n', 'p', 'q', 'r', 's', 't', 'v', 'w', 'x', 'y', 'z':
		return true
	}
	return false
}

// isVowel reports whether r is an English vowel (case-insensitive).
func (g *NameGenerator) isVowel(r rune) bool {
	switch unicode.ToLower(r) {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	}
	return false
}
