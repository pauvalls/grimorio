package services

import (
	"fmt"
	"os"
	"path/filepath"

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

	return campaign, nil
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
		CampaignID: campaignID,
		Number:     number,
		Title:      title,
		Content:    content,
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

// CompilePDF compiles campaign to PDF
func (s *CampaignService) CompilePDF(campaignID, title string) (string, error) {
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
	return comp.Compile(title)
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
