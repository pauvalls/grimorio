package services

import (
	"context"
	"sync/atomic"
	"testing"
)

// TestFinalizeChapter_CallsAuditor verifies that FinalizeChapter
// calls the bestiary auditor after the chapter is persisted.
// The audit is advisory — the finalize must succeed regardless
// of the auditor's verdict.
func TestFinalizeChapter_CallsAuditor(t *testing.T) {
	svc, _ := setupChapterPartsTest(t)
	spy := newSpyBestiaryAuditor()
	svc.SetBestiaryAuditor(spy)

	// Save 3 minimal parts so finalize has something to assemble.
	_, err := svc.SaveChapterPart("test-campaign", 1, "opener", "The heroes arrive at the dungeon entrance.")
	if err != nil {
		t.Fatalf("SaveChapterPart(opener) error: %v", err)
	}
	_, err = svc.SaveChapterPart("test-campaign", 1, "npcs", "## NPCs\n\n### Guard Captain\nA seasoned warrior.\n")
	if err != nil {
		t.Fatalf("SaveChapterPart(npcs) error: %v", err)
	}
	// 8 areas minimum so validation passes.
	areas1 := generateTestAreas(1, 4, 8)
	areas2 := generateTestAreas(5, 8, 8)
	if _, err := svc.SaveChapterPart("test-campaign", 1, "areas-1", areas1); err != nil {
		t.Fatalf("SaveChapterPart(areas-1) error: %v", err)
	}
	if _, err := svc.SaveChapterPart("test-campaign", 1, "areas-2", areas2); err != nil {
		t.Fatalf("SaveChapterPart(areas-2) error: %v", err)
	}
	if _, err := svc.SaveChapterPart("test-campaign", 1, "closing", "## Consequences\nThe dungeon is cleared.\n"); err != nil {
		t.Fatalf("SaveChapterPart(closing) error: %v", err)
	}

	result, err := svc.FinalizeChapter("test-campaign", 1, "The Dark Cavern")
	if err != nil {
		t.Fatalf("FinalizeChapter() unexpected error: %v", err)
	}
	if result.Status != "ok" {
		t.Errorf("Status = %q, want ok", result.Status)
	}
	if got := spy.chapterFinalizeCalls.Load(); got != 1 {
		t.Errorf("AuditChapterFinalize calls = %d, want 1", got)
	}
	if got, _ := spy.lastChapterContent.Load().(string); got == "" {
		t.Error("auditor was called with empty chapter content")
	}
	if got, _ := spy.lastCampaignID.Load().(string); got != "test-campaign" {
		t.Errorf("last campaign id = %q, want test-campaign", got)
	}
	// And the bestiary path should NOT have been called.
	if got := spy.bestiarySaveCalls.Load(); got != 0 {
		t.Errorf("AuditBestiarySave calls = %d, want 0 (chapter finalize should not call it)", got)
	}
}

// TestFinalizeChapter_NilAuditor_OK verifies that FinalizeChapter
// still works when no auditor is set (backward compatibility).
func TestFinalizeChapter_NilAuditor_OK(t *testing.T) {
	svc, _ := setupChapterPartsTest(t)
	// no SetBestiaryAuditor call — the field is nil

	_, err := svc.SaveChapterPart("test-campaign", 1, "opener", "The heroes arrive at the dungeon entrance.")
	if err != nil {
		t.Fatalf("SaveChapterPart(opener) error: %v", err)
	}
	_, err = svc.SaveChapterPart("test-campaign", 1, "npcs", "## NPCs\n\n### Guard Captain\nA seasoned warrior.\n")
	if err != nil {
		t.Fatalf("SaveChapterPart(npcs) error: %v", err)
	}
	areas1 := generateTestAreas(1, 4, 8)
	areas2 := generateTestAreas(5, 8, 8)
	if _, err := svc.SaveChapterPart("test-campaign", 1, "areas-1", areas1); err != nil {
		t.Fatalf("SaveChapterPart(areas-1) error: %v", err)
	}
	if _, err := svc.SaveChapterPart("test-campaign", 1, "areas-2", areas2); err != nil {
		t.Fatalf("SaveChapterPart(areas-2) error: %v", err)
	}
	if _, err := svc.SaveChapterPart("test-campaign", 1, "closing", "## Consequences\nThe dungeon is cleared.\n"); err != nil {
		t.Fatalf("SaveChapterPart(closing) error: %v", err)
	}

	if _, err := svc.FinalizeChapter("test-campaign", 1, "No Auditor"); err != nil {
		t.Fatalf("FinalizeChapter() with nil auditor: %v", err)
	}
}

// TestFinalizeChapter_AuditorPanicsButFinalizeSucceeds verifies
// that a panicking auditor cannot break the finalize. The auditor
// is advisory — a panic in it must be recovered and the finalize
// must still return success.
func TestFinalizeChapter_AuditorPanicsButFinalizeSucceeds(t *testing.T) {
	svc, _ := setupChapterPartsTest(t)
	svc.SetBestiaryAuditor(&panickyAuditor{})

	_, err := svc.SaveChapterPart("test-campaign", 1, "opener", "The heroes arrive at the dungeon entrance.")
	if err != nil {
		t.Fatalf("SaveChapterPart(opener) error: %v", err)
	}
	_, err = svc.SaveChapterPart("test-campaign", 1, "npcs", "## NPCs\n\n### Guard Captain\nA seasoned warrior.\n")
	if err != nil {
		t.Fatalf("SaveChapterPart(npcs) error: %v", err)
	}
	areas1 := generateTestAreas(1, 4, 8)
	areas2 := generateTestAreas(5, 8, 8)
	if _, err := svc.SaveChapterPart("test-campaign", 1, "areas-1", areas1); err != nil {
		t.Fatalf("SaveChapterPart(areas-1) error: %v", err)
	}
	if _, err := svc.SaveChapterPart("test-campaign", 1, "areas-2", areas2); err != nil {
		t.Fatalf("SaveChapterPart(areas-2) error: %v", err)
	}
	if _, err := svc.SaveChapterPart("test-campaign", 1, "closing", "## Consequences\nThe dungeon is cleared.\n"); err != nil {
		t.Fatalf("SaveChapterPart(closing) error: %v", err)
	}

	if _, err := svc.FinalizeChapter("test-campaign", 1, "Panicky Auditor"); err != nil {
		t.Fatalf("FinalizeChapter() failed even though auditor panicked: %v", err)
	}
}

// panickyAuditor panics on every call. The CampaignService must
// recover (or the test panics here) and the save must succeed.
type panickyAuditor struct{}

func (p *panickyAuditor) AuditBestiarySave(ctx context.Context, content string, campaignID string) {
	panic("simulated audit panic")
}

func (p *panickyAuditor) AuditChapterFinalize(ctx context.Context, chapterContent string, campaignID string) {
	panic("simulated audit panic")
}

// Force the use of context to keep imports stable.
var _ = atomic.Int32{}
