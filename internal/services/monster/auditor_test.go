package monster

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// nowFunc returns a clock; abstracted to allow tests to mock time.
var nowFunc = func() time.Time { return time.Now() }

func TestAuditCampaign_Nonexistent(t *testing.T) {
	t.Parallel()
	a := NewMonsterAuditor("")
	_, err := a.AuditCampaign("this-campaign-does-not-exist-12345")
	if err == nil {
		t.Error("AuditCampaign(nonexistent) returned no error, want not_found")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "not found") {
		t.Errorf("AuditCampaign(nonexistent) error = %v, want 'not found'", err)
	}
}

func TestAuditCampaign_RealCampaign(t *testing.T) {
	// Not parallel — uses the real home directory.
	campaigns := os.Getenv("HOME") + "/campaigns"
	if _, err := os.Stat(campaigns); err != nil {
		t.Skipf("no ~/campaigns directory: %v", err)
	}
	entries, err := os.ReadDir(campaigns)
	if err != nil || len(entries) == 0 {
		t.Skipf("no campaigns in %s: %v", campaigns, err)
	}
	// Use the first campaign that has a bestiary.
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		bestiaryPath := filepath.Join(campaigns, e.Name(), "bestiary", "bestiary.md")
		if _, err := os.Stat(bestiaryPath); err != nil {
			continue
		}
		a := NewMonsterAuditor("")
		report, err := a.AuditCampaign(e.Name())
		if err != nil {
			t.Errorf("AuditCampaign(%q) returned error: %v", e.Name(), err)
			continue
		}
		if report == nil {
			t.Errorf("AuditCampaign(%q) returned nil report", e.Name())
			continue
		}
		if report.CampaignID != e.Name() {
			t.Errorf("report.CampaignID = %q, want %q", report.CampaignID, e.Name())
		}
		if report.Summary.Total < 1 {
			t.Errorf("report.Summary.Total = %d, want >= 1", report.Summary.Total)
		}
		return
	}
	t.Skipf("no campaign with bestiary/bestiary.md in %s", campaigns)
}

func TestAuditCampaign_Synthetic(t *testing.T) {
	// Create a temp campaign with a synthetic bestiary.
	t.Parallel()
	dir := t.TempDir()
	bestiaryDir := filepath.Join(dir, "test-campaign", "bestiary")
	if err := os.MkdirAll(bestiaryDir, 0755); err != nil {
		t.Fatal(err)
	}
	bestiaryPath := filepath.Join(bestiaryDir, "bestiary.md")
	content := `# Bestiario: Test

## Test Goblin

*Small humanoid, neutral evil*

**Armor Class** 15
**Hit Points** 7
**Speed** 30 ft.
**Challenge** 1/4 (50 XP)

## Test Beholder

*Large aberration, lawful evil*

**Armor Class** 18
**Hit Points** 180
**Speed** 0 ft., fly 20 ft.
**Challenge** 13 (10,000 XP)
`
	if err := os.WriteFile(bestiaryPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	a := NewMonsterAuditor(dir)
	report, err := a.AuditCampaign("test-campaign")
	if err != nil {
		t.Fatalf("AuditCampaign returned error: %v", err)
	}
	if report.Summary.Total != 2 {
		t.Errorf("Summary.Total = %d, want 2", report.Summary.Total)
	}
}

func TestAuditCampaign_NotFoundError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := NewMonsterAuditor(dir)
	_, err := a.AuditCampaign("nonexistent")
	if err == nil {
		t.Fatal("AuditCampaign returned no error for nonexistent campaign")
	}
}

func TestAuditCampaign_Performance(t *testing.T) {
	// 100 monsters should be audited in < 5s.
	t.Parallel()
	dir := t.TempDir()
	bestiaryDir := filepath.Join(dir, "perf", "bestiary")
	if err := os.MkdirAll(bestiaryDir, 0755); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	b.WriteString("# Bestiario\n\n")
	for i := 0; i < 100; i++ {
		b.WriteString("## Monster ")
		b.WriteString(string(rune('A' + i%26)))
		b.WriteString(" ")
		b.WriteString(string(rune('0' + i/10)))
		b.WriteString("\n\n*Medium humanoid, neutral*\n\n")
		b.WriteString("**Armor Class** 13\n")
		b.WriteString("**Hit Points** 30\n")
		b.WriteString("**Speed** 30 ft.\n")
		b.WriteString("**Challenge** 1/2 (100 XP)\n\n")
	}
	if err := os.WriteFile(filepath.Join(bestiaryDir, "bestiary.md"), []byte(b.String()), 0644); err != nil {
		t.Fatal(err)
	}
	a := NewMonsterAuditor(dir)
	start := nowFunc()
	report, err := a.AuditCampaign("perf")
	elapsed := nowFunc().Sub(start)
	if err != nil {
		t.Fatalf("AuditCampaign returned error: %v", err)
	}
	if report.Summary.Total != 100 {
		t.Errorf("Summary.Total = %d, want 100", report.Summary.Total)
	}
	if elapsed > 5_000_000_000 { // 5 seconds in nanoseconds
		t.Errorf("AuditCampaign took %v for 100 monsters, want < 5s", elapsed)
	}
}
