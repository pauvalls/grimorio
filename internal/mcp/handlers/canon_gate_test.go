package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/services"
)

func setupCanonHandlersWithGate() (*CanonHandlers, *services.CanonService, *services.NarrativeStateService, *services.ConsistencyGateService) {
	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	canonSvc := services.NewCanonService(canonRepo, stateRepo)
	stateSvc := services.NewNarrativeStateService(stateRepo, canonRepo)
	validator := services.NewValidationEngine(canonSvc, stateSvc)
	gateSvc := services.NewConsistencyGateService(canonSvc, stateSvc, validator)

	handlers := NewCanonHandlers(canonSvc, stateSvc, validator, gateSvc)
	return handlers, canonSvc, stateSvc, gateSvc
}

func getResultText(result *mcp.CallToolResult) string {
	if len(result.Content) == 0 {
		return ""
	}
	if tc, ok := result.Content[0].(mcp.TextContent); ok {
		return tc.Text
	}
	return ""
}

func TestCanonHandlers_HandleProcessConsistencyGate_Approve(t *testing.T) {
	_, canonSvc, _, gateSvc := setupCanonHandlersWithGate()
	ctx := context.Background()

	// Setup campaign
	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	// Add NPC
	doc, _ := canonSvc.LoadCanon(ctx, "test-campaign")
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:         "npc-001",
		Name:       "Test NPC",
		Type:       domain.EntityTypeNPC,
		CanonState: domain.EntityStateAlive,
	})
	canonSvc.SaveCanon(ctx, doc)

	// Create handler
	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		batchID := getStringArg(args, "batch_id")
		fastMode := getBoolArg(args, "fast_mode")

		artifacts := []domain.ContentProposal{
			{
				ID:      "npc-001-ref",
				Type:    "npc",
				Content: "Test NPC is present.",
				EntityReferences: []domain.EntityReference{
					{EntityID: "npc-001"},
				},
			},
		}

		proposal := domain.BatchProposal{
			BatchID:    batchID,
			CampaignID: campaignID,
			Artifacts:  artifacts,
			Attempt:    1,
		}

		result, err := gateSvc.ProcessBatch(ctx, proposal, fastMode)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"campaign_id": "test-campaign",
		"batch_id":    "batch-001",
		"fast_mode":   false,
	}

	result, err := handler(ctx, request)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", getResultText(result))
	}

	var gateResult domain.GateResult
	if err := json.Unmarshal([]byte(getResultText(result)), &gateResult); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}
	if gateResult.Status != domain.GateStatusApproved {
		t.Fatalf("expected approved, got %s", gateResult.Status)
	}
}

func TestCanonHandlers_HandleProcessConsistencyGate_Reject(t *testing.T) {
	_, canonSvc, stateSvc, gateSvc := setupCanonHandlersWithGate()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	doc, _ := canonSvc.LoadCanon(ctx, "test-campaign")
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:         "npc-001",
		Name:       "Test NPC",
		Type:       domain.EntityTypeNPC,
		CanonState: domain.EntityStateAlive,
	})
	canonSvc.SaveCanon(ctx, doc)

	state, _ := stateSvc.Load(ctx, "test-campaign")
	state.DeadNPCs = append(state.DeadNPCs, domain.NPCDeathRecord{
		NPCID:   "npc-001",
		Name:    "Test NPC",
		Session: 1,
	})
	stateSvc.Save(ctx, state)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		batchID := getStringArg(args, "batch_id")
		fastMode := getBoolArg(args, "fast_mode")

		artifacts := []domain.ContentProposal{
			{
				ID:      "npc-001-ref",
				Type:    "npc",
				Content: "Test NPC gives a quest.",
				EntityReferences: []domain.EntityReference{
					{EntityID: "npc-001"},
				},
			},
		}

		proposal := domain.BatchProposal{
			BatchID:    batchID,
			CampaignID: campaignID,
			Artifacts:  artifacts,
			Attempt:    1,
		}

		result, err := gateSvc.ProcessBatch(ctx, proposal, fastMode)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"campaign_id": "test-campaign",
		"batch_id":    "batch-001",
		"fast_mode":   false,
	}

	result, err := handler(ctx, request)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", getResultText(result))
	}

	var gateResult domain.GateResult
	if err := json.Unmarshal([]byte(getResultText(result)), &gateResult); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}
	if gateResult.Status != domain.GateStatusRejected {
		t.Fatalf("expected rejected, got %s", gateResult.Status)
	}
	if gateResult.RetryPrompt == "" {
		t.Fatal("expected retry prompt in rejected result")
	}
}
