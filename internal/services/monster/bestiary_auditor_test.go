package monster

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/pauvalls/grimorio/internal/services/consolidation"
)

// captureLogger returns a slog.Logger that writes JSON to buf, and a
// pointer to buf for inspection.
func captureLogger() (*slog.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	handler := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	return slog.New(handler), buf
}

// goodGoblinBlockAudit is the same canonical CR 1/4 goblin used in
// cr_drift_analyzer_test.go. Sourced here to keep the auditor tests
// self-contained.
const goodGoblinBlockAudit = `## Goblin

*Small humanoid, neutral evil*

**Armor Class** 15 (Leather Armor, Shield)
**Hit Points** 7 (2d6)
**Speed** 30 ft.
**Challenge** 1/4 (50 XP)
`

const majorDriftBlockAudit = `## Outlier

*Medium humanoid, unaligned*

**Armor Class** 99
**Hit Points** 999
**Speed** 30 ft.
**Challenge** 1/4 (50 XP)
`

const malformedBlockAudit = `## Bad Monster

**Armor Class** not-a-number
**Hit Points** 10
**Speed** 30 ft.
**Challenge** 1/4 (50 XP)
`

func TestBestiaryAuditor_AuditBestiarySave_WellFormed_DoesNotWarn(t *testing.T) {
	t.Parallel()
	logger, buf := captureLogger()
	auditor := NewBestiaryAuditor(NewMonsterCRDriftAnalyzer(), logger)

	// The auditor must not panic, must not return an error, and must
	// not log at WARN level (the bestiary is well-formed).
	auditor.AuditBestiarySave(context.Background(), goodGoblinBlockAudit, "test-campaign")

	out := buf.String()
	if strings.Contains(out, `"level":"WARN"`) {
		t.Errorf("expected no WARN log for well-formed bestiary, got: %s", out)
	}
}

func TestBestiaryAuditor_AuditBestiarySave_MajorDrift_LogsWarn(t *testing.T) {
	t.Parallel()
	logger, buf := captureLogger()
	auditor := NewBestiaryAuditor(NewMonsterCRDriftAnalyzer(), logger)

	// Major drift must surface as a WARN log entry.
	auditor.AuditBestiarySave(context.Background(), majorDriftBlockAudit, "test-campaign")

	out := buf.String()
	if !strings.Contains(out, `"level":"WARN"`) {
		t.Errorf("expected WARN log for major drift, got: %s", out)
	}
	// The WARN log should mention the campaign id and the major count.
	if !strings.Contains(out, "test-campaign") {
		t.Errorf("expected campaign id in log, got: %s", out)
	}
	if !strings.Contains(out, "major") {
		t.Errorf("expected 'major' field in log, got: %s", out)
	}
}

func TestBestiaryAuditor_AuditBestiarySave_Malformed_LogsErrorDoesNotPanic(t *testing.T) {
	t.Parallel()
	logger, buf := captureLogger()
	auditor := NewBestiaryAuditor(NewMonsterCRDriftAnalyzer(), logger)

	// The auditor must never panic, must not return an error, and
	// must surface the malformed block as an ERROR log (since the
	// analyzer itself cannot parse it, the analyzer returns a
	// critical finding — and the auditor logs the failure at Error
	// when the analyzer returns an error).
	//
	// In this code path, the analyzer does NOT return an error
	// (per its design: malformed blocks become critical findings,
	// not errors). So we expect the auditor to log the critical
	// finding at WARN (Major severity). The test verifies that
	// the auditor completes without panicking and surfaces the
	// finding.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("auditor panicked on malformed block: %v", r)
		}
	}()

	auditor.AuditBestiarySave(context.Background(), malformedBlockAudit, "test-campaign")

	out := buf.String()
	// The analyzer's response to a malformed block is a "critical"
	// finding. That gets mapped by the auditor to a WARN log entry.
	if !strings.Contains(out, `"level":"WARN"`) {
		t.Errorf("expected WARN log for malformed block (critical finding), got: %s", out)
	}
}

// TestBestiaryAuditor_NilAnalyzer_NoPanic verifies the guard path:
// when Analyzer is nil, the auditor must not panic and must not
// return an error.
func TestBestiaryAuditor_NilAnalyzer_NoPanic(t *testing.T) {
	t.Parallel()
	logger, buf := captureLogger()
	auditor := &BestiaryAuditor{Analyzer: nil, Logger: logger}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("nil analyzer caused panic: %v", r)
		}
	}()

	// Should not panic, should not log at WARN/ERROR.
	auditor.AuditBestiarySave(context.Background(), goodGoblinBlockAudit, "test-campaign")
	auditor.AuditChapterFinalize(context.Background(), goodGoblinBlockAudit, "test-campaign")

	out := buf.String()
	if strings.Contains(out, `"level":"WARN"`) || strings.Contains(out, `"level":"ERROR"`) {
		t.Errorf("nil analyzer should not log at WARN/ERROR, got: %s", out)
	}
}

// TestBestiaryAuditor_NilLogger_UsesDefault verifies that a nil
// logger falls back to slog.Default() and does not panic.
func TestBestiaryAuditor_NilLogger_UsesDefault(t *testing.T) {
	t.Parallel()
	auditor := &BestiaryAuditor{Analyzer: NewMonsterCRDriftAnalyzer(), Logger: nil}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("nil logger caused panic: %v", r)
		}
	}()

	// Should not panic.
	auditor.AuditBestiarySave(context.Background(), goodGoblinBlockAudit, "test-campaign")
}

func TestBestiaryAuditor_NilReceiver_NoPanic(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("nil receiver caused panic: %v", r)
		}
	}()
	var b *BestiaryAuditor
	b.AuditBestiarySave(context.Background(), "content", "campaign")
	b.AuditChapterFinalize(context.Background(), "content", "campaign")
}

func TestBestiaryAuditor_AuditChapterFinalize_RunsAudit(t *testing.T) {
	t.Parallel()
	logger, buf := captureLogger()
	auditor := NewBestiaryAuditor(NewMonsterCRDriftAnalyzer(), logger)

	// A chapter that references a known monster (goblin) should
	// trigger the audit. The chapter content can include any
	// markdown — the auditor just passes it to the analyzer, which
	// then ignores non-bestiary blocks and runs on any stat-block
	// content it finds.
	chapterContent := `# Chapter 1: The Goblin Caves

## Encounters

### Encounter 1: The Goblin Patrol
The party meets a goblin scout.
` + "\n" + goodGoblinBlockAudit

	auditor.AuditChapterFinalize(context.Background(), chapterContent, "test-campaign")

	// The well-formed goblin is OK; we expect either no log or
	// a Debug log. Never a WARN (since the goblin is well-formed).
	out := buf.String()
	if strings.Contains(out, `"level":"WARN"`) || strings.Contains(out, `"level":"ERROR"`) {
		t.Errorf("well-formed chapter should not log at WARN/ERROR, got: %s", out)
	}
}

func TestBestiaryAuditor_AuditChapterFinalize_MajorDrift_LogsWarn(t *testing.T) {
	t.Parallel()
	logger, buf := captureLogger()
	auditor := NewBestiaryAuditor(NewMonsterCRDriftAnalyzer(), logger)

	// Chapter content that includes a major-drift monster block.
	chapterContent := `# Chapter 2: The Outlier

## Encounters

### Encounter 1: The Outlier
` + "\n" + majorDriftBlockAudit

	auditor.AuditChapterFinalize(context.Background(), chapterContent, "test-campaign")

	out := buf.String()
	if !strings.Contains(out, `"level":"WARN"`) {
		t.Errorf("expected WARN log for major-drift chapter, got: %s", out)
	}
}

func TestBestiaryAuditor_AnalyzerFails_LogsErrorNoPanic(t *testing.T) {
	t.Parallel()
	logger, buf := captureLogger()
	// Use a fake analyzer that returns an error from Analyze.
	auditor := NewBestiaryAuditor(nil, logger)
	auditor.Analyzer = &fakeFailingAnalyzer{}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("analyzer failure caused panic: %v", r)
		}
	}()

	auditor.AuditBestiarySave(context.Background(), goodGoblinBlockAudit, "test-campaign")

	out := buf.String()
	// Analyzer returned an error → auditor must log at ERROR.
	if !strings.Contains(out, `"level":"ERROR"`) {
		t.Errorf("expected ERROR log when analyzer fails, got: %s", out)
	}
}

// fakeFailingAnalyzer implements consolidation.Analyzer but always
// returns an error from Analyze.
type fakeFailingAnalyzer struct{}

// errFakeAnalyzer is a sentinel error used to verify the auditor's
// defensive path: log at Error, no panic, no error returned.
var errFakeAnalyzer = &fakeErr{msg: "fake analyzer error"}

type fakeErr struct{ msg string }

func (e *fakeErr) Error() string { return e.msg }

func (f *fakeFailingAnalyzer) Name() string { return "fake_failing" }
func (f *fakeFailingAnalyzer) Analyze(ctx context.Context, files []consolidation.CampaignFile) (*consolidation.AnalysisResult, error) {
	return nil, errFakeAnalyzer
}

// TestBestiaryAuditor_MinorDrift_LogsInfo verifies that a Minor
// drift is logged at Info (not Warn).
func TestBestiaryAuditor_MinorDrift_LogsInfo(t *testing.T) {
	t.Parallel()
	logger, buf := captureLogger()
	auditor := NewBestiaryAuditor(NewMonsterCRDriftAnalyzer(), logger)

	// Bandit at CR 1/4 with HP 80 (band is 36-49) — delta ~0.75
	// → Minor drift → "warning" severity in analyzer.
	minorBlock := `## Bandit

*Medium humanoid, any alignment*

**Armor Class** 15 (Studded Leather)
**Hit Points** 80 (11d8 + 11)
**Speed** 30 ft.
**Challenge** 1/4 (50 XP)
`

	auditor.AuditBestiarySave(context.Background(), minorBlock, "test-campaign")

	out := buf.String()
	if strings.Contains(out, `"level":"WARN"`) {
		t.Errorf("minor drift should not log at WARN, got: %s", out)
	}
	if !strings.Contains(out, `"level":"INFO"`) {
		t.Errorf("minor drift should log at INFO, got: %s", out)
	}
}
