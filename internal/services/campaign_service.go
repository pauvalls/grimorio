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
		RunningGuidance:   "Este capítulo presenta una aventura emocionante y memorable para los personajes jugadores y establece las bases fundamentales para el desarrollo continuo de toda la campaña. El Dungeon Master debe guiar cuidadosamente a los jugadores a través de todos los encuentros planificados, asegurándose constantemente de que cada jugador individual tenga oportunidad de brillar y contribuir significativamente al grupo. Los objetivos principales del capítulo deben cumplirse completamente antes de proceder al siguiente capítulo de la aventura completa. Si los personajes se estancan en algún punto crítico, proporciona pistas adicionales útiles a través de NPCs aliados o encuentros fortuitos bien diseñados. Mantén el ritmo general de la sesión perfectamente equilibrado entre combate táctico desafiante, exploración detallada del entorno e interacción social significativa con NPCs. Asegúrate siempre de que cada sesión termine con un gancho narrativo importante que motive poderosamente a los jugadores a continuar en la siguiente sesión. Las decisiones importantes tomadas por los personajes en este capítulo tendrán consecuencias significativas y duraderas en los capítulos siguientes de toda la campaña. Preparación previa incluye revisar mapas, handouts, y stats de NPCs. El timing sugerido divide el capítulo en apertura, exploración, encuentro principal y resolución. Los ganchos de personaje deben targetear backgrounds y clases específicas. Las decisiones críticas tienen puntos de ramificación con consecuencias inmediatas y a largo plazo. Los combates deben ajustarse en tiempo real según el desempeño del grupo. Las transiciones y cierre deben incluir un gancho al siguiente acto. Las notas de seguridad recuerdan herramientas como X-Card y puntos de shock específicos. Para la preparación de materiales, el DM debe tener listos los mapas de batalla con cuadrícula visible, los handouts impresos en papel de calidad, las estadísticas de NPCs en fichas separadas, y la playlist de música ambiental organizada por escenas. El repaso de consecuencias de sesiones anteriores es crítico: si los PJs hicieron aliados, estos aparecen con recursos; si hicieron enemigos, estos preparan emboscadas. El estado inicial del mundo refleja las acciones previas: facciones en diferentes niveles de alerta, economía afectada por eventos, NPCs con actitudes cambiadas. Durante la apertura, establecé el tono inmediatamente con descripciones sensoriales ricas. En exploración, incentivá la creatividad con múltiples soluciones. En el encuentro principal, describí acciones cinemáticamente. En resolución, siempre dejá un gancho. Los ajustes por grupo son obligatorios: 2-3 PJs reducen enemigos y CDs, 5-6 PJs agregan enemigos y acciones legendarias. El tesoro debe ser significativo pero no romper la economía. La transición al siguiente capítulo debe sentir como consecuencia natural, no como corte abrupto. La gestión del tiempo es esencial: si la sesión avanza rápido, agregá escenas opcionales de relleno; si va lenta, cortá descripciones menores y combates con minions. Los NPCs principales deben tener motivaciones claras y reaccionar consistentemente a las acciones de los PJs. Las facciones tienen agendas propias que continúan independientemente de la intervención de los jugadores. El ambiente se construye con detalles sensoriales: sonidos, olores, texturas, temperaturas. Las pistas deben ser redundantes: si una se pierde, hay otra vía. Los encuentros de combate deben tener terreno interesante, objetivos secundarios, y posibles resoluciones no violentas. La progresión de nivel debe sentirse merecida, no automática. Los secretos mejor guardados revelan información que cambia la comprensión de la trama. Los aliados verdaderos ayudan sin resolver todo por los PJs. Los villanos tienen razones que creen justificadas, no son malvados por deporte. La coherencia interna del mundo es fundamental: las reglas mágicas, las distancias de viaje, los tiempos de recuperación deben mantenerse consistentes. Los jugadores recuerdan inconsistencias y pierden inmersión. Documentá cambios importantes en el estado del mundo después de cada sesión. Las descripciones deben apelar a múltiples sentidos simultáneamente. Los diálogos de NPCs deben tener voz distintiva y vocabulario coherente con su background. Los combates épicos requieren descripciones que capturen la escala y el peligro. Los momentos tranquilos permiten desarrollo de personaje y roleplay profundo. El humor alivia tensión pero no debe socavar momentos serios. El pacing alterna intensidad alta y baja para evitar fatiga. La música ambiental refuerza el tono emocional de cada escena. Los efectos de sonido bien timingados crean impacto dramático. Las pausas silenciosas generan anticipación y tensión. El lenguaje corporal de NPCs comunica emociones sin palabras. Los objetos del entorno cuentan historias de eventos pasados. Recordá siempre: tu preparación es invisible para los jugadores, pero tu flexibilidad determina la calidad de su experiencia.",
		AssetHandoff:      "Los eventos de este capítulo conducen directamente al siguiente",
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
