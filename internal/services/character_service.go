package services

import (
	"fmt"
	"strings"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// CharacterService handles character business logic
type CharacterService struct {
	repo repository.CharacterRepository
}

// NewCharacterService creates a new character service
func NewCharacterService(repo repository.CharacterRepository) *CharacterService {
	return &CharacterService{repo: repo}
}

// hitDiceByClass maps class names to their hit dice
var hitDiceByClass = map[string]int{
	"barbaro": 12, "guerrero": 10, "paladin": 10, "ranger": 10,
	"clerigo": 8, "druida": 8, "monje": 8, "picaro": 8, "brujo": 8,
	"bardo": 6, "hechicero": 6, "mago": 6, "artifice": 6, "sangre": 6,
}

// classPrimaryStats defines the primary stats for each class (in priority order)
var classPrimaryStats = map[string][]string{
	"guerrero":   {"str", "con", "dex"},
	"barbaro":    {"str", "con", "dex"},
	"paladin":    {"str", "cha", "con"},
	"ranger":     {"dex", "wis", "con"},
	"clerigo":    {"wis", "con", "str"},
	"druida":     {"wis", "con", "dex"},
	"monje":      {"dex", "wis", "con"},
	"picaro":     {"dex", "con", "cha"},
	"brujo":      {"cha", "con", "dex"},
	"bardo":      {"cha", "dex", "con"},
	"hechicero":  {"cha", "con", "dex"},
	"mago":       {"int", "con", "dex"},
	"artifice":   {"int", "con", "dex"},
	"sangre":     {"con", "str", "dex"},
}

// raceBonuses defines racial ability score bonuses
var raceBonuses = map[string]map[string]int{
	"humano":     {"str": 1, "dex": 1, "con": 1, "int": 1, "wis": 1, "cha": 1},
	"elfo":       {"dex": 2, "int": 1},
	"enano":      {"con": 2, "str": 1},
	"mediano":    {"dex": 2, "cha": 1},
	"semielfo":   {"cha": 2, "dex": 1, "int": 1},
	"semiorco":   {"str": 2, "con": 1},
	"gnomo":      {"int": 2, "dex": 1},
	"tiefling":   {"cha": 2, "int": 1},
	"draconido":  {"str": 2, "cha": 1},
	"trasgo":     {"dex": 2, "con": 1},
	"firbolg":    {"wis": 2, "str": 1},
	"gith":       {"int": 2, "str": 1},
}

// classSkills defines class skill proficiencies
var classSkills = map[string][]string{
	"guerrero":   {"athletics", "intimidation", "survival", "perception"},
	"barbaro":    {"athletics", "intimidation", "survival", "nature"},
	"paladin":    {"athletics", "insight", "intimidation", "medicine", "persuasion", "religion"},
	"ranger":     {"athletics", "insight", "investigation", "nature", "perception", "stealth", "survival"},
	"clerigo":    {"history", "insight", "medicine", "persuasion", "religion"},
	"druida":     {"arcana", "animal_handling", "insight", "medicine", "nature", "perception", "survival"},
	"monje":      {"acrobatics", "athletics", "history", "insight", "religion", "stealth"},
	"picaro":     {"acrobatics", "athletics", "deception", "insight", "intimidation", "investigation", "perception", "performance", "persuasion", "sleight_of_hand", "stealth"},
	"brujo":      {"arcana", "deception", "history", "intimidation", "investigation", "nature", "religion"},
	"bardo":      {"acrobatics", "animal_handling", "arcana", "athletics", "deception", "history", "insight", "intimidation", "investigation", "medicine", "nature", "perception", "performance", "persuasion", "religion", "sleight_of_hand", "stealth", "survival"},
	"hechicero":  {"arcana", "deception", "insight", "intimidation", "persuasion", "religion"},
	"mago":       {"arcana", "history", "insight", "investigation", "medicine", "religion"},
	"artifice":   {"arcana", "history", "investigation", "medicine", "nature", "perception", "sleight_of_hand"},
	"sangre":     {"athletics", "intimidation", "survival", "medicine"},
}

// backgroundSkills defines background skill proficiencies
var backgroundSkills = map[string][]string{
	"soldado":           {"athletics", "intimidation"},
	"acolito":           {"insight", "religion"},
	"criminal":          {"deception", "stealth"},
	"sabio":             {"arcana", "history"},
	"noble":             {"history", "persuasion"},
	"artesano":          {"insight", "persuasion"},
	"marinero":          {"athletics", "perception"},
	"ermitano":          {"medicine", "religion"},
	"charlatan":         {"deception", "sleight_of_hand"},
	"heroe del pueblo":  {"animal_handling", "survival"},
	"animador":          {"acrobatics", "performance"},
	"gladiador":         {"athletics", "performance"},
	"mercenario":        {"athletics", "perception"},
	"pirata":            {"athletics", "perception"},
}

// classFeatures defines level 1 features for each class
var classFeatures = map[string][]string{
	"guerrero":   {"Fighting Style", "Second Wind"},
	"barbaro":    {"Rage", "Unarmored Defense"},
	"paladin":    {"Divine Sense", "Lay on Hands"},
	"ranger":     {"Favored Enemy", "Natural Explorer"},
	"clerigo":    {"Spellcasting", "Divine Domain"},
	"druida":     {"Druidic", "Spellcasting"},
	"monje":      {"Unarmored Defense", "Martial Arts"},
	"picaro":     {"Expertise", "Sneak Attack", "Thieves' Cant"},
	"brujo":      {"Otherworldly Patron", "Pact Magic"},
	"bardo":      {"Bardic Inspiration", "Spellcasting", "Jack of All Trades"},
	"hechicero":  {"Spellcasting", "Sorcerous Origin"},
	"mago":       {"Spellcasting", "Arcane Recovery"},
	"artifice":   {"Magical Tinkering", "Spellcasting"},
	"sangre":     {"Blood Magic", "Hemomancy"},
}

// classArmor defines base AC calculation for each class
var classArmor = map[string]struct {
	BaseAC     int
	UseDEX     bool
	DEXMax     int
	Unarmored  bool
}{
	"guerrero":   {BaseAC: 16, UseDEX: true, DEXMax: 2, Unarmored: false},  // chain mail
	"barbaro":    {BaseAC: 10, UseDEX: true, DEXMax: 99, Unarmored: true}, // unarmored: 10 + DEX + CON
	"paladin":    {BaseAC: 16, UseDEX: true, DEXMax: 2, Unarmored: false},  // chain mail
	"ranger":     {BaseAC: 14, UseDEX: true, DEXMax: 2, Unarmored: false},  // leather
	"clerigo":    {BaseAC: 14, UseDEX: true, DEXMax: 2, Unarmored: false},  // scale mail
	"druida":     {BaseAC: 13, UseDEX: true, DEXMax: 99, Unarmored: false}, // leather + shield
	"monje":      {BaseAC: 10, UseDEX: true, DEXMax: 99, Unarmored: true},  // unarmored: 10 + DEX + WIS
	"picaro":     {BaseAC: 12, UseDEX: true, DEXMax: 99, Unarmored: false}, // leather
	"brujo":      {BaseAC: 12, UseDEX: true, DEXMax: 99, Unarmored: false}, // leather
	"bardo":      {BaseAC: 12, UseDEX: true, DEXMax: 99, Unarmored: false}, // leather
	"hechicero":  {BaseAC: 10, UseDEX: true, DEXMax: 99, Unarmored: false}, // no armor
	"mago":       {BaseAC: 10, UseDEX: true, DEXMax: 99, Unarmored: false}, // no armor
	"artifice":   {BaseAC: 14, UseDEX: true, DEXMax: 2, Unarmored: false},  // scale mail
	"sangre":     {BaseAC: 12, UseDEX: true, DEXMax: 99, Unarmored: false}, // leather
}

// assignStats assigns the standard array (15,14,13,12,10,8) based on class priorities
func assignStats(class string) domain.Stats {
	standardArray := []int{15, 14, 13, 12, 10, 8}
	stats := domain.Stats{STR: 10, DEX: 10, CON: 10, INT: 10, WIS: 10, CHA: 10}

	// Get primary stats for class
	primaries := classPrimaryStats[class]
	if len(primaries) == 0 {
		primaries = []string{"str", "dex", "con"}
	}

	// Assign highest values to primary stats
	statMap := map[string]*int{
		"str": &stats.STR, "dex": &stats.DEX, "con": &stats.CON,
		"int": &stats.INT, "wis": &stats.WIS, "cha": &stats.CHA,
	}

	assigned := make(map[string]bool)
	valueIdx := 0

	// First assign to primary stats
	for _, stat := range primaries {
		if valueIdx < len(standardArray) {
			if ptr, ok := statMap[stat]; ok {
				*ptr = standardArray[valueIdx]
				assigned[stat] = true
				valueIdx++
			}
		}
	}

	// Then assign remaining values to other stats
	remainingStats := []string{"str", "dex", "con", "int", "wis", "cha"}
	for _, stat := range remainingStats {
		if !assigned[stat] && valueIdx < len(standardArray) {
			if ptr, ok := statMap[stat]; ok {
				*ptr = standardArray[valueIdx]
				assigned[stat] = true
				valueIdx++
			}
		}
	}

	return stats
}

// applyRaceBonuses applies racial ability score bonuses
func applyRaceBonuses(stats *domain.Stats, race string) {
	bonuses := raceBonuses[race]
	if bonuses == nil {
		return
	}

	for stat, bonus := range bonuses {
		switch stat {
		case "str":
			stats.STR += bonus
		case "dex":
			stats.DEX += bonus
		case "con":
			stats.CON += bonus
		case "int":
			stats.INT += bonus
		case "wis":
			stats.WIS += bonus
		case "cha":
			stats.CHA += bonus
		}
	}
}

// calculateHP calculates max HP based on class and CON modifier
func calculateHP(class string, con int, level int) int {
	hitDie := hitDiceByClass[class]
	if hitDie == 0 {
		hitDie = 8
	}

	conMod := domain.CalculateModifier(con)
	if conMod < 0 {
		conMod = 0
	}

	// First level: max hit die + CON mod
	hp := hitDie + conMod

	// Additional levels: average roll + CON mod
	for i := 2; i <= level; i++ {
		hp += (hitDie / 2) + 1 + conMod
	}

	return hp
}

// calculateAC calculates AC based on class, armor, and stats
func calculateAC(class string, stats domain.Stats) int {
	armor := classArmor[class]
	if armor.BaseAC == 0 {
		armor = classArmor["guerrero"] // default
	}

	dexMod := domain.CalculateModifier(stats.DEX)
	if dexMod < 0 {
		dexMod = 0
	}

	ac := armor.BaseAC

	if armor.UseDEX {
		if armor.DEXMax > 0 && armor.DEXMax < 99 {
			if dexMod > armor.DEXMax {
				dexMod = armor.DEXMax
			}
		}
		ac += dexMod
	}

	// Unarmored defense bonuses
	if armor.Unarmored {
		if class == "barbaro" {
			conMod := domain.CalculateModifier(stats.CON)
			if conMod > 0 {
				ac += conMod
			}
		} else if class == "monje" {
			wisMod := domain.CalculateModifier(stats.WIS)
			if wisMod > 0 {
				ac += wisMod
			}
		}
	}

	return ac
}

// selectSkills selects skills based on class and background
func selectSkills(class, background string) map[string]bool {
	skills := make(map[string]bool)

	// Add class skills (pick up to 2-4 depending on class)
	classSkillList := classSkills[class]
	classSkillCount := 2
	switch class {
	case "picaro":
		classSkillCount = 4
	case "ranger", "bardo":
		classSkillCount = 3
	}

	for i := 0; i < classSkillCount && i < len(classSkillList); i++ {
		skills[classSkillList[i]] = true
	}

	// Add background skills
	bgSkillList := backgroundSkills[background]
	for _, skill := range bgSkillList {
		skills[skill] = true
	}

	return skills
}

// getClassFeatures returns level 1 features for a class
func getClassFeatures(class string) []domain.Feature {
	featureNames := classFeatures[class]
	if featureNames == nil {
		return []domain.Feature{}
	}

	features := make([]domain.Feature, len(featureNames))
	for i, name := range featureNames {
		features[i] = domain.Feature{
			Name:   name,
			Source: "class",
		}
	}
	return features
}

// getClassInventory returns starting inventory for a class
func getClassInventory(class string) []domain.Item {
	// Basic starting equipment based on class
	inventory := []domain.Item{
		{Name: "Backpack", Quantity: 1, Type: "misc"},
		{Name: "Waterskin", Quantity: 1, Type: "misc"},
		{Name: "Rations", Quantity: 5, Type: "misc"},
		{Name: "Rope (50ft)", Quantity: 1, Type: "misc"},
	}

	switch class {
	case "guerrero", "paladin", "barbaro":
		inventory = append(inventory,
			domain.Item{Name: "Longsword", Quantity: 1, Type: "weapon"},
			domain.Item{Name: "Shield", Quantity: 1, Type: "armor"},
		)
	case "picaro", "brujo", "bardo":
		inventory = append(inventory,
			domain.Item{Name: "Rapier", Quantity: 1, Type: "weapon"},
			domain.Item{Name: "Leather Armor", Quantity: 1, Type: "armor"},
		)
	case "clerigo", "druida":
		inventory = append(inventory,
			domain.Item{Name: "Mace", Quantity: 1, Type: "weapon"},
			domain.Item{Name: "Holy Symbol", Quantity: 1, Type: "misc"},
		)
	case "mago", "hechicero", "artifice":
		inventory = append(inventory,
			domain.Item{Name: "Quarterstaff", Quantity: 1, Type: "weapon"},
			domain.Item{Name: "Spellbook", Quantity: 1, Type: "misc"},
		)
	case "ranger":
		inventory = append(inventory,
			domain.Item{Name: "Longbow", Quantity: 1, Type: "weapon"},
			domain.Item{Name: "Quiver (20 arrows)", Quantity: 1, Type: "misc"},
		)
	case "monje":
		inventory = append(inventory,
			domain.Item{Name: "Shortsword", Quantity: 1, Type: "weapon"},
		)
	}

	return inventory
}

// CreateCharacter creates a new character with proper stat generation
func (s *CharacterService) CreateCharacter(campaignID, name, race, class string, level int, background, alignment string) (*domain.Character, error) {
	if level <= 0 {
		level = 1
	}

	// Normalize inputs to lowercase for matching
	race = strings.ToLower(race)
	class = strings.ToLower(class)
	background = strings.ToLower(background)

	// Assign stats based on class priorities
	stats := assignStats(class)

	// Apply racial bonuses
	applyRaceBonuses(&stats, race)

	// Calculate HP and AC
	hp := calculateHP(class, stats.CON, level)
	ac := calculateAC(class, stats)

	// Select skills
	skills := selectSkills(class, background)

	// Get features and inventory
	features := getClassFeatures(class)
	inventory := getClassInventory(class)

	character := &domain.Character{
		CampaignID: campaignID,
		Name:       name,
		Race:       race,
		Class:      class,
		Level:      level,
		Background: background,
		Alignment:  alignment,
		Status:     "alive",
		Stats:      stats,
		HP: domain.HP{
			Current: hp,
			Maximum: hp,
		},
		AC:            ac,
		Proficiency:   2,
		Skills:        skills,
		Inventory:     inventory,
		Features:      features,
		Relationships: []domain.Relationship{},
	}

	if err := s.repo.Save(character); err != nil {
		return nil, fmt.Errorf("failed to create character: %w", err)
	}

	return character, nil
}

// GetCharacter retrieves a character by name
func (s *CharacterService) GetCharacter(campaignID, name string) (*domain.Character, error) {
	return s.repo.Read(campaignID, name)
}

// ListCharacters returns all characters in a campaign
func (s *CharacterService) ListCharacters(campaignID string) ([]domain.Character, error) {
	return s.repo.List(campaignID)
}

// UpdateCharacter updates a character
func (s *CharacterService) UpdateCharacter(character *domain.Character) error {
	return s.repo.Save(character)
}

// DeleteCharacter deletes a character
func (s *CharacterService) DeleteCharacter(campaignID, name string) error {
	return s.repo.Delete(campaignID, name)
}

// AddRelationship adds a relationship to a character
func (s *CharacterService) AddRelationship(campaignID, characterName string, rel domain.Relationship) error {
	character, err := s.repo.Read(campaignID, characterName)
	if err != nil {
		return fmt.Errorf("character not found: %w", err)
	}

	// Update existing or add new
	found := false
	for i, existing := range character.Relationships {
		if existing.EntityID == rel.EntityID {
			character.Relationships[i] = rel
			found = true
			break
		}
	}
	if !found {
		character.Relationships = append(character.Relationships, rel)
	}

	return s.repo.Save(character)
}
