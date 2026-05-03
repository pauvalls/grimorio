package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/config"
	"github.com/pauvalls/grimorio/internal/mcp/handlers"
	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/services"
)

// NewServer creates a new MCP server with all tools wired
func NewServer(cfg *config.Config) *server.MCPServer {
	s := server.NewMCPServer(
		"grimorio",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)

	// Initialize repositories
	campaignRepo := repository.NewFilesystemCampaignRepository(cfg.OutputDir)
	actRepo := repository.NewFilesystemActRepository(cfg.OutputDir)
	charRepo := repository.NewFilesystemCharacterRepository(cfg.OutputDir)
	npcRepo := repository.NewFilesystemNPCRepository(cfg.OutputDir)
	questRepo := repository.NewFilesystemQuestRepository(cfg.OutputDir)

	// Initialize services
	campaignService := services.NewCampaignService(
		campaignRepo, actRepo, charRepo, npcRepo, questRepo,
		cfg.OutputDir, cfg.PDFEngine,
	)
	characterService := services.NewCharacterService(charRepo)
	questService := services.NewQuestService(questRepo)
	assetService := services.NewAssetService(cfg.OutputDir, cfg.Config)

	// Initialize handlers
	campaignHandlers := handlers.NewCampaignHandlers(campaignService)
	characterHandlers := handlers.NewCharacterHandlers(characterService)
	questHandlers := handlers.NewQuestHandlers(questService)
	assetHandlers := handlers.NewAssetHandlers(assetService)

	// Register tools
	// Campaign management
	s.AddTool(mcp.NewTool("create_campaign",
		mcp.WithDescription("Create a new campaign directory structure"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Campaign name (kebab-case)"), mcp.Title("Campaign Name")),
		mcp.WithString("title", mcp.Description("Campaign title")),
		mcp.WithString("setting", mcp.Description("Brief setting description")),
	), campaignHandlers.HandleCreateCampaign())

	s.AddTool(mcp.NewTool("save_act",
		mcp.WithDescription("Save an act/chapter of the campaign"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("act_number", mcp.Required(), mcp.Description("Act number (1, 2, 3...)"), mcp.Title("Act Number")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Act title")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Full Markdown content of the act")),
	), campaignHandlers.HandleSaveAct())

	s.AddTool(mcp.NewTool("save_lore",
		mcp.WithDescription("Save world lore and history for the campaign"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Full Markdown content of the lore")),
	), campaignHandlers.HandleSaveLore())

	s.AddTool(mcp.NewTool("save_npcs",
		mcp.WithDescription("Save NPCs and factions for the campaign"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Full Markdown content with NPCs and factions")),
	), campaignHandlers.HandleSaveNPCs())

	s.AddTool(mcp.NewTool("save_encounters",
		mcp.WithDescription("Save combat encounters and challenges for the campaign"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Full Markdown content with encounters")),
	), campaignHandlers.HandleSaveEncounters())

	s.AddTool(mcp.NewTool("save_bestiary",
		mcp.WithDescription("Save monsters and creatures to the bestiary"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Full Markdown content with monster stat blocks")),
	), campaignHandlers.HandleSaveBestiary())

	s.AddTool(mcp.NewTool("save_maps",
		mcp.WithDescription("Save map descriptions and scene layouts for the campaign"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Full Markdown content with maps and scenes")),
	), campaignHandlers.HandleSaveMaps())

	s.AddTool(mcp.NewTool("compile_pdf",
		mcp.WithDescription("Compile all campaign markdown files into a styled PDF"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("title", mcp.Description("PDF title (defaults to campaign name)")),
	), campaignHandlers.HandleCompilePDF())

	s.AddTool(mcp.NewTool("get_template",
		mcp.WithDescription("Get the Markdown/CSS template for a specific section type"),
		mcp.WithString("type", mcp.Required(), mcp.Description("Template type: act, npc, monster, encounter, map, lore")),
	), campaignHandlers.HandleGetTemplate())

	// Character management
	s.AddTool(mcp.NewTool("generate_character",
		mcp.WithDescription("Generate a new player character with stats and abilities"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Character name")),
		mcp.WithString("race", mcp.Description("Character race (e.g., humano, elfo, enano)")),
		mcp.WithString("class", mcp.Description("Character class (e.g., guerrero, mago, picaro)")),
		mcp.WithNumber("level", mcp.Description("Character level (1-20)"), mcp.DefaultNumber(1)),
		mcp.WithString("background", mcp.Description("Character background")),
		mcp.WithString("alignment", mcp.Description("Character alignment (e.g., LG, NG, CG)")),
	), characterHandlers.HandleGenerateCharacter())

	s.AddTool(mcp.NewTool("get_character",
		mcp.WithDescription("Get a character's sheet and details"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Character name")),
	), characterHandlers.HandleGetCharacter())

	s.AddTool(mcp.NewTool("list_characters",
		mcp.WithDescription("List all characters in a campaign"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name")),
	), characterHandlers.HandleListCharacters())

	// Quest management
	s.AddTool(mcp.NewTool("create_personal_quest",
		mcp.WithDescription("Create a personal quest for a character"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name")),
		mcp.WithString("character_name", mcp.Description("Character this quest is for")),
		mcp.WithString("quest_title", mcp.Required(), mcp.Description("Quest title")),
		mcp.WithString("quest_type", mcp.Description("Quest type (redencion, venganza, descubrimiento, proteccion)")),
		mcp.WithString("hook", mcp.Description("How the quest is introduced")),
		mcp.WithString("stakes", mcp.Description("What's at stake")),
		mcp.WithString("reward", mcp.Description("Quest reward")),
	), questHandlers.HandleCreateQuest())

	s.AddTool(mcp.NewTool("update_quest_status",
		mcp.WithDescription("Update the status of a quest"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name")),
		mcp.WithString("quest_id", mcp.Required(), mcp.Description("Quest ID")),
		mcp.WithString("status", mcp.Required(), mcp.Description("New status (active, completed, failed, on_hold)")),
		mcp.WithString("notes", mcp.Description("Progress notes")),
	), questHandlers.HandleUpdateQuestStatus())

	s.AddTool(mcp.NewTool("list_quests",
		mcp.WithDescription("List all quests in a campaign"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name")),
	), questHandlers.HandleListQuests())

	// Asset generation
	s.AddTool(mcp.NewTool("generate_map",
		mcp.WithDescription("Generate a procedural SVG battle map"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name")),
		mcp.WithString("filename", mcp.Required(), mcp.Description("Output filename (without extension)")),
		mcp.WithString("style", mcp.Description("Map style: dungeon, landscape, city"), mcp.DefaultString("dungeon")),
		mcp.WithString("title", mcp.Description("Map title")),
		mcp.WithNumber("rooms", mcp.Description("Number of rooms (2-10)"), mcp.DefaultNumber(6)),
		mcp.WithString("labels", mcp.Description("Comma-separated room labels")),
	), assetHandlers.HandleGenerateMap())

	s.AddTool(mcp.NewTool("generate_divider",
		mcp.WithDescription("Generate a decorative SVG divider"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name")),
		mcp.WithString("filename", mcp.Required(), mcp.Description("Output filename")),
		mcp.WithString("style", mcp.Description("Divider style: ornate, simple, double"), mcp.DefaultString("ornate")),
		mcp.WithNumber("width", mcp.Description("Width in pixels"), mcp.DefaultNumber(600)),
	), assetHandlers.HandleGenerateDivider())

	s.AddTool(mcp.NewTool("generate_image",
		mcp.WithDescription("Generate an image using AI"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name")),
		mcp.WithString("filename", mcp.Required(), mcp.Description("Output filename")),
		mcp.WithString("prompt", mcp.Required(), mcp.Description("Image generation prompt")),
		mcp.WithString("type", mcp.Description("Image type: cover, portrait, illustration, scene"), mcp.DefaultString("illustration")),
	), assetHandlers.HandleGenerateImage())

	return s
}
