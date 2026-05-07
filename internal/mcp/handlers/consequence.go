package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/services"
)

// ConsequenceHandlers handles consequence engine MCP tools
type ConsequenceHandlers struct {
	engine       *services.ConsequenceEngine
	stateService *services.NarrativeStateService
}

// NewConsequenceHandlers creates new consequence handlers
func NewConsequenceHandlers(engine *services.ConsequenceEngine, stateService *services.NarrativeStateService) *ConsequenceHandlers {
	return &ConsequenceHandlers{engine: engine, stateService: stateService}
}

// HandleEvaluateConsequences handles the evaluate_consequences tool
func (h *ConsequenceHandlers) HandleEvaluateConsequences() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		if campaignID == "" {
			return mcp.NewToolResultError("campaign_id is required"), nil
		}

		// Load narrative state
		state, err := h.stateService.Load(ctx, campaignID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to load narrative state: %v", err)), nil
		}

		// Evaluate consequence rules
		eval, err := h.engine.Evaluate(ctx, campaignID, state)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to evaluate consequences: %v", err)), nil
		}

		result := map[string]any{
			"campaign_id":       campaignID,
			"session_num":       eval.SessionNum,
			"triggered_rules":   len(eval.TriggeredRules),
			"immediate_effects": len(eval.ImmediateEffects),
			"delayed_effects":   len(eval.DelayedEffects),
			"rules":             eval.TriggeredRules,
			"immediate":         eval.ImmediateEffects,
			"delayed":           eval.DelayedEffects,
		}

		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}
