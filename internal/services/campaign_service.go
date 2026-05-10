package services

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/pauvalls/grimorio/internal/compiler"
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// CampaignService handles campaign business logic
type CampaignService struct {
	campaignRepo repository.CampaignRepository
	actRepo      repository.ActRepository
	charRepo     repository.CharacterRepository
	npcRepo      repository.NPCRepository
	questRepo    repository.QuestRepository
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
	baseDir string,
	pdfEngine string,
) *CampaignService {
	return &CampaignService{
		campaignRepo: campaignRepo,
		actRepo:      actRepo,
		charRepo:     charRepo,
		npcRepo:      npcRepo,
		questRepo:    questRepo,
		baseDir:      baseDir,
		pdfEngine:    pdfEngine,
	}
}

// CreateCampaign creates a new campaign
func (s *CampaignService) CreateCampaign(name, title, setting string) (*domain.Campaign, error) {
	if title == "" {
		title = name
	}

	campaign := &domain.Campaign{
		Name:    name,
		Title:   title,
		Setting: setting,
		Status:  "active",
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

// SaveLore saves lore to a campaign
func (s *CampaignService) SaveLore(campaignID, content string) error {
	return s.saveMarkdownFile(campaignID, "", "lore.md", content)
}

// SaveNPCs saves NPCs to a campaign as markdown
func (s *CampaignService) SaveNPCs(campaignID, content string) error {
	return s.saveMarkdownFile(campaignID, "npcs", "npcs_and_factions.md", content)
}

// SaveEncounters saves encounters to a campaign as markdown
func (s *CampaignService) SaveEncounters(campaignID, content string) error {
	return s.saveMarkdownFile(campaignID, "encounters", "encounters.md", content)
}

// SaveBestiary saves bestiary to a campaign as markdown
func (s *CampaignService) SaveBestiary(campaignID, content string) error {
	return s.saveMarkdownFile(campaignID, "bestiary", "bestiary.md", content)
}

// SaveMaps saves maps to a campaign as markdown
func (s *CampaignService) SaveMaps(campaignID, content string) error {
	return s.saveMarkdownFile(campaignID, "maps", "maps.md", content)
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
