package services

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// SessionGenerator generates contextual session content (encounters, loot, NPC appearances).
type SessionGenerator struct {
	canonRepo repository.CanonRepository
	stateRepo repository.NarrativeStateRepository
}

// NewSessionGenerator creates a new SessionGenerator.
func NewSessionGenerator(canonRepo repository.CanonRepository, stateRepo repository.NarrativeStateRepository) *SessionGenerator {
	return &SessionGenerator{
		canonRepo: canonRepo,
		stateRepo: stateRepo,
	}
}

// GenerateEncounterRecommendations generates 2-4 encounter recommendations for a session.
func (s *SessionGenerator) GenerateEncounterRecommendations(ctx context.Context, campaignID string, sessionNum int) ([]domain.EncounterRecommendation, error) {
	state, err := s.stateRepo.Load(campaignID)
	if err != nil || state == nil {
		return s.generateGenericEncounters(sessionNum), nil
	}

	doc, err := s.canonRepo.Load(campaignID)
	if err != nil {
		return s.generateGenericEncounters(sessionNum), nil
	}

	// Determine party level from characters
	partyLevel := s.getPartyLevel(state, doc)

	// Generate encounters based on session number and active quests
	encounters := []domain.EncounterRecommendation{}

	// Combat encounter
	encounters = append(encounters, s.generateCombatEncounter(partyLevel, sessionNum, doc))

	// Social encounter
	encounters = append(encounters, s.generateSocialEncounter(partyLevel, sessionNum, state, doc))

	// Exploration encounter (for sessions 2+)
	if sessionNum >= 2 {
		encounters = append(encounters, s.generateExplorationEncounter(partyLevel, sessionNum, doc))
	}

	// Mixed encounter for later sessions
	if sessionNum >= 3 {
		encounters = append(encounters, s.generateMixedEncounter(partyLevel, sessionNum, state, doc))
	}

	return encounters, nil
}

// GenerateLootSuggestions generates tier-appropriate loot suggestions.
func (s *SessionGenerator) GenerateLootSuggestions(ctx context.Context, campaignID string, partyLevel int) ([]domain.LootSuggestion, error) {
	loot := []domain.LootSuggestion{}

	// Determine loot tier based on party level
	tier := s.getLootTier(partyLevel)

	// Generate 2-4 loot suggestions
	loot = append(loot, s.generateGoldReward(tier))
	loot = append(loot, s.generateMagicItem(tier))

	if tier >= 2 {
		loot = append(loot, s.generateConsumable(tier))
	}

	if tier >= 3 {
		loot = append(loot, s.generateRareItem(tier))
	}

	return loot, nil
}

// GenerateNPCAppearances generates NPC appearances relevant to the session.
func (s *SessionGenerator) GenerateNPCAppearances(ctx context.Context, campaignID string, sessionNum int) ([]domain.NPCAppearance, error) {
	state, err := s.stateRepo.Load(campaignID)
	if err != nil || state == nil {
		return []domain.NPCAppearance{}, nil
	}

	doc, err := s.canonRepo.Load(campaignID)
	if err != nil {
		return []domain.NPCAppearance{}, nil
	}

	appearances := []domain.NPCAppearance{}

	// Find quest-giver NPCs from active quests
	for _, quest := range state.ActiveQuests {
		if quest.GiverNPC != "" {
			npc := s.findNPCByID(doc, quest.GiverNPC)
			if npc != nil {
				appearances = append(appearances, domain.NPCAppearance{
					NPCID:      npc.ID,
					Name:       npc.Name,
					Role:       "Quest Giver",
					Context:    fmt.Sprintf("Relacionado con la quest activa: %s", quest.Name),
					Importance: "major",
				})
			}
		}
	}

	// Add 1-2 random NPCs from canon for flavor
	randomNPCs := s.getRandomNPCs(doc, 2)
	for _, npc := range randomNPCs {
		appearances = append(appearances, domain.NPCAppearance{
			NPCID:      npc.ID,
			Name:       npc.Name,
			Role:       npc.Role,
			Context:    s.generateNPCContext(npc, sessionNum),
			Importance: "minor",
		})
	}

	return appearances, nil
}

// Helper functions

func (s *SessionGenerator) getPartyLevel(state *domain.NarrativeState, doc *domain.CanonDocument) int {
	// Try to get party level from characters in the campaign
	// For now, estimate based on session number (typical progression)
	if state.CurrentSession >= 10 {
		return 10
	}
	if state.CurrentSession >= 5 {
		return 5
	}
	return 1 + state.CurrentSession
}

func (s *SessionGenerator) getLootTier(level int) int {
	if level >= 17 {
		return 4 // Tier 4 (17+)
	}
	if level >= 11 {
		return 3 // Tier 3 (11-16)
	}
	if level >= 5 {
		return 2 // Tier 2 (5-10)
	}
	return 1 // Tier 1 (1-4)
}

func (s *SessionGenerator) generateGenericEncounters(sessionNum int) []domain.EncounterRecommendation {
	encounters := []domain.EncounterRecommendation{
		{
			Name:        "Encuentro de Combate",
			CR:          fmt.Sprintf("%d-%d", sessionNum, sessionNum+1),
			Type:        "combat",
			Description: "Encuentro de combate balanceado para el nivel del partido.",
			Context:     "Encuentro genérico - ajustar según la narrativa actual.",
		},
		{
			Name:        "Encuentro Social",
			CR:          "N/A",
			Type:        "social",
			Description: "Interacción con NPCs importantes o facciones.",
			Context:     "Oportunidad para roleplay y desarrollo de la trama.",
		},
	}

	if sessionNum >= 2 {
		encounters = append(encounters, domain.EncounterRecommendation{
			Name:        "Encuentro de Exploración",
			CR:          "N/A",
			Type:        "exploration",
			Description: "Descubrimiento de ubicación, trampa o misterio.",
			Context:     "Momento para explorar el mundo y sus secretos.",
		})
	}

	return encounters
}

func (s *SessionGenerator) generateCombatEncounter(partyLevel, sessionNum int, doc *domain.CanonDocument) domain.EncounterRecommendation {
	cr := s.calculateCR(partyLevel)

	// Try to find monsters in canon that match the CR
	monsterName := s.findMonsterForCR(doc, cr)
	if monsterName == "" {
		monsterName = "Criaturas apropiadas"
	}

	return domain.EncounterRecommendation{
		Name:        fmt.Sprintf("Combate: %s", monsterName),
		CR:          cr,
		Type:        "combat",
		Description: fmt.Sprintf("Encuentro de combate contra %s.", monsterName),
		Context:     fmt.Sprintf("Encuentro balanceado para partido de nivel %d.", partyLevel),
	}
}

func (s *SessionGenerator) generateSocialEncounter(partyLevel, sessionNum int, state *domain.NarrativeState, doc *domain.CanonDocument) domain.EncounterRecommendation {
	// Find an NPC from active quests or canon
	npcName := "NPC Importante"
	if len(state.ActiveQuests) > 0 && state.ActiveQuests[0].GiverNPC != "" {
		npc := s.findNPCByID(doc, state.ActiveQuests[0].GiverNPC)
		if npc != nil {
			npcName = npc.Name
		}
	}

	return domain.EncounterRecommendation{
		Name:        fmt.Sprintf("Social: Encuentro con %s", npcName),
		CR:          "N/A",
		Type:        "social",
		Description: fmt.Sprintf("Interacción social con %s.", npcName),
		Context:     "Oportunidad para obtener información, negociar o desarrollar relaciones.",
	}
}

func (s *SessionGenerator) generateExplorationEncounter(partyLevel, sessionNum int, doc *domain.CanonDocument) domain.EncounterRecommendation {
	explorationTypes := []string{
		"Descubrimiento de ruinas antiguas",
		"Exploración de cueva misteriosa",
		"Navegación por territorio peligroso",
		"Investigación de fenómeno mágico",
	}

	idx := rand.Intn(len(explorationTypes))

	return domain.EncounterRecommendation{
		Name:        fmt.Sprintf("Exploración: %s", explorationTypes[idx]),
		CR:          "N/A",
		Type:        "exploration",
		Description: "Encuentro enfocado en exploración y descubrimiento.",
		Context:     "Momento para que el partido explore y descubra secretos del mundo.",
	}
}

func (s *SessionGenerator) generateMixedEncounter(partyLevel, sessionNum int, state *domain.NarrativeState, doc *domain.CanonDocument) domain.EncounterRecommendation {
	return domain.EncounterRecommendation{
		Name:        "Encuentro Mixto: Combate + Social",
		CR:          fmt.Sprintf("%d", partyLevel-1),
		Type:        "mixed",
		Description: "Encuentro que combina elementos de combate y interacción social.",
		Context:     "Situación compleja que requiere tanto habilidades de combate como diplomacia.",
	}
}

func (s *SessionGenerator) calculateCR(partyLevel int) string {
	// Simple CR calculation based on party level
	if partyLevel <= 2 {
		return "1/4-1/2"
	}
	if partyLevel <= 4 {
		return "1-2"
	}
	if partyLevel <= 6 {
		return "2-3"
	}
	if partyLevel <= 8 {
		return "3-4"
	}
	if partyLevel <= 10 {
		return "4-5"
	}
	if partyLevel <= 12 {
		return "5-7"
	}
	if partyLevel <= 14 {
		return "7-9"
	}
	if partyLevel <= 16 {
		return "9-11"
	}
	return "12+"
}

func (s *SessionGenerator) findMonsterForCR(doc *domain.CanonDocument, cr string) string {
	if doc == nil {
		return ""
	}

	// Search bestiary for monsters
	for _, entity := range doc.Entities {
		if entity.Type == domain.EntityTypeMonster {
			// Simple match - in production would parse CR properly
			return entity.Name
		}
	}
	return ""
}

func (s *SessionGenerator) findNPCByID(doc *domain.CanonDocument, npcID string) *domain.CanonEntity {
	if doc == nil {
		return nil
	}

	for _, entity := range doc.Entities {
		if entity.ID == npcID && entity.Type == domain.EntityTypeNPC {
			return &entity
		}
	}
	return nil
}

func (s *SessionGenerator) getRandomNPCs(doc *domain.CanonDocument, count int) []domain.CanonEntity {
	if doc == nil {
		return []domain.CanonEntity{}
	}

	var npcs []domain.CanonEntity
	for _, entity := range doc.Entities {
		if entity.Type == domain.EntityTypeNPC && entity.CanonState == domain.EntityStateAlive {
			npcs = append(npcs, entity)
		}
	}

	if len(npcs) == 0 {
		return []domain.CanonEntity{}
	}

	// Shuffle and take up to count
	rand.Seed(time.Now().UnixNano())
	result := []domain.CanonEntity{}
	for i := 0; i < count && i < len(npcs); i++ {
		idx := rand.Intn(len(npcs))
		result = append(result, npcs[idx])
		// Remove selected to avoid duplicates
		npcs = append(npcs[:idx], npcs[idx+1:]...)
	}

	return result
}

func (s *SessionGenerator) generateNPCContext(npc domain.CanonEntity, sessionNum int) string {
	contexts := []string{
		fmt.Sprintf("Aparece en la sesión %d proporcionando información o asistencia.", sessionNum),
		fmt.Sprintf("Puede ser encontrado en su ubicación habitual durante la sesión %d.", sessionNum),
		fmt.Sprintf("Posible encuentro casual durante la sesión %d.", sessionNum),
	}

	idx := rand.Intn(len(contexts))
	return contexts[idx]
}

func (s *SessionGenerator) generateGoldReward(tier int) domain.LootSuggestion {
	goldAmounts := map[int]string{
		1: "50-100 gp",
		2: "200-500 gp",
		3: "1000-2500 gp",
		4: "5000+ gp",
	}

	return domain.LootSuggestion{
		Name:        "Recompensa en Oro",
		Type:        "gold",
		Rarity:      "common",
		Description: fmt.Sprintf("Bolsa con %s", goldAmounts[tier]),
		Context:     "Recompensa estándar por completar la sesión o derrotar enemigos.",
	}
}

func (s *SessionGenerator) generateMagicItem(tier int) domain.LootSuggestion {
	items := map[int][]string{
		1: {"Poción de Curación", "Pergamino de Hechizo (nivel 1)"},
		2: {"Arma Mágica +1", "Armadura Mágica +1", "Anillo de Protección +1"},
		3: {"Arma Mágica +2", "Objeto Maravilloso Mayor", "Pergamino de Hechizo (nivel 3-4)"},
		4: {"Arma Mágica +3", "Objeto Maravilloso Legendario", "Artefacto Menor"},
	}

	itemList := items[tier]
	if itemList == nil {
		itemList = items[1]
	}

	idx := rand.Intn(len(itemList))
	rarity := s.getRarityForTier(tier)

	return domain.LootSuggestion{
		Name:        itemList[idx],
		Type:        "magical",
		Rarity:      rarity,
		Description: fmt.Sprintf("Objeto mágico apropiado para tier %d.", tier),
		Context:     "Recompensa significativa por logros importantes.",
	}
}

func (s *SessionGenerator) generateConsumable(tier int) domain.LootSuggestion {
	consumables := []string{
		"Poción de Velocidad",
		"Poción de Resistencia",
		"Antídoto",
		"Kit de Escalada",
		"Raciones de Viaje",
	}

	idx := rand.Intn(len(consumables))
	rarity := "common"
	if tier >= 3 {
		rarity = "uncommon"
	}

	return domain.LootSuggestion{
		Name:        consumables[idx],
		Type:        "consumable",
		Rarity:      rarity,
		Description: "Ítem consumible útil para la aventura.",
		Context:     "Suministros o consumibles mágicos para el partido.",
	}
}

func (s *SessionGenerator) generateRareItem(tier int) domain.LootSuggestion {
	rareItems := []string{
		"Objeto Maravilloso Raro",
		"Armadura Mágica +2",
		"Varita Mágica",
		"Amuleto de Protección",
	}

	idx := rand.Intn(len(rareItems))

	return domain.LootSuggestion{
		Name:        rareItems[idx],
		Type:        "magical",
		Rarity:      "rare",
		Description: "Objeto mágico raro de gran utilidad.",
		Context:     "Recompensa excepcional por logros extraordinarios.",
	}
}

func (s *SessionGenerator) getRarityForTier(tier int) string {
	switch tier {
	case 1:
		return "common"
	case 2:
		return "uncommon"
	case 3:
		return "rare"
	case 4:
		return "very_rare"
	default:
		return "common"
	}
}

// GetTierFromLevel returns the tier for a given character level.
func GetTierFromLevel(level int) string {
	if level >= 17 {
		return "Tier 4 (17-20)"
	}
	if level >= 11 {
		return "Tier 3 (11-16)"
	}
	if level >= 5 {
		return "Tier 2 (5-10)"
	}
	return "Tier 1 (1-4)"
}

// GetCRForLevel returns appropriate CR range for a party of given level.
func GetCRForLevel(partyLevel int) string {
	if partyLevel <= 2 {
		return "1/4-1/2"
	}
	if partyLevel <= 4 {
		return "1-2"
	}
	if partyLevel <= 6 {
		return "2-3"
	}
	if partyLevel <= 8 {
		return "3-4"
	}
	if partyLevel <= 10 {
		return "4-5"
	}
	if partyLevel <= 12 {
		return "5-7"
	}
	if partyLevel <= 14 {
		return "7-9"
	}
	if partyLevel <= 16 {
		return "9-11"
	}
	return "12+"
}

// FilterEncountersByType filters encounter recommendations by type.
func FilterEncountersByType(encounters []domain.EncounterRecommendation, encounterType string) []domain.EncounterRecommendation {
	var result []domain.EncounterRecommendation
	for _, e := range encounters {
		if strings.EqualFold(e.Type, encounterType) {
			result = append(result, e)
		}
	}
	return result
}
