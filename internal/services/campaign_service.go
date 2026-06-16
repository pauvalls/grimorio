package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/pauvalls/grimorio/internal/compiler"
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/generators"
	"github.com/pauvalls/grimorio/internal/repository"
)

// CampaignService handles campaign business logic
type CampaignService struct {
	canonRepo    repository.CanonRepository
	campaignRepo repository.CampaignRepository
	actRepo      repository.ActRepository
	charRepo     repository.CharacterRepository
	npcRepo      repository.NPCRepository
	questRepo    repository.QuestRepository
	monsterRepo  repository.MonsterRepository
	baseDir      string
	pdfEngine    string
}

// NewCampaignService creates a new campaign service
func NewCampaignService(
	campaignRepo repository.CampaignRepository,
	actRepo repository.ActRepository,
	charRepo repository.CharacterRepository,
	npcRepo repository.NPCRepository,
	questRepo repository.QuestRepository,
	canonRepo repository.CanonRepository,
	monsterRepo repository.MonsterRepository,
	baseDir string,
	pdfEngine string,
) *CampaignService {
	return &CampaignService{
		campaignRepo: campaignRepo,
		actRepo:      actRepo,
		charRepo:     charRepo,
		npcRepo:      npcRepo,
		questRepo:    questRepo,
		canonRepo:    canonRepo,
		monsterRepo:  monsterRepo,
		baseDir:      baseDir,
		pdfEngine:    pdfEngine,
	}
}

// GetBaseDir returns the base directory for campaign files
func (s *CampaignService) GetBaseDir() string {
	return s.baseDir
}

// CreateCampaign creates a new campaign with optional template.
func (s *CampaignService) CreateCampaign(name, title, setting, templateName string) (*domain.Campaign, error) {
	if title == "" {
		title = name
	}

	campaign := &domain.Campaign{
		Name:    name,
		Title:   title,
		Setting: setting,
		Status:  "active",
	}

	// Apply template defaults if provided
	if templateName != "" {
		tmpl, err := GetTemplate(templateName)
		if err == nil {
			campaign.Template = tmpl.Name
			ApplyTemplate(campaign, tmpl)
		}
	}

	if err := s.campaignRepo.Create(campaign); err != nil {
		return nil, fmt.Errorf("failed to create campaign: %w", err)
	}

	if err := s.generateSessionZero(campaign); err != nil {
		return campaign, fmt.Errorf("campaign created but session-zero generation failed: %w", err)
	}

	return campaign, nil
}

// generateSessionZero creates a default session-zero.md for the campaign
func (s *CampaignService) generateSessionZero(campaign *domain.Campaign) error {
	tmplStr, err := compiler.GetTemplate("session-zero")
	if err != nil {
		return err
	}

	tmpl, err := template.New("session-zero").Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("failed to parse session-zero template: %w", err)
	}

	data := struct {
		Name                 string
		Title                string
		Setting              string
		LevelRange           string
		Tone                 string
		HouseRules           string
		StartingLevel        string
		AllowedSources       string
		StatMethod           string
		SuggestedBackgrounds string
		ToneDescription      string
		ShockPoints          []domain.ShockPoint
	}{
		Name:                 campaign.Name,
		Title:                campaign.Title,
		Setting:              campaign.Setting,
		LevelRange:           "1-10",
		Tone:                 "heroic",
		HouseRules:           "Ninguna por defecto. Agregá las reglas de casa de tu mesa aquí.",
		StartingLevel:        "1",
		AllowedSources:       "Player's Handbook, Dungeon Master's Guide",
		StatMethod:           "Standard Array (15, 14, 13, 12, 10, 8) o Point Buy",
		SuggestedBackgrounds: "Acolyte, Criminal, Folk Hero, Sage, Soldier",
		ToneDescription:      "Una aventura épica donde los héroes enfrentan desafíos crecientes mientras descubren secretos del mundo.",
		ShockPoints: []domain.ShockPoint{
			{
				Type:        "Violencia",
				Severity:    "moderate",
				Description: "Combate fantástico, descripciones de heridas y sangre. No incluye tortura gráfica.",
				SafetyTools: []string{"X-Card", "Fade to black"},
			},
			{
				Type:        "Horror",
				Severity:    "mild",
				Description: "Imaginería perturbadora, criaturas aterradoras. Consultar límites del grupo.",
				SafetyTools: []string{"X-Card", "Lines and Veils"},
			},
			{
				Type:        "Muerte de Personajes",
				Severity:    "moderate",
				Description: "Los personajes pueden morir permanentemente. Las decisiones tienen consecuencias.",
				SafetyTools: []string{"X-Card", "Session Zero discussion"},
			},
		},
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute session-zero template: %w", err)
	}

	return s.saveMarkdownFile(campaign.Name, "", "session-zero.md", buf.String())
}

// GetCampaign retrieves a campaign by name
func (s *CampaignService) GetCampaign(name string) (*domain.Campaign, error) {
	return s.campaignRepo.Read(name)
}

// ListCampaigns returns all campaigns
func (s *CampaignService) ListCampaigns() ([]domain.CampaignSummary, error) {
	return s.campaignRepo.List()
}

// SaveLore saves lore to a campaign
func (s *CampaignService) SaveLore(campaignID, content string) error {
	return s.saveMarkdownFile(campaignID, "", "lore.md", content)
}

// SaveNPCs saves NPCs to a campaign as markdown and syncs to canon.json + JSON files
func (s *CampaignService) SaveNPCs(campaignID, content string) error {
	if !s.campaignRepo.Exists(campaignID) {
		return fmt.Errorf("campaign not found: %s", campaignID)
	}

	// 1. Parse markdown to extract entities
	parser := NewEntityParser()
	result, err := parser.ParseNPCs(content, campaignID)
	if err != nil {
		return fmt.Errorf("failed to parse NPCs from markdown: %w", err)
	}

	// Validate
	if len(result.NPCs) == 0 && len(result.Factions) == 0 {
		return fmt.Errorf("no NPCs or factions found in markdown - expected format: ## Name followed by - **Name** — description")
	}

	// 2. Write markdown (atomic: temp + rename)
	dir := filepath.Join(s.baseDir, campaignID, "npcs")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	tempPath := filepath.Join(dir, ".npcs_and_factions.md.tmp")
	finalPath := filepath.Join(dir, "npcs_and_factions.md")

	if err := os.WriteFile(tempPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write markdown: %w", err)
	}

	// 3. Sync to canon.json
	if err := s.syncCanonEntities(campaignID, result.NPCs, nil, result.Factions); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to sync canon: %w", err)
	}

	// 4. Write JSON files
	if err := s.writeNPCJSONFiles(campaignID, result.NPCs); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to write NPC JSON files: %w", err)
	}

	// 5. Atomic rename
	if err := os.Rename(tempPath, finalPath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to finalize markdown write: %w", err)
	}

	return nil
}

// SaveEncounters saves encounters to a campaign as markdown
func (s *CampaignService) SaveEncounters(campaignID, content string) error {
	if !s.campaignRepo.Exists(campaignID) {
		return fmt.Errorf("campaign not found: %s", campaignID)
	}

	// 1. Parse markdown to extract encounters
	parser := NewEntityParser()
	encounters, err := parser.ParseEncounters(content, campaignID)
	if err != nil {
		return fmt.Errorf("failed to parse encounters from markdown: %w", err)
	}

	// Validate
	if len(encounters) == 0 {
		return fmt.Errorf("no encounters found in markdown - expected format: ## Encuentro X: Nombre")
	}

	// 2. Write markdown (atomic: temp + rename)
	dir := filepath.Join(s.baseDir, campaignID, "encounters")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	tempPath := filepath.Join(dir, ".encounters.md.tmp")
	finalPath := filepath.Join(dir, "encounters.md")

	if err := os.WriteFile(tempPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write markdown: %w", err)
	}

	// 3. Atomic rename
	if err := os.Rename(tempPath, finalPath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to finalize markdown write: %w", err)
	}

	return nil
}

// SaveBestiary saves bestiary to a campaign as markdown and syncs to canon.json + JSON files
func (s *CampaignService) SaveBestiary(campaignID, content string) error {
	if !s.campaignRepo.Exists(campaignID) {
		return fmt.Errorf("campaign not found: %s", campaignID)
	}

	// 1. Parse markdown to extract monsters
	parser := NewEntityParser()
	monsters, err := parser.ParseMonsters(content, campaignID)
	if err != nil {
		return fmt.Errorf("failed to parse monsters from markdown: %w", err)
	}

	// Validate
	if len(monsters) == 0 {
		return fmt.Errorf("no monsters found in markdown - expected format: # Monster Name followed by stats")
	}

	// 2. Write markdown (atomic: temp + rename)
	dir := filepath.Join(s.baseDir, campaignID, "bestiary")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	tempPath := filepath.Join(dir, ".bestiary.md.tmp")
	finalPath := filepath.Join(dir, "bestiary.md")

	if err := os.WriteFile(tempPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write markdown: %w", err)
	}

	// 3. Sync to canon.json
	if err := s.syncCanonEntities(campaignID, nil, monsters, nil); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to sync canon: %w", err)
	}

	// 4. Write JSON files
	if err := s.writeMonsterJSONFiles(campaignID, monsters); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to write monster JSON files: %w", err)
	}

	// 5. Atomic rename
	if err := os.Rename(tempPath, finalPath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to finalize markdown write: %w", err)
	}

	return nil
}

// SaveAct saves an act to a campaign
func (s *CampaignService) SaveAct(campaignID string, number int, title, content string) error {
	if !s.campaignRepo.Exists(campaignID) {
		return fmt.Errorf("campaign not found: %s", campaignID)
	}

	act := &domain.Act{
		CampaignID:        campaignID,
		Number:            number,
		Title:             title,
		Content:           content,
		GameMode:          "investigacion",
		ChapterObjectives: []string{"Completar objetivos del capítulo", "Avanzar la trama principal"},
		EstimatedDuration: "2-3 sesiones",
		Tone:              "heroic",
		RunningGuidance:   "This chapter presents an exciting and memorable adventure for the player characters and establishes the fundamental foundation for the ongoing development of the entire campaign. The Dungeon Master must carefully guide the players through all planned encounters, constantly ensuring that each individual player has the opportunity to shine and contribute meaningfully to the group. The main objectives of the chapter must be fully completed before proceeding to the next chapter of the full adventure. If characters get stuck at any critical point, provide useful additional hints through allied NPCs or well-designed chance encounters. Maintain the overall session rhythm perfectly balanced between challenging tactical combat, detailed environment exploration, and meaningful social interaction with NPCs. Always make sure that each session ends with an important narrative hook that powerfully motivates the players to continue in the next session. Important decisions made by the characters in this chapter will have significant and lasting consequences in the following chapters of the entire campaign. Prior preparation includes reviewing maps, handouts, and NPC stats. The suggested timing divides the chapter into opening, exploration, main encounter, and resolution. Character hooks should target specific backgrounds and classes. Critical decisions have branching points with immediate and long-term consequences. Combats should be adjusted in real time based on group performance. Transitions and closing should include a hook to the next act. Safety notes remind tools like X-Card and specific shock points. For material preparation, the DM should have battle maps ready with visible grid, handouts printed on quality paper, NPC stats on separate cards, and the player map of the area clearly visible. The pacing of the chapter should be dynamic and adapt to the energy of the group, alternating moments of high tension with calmer moments of roleplay and exploration. Remember to ask for feedback at the end of the session about what worked well and what could be improved, and to take notes for the next session. The end of the chapter should provide a clear sense of accomplishment while opening new threads of intrigue that propel the campaign forward. Beyond these practical guidelines, the DM should also pay close attention to the emotional beats of the chapter, ensuring that moments of triumph feel earned and moments of loss resonate with the table. The narrative should respect the agency of the players, presenting meaningful choices with clear consequences rather than railroading them toward a predetermined outcome. When designing or running encounters, the DM should consider the strengths and weaknesses of the party composition, adjusting difficulty as needed to keep encounters engaging without being either trivial or overwhelming. Treasure and rewards should be tailored to the party's needs and interests, advancing both their mechanical power and their narrative investment. NPCs should have distinct voices, motivations, and mannerisms that make them memorable and easy to differentiate at the table. Foreshadowing and callbacks to earlier events can deepen the players' sense that their actions matter and that the world is alive and responsive. The DM should also be prepared to improvise when players take unexpected actions, using the established lore and NPCs to weave their choices back into the narrative rather than shutting them down. Finally, the chapter should end on a note that balances closure with anticipation, giving players a chance to celebrate their achievements while hinting at the larger challenges and mysteries that await in the next chapter. This balance of preparation, flexibility, attention to player experience, and respect for narrative consequences is what makes a D&D chapter not just a sequence of encounters, but a story that the table tells together and remembers long after the session ends. The DM's role is not just to referee rules, but to be the primary architect of shared experience, the curator of tone, and the guardian of fun. By keeping these principles in mind, the DM will deliver a chapter that feels cohesive, fair, dramatic, and deeply satisfying to all participants, setting the stage for everything that follows in the campaign. Throughout the session, the DM should also be mindful of the table's energy and engagement levels, taking breaks when needed, and adjusting the pace of the game to match the group's mood. Some nights will be filled with intense combat and tactical decision-making, while others will lean toward social intrigue, exploration, or character development. A skilled DM reads the room and adapts, knowing when to press forward with a tense encounter and when to slow down for a quiet moment of roleplay. The best sessions feel collaborative, with players and DM building on each other's contributions to create something none of them could have envisioned alone. By fostering this collaborative spirit, the DM helps ensure that the campaign remains a living, evolving story that continues to surprise and delight everyone at the table, session after session, chapter after chapter, all the way to its epic conclusion.",
		AssetHandoff:      "The events of this chapter lead directly to the next",
	}

	return s.actRepo.Save(act)
}

func (s *CampaignService) saveMarkdownFile(campaignID, subdir, filename, content string) error {
	if !s.campaignRepo.Exists(campaignID) {
		return fmt.Errorf("campaign not found: %s", campaignID)
	}

	dir := filepath.Join(s.baseDir, campaignID, subdir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// SaveChapter saves a self-contained chapter to a campaign.
// Writes to campaigns/{id}/chapters/chapter_{NN}.md
func (s *CampaignService) SaveChapter(campaignID string, chapterNum int, title, content string) error {
	if !s.campaignRepo.Exists(campaignID) {
		return fmt.Errorf("campaign not found: %s", campaignID)
	}

	// 1. Parse markdown to extract inline entities
	parser := NewEntityParser()
	result, err := parser.ParseChapter(content, campaignID, chapterNum)
	if err != nil {
		return fmt.Errorf("failed to parse chapter: %w", err)
	}

	// Validate: at least one area found
	if len(result.Areas) == 0 {
		return fmt.Errorf("no areas found in chapter - expected format: ### Área X: Nombre")
	}

	// 2. Write markdown file (atomic: temp + rename)
	dir := filepath.Join(s.baseDir, campaignID, "chapters")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	filename := fmt.Sprintf("chapter_%02d.md", chapterNum)
	tempPath := filepath.Join(dir, "."+filename+".tmp")
	finalPath := filepath.Join(dir, filename)

	if err := os.WriteFile(tempPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write markdown: %w", err)
	}

	// 3. Sync inline NPCs to canon.json
	if len(result.NPCs) > 0 {
		if err := s.syncCanonEntities(campaignID, result.NPCs, nil, nil); err != nil {
			_ = os.Remove(tempPath)
			return fmt.Errorf("failed to sync canon: %w", err)
		}
	}

	// 4. Atomic rename
	if err := os.Rename(tempPath, finalPath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to finalize markdown write: %w", err)
	}

	return nil
}

// SaveAreas saves areas to a campaign as markdown
func (s *CampaignService) SaveAreas(campaignID, chapterID, content string) error {
	if !s.campaignRepo.Exists(campaignID) {
		return fmt.Errorf("campaign not found: %s", campaignID)
	}

	// 1. Parse markdown to extract areas
	parser := NewEntityParser()
	areas, err := parser.ParseAreas(content, campaignID, chapterID)
	if err != nil {
		return fmt.Errorf("failed to parse areas from markdown: %w", err)
	}

	// Validate: at least one area found
	if len(areas) == 0 {
		return fmt.Errorf("no areas found in markdown - expected format: ### Área X: Nombre")
	}

	// 2. Write markdown file (atomic: temp + rename)
	dir := filepath.Join(s.baseDir, campaignID, "areas")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	filename := fmt.Sprintf("%s.md", chapterID)
	tempPath := filepath.Join(dir, "."+filename+".tmp")
	finalPath := filepath.Join(dir, filename)

	if err := os.WriteFile(tempPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write markdown: %w", err)
	}

	// 3. Atomic rename: temp → final
	if err := os.Rename(tempPath, finalPath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to finalize markdown write: %w", err)
	}

	return nil
}

// SaveMaps saves maps to a campaign as markdown
func (s *CampaignService) SaveMaps(campaignID, content string) error {
	return s.saveMarkdownFile(campaignID, "maps", "maps.md", content)
}

// SavePrologue saves a prologue to a campaign as markdown
func (s *CampaignService) SavePrologue(campaignID string, prologue *domain.Prologue) error {
	tmplStr, err := compiler.GetTemplate("prologue")
	if err != nil {
		return err
	}

	tmpl, err := template.New("prologue").Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("failed to parse prologue template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, prologue); err != nil {
		return fmt.Errorf("failed to execute prologue template: %w", err)
	}

	return s.saveMarkdownFile(campaignID, "", "prologue.md", buf.String())
}

// SaveIntroduction saves the campaign introduction/overview document
func (s *CampaignService) SaveIntroduction(campaignID, content string) error {
	return s.saveMarkdownFile(campaignID, "", "introduction.md", content)
}

// SaveSettingGuide saves the campaign setting guide (DM-only)
func (s *CampaignService) SaveSettingGuide(campaignID, content string) error {
	return s.saveMarkdownFile(campaignID, "", "setting-guide.md", content)
}

// SaveAppendices saves the campaign appendices (items, monsters, handouts)
func (s *CampaignService) SaveAppendices(campaignID, content string) error {
	return s.saveMarkdownFile(campaignID, "", "appendices.md", content)
}

// CompilePDF compiles campaign to PDF
func (s *CampaignService) CompilePDF(ctx context.Context, campaignID, title string) (string, error) {
	if !s.campaignRepo.Exists(campaignID) {
		return "", fmt.Errorf("campaign not found: %s", campaignID)
	}

	if title == "" {
		campaign, err := s.campaignRepo.Read(campaignID)
		if err != nil {
			title = campaignID
		} else {
			title = campaign.Title
		}
	}

	dir := filepath.Join(s.baseDir, campaignID)
	comp := compiler.New(dir, s.pdfEngine)
	return comp.Compile(ctx, title)
}

// GetTemplate returns a template by type
func (s *CampaignService) GetTemplate(tmplType string) (string, error) {
	return compiler.GetTemplate(tmplType)
}

// CampaignState returns the complete state of a campaign
func (s *CampaignService) CampaignState(campaignID string) (*domain.CampaignState, error) {
	if !s.campaignRepo.Exists(campaignID) {
		return nil, fmt.Errorf("campaign not found: %s", campaignID)
	}

	campaign, err := s.campaignRepo.Read(campaignID)
	if err != nil {
		return nil, err
	}

	acts, _ := s.actRepo.List(campaignID)
	characters, _ := s.charRepo.List(campaignID)
	npcs, _ := s.npcRepo.List(campaignID)
	quests, _ := s.questRepo.List(campaignID)

	var activeQuests, completedQuests []domain.Quest
	for _, q := range quests {
		switch q.Status {
		case domain.QuestStatusActive:
			activeQuests = append(activeQuests, q)
		case domain.QuestStatusCompleted:
			completedQuests = append(completedQuests, q)
		}
	}

	return &domain.CampaignState{
		Campaign:        *campaign,
		Acts:            acts,
		Characters:      characters,
		NPCs:            npcs,
		Quests:          quests,
		ActiveQuests:    activeQuests,
		CompletedQuests: completedQuests,
	}, nil
}

// GenerateAndRegisterNPCs generates NPCs from canon and saves them to the campaign
// This is a non-blocking operation - warnings are returned but do not cause failure
func (s *CampaignService) GenerateAndRegisterNPCs(ctx context.Context, campaignID string) ([]domain.NPC, []string, error) {
	generator := generators.NewNPCGenerator(s.canonRepo)
	npcs, warnings, err := generator.GenerateFromCanon(ctx, campaignID)
	if err != nil {
		return nil, warnings, fmt.Errorf("failed to generate NPCs: %w", err)
	}

	if len(npcs) == 0 {
		return npcs, warnings, nil
	}

	// Save each NPC to the repository
	for _, npc := range npcs {
		if saveErr := s.npcRepo.Save(&npc); saveErr != nil {
			warnings = append(warnings, fmt.Sprintf("failed to save NPC %s: %v", npc.Name, saveErr))
			// Continue with other NPCs - non-blocking
		}
	}

	return npcs, warnings, nil
}

// GenerateAndRegisterMonsters generates monsters from canon and saves them to the campaign
// This is a non-blocking operation - warnings are returned but do not cause failure
func (s *CampaignService) GenerateAndRegisterMonsters(ctx context.Context, campaignID string) ([]domain.Monster, []string, error) {
	generator := generators.NewMonsterGenerator(s.canonRepo)
	monsters, warnings, err := generator.GenerateFromCanon(ctx, campaignID)
	if err != nil {
		return nil, warnings, fmt.Errorf("failed to generate monsters: %w", err)
	}

	if len(monsters) == 0 {
		return monsters, warnings, nil
	}

	// Save each monster to the repository
	for _, monster := range monsters {
		if saveErr := s.monsterRepo.Save(&monster); saveErr != nil {
			warnings = append(warnings, fmt.Sprintf("failed to save monster %s: %v", monster.Name, saveErr))
			// Continue with other monsters - non-blocking
		}
	}

	return monsters, warnings, nil
}

// syncCanonEntities updates canon.json with parsed entities (upsert by ID)
func (s *CampaignService) syncCanonEntities(campaignID string, npcs []domain.NPC, monsters []domain.Monster, factions []domain.Faction) error {
	// Load existing canon document
	canonDoc, err := s.canonRepo.Load(campaignID)
	if err != nil {
		// Create new canon if doesn't exist
		canonDoc = &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    campaignID,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			Entities:      []domain.CanonEntity{},
		}
	}

	// Create index for O(1) upsert
	entityIndex := make(map[string]int)
	for i, e := range canonDoc.Entities {
		entityIndex[e.ID] = i
	}

	// Upsert NPCs
	for _, npc := range npcs {
		entity := npcToCanonEntity(npc)
		if idx, exists := entityIndex[entity.ID]; exists {
			canonDoc.Entities[idx] = entity
		} else {
			canonDoc.Entities = append(canonDoc.Entities, entity)
			entityIndex[entity.ID] = len(canonDoc.Entities) - 1
		}
	}

	// Upsert Monsters
	for _, monster := range monsters {
		entity := monsterToCanonEntity(monster)
		if idx, exists := entityIndex[entity.ID]; exists {
			canonDoc.Entities[idx] = entity
		} else {
			canonDoc.Entities = append(canonDoc.Entities, entity)
			entityIndex[entity.ID] = len(canonDoc.Entities) - 1
		}
	}

	// Upsert Factions
	for _, faction := range factions {
		entity := factionToCanonEntity(faction)
		if idx, exists := entityIndex[entity.ID]; exists {
			canonDoc.Entities[idx] = entity
		} else {
			canonDoc.Entities = append(canonDoc.Entities, entity)
			entityIndex[entity.ID] = len(canonDoc.Entities) - 1
		}
	}

	canonDoc.UpdatedAt = time.Now()
	return s.canonRepo.Save(campaignID, canonDoc)
}

// writeNPCJSONFiles writes individual JSON files for each NPC
func (s *CampaignService) writeNPCJSONFiles(campaignID string, npcs []domain.NPC) error {
	dir := filepath.Join(s.baseDir, campaignID, "npcs")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create npcs directory: %w", err)
	}

	for _, npc := range npcs {
		path := filepath.Join(dir, sanitizeFilename(npc.Name)+".json")
		if err := writeJSON(path, npc); err != nil {
			return fmt.Errorf("failed to save NPC %s: %w", npc.Name, err)
		}
	}
	return nil
}

// writeMonsterJSONFiles writes individual JSON files for each monster
func (s *CampaignService) writeMonsterJSONFiles(campaignID string, monsters []domain.Monster) error {
	dir := filepath.Join(s.baseDir, campaignID, "monsters")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create monsters directory: %w", err)
	}

	for _, monster := range monsters {
		path := filepath.Join(dir, sanitizeFilename(monster.Name)+".json")
		if err := writeJSON(path, monster); err != nil {
			return fmt.Errorf("failed to save monster %s: %w", monster.Name, err)
		}
	}
	return nil
}

// npcToCanonEntity converts an NPC to a CanonEntity
func npcToCanonEntity(npc domain.NPC) domain.CanonEntity {
	properties := map[string]any{
		"description": npc.Description,
		"faction":     npc.Faction,
		"role":        npc.Role,
	}
	if npc.Stats != nil {
		properties["hp"] = npc.Stats.HP
		properties["ac"] = npc.Stats.AC
	}

	return domain.CanonEntity{
		ID:          npc.ID,
		Name:        npc.Name,
		Type:        domain.EntityTypeNPC,
		Role:        npc.Role,
		CanonState:  domain.EntityStateAlive,
		Properties:  properties,
		Connections: []string{},
	}
}

// monsterToCanonEntity converts a Monster to a CanonEntity
func monsterToCanonEntity(monster domain.Monster) domain.CanonEntity {
	properties := map[string]any{
		"description": monster.Description,
		"cr":          monster.CR,
		"type":        monster.Type,
		"size":        monster.Size,
	}
	if monster.Stats.AC > 0 {
		properties["ac"] = monster.Stats.AC
	}
	if monster.Stats.HP > 0 {
		properties["hp"] = monster.Stats.HP
	}

	return domain.CanonEntity{
		ID:          monster.ID,
		Name:        monster.Name,
		Type:        domain.EntityTypeMonster,
		Role:        "monster",
		CanonState:  domain.EntityStateAlive,
		Properties:  properties,
		Connections: []string{},
	}
}

// factionToCanonEntity converts a Faction to a CanonEntity
func factionToCanonEntity(faction domain.Faction) domain.CanonEntity {
	properties := map[string]any{
		"description": faction.Description,
		"agenda":      faction.Agenda,
		"contact_npc": faction.ContactNPC,
		"tier":        faction.Tier,
		"is_secret":   faction.IsSecret,
	}
	if len(faction.Allies) > 0 {
		properties["allies"] = faction.Allies
	}
	if len(faction.Enemies) > 0 {
		properties["enemies"] = faction.Enemies
	}

	return domain.CanonEntity{
		ID:          faction.ID,
		Name:        faction.Name,
		Type:        domain.EntityTypeFaction,
		Role:        "faction",
		CanonState:  domain.EntityStateAlive,
		Properties:  properties,
		Connections: append(faction.Allies, faction.Enemies...),
	}
}

// sanitizeFilename converts a name to a safe filename
func sanitizeFilename(name string) string {
	id := strings.ToLower(name)
	id = strings.ReplaceAll(id, " ", "-")
	id = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(id, "")
	id = regexp.MustCompile(`-+`).ReplaceAllString(id, "-")
	id = strings.Trim(id, "-")
	return id
}

// writeJSON writes a value to a JSON file
func writeJSON(path string, value interface{}) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}
