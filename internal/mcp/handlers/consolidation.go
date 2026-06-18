package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/services"
)

// ConsolidationHandlers exposes the campaign consistency consolidation
// surface as MCP tools. The five tools are:
//   - consolidate_campaign
//   - detect_inconsistencies
//   - resolve_ambiguity
//   - regenerate_index
//   - verify_campaign_freshness
//
// All five delegate to a *services.ConsolidationAdapter which wraps
// the internal/services/consolidation engine.
type ConsolidationHandlers struct {
	adapter *services.ConsolidationAdapter
}

// NewConsolidationHandlers builds a ConsolidationHandlers bound to the
// given base directory (the root under which campaign subdirectories live).
func NewConsolidationHandlers(adapter *services.ConsolidationAdapter) *ConsolidationHandlers {
	return &ConsolidationHandlers{adapter: adapter}
}

// HandleConsolidateCampaign handles the consolidate_campaign tool.
func (h *ConsolidationHandlers) HandleConsolidateCampaign() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		campaignID := getStringArg(args, "campaign")
		if campaignID == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}
		opts := domain.ConsolidationOptions{
			AutoFix:                    getBoolArg(args, "auto_fix"),
			EntitySimilarityThreshold:  getFloat64Arg(args, "similarity_threshold"),
			BackupDir:                  getStringArg(args, "backup_dir"),
		}
		if opts.EntitySimilarityThreshold == 0 {
			opts.EntitySimilarityThreshold = 0.85
		}

		report, err := h.adapter.Consolidate(ctx, campaignID, opts)
		if err != nil {
			return ToToolResult(err), nil
		}
		return jsonResult(report)
	}
}

// HandleDetectInconsistencies handles the detect_inconsistencies tool
// (read-only detection, no file mutations).
func (h *ConsolidationHandlers) HandleDetectInconsistencies() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		campaignID := getStringArg(args, "campaign")
		if campaignID == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}

		report, err := h.adapter.Detect(ctx, campaignID)
		if err != nil {
			return ToToolResult(err), nil
		}
		return jsonResult(report)
	}
}

// HandleResolveAmbiguity handles the resolve_ambiguity tool.
func (h *ConsolidationHandlers) HandleResolveAmbiguity() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		campaignID := getStringArg(args, "campaign")
		questionID := getStringArg(args, "question_id")
		decision := getStringArg(args, "decision")
		if campaignID == "" || questionID == "" || decision == "" {
			return mcp.NewToolResultError("campaign, question_id and decision are required"), nil
		}

		if err := h.adapter.ResolveAmbiguity(ctx, campaignID, questionID, decision); err != nil {
			return ToToolResult(err), nil
		}
		return jsonResult(map[string]any{
			"resolved":   true,
			"campaign":   campaignID,
			"question_id": questionID,
			"decision":   decision,
		})
	}
}

// HandleRegenerateIndex handles the regenerate_index tool.
func (h *ConsolidationHandlers) HandleRegenerateIndex() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		campaignID := getStringArg(args, "campaign")
		if campaignID == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}
		if err := h.adapter.RegenerateIndex(ctx, campaignID); err != nil {
			return ToToolResult(err), nil
		}
		return jsonResult(map[string]any{
			"regenerated": true,
			"campaign":    campaignID,
			"index_path":  "INDEX.md",
		})
	}
}

// HandleVerifyCampaignFreshness handles the verify_campaign_freshness tool.
func (h *ConsolidationHandlers) HandleVerifyCampaignFreshness() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		campaignID := getStringArg(args, "campaign")
		if campaignID == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}
		fresh, err := h.adapter.VerifyFreshness(ctx, campaignID)
		if err != nil {
			return ToToolResult(err), nil
		}
		return jsonResult(fresh)
	}
}

// jsonResult marshals v to indented JSON and returns it as a tool text
// result. Errors are returned as a tool error result.
func jsonResult(v any) (*mcp.CallToolResult, error) {
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(bytes)), nil
}

// getFloat64Arg extracts a float64 argument from the args map.
func getFloat64Arg(args map[string]any, key string) float64 {
	if val, ok := args[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0
}
