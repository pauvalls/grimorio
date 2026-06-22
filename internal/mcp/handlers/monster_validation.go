// Package handlers - monster_validation.go exposes the three monster-CR
// MCP tools (validate_monster, suggest_monster_cr, audit_monster_cr).
//
// The three tools are kept in one file because they share the same
// backing services (*monster.MonsterValidator, *monster.MonsterAuditor,
// *monster.MonsterSuggester) and the same response shape philosophy
// (the *ValidationResult is serialized directly to JSON for the wire).
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/monster/rules/parser"
	"github.com/pauvalls/grimorio/internal/monster/rules/renderer"
	"github.com/pauvalls/grimorio/internal/services/monster"
)

// MonsterValidationHandlers bundles the three monster-CR MCP tools.
// The struct is intentionally narrow: it owns one base directory
// (the campaigns root, e.g. ~/campaigns) and three stateless service
// collaborators built lazily on construction.
type MonsterValidationHandlers struct {
	baseDir   string
	validator *monster.MonsterValidator
	auditor   *monster.MonsterAuditor
	suggester *monster.MonsterSuggester
}

// NewMonsterValidationHandlers builds a handler bound to the given
// base directory (the root under which campaign subdirectories live).
// An empty baseDir falls back to $HOME/campaigns to match the
// MonsterAuditor default.
func NewMonsterValidationHandlers(baseDir string) *MonsterValidationHandlers {
	if baseDir == "" {
		baseDir = os.Getenv("HOME") + "/campaigns"
	}
	return &MonsterValidationHandlers{
		baseDir:   baseDir,
		validator: monster.NewMonsterValidator(),
		auditor:   monster.NewMonsterAuditor(baseDir),
		suggester: monster.NewMonsterSuggester(),
	}
}

// HandleValidateMonster handles the validate_monster tool.
//
// Inputs: monster_name (optional, requires campaign), markdown
// (optional). Output: *ValidationResult as JSON.
func (h *MonsterValidationHandlers) HandleValidateMonster() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		monsterName := getStringArg(args, "monster_name")
		markdown := getStringArg(args, "markdown")
		campaign := getStringArg(args, "campaign")

		_, markdownProvided := args["markdown"]
		if monsterName == "" && !markdownProvided {
			return mcp.NewToolResultError("missing_input: monster_name or markdown required"), nil
		}

		var md string
		switch {
		case markdownProvided:
			md = markdown
		case monsterName != "":
			if campaign == "" {
				return mcp.NewToolResultError("missing_input: campaign is required when monster_name is provided"), nil
			}
			block, err := h.auditor.FindMonsterMarkdown(campaign, monsterName)
			if err != nil {
				if errors.Is(err, monster.ErrNotFound) {
					return mcp.NewToolResultError(fmt.Sprintf("not_found: monster %q not found in campaign %q", monsterName, campaign)), nil
				}
				return mcp.NewToolResultError(err.Error()), nil
			}
			md = block
		}

		m, perr := parser.ParseStatBlock(md)
		if perr != nil {
			var pe *parser.ParseError
			if errors.As(perr, &pe) {
				return mcp.NewToolResultError(fmt.Sprintf("parse_error: %s (line %d)", pe.Msg, pe.Line)), nil
			}
			return mcp.NewToolResultError(fmt.Sprintf("parse_error: %s", perr.Error())), nil
		}

		result := h.validator.Validate(m)
		return jsonResult(result)
	}
}

// HandleSuggestMonsterCR handles the suggest_monster_cr tool.
//
// Inputs: target_cr (required, 0..30), concept (optional),
// output ("markdown" | "json", default "markdown").
func (h *MonsterValidationHandlers) HandleSuggestMonsterCR() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		targetCR := getFloat64Arg(args, "target_cr")
		concept := getStringArg(args, "concept")
		output := getStringArg(args, "output")
		if output == "" {
			output = "markdown"
		}

		m, err := h.suggester.Suggest(targetCR, concept)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("unsupported_cr: %s", err.Error())), nil
		}

		switch output {
		case "json":
			bytes, jerr := json.MarshalIndent(m, "", "  ")
			if jerr != nil {
				return mcp.NewToolResultError(fmt.Sprintf("marshal: %s", jerr)), nil
			}
			return mcp.NewToolResultText(string(bytes)), nil
		case "markdown":
			fallthrough
		default:
			rendered, rerr := renderer.RenderStatBlock(m)
			if rerr != nil {
				return mcp.NewToolResultError(rerr.Error()), nil
			}
			return mcp.NewToolResultText(rendered), nil
		}
	}
}

// HandleAuditMonsterCR handles the audit_monster_cr tool.
//
// Input: campaign (required). Output: audit report JSON with
// "campaign" key (alias of AuditReport.CampaignID).
func (h *MonsterValidationHandlers) HandleAuditMonsterCR() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		campaign := getStringArg(args, "campaign")
		if campaign == "" {
			return mcp.NewToolResultError("missing_input: campaign is required"), nil
		}

		// Pre-flight: confirm the campaign dir exists so we can return a
		// clean not_found error before delegating to the auditor.
		if _, err := os.Stat(filepath.Join(h.baseDir, campaign)); err != nil {
			if os.IsNotExist(err) {
				return mcp.NewToolResultError(fmt.Sprintf("not_found: campaign %q not found", campaign)), nil
			}
			return mcp.NewToolResultError(err.Error()), nil
		}

		report, err := h.auditor.AuditCampaign(campaign)
		if err != nil {
			if errors.Is(err, monster.ErrNotFound) || strings.Contains(strings.ToLower(err.Error()), "not found") {
				return mcp.NewToolResultError(fmt.Sprintf("not_found: %s", err.Error())), nil
			}
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Map AuditReport to the spec's wire shape: rename
		// "campaign_id" to "campaign" without changing the
		// internal struct.
		wire := map[string]any{
			"campaign": report.CampaignID,
			"monsters": report.Monsters,
			"summary":  report.Summary,
		}
		return jsonResult(wire)
	}
}
