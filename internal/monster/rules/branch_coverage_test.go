package rules

import "testing"

// These tests target uncovered branches in the existing code to bring
// the package coverage above 95%.

func TestGetStatsForCR_ErrorPath(t *testing.T) {
	t.Parallel()
	// GetStatsForCR with a CR that's not in the master table.
	// The implementation snaps to nearest then re-looks up.
	// 0.3 is closest to 0.25 in the table.
	stats, err := GetStatsForCR(0.3)
	if err != nil {
		t.Fatalf("GetStatsForCR(0.3) returned error: %v", err)
	}
	if stats.PB != 2 {
		t.Errorf("PB for snapped CR 0.3 = %d, want 2", stats.PB)
	}
}

func TestXPForCR_SnapToNearest(t *testing.T) {
	t.Parallel()
	// 0.3 is closest to 0.25 → XP 50.
	if got := XPForCR(0.3); got != 50 {
		t.Errorf("XPForCR(0.3) = %d, want 50 (snapped to CR 1/4)", got)
	}
}

func TestDefensiveCRFromHP_SnapToNearest(t *testing.T) {
	t.Parallel()
	// 1000 HP is way above CR 30's 850 → returns 30.
	if got := DefensiveCRFromHP(1000); got != 30 {
		t.Errorf("DefensiveCRFromHP(1000) = %v, want 30", got)
	}
}

func TestOffensiveCRFromDPR_1000(t *testing.T) {
	t.Parallel()
	// 1000 DPR → CR 30.
	if got := OffensiveCRFromDPR(1000); got != 30 {
		t.Errorf("OffensiveCRFromDPR(1000) = %v, want 30", got)
	}
}

func TestAdjustCRByAC_ErrorPath(t *testing.T) {
	t.Parallel()
	// AdjustCRByAC for an out-of-range CR returns the same value.
	if got := AdjustCRByAC(-1, 15); got != -1 {
		t.Errorf("AdjustCRByAC(-1, 15) = %v, want -1", got)
	}
}

func TestAdjustCRByAttack_ErrorPath(t *testing.T) {
	t.Parallel()
	if got := AdjustCRByAttack(-1, 5); got != -1 {
		t.Errorf("AdjustCRByAttack(-1, 5) = %v, want -1", got)
	}
}

func TestAdjustCRBySaveDC_ErrorPath(t *testing.T) {
	t.Parallel()
	if got := AdjustCRBySaveDC(-1, 13); got != -1 {
		t.Errorf("AdjustCRBySaveDC(-1, 13) = %v, want -1", got)
	}
}

func TestParseCR_OutOfRangeFloat(t *testing.T) {
	t.Parallel()
	// 30.5 is out of range, even though 0..30 is the valid range.
	if _, err := ParseCR("30.5"); err == nil {
		t.Error("ParseCR(\"30.5\") returned no error, want error")
	}
}

func TestEffectiveHP_Clamped(t *testing.T) {
	t.Parallel()
	// If the multipliers stack to less than 1 (shouldn't happen with current
	// table but the clamp is defensive), we floor at 1.
	// CR 17+ has res=1, imm=1.25 → (1-1)+(1.25-1) = 0.25 added → ×1.25.
	got := EffectiveHP(20, 100, true, true)
	want := 125
	if got != want {
		t.Errorf("EffectiveHP(20, 100, true, true) = %d, want %d", got, want)
	}
}

func TestHasRangedDamage_NoMatch(t *testing.T) {
	t.Parallel()
	m := &Monster{
		CR: 5,
		Speed: map[SpeedKind]int{SpeedFly: 60},
		Actions: []Action{
			{Description: "Some kind of area attack centered on itself."},
		},
		BonusActions: nil,
	}
	// No "ranged" or "range " in description.
	if IsFlyingRangedUnderCR10(m) {
		t.Error("expected no flying bonus when no ranged action present")
	}
}
