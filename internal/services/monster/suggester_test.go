package monster

import (
	"strings"
	"testing"

	"github.com/pauvalls/grimorio/internal/monster/rules"
)

// rules_GetStatsForCR is a tiny shim so the suggester tests can compare
// against the canonical values without importing the full rules
// surface in test boilerplate.
func rules_GetStatsForCR(cr float64) (rules.CRStats, error) {
	return rules.GetStatsForCR(cr)
}

func TestSuggestStats_CR5(t *testing.T) {
	t.Parallel()
	s := NewMonsterSuggester()
	m, err := s.Suggest(5, "")
	if err != nil {
		t.Fatalf("Suggest(5) returned error: %v", err)
	}
	if m == nil {
		t.Fatal("Suggest(5) returned nil monster")
	}
	// HP must be in [131, 145].
	if m.HP < 131 || m.HP > 145 {
		t.Errorf("HP = %d, want in [131, 145]", m.HP)
	}
	// AC = 15 (canonical for CR 5).
	if m.AC != 15 {
		t.Errorf("AC = %d, want 15", m.AC)
	}
	// CR = 5.
	if m.CR != 5 {
		t.Errorf("CR = %v, want 5", m.CR)
	}
	// XP = 1800.
	if m.XP != 1800 {
		t.Errorf("XP = %d, want 1800", m.XP)
	}
}

func TestSuggestStats_CR24(t *testing.T) {
	t.Parallel()
	s := NewMonsterSuggester()
	m, err := s.Suggest(24, "")
	if err != nil {
		t.Fatalf("Suggest(24) returned error: %v", err)
	}
	if m.CR != 24 {
		t.Errorf("CR = %v, want 24", m.CR)
	}
	// HP must be in [536, 580].
	if m.HP < 536 || m.HP > 580 {
		t.Errorf("HP = %d, want in [536, 580]", m.HP)
	}
	// AC = 19.
	if m.AC != 19 {
		t.Errorf("AC = %d, want 19", m.AC)
	}
	// XP = 62000.
	if m.XP != 62000 {
		t.Errorf("XP = %d, want 62000", m.XP)
	}
}

func TestSuggestStats_CR0(t *testing.T) {
	t.Parallel()
	s := NewMonsterSuggester()
	m, err := s.Suggest(0, "")
	if err != nil {
		t.Fatalf("Suggest(0) returned error: %v", err)
	}
	if m.CR != 0 {
		t.Errorf("CR = %v, want 0", m.CR)
	}
	// HP must be in [1, 6].
	if m.HP < 1 || m.HP > 6 {
		t.Errorf("HP = %d, want in [1, 6]", m.HP)
	}
	// AC ≤ 13.
	if m.AC > 13 {
		t.Errorf("AC = %d, want ≤ 13", m.AC)
	}
}

func TestSuggestStats_CRSubInteger(t *testing.T) {
	t.Parallel()
	s := NewMonsterSuggester()
	m, err := s.Suggest(0.25, "")
	if err != nil {
		t.Fatalf("Suggest(0.25) returned error: %v", err)
	}
	if m.CR != 0.25 {
		t.Errorf("CR = %v, want 0.25", m.CR)
	}
	if m.HP < 36 || m.HP > 49 {
		t.Errorf("HP = %d, want in [36, 49]", m.HP)
	}
}

func TestSuggestStats_OutOfRange(t *testing.T) {
	t.Parallel()
	s := NewMonsterSuggester()
	_, err := s.Suggest(-1, "")
	if err == nil {
		t.Error("Suggest(-1) returned no error, want error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "cr") {
		t.Errorf("Suggest(-1) error = %v, want CR-related", err)
	}
	_, err = s.Suggest(31, "")
	if err == nil {
		t.Error("Suggest(31) returned no error, want error")
	}
}

func TestSuggestStats_AllCRs(t *testing.T) {
	t.Parallel()
	s := NewMonsterSuggester()
	// All integer CRs 0..30.
	for cr := 0.0; cr <= 30; cr++ {
		m, err := s.Suggest(cr, "")
		if err != nil {
			t.Errorf("Suggest(%v) returned error: %v", cr, err)
			continue
		}
		if m.CR != cr {
			t.Errorf("Suggest(%v).CR = %v, want %v", cr, m.CR, cr)
		}
		// Confirm HP is within the band.
		stats, _ := rules_GetStatsForCR(cr)
		if m.HP < stats.HPMin || m.HP > stats.HPMax {
			t.Errorf("Suggest(%v).HP = %d, want in [%d, %d]", cr, m.HP, stats.HPMin, stats.HPMax)
		}
		// AC matches the canonical value.
		if m.AC != stats.AC {
			t.Errorf("Suggest(%v).AC = %d, want %d", cr, m.AC, stats.AC)
		}
	}
}
