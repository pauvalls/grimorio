package services

import (
	"context"
	"fmt"
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
	// TODO: Implement repository method
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
