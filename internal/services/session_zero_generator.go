package services

import (
	"context"
	"fmt"
	"github.com/pauvalls/grimorio/internal/domain"
)

// SessionZeroGenerator generates Session Zero guides.
type SessionZeroGenerator struct {
	sessionRepo SessionZeroRepository
}

// NewSessionZeroGenerator creates a new SessionZeroGenerator.
func NewSessionZeroGenerator(sessionRepo SessionZeroRepository) *SessionZeroGenerator {
	return &SessionZeroGenerator{sessionRepo: sessionRepo}
}

// GenerateGuide generates a complete Session Zero guide.
func (s *SessionZeroGenerator) GenerateGuide(ctx context.Context, campaignID, campaignName, tone string, themes []string) (*domain.SessionZeroGuide, error) {
	guide := &domain.SessionZeroGuide{
		CampaignID:    campaignID,
		CampaignPitch: generateCampaignPitch(campaignName, tone),
		Tone:          tone,
		Themes:        themes,
		ContentWarnings: s.generateContentWarnings(themes),
		SessionExpectations: domain.SessionExpectations{
			SessionLength:    "3-4 hours",
			Frequency:        "Weekly",
			AttendancePolicy: "24-hour notice required",
			TableEtiquette:   []string{"Be respectful", "No phones at table", "Collaborative storytelling"},
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
func (s *SessionZeroGenerator) GenerateCharacterWorksheet(ctx context.Context, campaignID string, npcs []string, factions []string) (*domain.CharacterCreationGuide, error) {
	guide := generateCharacterCreationGuide()
	
	// Generate bond prompts tied to campaign NPCs
	if len(npcs) > 0 {
		guide.BondPrompts = append(guide.BondPrompts, 
			fmt.Sprintf("Which NPC (%s) do you have a connection with?", joinStrings(npcs, ", ")))
	}
	
	// Generate ideal prompts with conflict potential
	guide.IdealPrompts = append(guide.IdealPrompts,
		"What principle would you compromise your safety for?")
	
	// Generate flaw prompts (exploitable)
	guide.FlawPrompts = append(guide.FlawPrompts,
		"What secret could your enemies use against you?")
	
	// Party cohesion questions
	guide.PartyCohesionQuestions = []string{
		"How did you meet the other party members?",
		"What goal unites you?",
		"What would make you leave the party?",
	}

	return &guide, nil
}

// GetContentWarnings derives content warnings from campaign themes.
func (s *SessionZeroGenerator) GetContentWarnings(ctx context.Context, themes []string) []domain.ContentWarning {
	return s.generateContentWarnings(themes)
}

// ValidateSafetyTools checks if safety tools have proper instructions.
func (s *SessionZeroGenerator) ValidateSafetyTools(tools []domain.SafetyTool) (bool, error) {
	for _, tool := range tools {
		if tool.HowToUse == "" {
			return false, fmt.Errorf("safety tool %s missing usage instructions", tool.Name)
		}
	}
	return true, nil
}

func (s *SessionZeroGenerator) generateContentWarnings(themes []string) []domain.ContentWarning {
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
		"gore": {
			Topic:       "Gore",
			Severity:    "intense",
			Description: "Graphic descriptions of injuries and death",
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

func generateCampaignPitch(name, tone string) string {
	return fmt.Sprintf("Join the adventure in %s, a %s campaign filled with danger, mystery, and heroism.", name, tone)
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
		{
			Name:        "Script Change",
			Description: "Pause, rewind, or fast-forward scenes",
			HowToUse:    "Use hand signals or verbal cues to control scene pacing",
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
		{
			Name:        "Critical Hit House Rule",
			Description: "Critical hits deal maximum damage plus roll",
			Reason:      "Makes crits more exciting",
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
