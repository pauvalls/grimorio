package handlers

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/services"
)

func setupTimelineTest() *TimelineHandlers {
	narrativeStateRepo := repository.NewMemoryNarrativeStateRepository()
	canonRepo := repository.NewMemoryCanonRepository()
	narrativeStateService := services.NewNarrativeStateService(narrativeStateRepo, canonRepo)
	return NewTimelineHandlers(narrativeStateService)
}

func seedSessionLog(t *testing.T, nss *services.NarrativeStateService, campaignID string, sessions []domain.SessionRecord) {
	t.Helper()

	// Create narrative state directly in repo
	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     campaignID,
		CurrentSession: len(sessions),
		SessionLog:     sessions,
		LastUpdated:    time.Now(),
	}

	// Save via the service
	repo := repository.NewMemoryNarrativeStateRepository()
	_ = repo.Save(campaignID, state)
	_ = nss // We need to reload from the repo used by the handler
}

func TestHandleSessionTimeline_HappyPath(t *testing.T) {
	_ = setupTimelineTest()

	// Seed session data via the service's underlying repo
	// We need to access the repo that the handler's service uses.
	// Since setupTimelineTest creates a new service with its own repos,
	// let's use a different approach.

	// Create records directly
	now := time.Now()
	sessions := []domain.SessionRecord{
		{
			SessionNum: 1,
			Date:       now.AddDate(0, 0, -14),
			Summary:    "The party arrives in Phandalin and meets Sildar Hallwinter.",
			XPAwarded:  200,
			KeyDecisions: []domain.Decision{
				{ID: "d1", Description: "Help the townmaster", ChoiceMade: "Accepted the quest"},
			},
		},
		{
			SessionNum: 2,
			Date:       now.AddDate(0, 0, -7),
			Summary:    "The party explores the Cragmaw hideout and defeats the goblin chief.",
			XPAwarded:  350,
			KeyDecisions: []domain.Decision{
				{ID: "d2", Description: "Deal with prisoners", ChoiceMade: "Released them"},
			},
		},
	}

	// Save to the handler's service
	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "timeline-test",
		CurrentSession: 2,
		SessionLog:     sessions,
		LastUpdated:    now,
	}
	repository.NewMemoryNarrativeStateRepository().Save("timeline-test", state)

	// Reload: we need a fresh handler that uses the saved data
	nss := services.NewNarrativeStateService(
		repository.NewMemoryNarrativeStateRepository(),
		repository.NewMemoryCanonRepository(),
	)
	// Load the saved state from the first repo and save it to the second
	tmpRepo := repository.NewMemoryNarrativeStateRepository()
	savedState, _ := tmpRepo.Load("timeline-test")
	if savedState != nil {
		// State was saved to tmpRepo, need to use the same repo for the handler
	}
	_ = nss

	// Actually, let's just use the repo directly and create a handler pointing to it
	repo := repository.NewMemoryNarrativeStateRepository()
	repo.Save("timeline-test", &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "timeline-test",
		CurrentSession: 2,
		SessionLog:     sessions,
		LastUpdated:    now,
	})

	handler := NewTimelineHandlers(services.NewNarrativeStateService(repo, repository.NewMemoryCanonRepository())).HandleSessionTimeline()
	result, err := handler(context.Background(), newToolRequest("session_timeline", map[string]any{
		"campaign_id": "timeline-test",
	}))
	if err != nil {
		t.Fatalf("HandleSessionTimeline() error: %v", err)
	}

	if result.IsError {
		t.Fatalf("HandleSessionTimeline() returned error: %v", result.Content)
	}

	text := extractText(result)
	if !strings.Contains(text, "<!DOCTYPE html") && !strings.Contains(text, "<html") {
		t.Error("Response should contain HTML")
	}

	if !strings.Contains(text, "Session 1") {
		t.Error("HTML should contain 'Session 1'")
	}
	if !strings.Contains(text, "Session 2") {
		t.Error("HTML should contain 'Session 2'")
	}
	if !strings.Contains(text, "XP: 200") {
		t.Error("HTML should contain 'XP: 200'")
	}
	if !strings.Contains(text, "XP: 350") {
		t.Error("HTML should contain 'XP: 350'")
	}
	if !strings.Contains(text, "Help the townmaster") {
		t.Error("HTML should contain decision description")
	}
}

func TestHandleSessionTimeline_NoSessions(t *testing.T) {
	repo := repository.NewMemoryNarrativeStateRepository()
	now := time.Now()

	// Seed narrative state with empty session log
	repo.Save("empty-campaign", &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "empty-campaign",
		CurrentSession: 0,
		SessionLog:     []domain.SessionRecord{},
		LastUpdated:    now,
	})

	handler := NewTimelineHandlers(services.NewNarrativeStateService(repo, repository.NewMemoryCanonRepository())).HandleSessionTimeline()

	result, err := handler(context.Background(), newToolRequest("session_timeline", map[string]any{
		"campaign_id": "empty-campaign",
	}))
	if err != nil {
		t.Fatalf("HandleSessionTimeline() error: %v", err)
	}

	if result.IsError {
		t.Fatalf("HandleSessionTimeline() returned error: %v", result.Content)
	}

	text := extractText(result)
	if !strings.Contains(text, "No session data for campaign") {
		t.Errorf("Expected 'No session data for campaign', got: %s", text)
	}
}

func TestHandleSessionTimeline_MissingCampaignID(t *testing.T) {
	timeline := setupTimelineTest()
	handler := timeline.HandleSessionTimeline()

	result, err := handler(context.Background(), newToolRequest("session_timeline", map[string]any{}))
	if err != nil {
		t.Fatalf("HandleSessionTimeline() error: %v", err)
	}

	if !result.IsError {
		t.Error("Expected error for missing campaign_id")
	}
}

func TestSessionTimeline_SummaryTruncation(t *testing.T) {
	repo := repository.NewMemoryNarrativeStateRepository()
	now := time.Now()

	longSummary := strings.Repeat("A very long session summary that should definitely be truncated because it exceeds the 120 character limit. ", 5)

	repo.Save("trunc-test", &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "trunc-test",
		CurrentSession: 1,
		SessionLog: []domain.SessionRecord{
			{
				SessionNum: 1,
				Date:       now,
				Summary:    longSummary,
				XPAwarded:  100,
			},
		},
		LastUpdated: now,
	})

	handler := NewTimelineHandlers(services.NewNarrativeStateService(repo, repository.NewMemoryCanonRepository())).HandleSessionTimeline()
	result, err := handler(context.Background(), newToolRequest("session_timeline", map[string]any{
		"campaign_id": "trunc-test",
	}))
	if err != nil {
		t.Fatalf("HandleSessionTimeline() error: %v", err)
	}

	if result.IsError {
		t.Fatalf("HandleSessionTimeline() returned error: %v", result.Content)
	}

	text := extractText(result)
	// Should contain the ellipsis
	if !strings.Contains(text, "…") {
		t.Error("Truncated summary should contain ellipsis character")
	}
}

func TestSessionTimeline_NoDecisions(t *testing.T) {
	repo := repository.NewMemoryNarrativeStateRepository()
	now := time.Now()

	repo.Save("no-decision-test", &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "no-decision-test",
		CurrentSession: 1,
		SessionLog: []domain.SessionRecord{
			{
				SessionNum: 1,
				Date:       now,
				Summary:    "A session with no decisions.",
				XPAwarded:  100,
			},
		},
		LastUpdated: now,
	})

	handler := NewTimelineHandlers(services.NewNarrativeStateService(repo, repository.NewMemoryCanonRepository())).HandleSessionTimeline()
	result, err := handler(context.Background(), newToolRequest("session_timeline", map[string]any{
		"campaign_id": "no-decision-test",
	}))
	if err != nil {
		t.Fatalf("HandleSessionTimeline() error: %v", err)
	}

	if result.IsError {
		t.Fatalf("HandleSessionTimeline() returned error: %v", result.Content)
	}

	text := extractText(result)
	// Should NOT contain the details expandable section
	if strings.Contains(text, "<details") {
		t.Error("Session without decisions should not have details/expandable section")
	}
}

func TestRenderDecisionsBlock(t *testing.T) {
	empty := renderDecisionsBlock(nil)
	if empty != "" {
		t.Errorf("renderDecisionsBlock(nil) should return empty string, got: %s", empty)
	}

	decisions := []domain.Decision{
		{ID: "d1", Description: "Fight or flee", ChoiceMade: "Fight"},
	}
	html := renderDecisionsBlock(decisions)
	if !strings.Contains(html, "Fight or flee") {
		t.Error("renderDecisionsBlock should contain decision description")
	}
	if !strings.Contains(html, "Fight") {
		t.Error("renderDecisionsBlock should contain choice made")
	}
	if !strings.Contains(html, "<details") {
		t.Error("renderDecisionsBlock should use details/summary")
	}
}
