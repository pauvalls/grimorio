package services

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// FlowchartService generates campaign flowcharts in Mermaid and SVG
type FlowchartService struct {
	canonRepo repository.CanonRepository
	actRepo   repository.ActRepository
}

// NewFlowchartService creates a new flowchart service
func NewFlowchartService(canonRepo repository.CanonRepository, actRepo repository.ActRepository) *FlowchartService {
	return &FlowchartService{canonRepo: canonRepo, actRepo: actRepo}
}

// GenerateMermaid generates Mermaid flowchart syntax for the campaign
func (s *FlowchartService) GenerateMermaid(ctx context.Context, campaignID string, detailLevel string) (string, error) {
	if detailLevel != "overview" && detailLevel != "act" && detailLevel != "decision" {
		return "", fmt.Errorf("invalid detail_level: %s (must be overview, act, or decision)", detailLevel)
	}

	doc, err := s.canonRepo.Load(campaignID)
	if err != nil {
		// If canon doesn't exist, create a minimal one
		doc = &domain.CanonDocument{CampaignID: campaignID}
	}

	nodes := s.buildNodes(ctx, campaignID, doc, detailLevel)
	if len(nodes) == 0 {
		return "", fmt.Errorf("campaign has no acts to chart")
	}

	var sb strings.Builder
	sb.WriteString("flowchart TD\n")

	// Write node definitions
	for _, node := range nodes {
		label := strings.ReplaceAll(node.Label, `"`, `\"`)
		switch node.Type {
		case "act":
			sb.WriteString(fmt.Sprintf("    %s[%s]\n", node.ID, label))
		case "decision":
			sb.WriteString(fmt.Sprintf("    %s{%s}\n", node.ID, label))
		case "event":
			sb.WriteString(fmt.Sprintf("    %s(%s)\n", node.ID, label))
		default:
			sb.WriteString(fmt.Sprintf("    %s[%s]\n", node.ID, label))
		}
	}

	// Write edges
	for _, node := range nodes {
		for _, dep := range node.Dependencies {
			sb.WriteString(fmt.Sprintf("    %s --> %s\n", dep, node.ID))
		}
	}

	return sb.String(), nil
}

// GenerateSVG generates an SVG representation of the flowchart
func (s *FlowchartService) GenerateSVG(ctx context.Context, campaignID string, detailLevel string) (string, error) {
	doc, err := s.canonRepo.Load(campaignID)
	if err != nil {
		// If canon doesn't exist, create a minimal one
		doc = &domain.CanonDocument{CampaignID: campaignID}
	}

	nodes := s.buildNodes(ctx, campaignID, doc, detailLevel)
	if len(nodes) == 0 {
		return "", fmt.Errorf("campaign has no acts to chart")
	}

	return s.renderFlowchartSVG(nodes), nil
}

func (s *FlowchartService) buildNodes(ctx context.Context, campaignID string, doc *domain.CanonDocument, detailLevel string) []domain.FlowchartNode {
	// Build dead NPC set
	deadSet := make(map[string]bool)
	for _, e := range doc.Entities {
		if e.Type == domain.EntityTypeNPC && e.CanonState == domain.EntityStateDead {
			deadSet[e.ID] = true
		}
	}

	var nodes []domain.FlowchartNode
	nodeMap := make(map[string]domain.FlowchartNode)

	// Collect act nodes from canon
	for _, e := range doc.Entities {
		if e.Role == "act" || strings.HasPrefix(e.ID, "act-") {
			node := domain.FlowchartNode{
				ID:    sanitizeNodeID(e.ID),
				Label: e.Name,
				Type:  "act",
			}
			nodeMap[node.ID] = node
		}
	}

	// If no acts in canon, try loading from filesystem acts
	if len(nodeMap) == 0 && s.actRepo != nil {
		acts, err := s.actRepo.List(campaignID)
		if err == nil {
			for _, act := range acts {
				// Extract title from content (first H1 heading)
				title := fmt.Sprintf("Act %d", act.Number)
				lines := strings.Split(act.Content, "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "# ") {
						title = strings.TrimPrefix(line, "# ")
						title = strings.TrimSpace(title)
						break
					}
				}
				nodeID := sanitizeNodeID(fmt.Sprintf("act-%02d", act.Number))
				node := domain.FlowchartNode{
					ID:    nodeID,
					Label: title,
					Type:  "act",
				}
				// Link acts sequentially
				if act.Number > 1 {
					prevID := sanitizeNodeID(fmt.Sprintf("act-%02d", act.Number-1))
					node.Dependencies = append(node.Dependencies, prevID)
				}
				nodeMap[nodeID] = node
			}
		}
	}

	// If overview, only return act nodes with inter-act relationships
	if detailLevel == "overview" {
		for id, node := range nodeMap {
			for _, rel := range doc.Relationships {
				fromID := sanitizeNodeID(rel.From)
				if rel.To == id && nodeMap[fromID].ID != "" {
					node.Dependencies = append(node.Dependencies, fromID)
				}
			}
			nodeMap[id] = node
		}
		for _, n := range nodeMap {
			nodes = append(nodes, n)
		}
		sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
		return nodes
	}

	// For act and decision levels, include NPCs and other entities
	for _, e := range doc.Entities {
		if deadSet[e.ID] {
			continue
		}
		if e.Role == "act" || strings.HasPrefix(e.ID, "act-") {
			continue // already handled
		}

		nodeType := "event"
		if e.Type == domain.EntityTypeNPC {
			nodeType = "event"
		}

		node := domain.FlowchartNode{
			ID:    sanitizeNodeID(e.ID),
			Label: e.Name,
			Type:  nodeType,
		}
		nodeMap[node.ID] = node
	}

	// Build edges from relationships
	for id, node := range nodeMap {
		for _, rel := range doc.Relationships {
			fromID := sanitizeNodeID(rel.From)
			toID := sanitizeNodeID(rel.To)
			if toID == id && nodeMap[fromID].ID != "" {
				node.Dependencies = append(node.Dependencies, fromID)
			}
		}
		nodeMap[id] = node
	}

	// For decision level, add decision nodes from timeline events
	if detailLevel == "decision" {
		for _, ev := range doc.Timeline {
			if ev.IsRevealed {
				nodeID := sanitizeNodeID("event-" + ev.ID)
				node := domain.FlowchartNode{
					ID:    nodeID,
					Label: ev.Description,
					Type:  "event",
				}
				// Connect to involved entities
				for _, inv := range ev.Involved {
					invID := sanitizeNodeID(inv)
					if nodeMap[invID].ID != "" {
						node.Dependencies = append(node.Dependencies, invID)
					}
				}
				nodeMap[nodeID] = node
			}
		}
	}

	for _, n := range nodeMap {
		nodes = append(nodes, n)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	return nodes
}

func (s *FlowchartService) renderFlowchartSVG(nodes []domain.FlowchartNode) string {
	const nodeWidth = 140
	const nodeHeight = 50
	const hSpacing = 60
	const vSpacing = 80
	const margin = 40

	cols := 3
	if len(nodes) < cols {
		cols = len(nodes)
	}
	rows := (len(nodes) + cols - 1) / cols

	width := margin*2 + cols*nodeWidth + (cols-1)*hSpacing
	height := margin*2 + rows*nodeHeight + (rows-1)*vSpacing + 40

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" style="background:#faf8f5;" font-family="Arial,sans-serif" font-size="12">`,
		width, height, width, height))

	// Title
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="25" text-anchor="middle" font-size="16" font-weight="bold" fill="#5a3d2b">Campaign Flowchart</text>`,
		width/2))

	// Calculate positions
	positions := make(map[string]struct{ x, y int })
	for i, node := range nodes {
		col := i % cols
		row := i / cols
		x := margin + col*(nodeWidth+hSpacing)
		y := margin + 20 + row*(nodeHeight+vSpacing)
		positions[node.ID] = struct{ x, y int }{x, y}
	}

	// Draw edges first
	for _, node := range nodes {
		pos, ok := positions[node.ID]
		if !ok {
			continue
		}
		for _, dep := range node.Dependencies {
			depPos, ok := positions[dep]
			if !ok {
				continue
			}
			// Draw line from dep to node
			sb.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#8b7355" stroke-width="1.5" marker-end="url(#arrowhead)"/>`,
				depPos.x+nodeWidth/2, depPos.y+nodeHeight,
				pos.x+nodeWidth/2, pos.y))
		}
	}

	// Arrowhead marker
	sb.WriteString(`<defs><marker id="arrowhead" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto"><polygon points="0 0, 10 3.5, 0 7" fill="#8b7355"/></marker></defs>`)

	// Draw nodes
	for _, node := range nodes {
		pos, ok := positions[node.ID]
		if !ok {
			continue
		}

		var fill, stroke string
		switch node.Type {
		case "act":
			fill = "#c9ad6a"
			stroke = "#5a3d2b"
		case "decision":
			fill = "#e8d5b5"
			stroke = "#8b6914"
		default:
			fill = "#f5f0e6"
			stroke = "#8b7355"
		}

		sb.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="4" fill="%s" stroke="%s" stroke-width="1.5"/>`,
			pos.x, pos.y, nodeWidth, nodeHeight, fill, stroke))

		label := node.Label
		if len(label) > 20 {
			label = label[:17] + "..."
		}
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" dominant-baseline="central" fill="#2c1e14">%s</text>`,
			pos.x+nodeWidth/2, pos.y+nodeHeight/2, label))
	}

	sb.WriteString(`</svg>`)
	return sb.String()
}

func sanitizeNodeID(id string) string {
	// Mermaid IDs must be alphanumeric with limited special chars
	result := strings.ReplaceAll(id, " ", "_")
	result = strings.ReplaceAll(result, ".", "_")
	return result
}
