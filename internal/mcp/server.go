package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/paupena/grimorio/internal/compiler"
	"github.com/paupena/grimorio/internal/config"
	"github.com/paupena/grimorio/internal/image"
	"github.com/paupena/grimorio/internal/svg"
)

func NewServer(cfg *config.Config) *server.MCPServer {
	s := server.NewMCPServer(
		"grimorio",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)

	// Tool: create_campaign
	s.AddTool(mcp.NewTool("create_campaign",
		mcp.WithDescription("Create a new campaign directory structure"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Campaign name (kebab-case)"), mcp.Title("Campaign Name")),
		mcp.WithString("title", mcp.Description("Campaign title")),
		mcp.WithString("setting", mcp.Description("Brief setting description")),
	), handleCreateCampaign(cfg))

	// Tool: save_act
	s.AddTool(mcp.NewTool("save_act",
		mcp.WithDescription("Save an act/chapter of the campaign"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("act_number", mcp.Required(), mcp.Description("Act number (1, 2, 3...)"), mcp.Title("Act Number")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Act title")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Full Markdown content of the act")),
	), handleSaveAct(cfg))

	// Tool: save_npcs
	s.AddTool(mcp.NewTool("save_npcs",
		mcp.WithDescription("Save NPCs and factions to the campaign"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Markdown content with NPCs and factions")),
	), handleSaveNPCs(cfg))

	// Tool: save_bestiary
	s.AddTool(mcp.NewTool("save_bestiary",
		mcp.WithDescription("Save monsters and creatures to the bestiary"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Markdown content with monster stat blocks")),
	), handleSaveBestiary(cfg))

	// Tool: save_encounters
	s.AddTool(mcp.NewTool("save_encounters",
		mcp.WithDescription("Save combat encounters and challenges"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Markdown content with encounters")),
	), handleSaveEncounters(cfg))

	// Tool: save_maps
	s.AddTool(mcp.NewTool("save_maps",
		mcp.WithDescription("Save map descriptions and scene layouts"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Markdown content with maps and scenes")),
	), handleSaveMaps(cfg))

	// Tool: compile_pdf
	s.AddTool(mcp.NewTool("compile_pdf",
		mcp.WithDescription("Compile all campaign markdown files into a styled PDF"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("title", mcp.Description("PDF title (defaults to campaign name)")),
	), handleCompilePDF(cfg))

	// Tool: get_template
	s.AddTool(mcp.NewTool("get_template",
		mcp.WithDescription("Get the Markdown/CSS template for a specific section type"),
		mcp.WithString("type", mcp.Required(), mcp.Description("Template type: act, npc, monster, encounter, map, lore")),
	), handleGetTemplate())

	// Tool: generate_map (SVG procedural)
	s.AddTool(mcp.NewTool("generate_map",
		mcp.WithDescription("Generate a procedural SVG battle map. 100% local, no API key required"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("filename", mcp.Required(), mcp.Description("Output filename (without extension)")),
		mcp.WithString("style", mcp.Description("Map style: dungeon, landscape, city"), mcp.DefaultString("dungeon")),
		mcp.WithString("title", mcp.Description("Map title displayed on the image"), mcp.DefaultString("")),
		mcp.WithNumber("rooms", mcp.Description("Number of rooms/areas (2-10)"), mcp.DefaultNumber(6)),
		mcp.WithString("labels", mcp.Description("Comma-separated room labels (e.g. 'Entrance,Tavern,Boss')")),
	), handleGenerateMap(cfg))

	// Tool: generate_divider (SVG procedural)
	s.AddTool(mcp.NewTool("generate_divider",
		mcp.WithDescription("Generate a decorative SVG divider/separator for sections"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("filename", mcp.Required(), mcp.Description("Output filename (without extension)")),
		mcp.WithString("style", mcp.Description("Divider style: ornate, simple, double"), mcp.DefaultString("ornate")),
		mcp.WithNumber("width", mcp.Description("Width in pixels"), mcp.DefaultNumber(600)),
	), handleGenerateDivider(cfg))

	// Tool: generate_image (AI image generation — free by default via Pollinations.ai)
	s.AddTool(mcp.NewTool("generate_image",
		mcp.WithDescription("Generate an image using AI. Free by default via Pollinations.ai (no API key required). Optional: set image_provider to 'dalle' and OPENAI_API_KEY for higher quality. Use for cover art, NPC portraits, monster illustrations"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("filename", mcp.Required(), mcp.Description("Output filename (without extension)")),
		mcp.WithString("prompt", mcp.Required(), mcp.Description("Detailed image generation prompt")),
		mcp.WithString("type", mcp.Description("Image type: cover, portrait, illustration, scene"), mcp.DefaultString("illustration")),
	), handleGenerateImage(cfg))

	return s
}

func campaignDir(cfg *config.Config, name string) string {
	return filepath.Join(cfg.OutputDir, name)
}

func ensureSubdir(cfg *config.Config, campaign, subdir string) error {
	dir := filepath.Join(campaignDir(cfg, campaign), subdir)
	return os.MkdirAll(dir, 0755)
}

func getArg(args map[string]any, key string) string {
	if val, ok := args[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

func handleCreateCampaign(cfg *config.Config) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		name := getArg(args, "name")
		title := getArg(args, "title")
		setting := getArg(args, "setting")

		if name == "" {
			return mcp.NewToolResultError("campaign name is required"), nil
		}

		dir := campaignDir(cfg, name)
		subdirs := []string{"acts", "npcs", "bestiary", "encounters", "maps", "assets"}
		for _, sub := range subdirs {
			if err := os.MkdirAll(filepath.Join(dir, sub), 0755); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to create directory: %v", err)), nil
			}
		}

		manifest := fmt.Sprintf("# %s\n\n**Setting:** %s\n\nGenerated by Campaign AI\n", title, setting)
		if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(manifest), 0644); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to write manifest: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Campaign '%s' created at %s", name, dir)), nil
	}
}

func handleSaveAct(cfg *config.Config) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		campaign := getArg(args, "campaign")
		actNum := getArg(args, "act_number")
		title := getArg(args, "title")
		content := getArg(args, "content")

		if err := ensureSubdir(cfg, campaign, "acts"); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create acts directory: %v", err)), nil
		}
		dir := campaignDir(cfg, campaign)
		filename := filepath.Join(dir, "acts", fmt.Sprintf("act_%s_%s.md", actNum, sanitize(title)))
		
		header := fmt.Sprintf("# Acto %s: %s\n\n", actNum, title)
		if err := os.WriteFile(filename, []byte(header+content), 0644); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to save act: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Act %s saved to %s", actNum, filename)), nil
	}
}

func handleSaveNPCs(cfg *config.Config) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		campaign := getArg(args, "campaign")
		content := getArg(args, "content")

		if err := ensureSubdir(cfg, campaign, "npcs"); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create npcs directory: %v", err)), nil
		}
		dir := campaignDir(cfg, campaign)
		filename := filepath.Join(dir, "npcs", "npcs_and_factions.md")
		
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to save NPCs: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("NPCs saved to %s", filename)), nil
	}
}

func handleSaveBestiary(cfg *config.Config) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		campaign := getArg(args, "campaign")
		content := getArg(args, "content")

		if err := ensureSubdir(cfg, campaign, "bestiary"); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create bestiary directory: %v", err)), nil
		}
		dir := campaignDir(cfg, campaign)
		filename := filepath.Join(dir, "bestiary", "bestiary.md")
		
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to save bestiary: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Bestiary saved to %s", filename)), nil
	}
}

func handleSaveEncounters(cfg *config.Config) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		campaign := getArg(args, "campaign")
		content := getArg(args, "content")

		if err := ensureSubdir(cfg, campaign, "encounters"); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create encounters directory: %v", err)), nil
		}
		dir := campaignDir(cfg, campaign)
		filename := filepath.Join(dir, "encounters", "encounters.md")
		
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to save encounters: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Encounters saved to %s", filename)), nil
	}
}

func handleSaveMaps(cfg *config.Config) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		campaign := getArg(args, "campaign")
		content := getArg(args, "content")

		if err := ensureSubdir(cfg, campaign, "maps"); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create maps directory: %v", err)), nil
		}
		dir := campaignDir(cfg, campaign)
		filename := filepath.Join(dir, "maps", "maps_and_scenes.md")
		
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to save maps: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Maps saved to %s", filename)), nil
	}
}

func handleCompilePDF(cfg *config.Config) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		campaign := getArg(args, "campaign")
		title := getArg(args, "title")
		if title == "" {
			title = campaign
		}

		dir := campaignDir(cfg, campaign)
		comp := compiler.New(dir, cfg.PDFEngine)
		pdfPath, err := comp.Compile(title)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to compile PDF: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("PDF compiled: %s", pdfPath)), nil
	}
}

func handleGetTemplate() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		tmplType := getArg(args, "type")
		
		template, err := compiler.GetTemplate(tmplType)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(template), nil
	}
}

func sanitize(s string) string {
	result := []rune(s)
	for i, r := range result {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			result[i] = '_'
		}
	}
	return string(result)
}

func handleGenerateMap(cfg *config.Config) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		campaign := getArg(args, "campaign")
		filename := getArg(args, "filename")
		style := getArg(args, "style")
		title := getArg(args, "title")
		labels := getArg(args, "labels")

		if campaign == "" || filename == "" {
			return mcp.NewToolResultError("campaign and filename are required"), nil
		}

		svgCfg := svg.DefaultBattleMapConfig()
		if style != "" {
			svgCfg.Style = svg.MapStyle(style)
		}
		if title != "" {
			svgCfg.Title = title
		}

		numRooms := 6
		if v, ok := args["rooms"].(float64); ok {
			numRooms = int(v)
		}
		svgCfg.NumRooms = numRooms

		dir := campaignDir(cfg, campaign)
		assetsDir := filepath.Join(dir, "assets")
		if err := os.MkdirAll(assetsDir, 0755); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create assets directory: %v", err)), nil
		}

		if labels != "" {
			svgCfg.Labels = splitLabels(labels)
		}

		svgContent := svg.GenerateBattleMap(svgCfg)
		outputPath := filepath.Join(assetsDir, filename+".svg")
		if err := os.WriteFile(outputPath, []byte(svgContent), 0644); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to save map: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Map generated: %s (%s style, %d rooms)", outputPath, style, numRooms)), nil
	}
}

func splitLabels(labels string) []string {
	var result []string
	for _, l := range splitCSV(labels) {
		if l != "" {
			result = append(result, l)
		}
	}
	return result
}

func splitCSV(s string) []string {
	var result []string
	current := ""
	for _, r := range s {
		if r == ',' {
			result = append(result, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func handleGenerateDivider(cfg *config.Config) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		campaign := getArg(args, "campaign")
		filename := getArg(args, "filename")
		style := getArg(args, "style")

		if campaign == "" || filename == "" {
			return mcp.NewToolResultError("campaign and filename are required"), nil
		}

		width := 600
		if v, ok := args["width"].(float64); ok {
			width = int(v)
		}
		if style == "" {
			style = "ornate"
		}

		dir := campaignDir(cfg, campaign)
		assetsDir := filepath.Join(dir, "assets")
		if err := os.MkdirAll(assetsDir, 0755); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create assets directory: %v", err)), nil
		}

		svgContent := svg.GenerateDivider(width, style)
		outputPath := filepath.Join(assetsDir, filename+".svg")
		if err := os.WriteFile(outputPath, []byte(svgContent), 0644); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to save divider: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Divider generated: %s (%s style)", outputPath, style)), nil
	}
}

func handleGenerateImage(cfg *config.Config) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		campaign := getArg(args, "campaign")
		filename := getArg(args, "filename")
		prompt := getArg(args, "prompt")
		imgType := getArg(args, "type")

		if campaign == "" || filename == "" || prompt == "" {
			return mcp.NewToolResultError("campaign, filename, and prompt are required"), nil
		}

		provider, err := image.NewProvider(cfg.Config)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to initialize image provider: %v", err)), nil
		}

		dir := campaignDir(cfg, campaign)
		assetsDir := filepath.Join(dir, "assets")
		if err := os.MkdirAll(assetsDir, 0755); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create assets directory: %v", err)), nil
		}

		outputPath := filepath.Join(assetsDir, filename+".png")
		if err := image.GenerateAndSave(provider, prompt, outputPath); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("%s generation failed: %v", provider.Name(), err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Image generated via %s: %s (type: %s)", provider.Name(), outputPath, imgType)), nil
	}
}
