package services

import (
	"context"
	"fmt"
	"github.com/pauvalls/grimorio/internal/domain"
)

// SessionZeroService generates Session Zero guides and character worksheets.
type SessionZeroService struct {
	sessionRepo SessionZeroRepository
}

// SessionZeroRepository defines repository interface for Session Zero guides.
type SessionZeroRepository interface {
	Create(ctx context.Context, campaignID string, guide *domain.SessionZeroGuide) error
	Read(ctx context.Context, campaignID string) (*domain.SessionZeroGuide, error)
	Update(ctx context.Context, campaignID string, guide *domain.SessionZeroGuide) error
	Delete(ctx context.Context, campaignID string) error
}

// NewSessionZeroService creates a new SessionZeroService.
func NewSessionZeroService(sessionRepo SessionZeroRepository) *SessionZeroService {
	return &SessionZeroService{sessionRepo: sessionRepo}
}

// GenerateGuide generates a complete Session Zero guide for a campaign.
func (s *SessionZeroService) GenerateGuide(ctx context.Context, campaignID string, campaignName string, tone string, themes []string) (*domain.SessionZeroGuide, error) {
	guide := &domain.SessionZeroGuide{
		CampaignID:    campaignID,
		CampaignPitch: generateCampaignPitch(campaignName, tone),
		Tone:          tone,
		Themes:        themes,
		ContentWarnings: generateContentWarnings(themes),
		SessionExpectations: domain.SessionExpectations{
			SessionLength:   "3-4 hours",
			Frequency:       "Weekly",
			AttendancePolicy: "24-hour notice required",
			TableEtiquette:  []string{"Be respectful", "No phones at table", "Collaborative storytelling"},
		},
		SafetyTools:       generateSafetyTools(),
		CharacterCreation: generateCharacterCreationGuide(),
		HouseRules:        generateHouseRules(),
		Agenda:            generateSessionAgenda(),
	}

	if err := guide.Validate(); err != nil {
		return nil, fmt.Errorf("generated Session Zero guide validation failed: %w", err)
	}

	return guide, nil
}

// GenerateCharacterWorksheet generates a character creation worksheet.
func (s *SessionZeroService) GenerateCharacterWorksheet(ctx context.Context, campaignID string) (*domain.CharacterCreationGuide, error) {
	guide := generateCharacterCreationGuide()
	return &guide, nil
}

// GetContentWarnings derives content warnings from campaign themes.
func (s *SessionZeroService) GetContentWarnings(ctx context.Context, campaignID string, themes []string) ([]domain.ContentWarning, error) {
	return generateContentWarnings(themes), nil
}

// ValidateSafetyTools checks if safety tools have proper instructions.
func (s *SessionZeroService) ValidateSafetyTools(ctx context.Context, tools []domain.SafetyTool) (bool, error) {
	for _, tool := range tools {
		if tool.HowToUse == "" {
			return false, fmt.Errorf("safety tool %s missing usage instructions", tool.Name)
		}
	}
	return true, nil
}

// Helper functions

func generateCampaignPitch(name, tone string) string {
	return fmt.Sprintf("Join the adventure in %s, a %s campaign filled with danger, mystery, and heroism.", name, tone)
}

func generateContentWarnings(themes []string) []domain.ContentWarning {
	warnings := []domain.ContentWarning{}

	warningMap := map[string]domain.ContentWarning{
		"horror": {
			Topic:       "Horror elements",
			Severity:    "moderate",
			Description: "Includes psychological horror and disturbing imagery",
			Avoidable:   true,
		},
		"violence": {
			Topic:       "Violence",
			Severity:    "moderate",
			Description: "Combat and fantasy violence",
			Avoidable:   false,
		},
		"dark themes": {
			Topic:       "Dark themes",
			Severity:    "intense",
			Description: "Explores moral ambiguity and difficult choices",
			Avoidable:   true,
		},
	}

	for _, theme := range themes {
		if warning, ok := warningMap[theme]; ok {
			warnings = append(warnings, warning)
		}
	}

	return warnings
}

func generateSafetyTools() []domain.SafetyTool {
	return []domain.SafetyTool{
		{
			Name:        "X-Card",
			Description: "A tool to discreetly indicate discomfort",
			HowToUse:    "Tap the X-card to skip or fade-to-black uncomfortable content",
		},
		{
			Name:        "Lines and Veils",
			Description: "Establish boundaries before play begins",
			HowToUse:    "Discuss and agree on content to exclude (lines) or fade-to-black (veils)",
		},
	}
}

func generateCharacterCreationGuide() domain.CharacterCreationGuide {
	return domain.CharacterCreationGuide{
		Level:              1,
		AbilityScoreMethod: "4d6 drop lowest",
		AllowedSources:     []string{"PHB", "XGtE", "TCoE"},
		BannedSources:      []string{},
		BackgroundGuidance: "Choose a background that ties you to the campaign world",
		BondPrompts: []string{
			"What connects you to this region?",
			"Who do you care about most?",
			"What debt do you owe?",
		},
		IdealPrompts: []string{
			"What principle guides your actions?",
			"What would you die for?",
		},
		FlawPrompts: []string{
			"What weakness could enemies exploit?",
			"What secret are you hiding?",
		},
		PartyCohesionQuestions: []string{
			"How did you meet the other party members?",
			"What goal unites you?",
		},
	}
}

func generateHouseRules() []domain.HouseRule {
	return []domain.HouseRule{
		{
			Name:        "Flanking",
			Description: "Advantage on melee attacks when flanking an enemy",
			Reason:      "Rewards tactical positioning",
		},
		{
			Name:        "Potion Bonus Action",
			Description: "Drinking a potion uses a bonus action",
			Reason:      "Makes potions more viable in combat",
		},
	}
}

func generateSessionAgenda() []domain.SessionAgendaItem {
	return []domain.SessionAgendaItem{
		{Topic: "Welcome & Introductions", DurationMinutes: 15, Description: "Meet the group"},
		{Topic: "Campaign Pitch", DurationMinutes: 20, Description: "DM presents the campaign"},
		{Topic: "Session Zero Guide", DurationMinutes: 20, Description: "Review expectations and safety tools"},
		{Topic: "Character Creation", DurationMinutes: 45, Description: "Create characters together"},
		{Topic: "Party Cohesion", DurationMinutes: 20, Description: "Establish party bonds"},
		{Topic: "Schedule & Wrap-up", DurationMinutes: 10, Description: "Set meeting times"},
	}
}
