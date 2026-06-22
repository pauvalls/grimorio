package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/config"
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/image"
	"github.com/pauvalls/grimorio/internal/mcp/handlers"
	"github.com/pauvalls/grimorio/internal/repository"
	fsrepo "github.com/pauvalls/grimorio/internal/repository/fs"
	"github.com/pauvalls/grimorio/internal/services"
	"github.com/pauvalls/grimorio/internal/services/monster"
	"github.com/pauvalls/grimorio/internal/tts/piper"
)

// NewServer creates a new MCP server with all tools wired.
// It returns the server and a shutdown function for graceful cleanup.
func NewServer(cfg *config.Config) (*server.MCPServer, func() error) {
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

	// V3 filesystem repositories
	monsterRepo := fsrepo.NewFilesystemMonsterRepository(cfg.OutputDir)
	encounterRepo := fsrepo.NewFilesystemEncounterRepository(cfg.OutputDir)
	areaRepoV3 := fsrepo.NewFilesystemAreaRepositoryV3(cfg.OutputDir)
	handoutRepoV3 := fsrepo.NewFilesystemHandoutRepositoryV3(cfg.OutputDir)
	milestoneXpRepo := fsrepo.NewFilesystemMilestoneXPRepository(cfg.OutputDir)
	tacticsRepo := fsrepo.NewFilesystemTacticsRepository(cfg.OutputDir)
	playerMapRepo := fsrepo.NewFilesystemPlayerMapRepository(cfg.OutputDir)
	checkpointRepo := repository.NewCheckpointRepository(cfg.OutputDir)
	auditRepo := repository.NewAuditLogRepository(cfg.OutputDir)

	// Initialize services
	campaignService := services.NewCampaignService(
		campaignRepo, actRepo, charRepo, npcRepo, questRepo,
		canonRepo,
		&monsterRepoWrapper{fs: monsterRepo},
		cfg.OutputDir, cfg.PDFEngine,
	)
	characterService := services.NewCharacterService(charRepo)
	questService := services.NewQuestService(questRepo)
	assetService := services.NewAssetService(cfg.OutputDir, cfg.Config)
	// Wire image-generation cache (Fase 4). The cache is best-effort:
	// a failure here must not break MCP server startup, so we log and
	// continue with an un-cached service.
	if imgCache, err := image.NewImageCache(cfg.ImageCacheDir, cfg.ImageCacheSize); err == nil {
		assetService = assetService.WithCache(imgCache)
	} else {
		fmt.Fprintf(os.Stderr, "warning: image cache disabled: %v\n", err)
	}
	canonService := services.NewCanonService(canonRepo, narrativeStateRepo, checkpointRepo)

	// Initialize structured logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	// Wire the advisory CR audit hook. The hook is strictly
	// advisory: it logs at WARN for major drift and never
	// blocks the save or the chapter finalize. The audit goes
	// through the same structured logger so DMs can grep the
	// MCP stderr for "cr audit" findings.
	campaignService.SetBestiaryAuditor(
		monster.NewBestiaryAuditor(monster.NewMonsterCRDriftAnalyzer(), logger),
	)

	// Degraded mode: if CANON_LEGACY_MODE is set or repo initialization fails
	if os.Getenv("CANON_LEGACY_MODE") == "1" {
		logger.Warn("CANON_LEGACY_MODE is enabled. Canon consistency gates will be bypassed.", "component", "canon")
		canonService.SetDegraded(true)
	}

	narrativeStateService := services.NewNarrativeStateService(narrativeStateRepo, canonRepo)
	validationEngine := services.NewValidationEngine(canonService, narrativeStateService, factionRepo, cfg.OutputDir)
	consistencyGateService := services.NewConsistencyGateService(canonService, narrativeStateService, validationEngine, checkpointRepo, auditRepo)
	factionService := services.NewFactionService(canonRepo, factionRepo)
	tableService := services.NewRandomTableService(canonRepo)
	handoutService := services.NewHandoutService(questRepo, canonRepo)
	consequenceEngine := services.NewConsequenceEngine(canonRepo)
	adaptationPatchService := services.NewAdaptationPatchService(actRepo, canonRepo)
	sessionPrepService := services.NewSessionPrepService(canonRepo, narrativeStateRepo, factionRepo)
	dmContextService := services.NewDMContextService(
		canonRepo, narrativeStateRepo, charRepo, npcRepo, questRepo,
		monsterRepo, areaRepoV3, factionRepo, sessionPrepService, cfg.OutputDir,
	)
	flowchartService := services.NewFlowchartService(canonRepo, actRepo)
	hookService := services.NewPlayerHookService(charRepo, canonRepo)
	prologueService := services.NewPrologueService(cfg.OutputDir, canonRepo)
	_ = adaptationPatchService

	// V3 services
	playerMapService := services.NewPlayerMapService(playerMapRepo)
	handoutServiceV3 := services.NewHandoutServiceV3(handoutRepoV3)
	milestoneService := services.NewMilestoneService(milestoneXpRepo)
	tacticsService := services.NewTacticsService(
		monsterRepo,
		encounterRepo,
		areaRepoV3,
		tacticsRepo,
	)

	// Fase 3 product features
	treasureService := services.NewTreasureService()
	campaignHealthCheck := services.NewCampaignHealthCheck(
		canonRepo, narrativeStateRepo, questRepo, factionRepo, npcRepo, cfg.OutputDir,
	)
	campaignHealthScore := services.NewCampaignHealthScore(campaignHealthCheck, validationEngine)
	exportHandlers := handlers.NewExportHandlers(cfg.OutputDir, cfg.PDFEngine)

	// Consolidation: the cross-file consistency engine that wires into
	// validation_engine, campaign_health, and the MCP consolidation tools.
	consolidationAdapter := services.NewConsolidationAdapter(cfg.OutputDir)
	consolidationHandlers := handlers.NewConsolidationHandlers(consolidationAdapter)

	// TTS initialization
	var ttsService *services.TTSService
	var ttsHandlers *handlers.TTSHandlers

	if cfg.TTS.Enabled {
		piperCfg := piper.Config{
			ModelPath:          cfg.TTS.Piper.ModelPath,
			ConfigPath:         cfg.TTS.Piper.ConfigPath,
			Port:               cfg.TTS.Piper.Port,
			Host:               cfg.TTS.Piper.Host,
			LengthScale:        cfg.TTS.Piper.LengthScale,
			Volume:             cfg.TTS.Piper.Volume,
			MaxRestarts:        cfg.TTS.Piper.MaxRestarts,
			HealthcheckTimeout: cfg.TTS.Piper.HealthcheckTimeout,
		}

		lifecycle := piper.NewLifecycleManager(piperCfg)
		if lifecycle.IsInstalled() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := lifecycle.Start(ctx); err != nil {
				logger.Warn("Piper TTS failed to start", "error", err, "component", "tts")
			}
			cancel()
		} else {
			logger.Info("Piper TTS not installed. TTS will be unavailable.", "component", "tts")
		}

		client := piper.NewClient(piperCfg.Host, piperCfg.Port)
		player := piper.NewPlayer(cfg.TTS.Audio.Player, cfg.TTS.Audio.Device)
		filter := &piper.DefaultTextFilter{}
		narrator := piper.NewNarrator(filter, client, player, cfg.TTS.Chunker.MaxChunkSize)

		ttsCacheDir := filepath.Join(cfg.OutputDir, ".tts-cache")
		voiceStore := services.NewFileCampaignVoiceStore(ttsCacheDir)

		ttsService = services.NewTTSService(narrator, lifecycle, voiceStore, cfg.TTS.Enabled)
		ttsHandlers = handlers.NewTTSHandlers(ttsService)
	} else {
		// TTS disabled: create service with nil dependencies so handlers report unavailable
		ttsService = services.NewTTSService(nil, nil, nil, false)
		ttsHandlers = handlers.NewTTSHandlers(ttsService)
	}

	// Initialize handlers
	campaignHandlers := handlers.NewCampaignHandlers(campaignService)
	chapterPartHandlers := handlers.NewChapterPartHandlers(campaignService)
	characterHandlers := handlers.NewCharacterHandlers(characterService)
	questHandlers := handlers.NewQuestHandlers(questService)
	assetHandlers := handlers.NewAssetHandlers(assetService)
	canonHandlers := handlers.NewCanonHandlers(canonService, narrativeStateService, validationEngine, consistencyGateService, campaignService)
	factionHandlers := handlers.NewFactionHandlers(factionService)
	tableHandlers := handlers.NewTableHandlers(tableService)
	handoutHandlers := handlers.NewHandoutHandlers(handoutService)
	consequenceHandlers := handlers.NewConsequenceHandlers(consequenceEngine, narrativeStateService)
	sessionPrepHandlers := handlers.NewSessionPrepHandlers(sessionPrepService)
	dmContextHandlers := handlers.NewDMContextHandlers(dmContextService)
	flowchartHandlers := handlers.NewFlowchartHandlers(flowchartService)
	hookHandlers := handlers.NewHookHandlers(hookService)
	prologueHandlers := handlers.NewPrologueHandlers(prologueService, campaignService)

	// V3 handlers
	tacticsHandlers := handlers.NewTacticsHandlers(tacticsService, encounterRepo, areaRepoV3)
	milestoneHandlers := handlers.NewMilestoneHandlers(milestoneService)
	playerMapHandlers := handlers.NewPlayerMapHandlers(playerMapService)
	handoutV3Handlers := handlers.NewHandoutV3Handlers(handoutServiceV3)

	// Name generation handler
	namegenHandlers := handlers.NewNamegenHandlers()

	// Visualization and dashboard handlers
	vizHandlers := handlers.NewVizHandlers(canonService)
	dashboardHandlers := handlers.NewDashboardHandlers(factionService, canonService)
	timelineHandlers := handlers.NewTimelineHandlers(narrativeStateService)

	// Fase 3 handlers
	healthHandlers := handlers.NewHealthHandlers(campaignHealthScore)
	treasureHandlers := handlers.NewTreasureHandlers(treasureService)

	// Monster design engine (mde-004) — 3 tools for CR validation,
	// suggestion, and campaign-wide audit.
	monsterValidationHandlers := handlers.NewMonsterValidationHandlers(cfg.OutputDir)

	// Register tools
	// Campaign management
	s.AddTool(mcp.NewTool("create_campaign",
		mcp.WithDescription("Create a new campaign directory structure"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Campaign name (kebab-case)"), mcp.Title("Campaign Name")),
		mcp.WithString("title", mcp.Description("Campaign title")),
		mcp.WithString("setting", mcp.Description("Brief setting description")),
		mcp.WithString("template", mcp.Description("Campaign template preset: Urban Fantasy, Gothic Horror, Maritime Adventure, Dungeon Crawl, Political Intrigue")),
	), campaignHandlers.HandleCreateCampaign())

	s.AddTool(mcp.NewTool("save_chapter",
		mcp.WithDescription("Save a self-contained chapter with inline NPCs, encounters, and areas (WotC format: 10-15 areas per chapter)"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithNumber("chapter_number", mcp.Required(), mcp.Description("Chapter number (1, 2, 3...)"), mcp.Title("Chapter Number")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Chapter title")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Full Markdown content with inline NPCs, encounters, and numbered areas (WotC format)")),
	), campaignHandlers.HandleSaveChapter())

	s.AddTool(mcp.NewTool("save_chapter_part",
		mcp.WithDescription("Save a chapter part to draft directory for sequential generation. Call 6 times (opener, npcs, encounters, areas-1, areas-2, closing), then call finalize_chapter."),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithNumber("chapter_number", mcp.Required(), mcp.Description("Chapter number (0 for prologue, 1-N for chapters)")),
		mcp.WithString("part_name", mcp.Required(), mcp.Description("Part name: opener, npcs, encounters, areas-1, areas-2, closing")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Markdown content for this part")),
	), chapterPartHandlers.HandleSaveChapterPart())

	s.AddTool(mcp.NewTool("finalize_chapter",
		mcp.WithDescription("Assemble draft parts into final chapter, validate, and sync canon. Call after all save_chapter_part calls."),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithNumber("chapter_number", mcp.Required(), mcp.Description("Chapter number")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Chapter title")),
	), chapterPartHandlers.HandleFinalizeChapter())

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
		mcp.WithString("type", mcp.Required(), mcp.Description("Template type: areas, npc, monster, encounter, map, lore")),
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

	s.AddTool(mcp.NewTool("generate_character_hooks",
		mcp.WithDescription("Generate personalized plot hooks for all player characters in a campaign. Returns hooks organized by character and by area for easy integration."),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
	), hookHandlers.HandleGenerateCharacterHooks())

	s.AddTool(mcp.NewTool("generate_names",
		mcp.WithDescription("Generate fantasy names using syllable-based generation"),
		mcp.WithString("category", mcp.Required(), mcp.Description("Name category: character, npc, city, tavern, monster, faction, item")),
		mcp.WithString("style", mcp.Description("Cultural style: generic_fantasy, elven, dwarven, orcish, human_medieval"), mcp.DefaultString("generic_fantasy")),
		mcp.WithNumber("count", mcp.Required(), mcp.Description("Number of names to generate (1-50)")),
		mcp.WithNumber("seed", mcp.Description("Random seed for reproducibility")),
	), namegenHandlers.HandleGenerateNames())

	s.AddTool(mcp.NewTool("grimorio_generate_prologue",
		mcp.WithDescription("Generate a 4-part narrative prologue for a campaign."),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("tone", mcp.Description("Tone: grim, heroic, mystery, horror (default: heroic)")),
		mcp.WithArray("character_hooks", mcp.Description("Optional character hook strings to weave into the prologue")),
		mcp.WithBoolean("regenerate", mcp.Description("Force regeneration even if prologue already exists"), mcp.DefaultBool(false)),
	), prologueHandlers.HandleGeneratePrologue())

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
		mcp.WithString("brief_description", mcp.Description("Brief campaign description or story concept")),
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
		mcp.WithArray("revealed_clues", mcp.Description("Clues revealed this session (strings or objects with id/description/source_act/is_critical)")),
		mcp.WithArray("completed_quests", mcp.Description("Quest IDs completed this session")),
		mcp.WithArray("dead_npcs", mcp.Description("NPCs who died this session (strings or objects with npc_id/name)")),
		mcp.WithArray("key_decisions", mcp.Description("Key decisions made this session (strings or objects with id/description/choice_made/impact_scope)")),
		mcp.WithArray("active_quests", mcp.Description("Active quests (objects with id/name/status/source_act)")),
		mcp.WithArray("key_items", mcp.Description("Key items acquired this session (objects with id/name/holder/session_found)")),
		mcp.WithString("session_summary", mcp.Description("Summary of what happened this session")),
		mcp.WithNumber("xp_awarded", mcp.Description("XP awarded this session")),
		mcp.WithString("xp_reason", mcp.Description("XP reason: combat, roleplay, milestone, exploration")),
		mcp.WithArray("loot_acquired", mcp.Description("Loot acquired this session")),
		mcp.WithString("dm_notes", mcp.Description("DM notes for this session")),
		mcp.WithString("current_location", mcp.Description("Current party location")),
		mcp.WithString("current_chapter_id", mcp.Description("Current chapter ID (e.g., chapter-1)")),
		mcp.WithArray("completed_chapters", mcp.Description("Chapter IDs completed this session")),
		mcp.WithArray("pc_status", mcp.Description("PC health status (objects with name, hp_current, hp_max, conditions)")),
		mcp.WithString("default_source_act", mcp.Description("Default source act for string clues (e.g., act-1)")),
		mcp.WithString("default_choice_made", mcp.Description("Default choice made for string decisions")),
		mcp.WithString("default_impact_scope", mcp.Description("Default impact scope for string decisions")),
		mcp.WithArray("critical_clue_indices", mcp.Description("0-based indices of critical clues in revealed_clues array")),
		mcp.WithBoolean("replace_session", mcp.Description("Replace existing session log entry instead of appending")),
		mcp.WithBoolean("sync_to_canon", mcp.Description("Sync state changes to canon document (default: false)")),
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
		mcp.WithBoolean("with_scenarios", mcp.Description("Include encounter, loot, and NPC scenario recommendations"), mcp.DefaultBool(false)),
	), sessionPrepHandlers.HandleGenerateSessionPrep())

	s.AddTool(mcp.NewTool("generate_flowchart",
		mcp.WithDescription("Generate a campaign flowchart in Mermaid and SVG"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("detail_level", mcp.Description("Detail level: overview, act, decision"), mcp.DefaultString("overview")),
	), flowchartHandlers.HandleGenerateFlowchart())

	// DM context tool
	s.AddTool(mcp.NewTool("dm_session_context",
		mcp.WithDescription("Aggregate all campaign data into a single JSON payload for the AI Dungeon Master"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithNumber("session_num", mcp.Description("Session number (defaults to current session)")),
		mcp.WithBoolean("include_prologue", mcp.Description("Include prologue in payload"), mcp.DefaultBool(true)),
		mcp.WithBoolean("include_pdf_text", mcp.Description("Extract and include text from compiled PDF if available"), mcp.DefaultBool(false)),
	), dmContextHandlers.HandleDMContext())

	// V3 tools
	s.AddTool(mcp.NewTool("grimorio_generate_tactics",
		mcp.WithDescription("Generate combat tactics for all monsters in an encounter"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("encounter_id", mcp.Required(), mcp.Description("Encounter ID")),
		mcp.WithString("area_id", mcp.Description("Area ID for environmental tactics")),
	), tacticsHandlers.HandleGenerateTactics())

	s.AddTool(mcp.NewTool("grimorio_get_tactics",
		mcp.WithDescription("Retrieve generated tactics for an encounter"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("encounter_id", mcp.Required(), mcp.Description("Encounter ID")),
	), tacticsHandlers.HandleGetTactics())

	s.AddTool(mcp.NewTool("grimorio_generate_xp_table",
		mcp.WithDescription("Generate a milestone XP table for a chapter"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithNumber("chapter_number", mcp.Required(), mcp.Description("Chapter number")),
		mcp.WithNumber("level_min", mcp.Description("Minimum level for this chapter"), mcp.DefaultNumber(1)),
		mcp.WithNumber("level_max", mcp.Description("Maximum level for this chapter"), mcp.DefaultNumber(5)),
	), milestoneHandlers.HandleGenerateXPTable())

	s.AddTool(mcp.NewTool("grimorio_track_party_progress",
		mcp.WithDescription("Track party progress and calculate current level from XP"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("party_id", mcp.Required(), mcp.Description("Party identifier")),
	), milestoneHandlers.HandleTrackPartyProgress())

	s.AddTool(mcp.NewTool("grimorio_generate_player_map",
		mcp.WithDescription("Generate a player-facing map variant with secrets redacted"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("dm_map_id", mcp.Required(), mcp.Description("DM map ID (source)")),
		mcp.WithString("area_id", mcp.Description("Area ID to associate the player map with")),
	), playerMapHandlers.HandleGeneratePlayerMap())

	s.AddTool(mcp.NewTool("grimorio_export_handout",
		mcp.WithDescription("Export a handout in the specified format"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("handout_id", mcp.Required(), mcp.Description("Handout ID")),
		mcp.WithString("format", mcp.Description("Export format (text, pdf)"), mcp.DefaultString("text")),
	), handoutV3Handlers.HandleExportHandout())

	// Visualization tools
	s.AddTool(mcp.NewTool("visualize_relationship_graph",
		mcp.WithDescription("Visualize the entity relationship graph for a campaign as an interactive D3.js force-directed graph"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
	), vizHandlers.HandleGenerateRelationshipGraph())

	s.AddTool(mcp.NewTool("faction_reputation_dashboard",
		mcp.WithDescription("Show a visual dashboard of all faction reputations with color-coded badges and trend sparklines"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
	), dashboardHandlers.HandleFactionDashboard())

	s.AddTool(mcp.NewTool("session_timeline",
		mcp.WithDescription("Show a vertical timeline of all recorded sessions with expandable key decisions"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
	), timelineHandlers.HandleSessionTimeline())

	// Fase 3 product feature tools
	s.AddTool(mcp.NewTool("campaign_health_dashboard",
		mcp.WithDescription("Compute and return campaign health scores (0-100) for OverallHealth, CanonCompleteness, NarrativeCoherence, FactionBalance, WotCCompliance, and HookCoverage"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
	), healthHandlers.HandleCampaignHealthDashboard())

	s.AddTool(mcp.NewTool("generate_treasure",
		mcp.WithDescription("Generate SRD-compliant treasure (individual or hoard)"),
		mcp.WithString("type", mcp.Description("Treasure type: individual or hoard"), mcp.DefaultString("individual")),
		mcp.WithNumber("cr", mcp.Description("Challenge Rating for individual treasure (0-30)"), mcp.DefaultNumber(1)),
		mcp.WithNumber("tier", mcp.Description("Treasure tier for hoards (1-4)"), mcp.DefaultNumber(1)),
	), treasureHandlers.HandleGenerateTreasure())

	s.AddTool(mcp.NewTool("export_campaign",
		mcp.WithDescription("Export campaign to PDF, Markdown, or EPUB"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("format", mcp.Description("Export format: pdf, markdown, epub"), mcp.DefaultString("pdf")),
		mcp.WithString("title", mcp.Description("Export title (defaults to campaign name)")),
	), exportHandlers.HandleExportCampaign())

	// Monster design engine tools (mde-004)
	s.AddTool(mcp.NewTool("validate_monster",
		mcp.WithDescription("Validate a monster (by name from a campaign, or by raw markdown) and return a ValidationResult"),
		mcp.WithString("monster_name", mcp.Description("Name of a monster in the campaign bestiary (optional if markdown is provided)")),
		mcp.WithString("markdown", mcp.Description("Raw markdown stat block to validate (optional if monster_name is provided)")),
		mcp.WithString("campaign", mcp.Description("Campaign ID (required if monster_name is provided)")),
	), monsterValidationHandlers.HandleValidateMonster())

	s.AddTool(mcp.NewTool("suggest_monster_cr",
		mcp.WithDescription("Given a target CR and an optional concept, return a Monster skeleton (as markdown or JSON)"),
		mcp.WithNumber("target_cr", mcp.Required(), mcp.Description("Target CR (0-30, including sub-integers 0.125, 0.25, 0.5)")),
		mcp.WithString("concept", mcp.Description("Optional concept (e.g. 'fire-breathing dragon')")),
		mcp.WithString("output", mcp.Description("Output format: markdown, json (default: markdown)")),
	), monsterValidationHandlers.HandleSuggestMonsterCR())

	s.AddTool(mcp.NewTool("audit_monster_cr",
		mcp.WithDescription("Audit an entire campaign's bestiary and return a summary + per-monster validation"),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign ID")),
	), monsterValidationHandlers.HandleAuditMonsterCR())

	// Consolidation tools — cross-file consistency engine
	s.AddTool(mcp.NewTool("consolidate_campaign",
		mcp.WithDescription("Run the full consolidation engine: detect entity, lore, stat-block, event, file, and map drift; apply safe fixes; return a ConsolidationReport."),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithBoolean("auto_fix", mcp.Description("Apply safe fixes automatically (default: false)"), mcp.DefaultBool(false)),
		mcp.WithNumber("similarity_threshold", mcp.Description("Entity merge threshold (0.0-1.0, default 0.85)"), mcp.DefaultNumber(0.85)),
		mcp.WithString("backup_dir", mcp.Description("Override backup directory (default: .consolidation/backups/<timestamp>)")),
	), consolidationHandlers.HandleConsolidateCampaign())

	s.AddTool(mcp.NewTool("detect_inconsistencies",
		mcp.WithDescription("Read-only drift detection across all campaign markdown files. Returns a ConsolidationReport without mutating files."),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
	), consolidationHandlers.HandleDetectInconsistencies())

	s.AddTool(mcp.NewTool("resolve_ambiguity",
		mcp.WithDescription("Resolve a specific AmbiguityQuestion by ID with a user/agent decision."),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
		mcp.WithString("question_id", mcp.Required(), mcp.Description("AmbiguityQuestion.ID from a prior detect_inconsistencies or consolidate_campaign report")),
		mcp.WithString("decision", mcp.Required(), mcp.Description("Resolution: one of the question's Options")),
	), consolidationHandlers.HandleResolveAmbiguity())

	s.AddTool(mcp.NewTool("regenerate_index",
		mcp.WithDescription("Generate or refresh INDEX.md with breadcrumbs and verified links to every source file in the campaign."),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
	), consolidationHandlers.HandleRegenerateIndex())

	s.AddTool(mcp.NewTool("verify_campaign_freshness",
		mcp.WithDescription("Compare campaign.md and INDEX.md against source files and report whether they are stale."),
		mcp.WithString("campaign", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
	), consolidationHandlers.HandleVerifyCampaignFreshness())

	// TTS tools
	s.AddTool(mcp.NewTool("set_dm_mode",
		mcp.WithDescription("Set the DM output mode to written or TTS"),
		mcp.WithString("mode", mcp.Required(), mcp.Description("Mode: written or tts")),
	), ttsHandlers.HandleSetDMMode())

	s.AddTool(mcp.NewTool("assign_npc_voice",
		mcp.WithDescription("Assign a voice prompt to an NPC for consistent TTS dialogue"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name")),
		mcp.WithString("npc_name", mcp.Required(), mcp.Description("NPC name")),
		mcp.WithString("voice_prompt", mcp.Required(), mcp.Description("Voice description prompt")),
	), ttsHandlers.HandleAssignNPCVoice())

	s.AddTool(mcp.NewTool("tts_control",
		mcp.WithDescription("Control TTS playback (skip, stop, pause, resume)"),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action: skip, stop, pause, resume")),
	), ttsHandlers.HandleTTSControl())

	s.AddTool(mcp.NewTool("list_tts_voices",
		mcp.WithDescription("List assigned NPC voices for a campaign"),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign name")),
	), ttsHandlers.HandleListTTSVoices())

	s.AddTool(mcp.NewTool("get_tts_status",
		mcp.WithDescription("Get current TTS system status"),
	), ttsHandlers.HandleGetTTSStatus())

	s.AddTool(mcp.NewTool("tts_speak",
		mcp.WithDescription("Speak text aloud using Piper TTS. Displays the text on screen AND narrates it via voice. Use after generating narrative text to have it spoken."),
		mcp.WithString("text", mcp.Required(), mcp.Description("Text to speak aloud (e.g., narrative description, NPC dialogue)")),
	), ttsHandlers.HandleTTSSpeak())

	shutdown := func() error {
		if ttsService != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			return ttsService.Shutdown(ctx)
		}
		return nil
	}

	return s, shutdown
}

// monsterRepoWrapper adapts fs.FilesystemMonsterRepository to repository.MonsterRepository
type monsterRepoWrapper struct {
	fs *fsrepo.FilesystemMonsterRepository
}

func (w *monsterRepoWrapper) Save(monster *domain.Monster) error {
	return w.fs.Save(context.Background(), monster.CampaignID, monster)
}

func (w *monsterRepoWrapper) Read(campaignID, name string) (*domain.Monster, error) {
	return w.fs.Read(context.Background(), campaignID, name)
}

func (w *monsterRepoWrapper) List(campaignID string) ([]domain.Monster, error) {
	ptrs, err := w.fs.List(context.Background(), campaignID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Monster, len(ptrs))
	for i, p := range ptrs {
		result[i] = *p
	}
	return result, nil
}

func (w *monsterRepoWrapper) Delete(campaignID, name string) error {
	return w.fs.Delete(context.Background(), campaignID, name)
}
