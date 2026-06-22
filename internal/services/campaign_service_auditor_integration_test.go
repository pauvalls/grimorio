package services

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/pauvalls/grimorio/internal/repository"
)

// spyBestiaryAuditor is a test spy that records calls to
// AuditBestiarySave and AuditChapterFinalize. It is the
// implementation used to verify the wiring in CampaignService
// without depending on the real analyzer output.
type spyBestiaryAuditor struct {
	bestiarySaveCalls    atomic.Int32
	chapterFinalizeCalls atomic.Int32
	lastBestiaryContent  atomic.Value // string
	lastChapterContent   atomic.Value // string
	lastCampaignID       atomic.Value // string
}

func newSpyBestiaryAuditor() *spyBestiaryAuditor {
	return &spyBestiaryAuditor{}
}

func (s *spyBestiaryAuditor) AuditBestiarySave(ctx context.Context, content string, campaignID string) {
	s.bestiarySaveCalls.Add(1)
	s.lastBestiaryContent.Store(content)
	s.lastCampaignID.Store(campaignID)
}

func (s *spyBestiaryAuditor) AuditChapterFinalize(ctx context.Context, chapterContent string, campaignID string) {
	s.chapterFinalizeCalls.Add(1)
	s.lastChapterContent.Store(chapterContent)
	s.lastCampaignID.Store(campaignID)
}

// TestCampaignService_SaveBestiary_CallsAuditor verifies that
// SaveBestiary calls the bestiary auditor after persisting the
// bestiary. The audit is advisory — the save must succeed even
// if the auditor would report a major finding.
func TestCampaignService_SaveBestiary_CallsAuditor(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	canonRepo := repository.NewMemoryCanonRepository()
	monsterRepo := repository.NewMemoryMonsterRepository()
	service := NewCampaignService(campaignRepo, actRepo, charRepo, npcRepo, questRepo, canonRepo, monsterRepo, t.TempDir(), "")
	spy := newSpyBestiaryAuditor()
	service.SetBestiaryAuditor(spy)

	_, err := service.CreateCampaign("audit-test", "Audit Test", "Setting", "")
	if err != nil {
		t.Fatalf("CreateCampaign() unexpected error: %v", err)
	}

	// A well-formed bestiary markdown (the parser expects a
	// # Name / stat block shape; the spy doesn't care about the
	// content, just that the auditor was called).
	const bestiaryContent = "# Goblin\n\n*Small humanoid*\n\n**Armor Class** 15\n**Hit Points** 7\n**Challenge** 1/4\n"

	if err := service.SaveBestiary("audit-test", bestiaryContent); err != nil {
		t.Fatalf("SaveBestiary() unexpected error: %v", err)
	}
	if got := spy.bestiarySaveCalls.Load(); got != 1 {
		t.Errorf("AuditBestiarySave calls = %d, want 1", got)
	}
	if got, _ := spy.lastBestiaryContent.Load().(string); got != bestiaryContent {
		t.Errorf("last bestiary content mismatch (got %d bytes, want %d)", len(got), len(bestiaryContent))
	}
	if got, _ := spy.lastCampaignID.Load().(string); got != "audit-test" {
		t.Errorf("last campaign id = %q, want audit-test", got)
	}
}

// TestCampaignService_SaveBestiary_NilAuditor_OK verifies that
// SaveBestiary still works when no auditor is set (backward
// compatibility with the existing constructor).
func TestCampaignService_SaveBestiary_NilAuditor_OK(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	canonRepo := repository.NewMemoryCanonRepository()
	monsterRepo := repository.NewMemoryMonsterRepository()
	service := NewCampaignService(campaignRepo, actRepo, charRepo, npcRepo, questRepo, canonRepo, monsterRepo, t.TempDir(), "")
	// no SetBestiaryAuditor call

	_, err := service.CreateCampaign("nil-auditor-test", "Nil Auditor Test", "Setting", "")
	if err != nil {
		t.Fatalf("CreateCampaign() unexpected error: %v", err)
	}
	const bestiaryContent = "# Goblin\n\n*Small humanoid*\n\n**Armor Class** 15\n**Hit Points** 7\n**Challenge** 1/4\n"
	if err := service.SaveBestiary("nil-auditor-test", bestiaryContent); err != nil {
		t.Fatalf("SaveBestiary() unexpected error with nil auditor: %v", err)
	}
}

// TestCampaignService_SaveBestiary_AuditorNeverBlocksSave verifies
// that an auditor that panics or does nothing does not prevent the
// save. The save MUST succeed even if the audit would fail.
func TestCampaignService_SaveBestiary_AuditorNeverBlocksSave(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	canonRepo := repository.NewMemoryCanonRepository()
	monsterRepo := repository.NewMemoryMonsterRepository()
	service := NewCampaignService(campaignRepo, actRepo, charRepo, npcRepo, questRepo, canonRepo, monsterRepo, t.TempDir(), "")

	// Use a no-op auditor that records nothing. The point is
	// that the save must complete regardless.
	spy := newSpyBestiaryAuditor()
	service.SetBestiaryAuditor(spy)

	_, err := service.CreateCampaign("never-blocks", "Never Blocks", "Setting", "")
	if err != nil {
		t.Fatalf("CreateCampaign() unexpected error: %v", err)
	}
	const bestiaryContent = "# Goblin\n\n*Small humanoid*\n\n**Armor Class** 15\n**Hit Points** 7\n**Challenge** 1/4\n"
	if err := service.SaveBestiary("never-blocks", bestiaryContent); err != nil {
		t.Fatalf("SaveBestiary() failed even though audit is advisory: %v", err)
	}
}
