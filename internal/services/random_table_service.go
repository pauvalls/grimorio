package services

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// RandomTableService generates contextual random tables from canon data
type RandomTableService struct {
	canonRepo repository.CanonRepository
}

// NewRandomTableService creates a new random table service
func NewRandomTableService(canonRepo repository.CanonRepository) *RandomTableService {
	return &RandomTableService{canonRepo: canonRepo}
}

// GenerateTable creates a weighted random table from canon facts
func (s *RandomTableService) GenerateTable(ctx context.Context, campaignID string, tableType domain.TableType, context domain.TableContext) (*domain.RandomTable, error) {
	if !domain.IsValidTableType(string(tableType)) {
		return nil, fmt.Errorf("invalid table type: %s", tableType)
	}

	doc, err := s.canonRepo.Load(campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to load canon: %w", err)
	}

	tbl := &domain.RandomTable{
		CampaignID: campaignID,
		TableType:  tableType,
		Context:    context,
		Entries:    []domain.RandomTableEntry{},
	}

	// Parse level range
	minLevel, maxLevel := parseLevelRange(context.LevelRange)
	if minLevel == 0 && maxLevel == 0 {
		minLevel, maxLevel = parseLevelRangeFromDoc(doc)
	}

	// Filter facts by context
	filtered := s.filterFacts(doc.Facts, tableType, context, minLevel, maxLevel)

	if len(filtered) == 0 {
		return tbl, nil // empty table, no panic
	}

	// Cap at 100 entries
	if len(filtered) > 100 {
		filtered = filtered[:100]
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for _, fact := range filtered {
		weight := rng.Intn(10) + 1
		tbl.Entries = append(tbl.Entries, domain.RandomTableEntry{
			Weight:      weight,
			Description: fact.Statement,
			SourceFact:  fact.ID,
		})
	}

	return tbl, nil
}

func (s *RandomTableService) filterFacts(facts []domain.CanonFact, tableType domain.TableType, context domain.TableContext, minLevel, maxLevel int) []domain.CanonFact {
	var filtered []domain.CanonFact
	for _, fact := range facts {
		// Match category to table type
		if !categoryMatchesTableType(fact.Category, tableType) {
			continue
		}

		// Match setting context
		if context.SettingType != "" && !strings.Contains(strings.ToLower(fact.Statement), context.SettingType) {
			continue
		}

		// Match level range (crude heuristic: check if statement mentions CR within range)
		if minLevel > 0 && maxLevel > 0 {
			if !crInRange(fact.Statement, minLevel, maxLevel) {
				// Keep facts that don't mention CR at all
				if extractCRs(fact.Statement) != nil && len(extractCRs(fact.Statement)) > 0 {
					continue
				}
			}
		}

		filtered = append(filtered, fact)
	}
	return filtered
}

func categoryMatchesTableType(category string, tableType domain.TableType) bool {
	switch tableType {
	case domain.TableTypeEncounter:
		return category == "creature" || category == "encounter"
	case domain.TableTypeRumor:
		return category == "lore" || category == "rumor"
	case domain.TableTypeWeather:
		return category == "weather" || category == "environment"
	case domain.TableTypeTreasure:
		return category == "item" || category == "treasure"
	default:
		return true
	}
}

func parseLevelRange(levelRange string) (int, int) {
	parts := strings.Split(levelRange, "-")
	if len(parts) != 2 {
		return 0, 0
	}
	min, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	max, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
	return min, max
}

func parseLevelRangeFromDoc(doc *domain.CanonDocument) (int, int) {
	for _, entity := range doc.Entities {
		if entity.Role == "mcguffin" {
			if lr, ok := entity.Properties["level_range"].(string); ok {
				return parseLevelRange(lr)
			}
		}
	}
	return 1, 20
}

func extractCRs(statement string) []int {
	var crs []int
	re := strings.NewReplacer("CR", "", ":", "", "-", " ")
	cleaned := re.Replace(statement)
	words := strings.Fields(cleaned)
	for _, w := range words {
		if n, err := strconv.Atoi(w); err == nil && n > 0 {
			crs = append(crs, n)
		}
	}
	return crs
}

func crInRange(statement string, minLevel, maxLevel int) bool {
	crs := extractCRs(statement)
	if len(crs) == 0 {
		return true // no CR mentioned, keep it
	}
	for _, cr := range crs {
		if cr >= minLevel-2 && cr <= maxLevel+3 {
			return true
		}
	}
	return false
}
