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

// VizHandlers handles the relationship graph visualization tool.
type VizHandlers struct {
	canonService *services.CanonService
}

// NewVizHandlers creates a new VizHandlers.
func NewVizHandlers(canonService *services.CanonService) *VizHandlers {
	return &VizHandlers{canonService: canonService}
}

// HandleGenerateRelationshipGraph returns a tool handler that renders a D3.js force-directed graph.
func (h *VizHandlers) HandleGenerateRelationshipGraph() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		if campaignID == "" {
			return mcp.NewToolResultError("campaign_id is required"), nil
		}

		graph, err := h.canonService.GetRelationshipGraph(ctx, campaignID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if len(graph.Nodes) == 0 {
			return mcp.NewToolResultText("No entities found"), nil
		}

		html := renderRelationshipGraph(graph)
		return mcp.NewToolResultText(html), nil
	}
}

// entityColor returns the HTML color for an entity type.
func entityColor(entityType domain.EntityType) string {
	switch entityType {
	case domain.EntityTypeNPC:
		return "#4A90D9" // blue
	case domain.EntityTypeLocation:
		return "#27AE60" // green
	case domain.EntityTypeFaction:
		return "#E67E22" // orange
	case domain.EntityTypeItem:
		return "#8E44AD" // purple
	case domain.EntityTypeMonster:
		return "#C0392B" // red
	default:
		return "#95A5A6" // grey
	}
}

// edgeColor returns the HTML color for a relationship type.
func edgeColor(relType domain.RelationshipType) string {
	switch relType {
	case domain.RelationshipTypeAlly:
		return "#27AE60" // green
	case domain.RelationshipTypeEnemy:
		return "#E74C3C" // red
	case domain.RelationshipTypeRival:
		return "#F39C12" // orange
	case domain.RelationshipTypeIndebted:
		return "#95A5A6" // grey
	case domain.RelationshipTypeBloodOath:
		return "#E91E63" // pink
	default:
		return "#95A5A6"
	}
}

// renderRelationshipGraph produces an HTML page with either D3.js or static SVG.
func renderRelationshipGraph(graph *domain.RelationshipGraph) string {
	if len(graph.Nodes) > 50 {
		return renderStaticSVG(graph)
	}
	return renderD3Graph(graph)
}

// renderD3Graph produces an interactive D3.js force-directed graph.
func renderD3Graph(graph *domain.RelationshipGraph) string {
	var nodesJSON, edgesJSON strings.Builder
	nodesJSON.WriteString("[")
	for i, n := range graph.Nodes {
		if i > 0 {
			nodesJSON.WriteString(",")
		}
		fmt.Fprintf(&nodesJSON, `{"id":"%s","name":"%s","type":"%s"}`, n.ID, n.Name, n.Type)
	}
	nodesJSON.WriteString("]")

	edgesJSON.WriteString("[")
	for i, e := range graph.Edges {
		if i > 0 {
			edgesJSON.WriteString(",")
		}
		fmt.Fprintf(&edgesJSON, `{"source":"%s","target":"%s","type":"%s","strength":%d}`,
			e.From, e.To, e.Type, e.Strength)
	}
	edgesJSON.WriteString("]")

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="utf-8"><title>Relationship Graph</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#1a1a2e;color:#eee}
svg{width:100%%;height:100vh}
.node circle{stroke:#fff;stroke-width:1.5px;cursor:pointer}
.node text{font-size:10px;fill:#eee;pointer-events:none}
.link{stroke-opacity:0.6}
.tooltip{position:absolute;background:rgba(0,0,0,0.85);color:#fff;padding:8px 12px;border-radius:4px;font-size:12px;pointer-events:none;display:none;z-index:100}
.legend{position:absolute;top:10px;right:10px;background:rgba(0,0,0,0.7);padding:10px;border-radius:4px;font-size:11px;z-index:50}
.legend-item{display:flex;align-items:center;margin:3px 0}
.legend-color{width:12px;height:12px;border-radius:50%%;margin-right:6px}
</style>
</head>
<body>
<div id="tooltip" class="tooltip"></div>
<div class="legend">
<div style="margin-bottom:4px;font-weight:bold">Nodes</div>
<div class="legend-item"><div class="legend-color" style="background:#4A90D9"></div>NPC</div>
<div class="legend-item"><div class="legend-color" style="background:#27AE60"></div>Location</div>
<div class="legend-item"><div class="legend-color" style="background:#E67E22"></div>Faction</div>
<div class="legend-item"><div class="legend-color" style="background:#8E44AD"></div>Item</div>
<div class="legend-item"><div class="legend-color" style="background:#C0392B"></div>Monster</div>
<div style="margin:6px 0 4px;font-weight:bold">Edges</div>
<div class="legend-item"><div class="legend-color" style="background:#27AE60"></div>Ally</div>
<div class="legend-item"><div class="legend-color" style="background:#E74C3C"></div>Enemy</div>
<div class="legend-item"><div class="legend-color" style="background:#F39C12"></div>Rival</div>
<div class="legend-item"><div class="legend-color" style="background:#95A5A6"></div>Indebted</div>
<div class="legend-item"><div class="legend-color" style="background:#E91E63"></div>Blood Oath</div>
</div>
<script src="https://d3js.org/d3.v7.min.js"></script>
<script>
const nodes = %s;
const edges = %s;

const width = window.innerWidth;
const height = window.innerHeight;

const svg = d3.select("body").append("svg")
    .attr("width", width)
    .attr("height", height);

const tooltip = d3.select("#tooltip");

const simulation = d3.forceSimulation(nodes)
    .force("link", d3.forceLink(edges).id(d => d.id).distance(100))
    .force("charge", d3.forceManyBody().strength(-300))
    .force("center", d3.forceCenter(width/2, height/2))
    .force("collision", d3.forceCollide().radius(30));

const link = svg.append("g")
    .selectAll("line")
    .data(edges)
    .join("line")
    .attr("class", "link")
    .attr("stroke", d => {
        const colors = {ally:"#27AE60",enemy:"#E74C3C",rival:"#F39C12",indebted:"#95A5A6",blood_oath:"#E91E63"};
        return colors[d.type]||"%s";
    })
    .attr("stroke-width", d => Math.max(1, Math.abs(d.strength)));

const node = svg.append("g")
    .selectAll("g")
    .data(nodes)
    .join("g")
    .attr("class", "node")
    .call(d3.drag()
        .on("start", (event, d) => { if(!event.active) simulation.alphaTarget(0.3).restart(); d.fx = d.x; d.fy = d.y; })
        .on("drag", (event, d) => { d.fx = event.x; d.fy = event.y; })
        .on("end", (event, d) => { if(!event.active) simulation.alphaTarget(0); d.fx = null; d.fy = null; })
    );

node.append("circle")
    .attr("r", 8)
    .attr("fill", d => {
        const colors = {npc:"#4A90D9",location:"#27AE60",faction:"#E67E22",item:"#8E44AD",monster:"#C0392B"};
        return colors[d.type]||"%s";
    });

node.append("text")
    .attr("dx", 12)
    .attr("dy", 4)
    .text(d => d.name);

node.on("mouseover", (event, d) => {
    tooltip.style("display","block")
        .html("<strong>"+d.name+"</strong><br>Type: "+d.type)
        .style("left",(event.pageX+10)+"px")
        .style("top",(event.pageY-10)+"px");
})
.on("mouseout", () => tooltip.style("display","none"));

link.on("mouseover", (event, d) => {
    tooltip.style("display","block")
        .html("<strong>"+d.type+"</strong><br>Strength: "+d.strength)
        .style("left",(event.pageX+10)+"px")
        .style("top",(event.pageY-10)+"px");
})
.on("mouseout", () => tooltip.style("display","none"));

simulation.on("tick", () => {
    link.attr("x1", d => d.source.x).attr("y1", d => d.source.y)
        .attr("x2", d => d.target.x).attr("y2", d => d.target.y);
    node.attr("transform", d => "translate("+d.x+","+d.y+")");
});
</script>
</body>
</html>`,
		nodesJSON.String(), edgesJSON.String(), "#95A5A6", "#95A5A6")
}

// renderStaticSVG produces a static SVG with a legend (fallback for 50+ nodes).
func renderStaticSVG(graph *domain.RelationshipGraph) string {
	// Group nodes by type
	typeCount := make(map[domain.EntityType]int)
	for _, n := range graph.Nodes {
		typeCount[n.Type]++
	}
	var stats strings.Builder
	typeNames := []domain.EntityType{
		domain.EntityTypeNPC, domain.EntityTypeLocation,
		domain.EntityTypeFaction, domain.EntityTypeItem, domain.EntityTypeMonster,
	}
	for _, t := range typeNames {
		if count := typeCount[t]; count > 0 {
			fmt.Fprintf(&stats, "<div style=\"color:%s\">%s: %d</div>", entityColor(t), t, count)
		}
	}

	// Count relationship types
	relCount := make(map[domain.RelationshipType]int)
	for _, e := range graph.Edges {
		relCount[e.Type]++
	}
	var relStats strings.Builder
	relTypes := []domain.RelationshipType{
		domain.RelationshipTypeAlly, domain.RelationshipTypeEnemy,
		domain.RelationshipTypeRival, domain.RelationshipTypeIndebted,
		domain.RelationshipTypeBloodOath,
	}
	for _, t := range relTypes {
		if count := relCount[t]; count > 0 {
			fmt.Fprintf(&relStats, "<div style=\"color:%s\">%s: %d</div>", edgeColor(t), t, count)
		}
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="utf-8"><title>Relationship Graph (Static)</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#1a1a2e;color:#eee;padding:20px}
h1{font-size:18px;margin-bottom:10px}
.container{display:flex;gap:40px}
.column{background:rgba(255,255,255,0.05);padding:16px;border-radius:6px;min-width:200px}
.column h2{font-size:14px;margin-bottom:8px;color:#aaa}
</style>
</head>
<body>
<h1>Relationship Graph — %d entities, %d relationships (static view — too many nodes for interactive)</h1>
<div class="container">
<div class="column"><h2>Entities by Type</h2>%s</div>
<div class="column"><h2>Relationships by Type</h2>%s</div>
</div>
</body>
</html>`,
		len(graph.Nodes), len(graph.Edges),
		stats.String(), relStats.String())
}
