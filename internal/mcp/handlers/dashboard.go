package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/services"
)

// DashboardHandlers handles the faction reputation dashboard tool.
type DashboardHandlers struct {
	factionService *services.FactionService
	canonService   *services.CanonService
}

// NewDashboardHandlers creates a new DashboardHandlers.
func NewDashboardHandlers(factionService *services.FactionService, canonService *services.CanonService) *DashboardHandlers {
	return &DashboardHandlers{
		factionService: factionService,
		canonService:   canonService,
	}
}

// HandleFactionDashboard returns a tool handler that renders a faction reputation dashboard.
func (h *DashboardHandlers) HandleFactionDashboard() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		if campaignID == "" {
			return mcp.NewToolResultError("campaign_id is required"), nil
		}

		matrix, err := h.factionService.GetPlayerReputationMatrix(ctx, campaignID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if len(matrix.Entries) == 0 {
			return mcp.NewToolResultText("No faction data"), nil
		}

		// Get faction names from canon
		factionNames := h.getFactionNames(ctx, campaignID)

		html := renderFactionDashboard(matrix, factionNames)
		return mcp.NewToolResultText(html), nil
	}
}

// getFactionNames loads canon entities and builds a factionID → name map.
func (h *DashboardHandlers) getFactionNames(ctx context.Context, campaignID string) map[string]string {
	result := make(map[string]string)
	if h.canonService == nil {
		return result
	}
	graph, err := h.canonService.GetRelationshipGraph(ctx, campaignID)
	if err != nil {
		return result
	}
	for _, n := range graph.Nodes {
		if n.Type == domain.EntityTypeFaction {
			result[n.ID] = n.Name
		}
	}
	return result
}

// renderFactionDashboard produces an HTML dashboard with faction reputations.
func renderFactionDashboard(matrix *domain.FactionReputationMatrix, factionNames map[string]string) string {
	var rows strings.Builder

	for _, entry := range matrix.Entries {
		name := factionNames[entry.FactionID]
		if name == "" {
			name = entry.FactionID
		}
		status := repStatusLabel(entry.Score)
		color := repScoreColor(entry.Score)
		textColor := textColorForRepBg(color)

		sparkline := renderSparkline(entry.History)

		fmt.Fprintf(&rows, `<tr>
	<td><strong>%s</strong></td>
	<td style="text-align:right">%d</td>
	<td><span style="display:inline-block;padding:2px 10px;border-radius:10px;background:%s;color:%s;font-weight:bold;font-size:12px">%s</span></td>
	<td style="text-align:center">%s</td>
</tr>
`, name, entry.Score, color, textColor, status, sparkline)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<head><meta charset="utf-8"><title>Faction Reputation Dashboard</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#1a1a2e;color:#eee;padding:20px}
h1{font-size:20px;margin-bottom:16px}
table{width:100%%;border-collapse:collapse}
th,td{padding:10px 12px;text-align:left;border-bottom:1px solid rgba(255,255,255,0.1)}
th{color:#aaa;font-size:12px;text-transform:uppercase;letter-spacing:0.5px}
tr:hover{background:rgba(255,255,255,0.03)}
</style>
</head>
<body>
<h1>Faction Reputation Dashboard</h1>
<table>
<thead><tr><th>Faction</th><th>Score</th><th>Status</th><th>Trend</th></tr></thead>
<tbody>%s</tbody>
</table>
</body>
</html>`, rows.String())
}

// repScoreColor returns the HTML color for a reputation score.
func repScoreColor(score int8) string {
	switch {
	case score >= 50:
		return "#FFD700" // gold — allied
	case score >= 20:
		return "#27AE60" // green — friendly
	case score >= -19:
		return "#95A5A6" // grey — neutral
	case score >= -49:
		return "#F39C12" // orange — unfriendly
	default:
		return "#E74C3C" // red — hostile
	}
}

// repStatusLabel returns a human-readable label for a reputation score.
func repStatusLabel(score int8) string {
	switch {
	case score >= 50:
		return "Allied"
	case score >= 20:
		return "Friendly"
	case score >= -19:
		return "Neutral"
	case score >= -49:
		return "Unfriendly"
	default:
		return "Hostile"
	}
}

// textColorForRepBg returns black or white depending on background brightness.
func textColorForRepBg(bg string) string {
	switch bg {
	case "#FFD700", "#F39C12":
		return "#000"
	default:
		return "#fff"
	}
}

// renderSparkline renders an inline SVG bar chart showing recent reputation changes.
func renderSparkline(history []domain.ReputationEvent) string {
	if len(history) == 0 {
		return `<svg width="60" height="20" viewBox="0 0 60 20"><rect x="0" y="9" width="8" height="2" fill="#555" rx="1"/></svg>`
	}

	// Show last 6 events as bars
	maxEvents := 6
	start := 0
	if len(history) > maxEvents {
		start = len(history) - maxEvents
	}
	recent := history[start:]

	// Find max absolute value for scaling
	maxAbs := int8(1)
	for _, e := range recent {
		delta := e.Delta
		if delta < 0 {
			delta = -delta
		}
		if delta > maxAbs {
			maxAbs = delta
		}
	}

	var bars strings.Builder
	barWidth := 8
	gap := 2
	totalWidth := len(recent)*barWidth + (len(recent)-1)*gap
	if totalWidth < 20 {
		totalWidth = 20
	}
	svgHeight := 24
	midY := svgHeight / 2

	for i, e := range recent {
		x := i * (barWidth + gap)
		barH := int(e.Delta) * (midY - 2) / int(maxAbs)
		if barH < 0 {
			barH = -barH
		}
		if barH < 1 {
			barH = 1
		}
		if barH > midY-2 {
			barH = midY - 2
		}

		var barColor string
		if e.Delta >= 0 {
			barColor = "#27AE60" // green for positive
			y := midY - barH
			fmt.Fprintf(&bars, `<rect x="%d" y="%d" width="%d" height="%d" fill="%s" rx="1"/>`, x, y, barWidth, barH, barColor)
		} else {
			barColor = "#E74C3C" // red for negative
			fmt.Fprintf(&bars, `<rect x="%d" y="%d" width="%d" height="%d" fill="%s" rx="1"/>`, x, midY, barWidth, barH, barColor)
		}
	}

	// Draw baseline
	fmt.Fprintf(&bars, `<line x1="0" y1="%d" x2="%d" y2="%d" stroke="#444" stroke-width="1"/>`, midY, totalWidth, midY)

	return fmt.Sprintf(`<svg width="%d" height="%d" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">%s</svg>`,
		totalWidth, svgHeight, totalWidth, svgHeight, bars.String())
}
