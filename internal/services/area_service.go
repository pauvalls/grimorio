package services

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

// AreaService generates unified WotC-format areas.
type AreaService struct {
	areaRepo AreaRepositoryV3
}

// AreaRepositoryV3 defines repository interface for V3 areas.
type AreaRepositoryV3 interface {
	Create(ctx context.Context, campaignID string, area *domain.Area) error
	Read(ctx context.Context, campaignID string, areaID string) (*domain.Area, error)
	Update(ctx context.Context, campaignID string, area *domain.Area) error
	Delete(ctx context.Context, campaignID string, areaID string) error
	GetByChapter(ctx context.Context, campaignID string, chapterID string) ([]*domain.Area, error)
}

// NewAreaService creates a new AreaService.
func NewAreaService(areaRepo AreaRepositoryV3) *AreaService {
	return &AreaService{areaRepo: areaRepo}
}

// GenerateArea generates a single area with sequential numbering.
func (s *AreaService) GenerateArea(ctx context.Context, chapterID string, areaNumber int, levelRange domain.LevelRange) (*domain.Area, error) {
	if err := domain.ValidateAreaNumber(areaNumber); err != nil {
		return nil, err
	}
	if err := domain.ValidateLevelRange(levelRange); err != nil {
		return nil, err
	}

	area := &domain.Area{
		ID:          fmt.Sprintf("area_%s_%d", chapterID, areaNumber),
		ChapterID:   chapterID,
		AreaNumber:  areaNumber,
		Title:       fmt.Sprintf("Area %d", areaNumber),
		Summary:     "A mysterious location awaits exploration.",
		Description: "This area contains dangers and treasures.",
		LevelRange:  levelRange,
		Features: []domain.AreaFeature{
			{Type: "room", Name: "Main Chamber", Description: "A large room with ancient markings", Hidden: false},
		},
		Encounters: []domain.AreaEncounter{
			{EncounterID: "enc_1", Trigger: "Upon entering", CRTotal: 2.0, XPValue: 450},
		},
		NPCs:        []domain.AreaNPC{},
		Treasure:    []domain.Treasure{},
		Development: "The area remains unchanged after the party leaves.",
		DMSidebars:  []domain.DMSidebar{},
		Maps:        []domain.MapReference{},
	}

	if err := area.Validate(); err != nil {
		return nil, fmt.Errorf("generated area validation failed: %w", err)
	}

	return area, nil
}

// GenerateChapterAreas generates all areas for a chapter (10-15 areas).
func (s *AreaService) GenerateChapterAreas(ctx context.Context, chapterID string, levelRange domain.LevelRange, count int) ([]*domain.Area, error) {
	if count < 10 || count > 15 {
		count = 12 // Default to 12 areas
	}

	areas := make([]*domain.Area, 0, count)
	for i := 1; i <= count; i++ {
		area, err := s.GenerateArea(ctx, chapterID, i, levelRange)
		if err != nil {
			return nil, err
		}
		areas = append(areas, area)
	}

	return areas, nil
}

// GetAreaByNumber retrieves an area by chapter and number.
func (s *AreaService) GetAreaByNumber(ctx context.Context, campaignID string, chapterNumber, areaNumber int) (*domain.Area, error) {
	chapterID := fmt.Sprintf("chapter_%d", chapterNumber)
	areas, err := s.areaRepo.GetByChapter(ctx, campaignID, chapterID)
	if err != nil {
		return nil, err
	}
	for _, area := range areas {
		if area.AreaNumber == areaNumber {
			return area, nil
		}
	}
	return nil, nil
}

// ValidateAreaLevel checks if area encounters are appropriate for party level.
func (s *AreaService) ValidateAreaLevel(ctx context.Context, area *domain.Area, partyLevel int) (bool, error) {
	if area == nil {
		return false, fmt.Errorf("area cannot be nil")
	}
	if partyLevel < area.LevelRange.Min || partyLevel > area.LevelRange.Max {
		return false, nil
	}
	return true, nil
}

// GetAreasByChapter retrieves all areas for a chapter.
func (s *AreaService) GetAreasByChapter(ctx context.Context, campaignID string, chapterID string) ([]*domain.Area, error) {
	return s.areaRepo.GetByChapter(ctx, campaignID, chapterID)
}

// areaTemplate defines a template for area generation
type areaTemplate struct {
	Name           string
	SettingType    string
	Description    string
	FeatureTypes   []string
	EncounterTypes []domain.TableType
	NPCProbabilities map[string]float64 // attitude -> probability
}

// selectTemplate selects a template based on settingType
func (s *AreaService) selectTemplate(settingType string) *areaTemplate {
	templates := map[string]*areaTemplate{
		"wilderness": {
			Name:        "Wilderness Encounter Zone",
			SettingType: "wilderness",
			Description: "An outdoor location with natural hazards and wildlife.",
			FeatureTypes: []string{"room", "hazard", "clue"},
			EncounterTypes: []domain.TableType{domain.TableTypeEncounter, domain.TableTypeWeather},
			NPCProbabilities: map[string]float64{
				"helpful": 0.3,
				"neutral": 0.5,
				"hostile": 0.2,
			},
		},
		"urban": {
			Name:        "Urban District",
			SettingType: "urban",
			Description: "A city or town location with social encounters and shops.",
			FeatureTypes: []string{"room", "clue", "secret"},
			EncounterTypes: []domain.TableType{domain.TableTypeRumor, domain.TableTypeEncounter},
			NPCProbabilities: map[string]float64{
				"helpful": 0.5,
				"neutral": 0.4,
				"hostile": 0.1,
			},
		},
		"dungeon": {
			Name:        "Dungeon Complex",
			SettingType: "dungeon",
			Description: "An indoor location with traps, monsters, and treasure.",
			FeatureTypes: []string{"room", "passage", "trap", "treasure", "secret"},
			EncounterTypes: []domain.TableType{domain.TableTypeEncounter, domain.TableTypeTreasure},
			NPCProbabilities: map[string]float64{
				"helpful": 0.1,
				"neutral": 0.2,
				"hostile": 0.7,
			},
		},
		"social": {
			Name:        "Social Encounter Location",
			SettingType: "social",
			Description: "A court, tavern, or meeting hall focused on NPC interactions.",
			FeatureTypes: []string{"room", "clue"},
			EncounterTypes: []domain.TableType{domain.TableTypeRumor},
			NPCProbabilities: map[string]float64{
				"helpful": 0.6,
				"neutral": 0.3,
				"hostile": 0.1,
			},
		},
		"mixed": {
			Name:        "Mixed Exploration/Combat",
			SettingType: "mixed",
			Description: "A balanced location with both exploration and combat challenges.",
			FeatureTypes: []string{"room", "hazard", "clue", "treasure"},
			EncounterTypes: []domain.TableType{domain.TableTypeEncounter, domain.TableTypeRumor, domain.TableTypeTreasure},
			NPCProbabilities: map[string]float64{
				"helpful": 0.4,
				"neutral": 0.4,
				"hostile": 0.2,
			},
		},
	}

	if template, ok := templates[settingType]; ok {
		return template
	}

	// Default to wilderness if no match
	return templates["wilderness"]
}

// GenerateAreaWithContext generates a context-aware area with encounters, NPCs, and treasure.
func (s *AreaService) GenerateAreaWithContext(
	ctx context.Context,
	campaignID string,
	chapterID string,
	areaNumber int,
	locationHint string,
	settingType string,
	partyLevel int,
	factionContext *domain.FactionReputationMatrix,
	narrativeState *domain.NarrativeState,
) (*domain.Area, error) {
	if err := domain.ValidateAreaNumber(areaNumber); err != nil {
		return nil, err
	}

	levelRange := domain.LevelRange{
		Min: max(1, partyLevel-2),
		Max: min(20, partyLevel+3),
	}
	if err := domain.ValidateLevelRange(levelRange); err != nil {
		return nil, err
	}

	// Select template based on settingType
	template := s.selectTemplate(settingType)

	// Initialize random table service (inject as dependency in production)
	// For now, this is a placeholder - in production, inject via constructor
	randomTableService := &RandomTableService{}

	// Generate encounters using random table
	encounters, err := s.generateEncounters(randomTableService, campaignID, locationHint, settingType, partyLevel, factionContext, narrativeState)
	if err != nil {
		return nil, fmt.Errorf("failed to generate encounters: %w", err)
	}

	// Generate treasure using random table
	treasure, err := s.generateTreasure(randomTableService, campaignID, locationHint, partyLevel, factionContext, narrativeState)
	if err != nil {
		return nil, fmt.Errorf("failed to generate treasure: %w", err)
	}

	// Generate NPCs with faction awareness
	npcs := s.generateNPCs(randomTableService, campaignID, locationHint, partyLevel, factionContext, narrativeState)

	// Generate features (3-5 minimum)
	features := s.generateFeatures(template, locationHint, partyLevel)

	// Generate boxed text (100-600 words)
	boxedText := s.generateBoxedText(template, locationHint, settingType)

	// Generate development text
	developmentText := s.generateDevelopmentText(template, narrativeState)

	area := &domain.Area{
		ID:              fmt.Sprintf("area_%s_%d", chapterID, areaNumber),
		ChapterID:       chapterID,
		AreaNumber:      areaNumber,
		Title:           s.generateTitle(template, locationHint),
		Summary:         fmt.Sprintf("A %s location for level %d adventurers.", settingType, partyLevel),
		Description:     template.Description,
		LevelRange:      levelRange,
		Features:        features,
		Encounters:      encounters,
		NPCs:            npcs,
		Treasure:        treasure,
		Development:     developmentText,
		DMSidebars:      []domain.DMSidebar{},
		PlayerReadAloud: boxedText,
		Maps:            []domain.MapReference{},
	}

	if err := area.Validate(); err != nil {
		return nil, fmt.Errorf("generated area validation failed: %w", err)
	}

	return area, nil
}

// generateEncounters generates 2-4 encounters using random table
func (s *AreaService) generateEncounters(
	randomTableService *RandomTableService,
	campaignID, locationHint, settingType string,
	partyLevel int,
	factionContext *domain.FactionReputationMatrix,
	narrativeState *domain.NarrativeState,
) ([]domain.AreaEncounter, error) {
	// Placeholder implementation - returns sample encounters
	// In production, this would call randomTableService.GenerateTable()
	numEncounters := 2 + rand.Intn(3) // 2, 3, or 4
	encounters := make([]domain.AreaEncounter, 0, numEncounters)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < numEncounters; i++ {
		crTotal := float64(partyLevel) * (1.0 + rng.Float64()*0.5) // partyLevel × 1.0 to 1.5
		xpValue := int(crTotal * 100) // Simplified XP calculation

		encounters = append(encounters, domain.AreaEncounter{
			EncounterID: fmt.Sprintf("enc_%d", i+1),
			Trigger:     fmt.Sprintf("As the party ventures into the %s, they encounter...", locationHint),
			CRTotal:     crTotal,
			XPValue:     xpValue,
		})
	}

	return encounters, nil
}

// generateTreasure generates 1-3 treasure entries
func (s *AreaService) generateTreasure(
	randomTableService *RandomTableService,
	campaignID, locationHint string,
	partyLevel int,
	factionContext *domain.FactionReputationMatrix,
	narrativeState *domain.NarrativeState,
) ([]domain.Treasure, error) {
	// Placeholder implementation
	numTreasure := 1 + rand.Intn(3) // 1, 2, or 3
	treasure := make([]domain.Treasure, 0, numTreasure)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < numTreasure; i++ {
		value := rng.Intn(10)*10 + 10 // 10-100 gold
		if partyLevel >= 5 && rng.Intn(100) < 20 {
			// 20% chance for magic item at level 5+
			treasure = append(treasure, domain.Treasure{
				Name:        fmt.Sprintf("Magic Item %d", i+1),
				Description: "A magical item of some sort",
				ValueGP:     value * 10,
				MagicItemID: fmt.Sprintf("magic_%d", i+1),
			})
		} else {
			treasure = append(treasure, domain.Treasure{
				Name:        fmt.Sprintf("Gold Pouch %d", i+1),
				Description: "A pouch containing gold coins",
				ValueGP:     value,
			})
		}
	}

	return treasure, nil
}

// generateNPCs generates 0-2 NPCs based on faction context
func (s *AreaService) generateNPCs(
	randomTableService *RandomTableService,
	campaignID, locationHint string,
	partyLevel int,
	factionContext *domain.FactionReputationMatrix,
	narrativeState *domain.NarrativeState,
) []domain.AreaNPC {
	// Placeholder implementation
	// In production, this would use faction context to determine NPC attitude
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 50% chance of having an NPC
	if rng.Intn(100) >= 50 {
		return []domain.AreaNPC{}
	}

	// Determine attitude based on faction (simplified)
	attitude := "neutral"
	if factionContext != nil && len(factionContext.Entries) > 0 {
		avgScore := int8(0)
		for _, entry := range factionContext.Entries {
			avgScore += entry.Score
		}
		avgScore /= int8(len(factionContext.Entries))
		if avgScore >= 30 {
			attitude = "helpful"
		} else if avgScore <= -30 {
			attitude = "hostile"
		}
	}

	return []domain.AreaNPC{
		{
			NPCID:           "npc_generated_1",
			Role:            "Guide or informant",
			InteractionNotes: fmt.Sprintf("Attitude: %s. The NPC approaches with a cautious greeting.", attitude),
		},
	}
}

// generateFeatures generates 3-5 features for the area
func (s *AreaService) generateFeatures(template *areaTemplate, locationHint string, partyLevel int) []domain.AreaFeature {
	numFeatures := 3 + rand.Intn(3) // 3, 4, or 5
	features := make([]domain.AreaFeature, 0, numFeatures)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	featureNames := []string{
		"Main Chamber", "Guard Post", "Hidden Alcove", "Ancient Ruin", "Natural Clearing",
		"Secret Passage", "Treasure Vault", "Trap Room", "Observation Deck", "Sacred Shrine",
	}

	for i := 0; i < numFeatures; i++ {
		isHidden := rng.Intn(100) < 20 // 20% chance for hidden features
		dcValue := 10 + rng.Intn(10) + partyLevel/2 // DC 10-20 based on level

		featureType := "room"
		if len(template.FeatureTypes) > 0 {
			featureType = template.FeatureTypes[rng.Intn(len(template.FeatureTypes))]
		}

		features = append(features, domain.AreaFeature{
			Type:        featureType,
			Name:        featureNames[rng.Intn(len(featureNames))],
			Description: fmt.Sprintf("A %s feature in the %s area.", featureType, locationHint),
			DC:          &dcValue,
			Hidden:      isHidden,
		})
	}

	return features
}

// generateBoxedText generates 100-600 word boxed text for the area
func (s *AreaService) generateBoxedText(template *areaTemplate, locationHint string, settingType string) string {
	// Placeholder implementation - generates simple boxed text
	// In production, this would use LLM or more sophisticated generation
	return fmt.Sprintf(`The %s stretches before you, a %s environment filled with mystery and danger. 

The air is thick with anticipation as you venture deeper into the %s. Strange sounds echo through the %s, and shadows dance in the periphery of your vision.

Ahead, you can see the outline of structures that hint at the adventures that await. The path forward is clear, though what lies beyond remains shrouded in mystery.

Your journey through this %s location is about to begin.`,
		locationHint, settingType, locationHint, locationHint, settingType)
}

// generateDevelopmentText generates development text for the area
func (s *AreaService) generateDevelopmentText(template *areaTemplate, narrativeState *domain.NarrativeState) string {
	branches := []string{
		fmt.Sprintf("IF the party alerts the local inhabitants THEN the area becomes hostile: all NPCs are wary, prices increase by 50%%"),
		fmt.Sprintf("IF the party leaves without causing trouble THEN they gain a safe haven: can rest here without danger"),
	}

	// Add narrative-state-aware branch if clues are revealed
	if narrativeState != nil && len(narrativeState.RevealedClues) > 0 {
		clue := narrativeState.RevealedClues[0]
		branches = append(branches,
			fmt.Sprintf("IF the party uses the clue about '%s' THEN they discover a hidden aspect of the area", clue.Description))
	}

	return strings.Join(branches, "; ")
}

// generateTitle generates a title for the area
func (s *AreaService) generateTitle(template *areaTemplate, locationHint string) string {
	if locationHint == "" {
		return template.Name
	}
	// Capitalize first letter
	if len(locationHint) > 0 {
		locationHint = strings.ToUpper(string(locationHint[0])) + locationHint[1:]
	}
	return fmt.Sprintf("The %s %s", locationHint, template.Name)
}
