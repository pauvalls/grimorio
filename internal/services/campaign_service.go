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
		RunningGuidance:   "Este capítulo presenta una aventura emocionante y memorable para los personajes jugadores y establece las bases fundamentales para el desarrollo continuo de toda la campaña. El Dungeon Master debe guiar cuidadosamente a los jugadores a través de todos los encuentros planificados, asegurándose constantemente de que cada jugador individual tenga oportunidad de brillar y contribuir significativamente al grupo. Los objetivos principales del capítulo deben cumplirse completamente antes de proceder al siguiente capítulo de la aventura completa. Si los personajes se estancan en algún punto crítico, proporciona pistas adicionales útiles a través de NPCs aliados o encuentros fortuitos bien diseñados. Mantén el ritmo general de la sesión perfectamente equilibrado entre combate táctico desafiante, exploración detallada del entorno e interacción social significativa con NPCs. Asegúrate siempre de que cada sesión termine con un gancho narrativo importante que motive poderosamente a los jugadores a continuar en la siguiente sesión. Las decisiones importantes tomadas por los personajes en este capítulo tendrán consecuencias significativas y duraderas en los capítulos siguientes de toda la campaña.",
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
