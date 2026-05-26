package services

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode"

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

	// Build narrative maps once for efficiency
	revealedClueMap := buildRevealedClueMap(context.NarrativeState)

	for _, fact := range filtered {
		baseWeight := rng.Intn(10) + 1

		// Apply faction weight modifier
		weighted := applyFactionWeightModifier(baseWeight, fact, context.FactionContext)

		// Apply narrative weight modifier (revealed clues boost)
		weighted = applyNarrativeWeightModifier(weighted, fact, revealedClueMap)

		tbl.Entries = append(tbl.Entries, domain.RandomTableEntry{
			Weight:      weighted,
			Description: fact.Statement,
			SourceFact:  fact.ID,
		})
	}

	return tbl, nil
}

func (s *RandomTableService) filterFacts(facts []domain.CanonFact, tableType domain.TableType, context domain.TableContext, minLevel, maxLevel int) []domain.CanonFact {
	var filtered []domain.CanonFact

	// Build lookup maps for performance
	deadNPCMap := buildDeadNPCMap(context.NarrativeState)

	// Extract location keywords
	locationKeywords := extractLocationKeywords(context.LocationHint)

	for _, fact := range facts {
		// 1. Check narrative state exclusions (dead NPCs)
		if shouldExcludeFact(fact, deadNPCMap, tableType) {
			continue
		}

		// 2. Match category to table type
		if !categoryMatchesTableType(fact.Category, tableType) {
			continue
		}

		// 3. Match location hint (fuzzy matching)
		if !matchesLocation(fact, locationKeywords) {
			continue
		}

		// 4. Match setting context
		if context.SettingType != "" && !strings.Contains(strings.ToLower(fact.Statement), context.SettingType) {
			continue
		}

		// 5. Match level range
		if minLevel > 0 && maxLevel > 0 {
			if !crInRange(fact.Statement, minLevel, maxLevel) {
				if extractCRs(fact.Statement) != nil && len(extractCRs(fact.Statement)) > 0 {
					continue
				}
			}
		}

		filtered = append(filtered, fact)
	}

	// Fallback: if no matches, return all facts with warning
	if len(filtered) == 0 && len(facts) > 0 {
		// Log warning (use logger in production)
		fmt.Printf("Warning: No facts matched location hint '%s', using unfiltered results\n", context.LocationHint)
		return facts
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

// extractLocationKeywords extracts keywords from LocationHint
func extractLocationKeywords(locationHint string) []string {
	if locationHint == "" {
		return nil
	}

	// Convert to lowercase and split by spaces/punctuation
	normalized := strings.ToLower(locationHint)
	words := strings.FieldsFunc(normalized, func(r rune) bool {
		return unicode.IsSpace(r) || r == ',' || r == '.'
	})

	// Filter common stop words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "in": true,
		"at": true, "on": true, "with": true, "and": true,
	}

	var keywords []string
	for _, word := range words {
		if !stopWords[word] && len(word) >= 3 {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

// matchesLocation checks if a fact matches location keywords
func matchesLocation(fact domain.CanonFact, keywords []string) bool {
	if len(keywords) == 0 {
		return true // No location hint, match all
	}

	// Check category (exact match)
	categoryLower := strings.ToLower(fact.Category)
	for _, keyword := range keywords {
		if categoryLower == keyword {
			return true
		}
	}

	// Check statement (keyword presence, case-insensitive)
	statementLower := strings.ToLower(fact.Statement)
	for _, keyword := range keywords {
		if strings.Contains(statementLower, keyword) {
			return true
		}
	}

	return false
}

// Reputation thresholds for faction weight modifiers
const (
	ReputationHostile    = -30
	ReputationUnfriendly = -10
	ReputationFriendly   = 30
	ReputationAllied     = 80
)

// getReputationStatus returns the reputation status for a score
func getReputationStatus(score int8) string {
	if score <= ReputationHostile {
		return "hostile"
	} else if score <= ReputationUnfriendly {
		return "unfriendly"
	} else if score < ReputationFriendly {
		return "neutral"
	} else if score < ReputationAllied {
		return "friendly"
	}
	return "allied"
}

// isHostileFact determines if a fact represents a hostile encounter
func isHostileFact(fact domain.CanonFact) bool {
	hostileKeywords := []string{
		"ambush", "attack", "enemy", "hostile", "threat",
		"danger", "assault", "raid",
	}
	statementLower := strings.ToLower(fact.Statement)
	for _, keyword := range hostileKeywords {
		if strings.Contains(statementLower, keyword) {
			return true
		}
	}
	return false
}

// isHelpfulFact determines if a fact represents a helpful interaction
func isHelpfulFact(fact domain.CanonFact) bool {
	helpfulKeywords := []string{
		"help", "ally", "friend", "offer", "assist",
		"support", "aid", "discount", "information",
	}
	statementLower := strings.ToLower(fact.Statement)
	for _, keyword := range helpfulKeywords {
		if strings.Contains(statementLower, keyword) {
			return true
		}
	}
	return false
}

// extractFactionFromFact extracts faction ID from fact statement
func extractFactionFromFact(fact domain.CanonFact) string {
	statementLower := strings.ToLower(fact.Statement)

	// Heuristic: look for faction patterns in statement
	if strings.Contains(statementLower, "thieves guild") {
		return "faction_thieves_guild"
	}
	if strings.Contains(statementLower, "merchant guild") {
		return "faction_merchant_guild"
	}
	if strings.Contains(statementLower, "fighters guild") {
		return "faction_fighters_guild"
	}
	if strings.Contains(statementLower, "mages guild") {
		return "faction_mages_guild"
	}
	if strings.Contains(statementLower, "guards") || strings.Contains(statementLower, "city watch") {
		return "faction_city_watch"
	}

	return ""
}

// applyFactionWeightModifier applies weight modifier based on faction reputation
func applyFactionWeightModifier(
	baseWeight int,
	fact domain.CanonFact,
	factionContext *domain.FactionReputationMatrix,
) int {
	if factionContext == nil {
		return baseWeight
	}

	// Extract faction from fact
	factionID := extractFactionFromFact(fact)
	if factionID == "" {
		return baseWeight // No faction association
	}

	// Find reputation entry
	var entry *domain.ReputationEntry
	for i := range factionContext.Entries {
		if factionContext.Entries[i].FactionID == factionID {
			entry = &factionContext.Entries[i]
			break
		}
	}

	if entry == nil {
		return baseWeight // No reputation data for this faction
	}

	// Determine if fact is hostile or helpful
	isHostile := isHostileFact(fact)
	isHelpful := isHelpfulFact(fact)

	// Apply modifier based on reputation status
	status := getReputationStatus(entry.Score)
	var modifier float64

	switch status {
	case "hostile":
		if isHostile {
			modifier = 1.5 // +50%
		} else if isHelpful {
			modifier = 0.2 // -80%
		} else {
			modifier = 1.0
		}
	case "unfriendly":
		if isHostile {
			modifier = 1.2 // +20%
		} else if isHelpful {
			modifier = 0.6 // -40%
		} else {
			modifier = 1.0
		}
	case "neutral":
		modifier = 1.0 // No modifier
	case "friendly":
		if isHostile {
			modifier = 0.5 // -50%
		} else if isHelpful {
			modifier = 1.4 // +40%
		} else {
			modifier = 1.0
		}
	case "allied":
		if isHostile {
			modifier = 0.2 // -80%
		} else if isHelpful {
			modifier = 1.6 // +60%
		} else {
			modifier = 1.0
		}
	}

	modifiedWeight := int(float64(baseWeight) * modifier)
	if modifiedWeight < 1 {
		modifiedWeight = 1 // Clamp to minimum 1
	}

	return modifiedWeight
}

// buildDeadNPCMap creates O(1) lookup map for dead NPCs
func buildDeadNPCMap(narrativeState *domain.NarrativeState) map[string]bool {
	if narrativeState == nil {
		return nil
	}

	deadNPCMap := make(map[string]bool, len(narrativeState.DeadNPCs))
	for _, death := range narrativeState.DeadNPCs {
		deadNPCMap[death.NPCID] = true
		deadNPCMap[strings.ToLower(death.Name)] = true // Also match by name
	}
	return deadNPCMap
}

// buildRevealedClueMap creates O(1) lookup for revealed clues
func buildRevealedClueMap(narrativeState *domain.NarrativeState) map[string]bool {
	if narrativeState == nil {
		return nil
	}

	clueMap := make(map[string]bool, len(narrativeState.RevealedClues))
	for _, clue := range narrativeState.RevealedClues {
		clueMap[clue.ID] = true
		clueMap[strings.ToLower(clue.Description)] = true
	}
	return clueMap
}

// shouldExcludeFact checks if fact should be excluded based on narrative state
func shouldExcludeFact(fact domain.CanonFact, deadNPCMap map[string]bool, tableType domain.TableType) bool {
	if deadNPCMap == nil {
		return false
	}

	// Check if fact references any dead NPC
	if factReferencesDeadNPC(fact, deadNPCMap) {
		// Dead NPCs are ALWAYS excluded from encounter and rumor tables
		if tableType == domain.TableTypeEncounter || tableType == domain.TableTypeRumor {
			return true
		}
		// For weather/treasure, only exclude if directly named
		if isDirectNPCReference(fact) {
			return true
		}
	}

	return false
}

// factReferencesDeadNPC checks if fact references a dead NPC
func factReferencesDeadNPC(fact domain.CanonFact, deadNPCMap map[string]bool) bool {
	statementLower := strings.ToLower(fact.Statement)

	// Check for NPC ID pattern: "npc_xxx"
	if strings.Contains(statementLower, "npc_") {
		parts := strings.Fields(statementLower)
		for _, part := range parts {
			if strings.HasPrefix(part, "npc_") {
				if deadNPCMap[part] {
					return true
				}
			}
		}
	}

	// Check for NPC names (simplified: exact word match)
	for deadName := range deadNPCMap {
		if strings.Contains(statementLower, deadName) {
			return true
		}
	}

	return false
}

// isDirectNPCReference checks if fact directly names an NPC (vs indirect reference)
func isDirectNPCReference(fact domain.CanonFact) bool {
	// Direct: "Gareth the Guard patrols..."
	// Indirect: "guard's equipment found" (no name)
	statementLower := strings.ToLower(fact.Statement)
	return strings.Contains(statementLower, " the ") ||
		strings.Contains(statementLower, " sir ") ||
		strings.Contains(statementLower, " lady ") ||
		strings.Contains(statementLower, " captain ")
}

// applyNarrativeWeightModifier applies weight modifier based on revealed clues
func applyNarrativeWeightModifier(baseWeight int, fact domain.CanonFact, revealedClueMap map[string]bool) int {
	if revealedClueMap == nil {
		return baseWeight
	}

	statementLower := strings.ToLower(fact.Statement)

	// Check if fact references a revealed clue
	for clueDesc := range revealedClueMap {
		if strings.Contains(statementLower, strings.ToLower(clueDesc)) {
			return baseWeight + 3 // Boost weight by +3
		}
	}

	return baseWeight
}
