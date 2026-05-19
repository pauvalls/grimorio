package handlers

import (
	"context"
	"fmt"
	"html"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/services"
)

// TimelineHandlers handles the session timeline tool.
type TimelineHandlers struct {
	narrativeStateService *services.NarrativeStateService
}

// NewTimelineHandlers creates a new TimelineHandlers.
func NewTimelineHandlers(narrativeStateService *services.NarrativeStateService) *TimelineHandlers {
	return &TimelineHandlers{narrativeStateService: narrativeStateService}
}

// HandleSessionTimeline returns a tool handler that renders a session timeline.
func (h *TimelineHandlers) HandleSessionTimeline() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		if campaignID == "" {
			return mcp.NewToolResultError("campaign_id is required"), nil
		}

		state, err := h.narrativeStateService.Load(ctx, campaignID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if len(state.SessionLog) == 0 {
			return mcp.NewToolResultText("No session data for campaign"), nil
		}

		html := renderSessionTimeline(state.SessionLog)
		return mcp.NewToolResultText(html), nil
	}
}

// renderSessionTimeline produces HTML for a vertical session timeline.
func renderSessionTimeline(sessions []domain.SessionRecord) string {
	var entries strings.Builder

	for _, s := range sessions {
		dateStr := s.Date.Format("2006-01-02")
		if s.Date.IsZero() {
			dateStr = "TBD"
		}

		summary := html.EscapeString(s.Summary)
		if len(summary) > 120 {
			summary = summary[:120] + "…"
		}
		fullSummary := html.EscapeString(s.Summary)
		if fullSummary == "" {
			fullSummary = "No summary recorded"
		}

		decisionsHTML := renderDecisionsBlock(s.KeyDecisions)

		fmt.Fprintf(&entries, `
	<div class="timeline-entry">
		<div class="timeline-dot"></div>
		<div class="timeline-card">
			<div class="session-header">
				<span class="session-num">Session %d</span>
				<span class="session-date">%s</span>
			</div>
			<div class="session-summary" title="%s">%s</div>
			<div class="session-meta">
				<span class="meta-item">XP: %d</span>
				<span class="meta-item">Decisions: %d</span>
			</div>
			%s
		</div>
	</div>`,
			s.SessionNum, dateStr, fullSummary, summary,
			s.XPAwarded, len(s.KeyDecisions),
			decisionsHTML)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="utf-8"><title>Session Timeline</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#1a1a2e;color:#eee;padding:20px}
h1{font-size:20px;margin-bottom:20px}
.timeline{position:relative;max-width:700px;margin:0 auto;padding-left:30px}
.timeline::before{content:'';position:absolute;left:12px;top:0;bottom:0;width:2px;background:#3a3a5e}
.timeline-entry{position:relative;margin-bottom:20px}
.timeline-dot{position:absolute;left:-18px;top:8px;width:14px;height:14px;border-radius:50%%;background:#4A90D9;border:2px solid #1a1a2e}
.timeline-card{background:rgba(255,255,255,0.05);border-radius:8px;padding:14px 16px}
.session-header{display:flex;justify-content:space-between;align-items:center;margin-bottom:6px}
.session-num{font-weight:bold;font-size:14px}
.session-date{color:#aaa;font-size:12px}
.session-summary{font-size:13px;color:#ccc;line-height:1.5;margin-bottom:8px}
.session-meta{display:flex;gap:12px;font-size:11px;color:#888}
.meta-item{background:rgba(255,255,255,0.08);padding:2px 8px;border-radius:4px}
.decisions{margin-top:10px}
.decisions summary{font-size:12px;color:#4A90D9;cursor:pointer;padding:4px 0}
.decisions-list{padding:8px 0 4px 12px}
.decision-item{margin-bottom:4px;font-size:12px;color:#bbb;border-left:2px solid #F39C12;padding-left:8px;line-height:1.4}
.decision-item strong{color:#eee}
</style>
</head>
<body>
<h1>Session Timeline</h1>
<div class="timeline">%s
</div>
</body>
</html>`, entries.String())
}

// renderDecisionsBlock returns HTML for the key decisions expandable section.
func renderDecisionsBlock(decisions []domain.Decision) string {
	if len(decisions) == 0 {
		return ""
	}

	var items strings.Builder
	for _, d := range decisions {
		desc := html.EscapeString(d.Description)
		choice := html.EscapeString(d.ChoiceMade)
		if choice != "" {
			fmt.Fprintf(&items, `<div class="decision-item"><strong>%s</strong> — %s</div>`, desc, choice)
		} else {
			fmt.Fprintf(&items, `<div class="decision-item">%s</div>`, desc)
		}
	}

	return fmt.Sprintf(`<details class="decisions">
	<summary>Key Decisions (%d)</summary>
	<div class="decisions-list">%s</div>
</details>`, len(decisions), items.String())
}
