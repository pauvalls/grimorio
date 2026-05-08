package mcp

import (
	"log"
	"os"

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
	canonRepo := repository.NewFilesystemCanonRepository(cfg.OutputDir)
	narrativeStateRepo := repository.NewFilesystemNarrativeStateRepository(cfg.OutputDir)
	factionRepo := repository.NewFilesystemFactionRepository(cfg.OutputDir)

	// Initialize services
	campaignService := services.NewCampaignService(
		campaignRepo, actRepo, charRepo, npcRepo, questRepo,
		cfg.OutputDir, cfg.PDFEngine,
	)
	characterService := services.NewCharacterService(charRepo)
	questService := services.NewQuestService(questRepo)
	assetService := services.NewAssetService(cfg.OutputDir, cfg.Config)
	canonService := services.NewCanonService(canonRepo, narrativeStateRepo)

	// Degraded mode: if CANON_LEGACY_MODE is set or repo initialization fails
	if os.Getenv("CANON_LEGACY_MODE") == "1" {
		log.Println("WARNING: CANON_LEGACY_MODE is enabled. Canon consistency gates will be bypassed.")
		canonService.SetDegraded(true)
	}

	narrativeStateService := services.NewNarrativeStateService(narrativeStateRepo, canonRepo)
	validationEngine := services.NewValidationEngine(canonService, narrativeStateService, factionRepo)
	consistencyGateService := services.NewConsistencyGateService(canonService, narrativeStateService, validationEngine)
	factionService := services.NewFactionService(canonRepo, factionRepo)
	tableService := services.NewRandomTableService(canonRepo)
	handoutService := services.NewHandoutService(questRepo, canonRepo)
	consequenceEngine := services.NewConsequenceEngine(canonRepo)
	adaptationPatchService := services.NewAdaptationPatchService(actRepo, canonRepo)
	sessionPrepService := services.NewSessionPrepService(canonRepo, narrativeStateRepo)
	flowchartService := services.NewFlowchartService(canonRepo, actRepo)
	_ = adaptationPatchService

	// Initialize handlers
	campaignHandlers := handlers.NewCampaignHandlers(campaignService)
	characterHandlers := handlers.NewCharacterHandlers(characterService)
	questHandlers := handlers.NewQuestHandlers(questService)
	assetHandlers := handlers.NewAssetHandlers(assetService)
	canonHandlers := handlers.NewCanonHandlers(canonService, narrativeStateService, validationEngine, consistencyGateService)
	factionHandlers := handlers.NewFactionHandlers(factionService)
	tableHandlers := handlers.NewTableHandlers(tableService)
	handoutHandlers := handlers.NewHandoutHandlers(handoutService)
	consequenceHandlers := handlers.NewConsequenceHandlers(consequenceEngine, narrativeStateService)
	sessionPrepHandlers := handlers.NewSessionPrepHandlers(sessionPrepService)
	flowchartHandlers := handlers.NewFlowchartHandlers(flowchartService)

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

	s.AddTool(mcp.NewTool("save_introduction",
		mcp.WithDescription("Save campaign introduction/overview document"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Full Markdown content of the introduction")),
	), campaignHandlers.HandleSaveIntroduction())

	s.AddTool(mcp.NewTool("save_setting_guide",
		mcp.WithDescription("Save campaign setting guide (DM-only, with spoilers)"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Full Markdown content of the setting guide")),
	), campaignHandlers.HandleSaveSettingGuide())

	s.AddTool(mcp.NewTool("save_appendices",
		mcp.WithDescription("Save campaign appendices (items, monsters, handouts)"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Full Markdown content of the appendices")),
	), campaignHandlers.HandleSaveAppendices())

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

	s.AddTool(mcp.NewTool("save_characters",
		mcp.WithDescription("Save multiple characters to a campaign"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name")),
		mcp.WithArray("characters", mcp.Required(), mcp.Description("Array of character objects with name, race, class, level, background, alignment")),
	), characterHandlers.HandleSaveCharacters())

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
		mcp.WithDescription("Generate a procedural SVG battle map and optionally link it to a markdown file"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name")),
		mcp.WithString("filename", mcp.Required(), mcp.Description("Output filename (without extension)")),
		mcp.WithString("style", mcp.Description("Map style: dungeon, landscape, city"), mcp.DefaultString("dungeon")),
		mcp.WithString("title", mcp.Description("Map title")),
		mcp.WithNumber("rooms", mcp.Description("Number of rooms (2-10)"), mcp.DefaultNumber(6)),
		mcp.WithString("labels", mcp.Description("Comma-separated room labels")),
		mcp.WithString("markdown_file", mcp.Description("Optional: markdown file to insert image reference (e.g., 'maps/maps.md')")),
		mcp.WithString("section", mcp.Description("Optional: section heading where to insert the image reference")),
		mcp.WithString("alt", mcp.Description("Optional: alt text for the image (defaults to title or filename)")),
	), assetHandlers.HandleGenerateMap())

	s.AddTool(mcp.NewTool("generate_divider",
		mcp.WithDescription("Generate a decorative SVG divider and optionally link it to a markdown file"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name")),
		mcp.WithString("filename", mcp.Required(), mcp.Description("Output filename")),
		mcp.WithString("style", mcp.Description("Divider style: ornate, simple, double"), mcp.DefaultString("ornate")),
		mcp.WithNumber("width", mcp.Description("Width in pixels"), mcp.DefaultNumber(600)),
		mcp.WithString("markdown_file", mcp.Description("Optional: markdown file to insert image reference")),
		mcp.WithString("section", mcp.Description("Optional: section heading where to insert the image reference")),
		mcp.WithString("alt", mcp.Description("Optional: alt text for the image (defaults to filename)")),
	), assetHandlers.HandleGenerateDivider())

	s.AddTool(mcp.NewTool("generate_image",
		mcp.WithDescription("Generate an image using AI and optionally link it to a markdown file"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name")),
		mcp.WithString("filename", mcp.Required(), mcp.Description("Output filename")),
		mcp.WithString("prompt", mcp.Required(), mcp.Description("Image generation prompt")),
		mcp.WithString("type", mcp.Description("Image type: cover, portrait, illustration, scene"), mcp.DefaultString("illustration")),
		mcp.WithString("markdown_file", mcp.Description("Optional: markdown file to insert image reference (e.g., 'npcs/npcs_and_factions.md')")),
		mcp.WithString("section", mcp.Description("Optional: section heading where to insert the image reference")),
		mcp.WithString("alt", mcp.Description("Optional: alt text for the image (defaults to filename)")),
	), assetHandlers.HandleGenerateImage())

	// Canon and narrative coherence tools
	s.AddTool(mcp.NewTool("generate_adventure_bible",
		mcp.WithDescription("Generate the canon initial document for a campaign from a brief"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("level_range", mcp.Description("Level range (e.g., 1-10)")),
		mcp.WithString("tone", mcp.Description("Campaign tone (grim, whimsical, heroic, horror, political, mystery)")),
		mcp.WithString("setting_type", mcp.Description("Setting type (urban, wilderness, dungeon, maritime, planar)")),
		mcp.WithArray("themes", mcp.Description("Campaign themes")),
		mcp.WithString("villain_type", mcp.Description("Type of main villain")),
		mcp.WithString("mcguffin_type", mcp.Description("Type of McGuffin")),
	), canonHandlers.HandleGenerateAdventureBible())

	s.AddTool(mcp.NewTool("validate_canon",
		mcp.WithDescription("Validate a content proposal against the campaign canon"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("proposal_id", mcp.Required(), mcp.Description("Proposal ID")),
		mcp.WithString("proposal_type", mcp.Required(), mcp.Description("Type: act, quest, encounter, npc, lore, item")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Content markdown or JSON to validate")),
		mcp.WithString("faction_context", mcp.Description("Optional faction context for reputation-aware validation")),
	), canonHandlers.HandleValidateCanon())

	s.AddTool(mcp.NewTool("update_narrative_state",
		mcp.WithDescription("Update the narrative state after a session"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithNumber("session_num", mcp.Required(), mcp.Description("Session number (1, 2, 3...)")),
		mcp.WithArray("revealed_clues", mcp.Description("Clues revealed this session")),
		mcp.WithArray("completed_quests", mcp.Description("Quest IDs completed this session")),
		mcp.WithArray("dead_npcs", mcp.Description("NPCs who died this session")),
		mcp.WithArray("key_decisions", mcp.Description("Key decisions made this session")),
	), canonHandlers.HandleUpdateNarrativeState())

	s.AddTool(mcp.NewTool("check_consistency",
		mcp.WithDescription("Check campaign consistency across all artifacts"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("scope", mcp.Description("Scope: full, lore_only, acts_only, npcs_only, quests_only"), mcp.DefaultString("full")),
	), canonHandlers.HandleCheckConsistency())

	s.AddTool(mcp.NewTool("process_consistency_gate",
		mcp.WithDescription("Process a batch of content proposals through the consistency gate"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("batch_id", mcp.Required(), mcp.Description("Batch identifier (e.g., batch-1, batch-2)")),
		mcp.WithArray("proposals", mcp.Required(), mcp.Description("Array of content proposals with id, type, content, and optional entity_references")),
		mcp.WithNumber("attempt", mcp.Description("Attempt number (1-3)"), mcp.DefaultNumber(1)),
		mcp.WithBoolean("fast_mode", mcp.Description("Skip non-critical validations for speed"), mcp.DefaultBool(false)),
	), canonHandlers.HandleProcessConsistencyGate())

	// Faction and living world tools
	s.AddTool(mcp.NewTool("update_faction_reputation",
		mcp.WithDescription("Update faction reputation and propagate to allies/enemies"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("faction_id", mcp.Required(), mcp.Description("Target faction ID")),
		mcp.WithString("party_id", mcp.Required(), mcp.Description("Party identifier")),
		mcp.WithNumber("delta", mcp.Required(), mcp.Description("Reputation delta (-100 to 100)")),
		mcp.WithString("reason", mcp.Required(), mcp.Description("Human-readable reason")),
	), factionHandlers.HandleUpdateFactionReputation())

	s.AddTool(mcp.NewTool("generate_random_tables",
		mcp.WithDescription("Generate contextual random tables from canon data"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("table_type", mcp.Required(), mcp.Description("Table type: encounter, rumor, weather, treasure")),
		mcp.WithString("level_range", mcp.Description("Level range filter (e.g., 5-8)")),
		mcp.WithString("setting_type", mcp.Description("Setting type filter")),
		mcp.WithNumber("party_size", mcp.Description("Party size")),
		mcp.WithString("location_hint", mcp.Description("Location hint")),
	), tableHandlers.HandleGenerateRandomTables())

	s.AddTool(mcp.NewTool("generate_handouts",
		mcp.WithDescription("Generate player and/or DM handouts"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("handout_type", mcp.Required(), mcp.Description("Handout type: summary, encounter, quest, lore, faction")),
		mcp.WithArray("content_refs", mcp.Required(), mcp.Description("Content IDs to include")),
		mcp.WithString("version", mcp.Description("Version: player, dm, both (default)")),
	), handoutHandlers.HandleGenerateHandouts())

	s.AddTool(mcp.NewTool("evaluate_consequences",
		mcp.WithDescription("Evaluate consequence rules against narrative state"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
	), consequenceHandlers.HandleEvaluateConsequences())

	s.AddTool(mcp.NewTool("generate_session_prep",
		mcp.WithDescription("Generate a DM prep sheet for the next session"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithNumber("session_num", mcp.Description("Session number (defaults to current+1)")),
	), sessionPrepHandlers.HandleGenerateSessionPrep())

	s.AddTool(mcp.NewTool("generate_flowchart",
		mcp.WithDescription("Generate a campaign flowchart in Mermaid and SVG"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("detail_level", mcp.Description("Detail level: overview, act, decision"), mcp.DefaultString("overview")),
	), flowchartHandlers.HandleGenerateFlowchart())

	return s
}
