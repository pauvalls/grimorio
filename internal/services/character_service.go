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

// SaveCharacter saves a character (create or update)
func (s *CharacterService) SaveCharacter(character *domain.Character) error {
	if character == nil {
		return fmt.Errorf("character is required")
	}
	if character.CampaignID == "" {
		return fmt.Errorf("campaign ID is required")
	}
	if character.Name == "" {
		return fmt.Errorf("character name is required")
	}
	if character.Status == "" {
		character.Status = "alive"
	}
	return s.repo.Save(character)
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

// GenerateWithBackstory creates a character with expanded narrative details.
func (s *CharacterService) GenerateWithBackstory(campaignID, name, race, class string, level int, background, alignment string) (*domain.Character, error) {
	character, err := s.CreateCharacter(campaignID, name, race, class, level, background, alignment)
	if err != nil {
		return nil, err
	}

	// Add backstory hooks
	character.BackstoryHooks = s.generateBackstoryHooks(class, background, alignment)

	// Add secrets
	character.Secrets = s.generateSecrets(background, alignment)

	// Add goals
	character.Goals = s.generateGoals(class, alignment)

	// Expand personality
	character.Personality = s.generatePersonalityDepth(class, background, alignment)

	// Generate spells for spellcasting classes
	if s.isSpellcastingClass(class) {
		character.Spells = s.generateSpellsForClass(class, level)
	}

	if err := s.repo.Save(character); err != nil {
		return nil, fmt.Errorf("failed to save character with backstory: %w", err)
	}

	return character, nil
}

// AddBackstoryHooks adds custom backstory hooks to a character.
func (s *CharacterService) AddBackstoryHooks(campaignID, characterName string, hooks []string) error {
	character, err := s.repo.Read(campaignID, characterName)
	if err != nil {
		return fmt.Errorf("character not found: %w", err)
	}

	character.BackstoryHooks = append(character.BackstoryHooks, hooks...)
	return s.repo.Save(character)
}

// GenerateSpells generates appropriate spells for a spellcasting character.
func (s *CharacterService) GenerateSpells(campaignID, characterName string) error {
	character, err := s.repo.Read(campaignID, characterName)
	if err != nil {
		return fmt.Errorf("character not found: %w", err)
	}

	if !s.isSpellcastingClass(character.Class) {
		return nil // Not a spellcasting class
	}

	character.Spells = s.generateSpellsForClass(character.Class, character.Level)
	return s.repo.Save(character)
}

// GeneratePersonalityDepth generates detailed personality traits for a character.
func (s *CharacterService) GeneratePersonalityDepth(campaignID, characterName string) error {
	character, err := s.repo.Read(campaignID, characterName)
	if err != nil {
		return fmt.Errorf("character not found: %w", err)
	}

	character.Personality = s.generatePersonalityDepth(character.Class, character.Background, character.Alignment)
	return s.repo.Save(character)
}

// Helper functions for character generation

func (s *CharacterService) generateBackstoryHooks(class, background, alignment string) []string {
	hooks := []string{}

	// Class-based hooks
	classHooks := map[string][]string{
		"guerrero": {"Veterano de una guerra olvidada", "Busca redención por un fallo en el combate"},
		"barbaro":  {"Exiliado de su tribu", "Busca un nuevo propósito tras perder su clan"},
		"mago":     {"Estudiante de una academia arcana", "Investiga un misterio mágico ancestral"},
		"clerigo":  {"Elegido por su deidad para una misión", "Busca restaurar la fe en su templo"},
		"picaro":   {"Huyó de una vida criminal", "Busca limpiar su nombre"},
	}

	if hooks, ok := classHooks[class]; ok {
		hooks = append(hooks, hooks...)
	}

	// Background-based hooks
	bgHooks := map[string][]string{
		"soldado": {"Deuda de honor con un compañero caído", "Conoce tácticas militares secretas"},
		"acolito": {"Guarda un secreto de su orden religiosa", "Tiene conexiones con otros clérigos"},
		"criminal": {"Perseguido por su antigua organización", "Conoce los bajos fondos de la ciudad"},
		"sabio":   {"Investiga un conocimiento prohibido", "Mentor desaparecido misteriosamente"},
	}

	if hooks, ok := bgHooks[background]; ok {
		hooks = append(hooks, hooks...)
	}

	// Ensure at least 2 hooks
	if len(hooks) < 2 {
		hooks = append(hooks, "Vinculado al destino de la región", "Busca algo que perdió hace mucho tiempo")
	}

	return hooks[:min(len(hooks), 4)]
}

func (s *CharacterService) generateSecrets(background, alignment string) []string {
	secrets := []string{}

	// Alignment-based secrets
	if strings.Contains(alignment, "evil") || strings.Contains(alignment, "maligno") {
		secrets = append(secrets, "Oculta actos oscuros del pasado")
	}
	if strings.Contains(alignment, "chaotic") || strings.Contains(alignment, "caotico") {
		secrets = append(secrets, "Desconfía de la autoridad y tiene planes propios")
	}

	// Background-based secrets
	secretMap := map[string][]string{
		"criminal": {"Identidad falsa", "Recompensa por su captura"},
		"noble":    {"Título usurpado", "Familia en peligro"},
		"sabio":    {"Conocimiento demasiado peligroso", "Experimentos cuestionables"},
	}

	if secrets, ok := secretMap[background]; ok {
		secrets = append(secrets, secrets...)
	}

	if len(secrets) == 0 {
		secrets = append(secrets, "Guarda un secreto personal sin revelar")
	}

	return secrets[:min(len(secrets), 2)]
}

func (s *CharacterService) generateGoals(class, alignment string) []string {
	goals := []string{}

	// Class-based goals
	goalMap := map[string][]string{
		"guerrero": {"Convertirse en el mejor guerrero", "Encontrar un arma legendaria"},
		"mago":     {"Dominar la magia ancestral", "Descubrir secretos arcanos perdidos"},
		"clerigo":  {"Expandir la fe de su deidad", "Proteger a los inocentes"},
		"picaro":   {"Acumular riqueza suficiente", "Vivir sin ataduras"},
	}

	if goals, ok := goalMap[class]; ok {
		goals = append(goals, goals...)
	}

	// Alignment-based goals
	if strings.Contains(alignment, "good") || strings.Contains(alignment, "bueno") {
		goals = append(goals, "Hacer del mundo un lugar mejor")
	}
	if strings.Contains(alignment, "lawful") || strings.Contains(alignment, "legal") {
		goals = append(goals, "Restaurar el orden donde haya caos")
	}

	if len(goals) == 0 {
		goals = append(goals, "Encontrar su lugar en el mundo")
	}

	return goals[:min(len(goals), 3)]
}

func (s *CharacterService) generatePersonalityDepth(class, background, alignment string) domain.Personality {
	personality := domain.Personality{}

	// Class-based traits
	traitMap := map[string][]string{
		"guerrero": {"Valiente en batalla", "Leal a sus compañeros", "Disciplinado"},
		"barbaro":  {"Feroz en combate", "Apasionado", "Directo y honesto"},
		"mago":     {"Curioso intelectual", "Metódico", "Reservado"},
		"clerigo":  {"Devoto", "Compasivo", "Principista"},
		"picaro":   {"Astuto", "Carismático", "Oportunista"},
	}

	if traits, ok := traitMap[class]; ok {
		personality.Traits = traits
	} else {
		personality.Traits = []string{"Determinado", "Adaptable"}
	}

	// Ideals based on alignment
	idealMap := map[string][]string{
		"lawful good":     {"Justicia", "El orden protege a los inocentes"},
		"neutral good":    {"Bondad", "Ayudar a otros es lo más importante"},
		"chaotic good":    {"Libertad", "Las cadenas deben romperse"},
		"lawful neutral":  {"Orden", "La ley es lo único que importa"},
		"true neutral":    {"Equilibrio", "Ni bien ni mal, solo equilibrio"},
		"chaotic neutral": {"Libertad", "Nadie me dice qué hacer"},
		"lawful evil":     {"Tiranía", "El poder justifica los medios"},
		"neutral evil":    {"Egoísmo", "Mis deseos sobre todo"},
		"chaotic evil":    {"Destrucción", "El caos es la única verdad"},
	}

	personality.Ideals = []string{"Propósito", "Busca cumplir su destino"}
	if ideals, ok := idealMap[strings.ToLower(alignment)]; ok {
		personality.Ideals = ideals
	}

	// Bonds
	bondMap := map[string][]string{
		"soldado": {"Mi unidad militar", "Mi tierra natal"},
		"acolito": {"Mi templo", "Mi deidad"},
		"criminal": {"Mi familia criminal", "Mi libertad"},
		"sabio":   {"Mi investigación", "Mi mentor"},
	}

	personality.Bonds = []string{"Mis compañeros de aventura", "Mi propósito"}
	if bonds, ok := bondMap[background]; ok {
		personality.Bonds = append(personality.Bonds, bonds...)
	}

	// Flaws
	flawMap := map[string][]string{
		"guerrero": {"Confío demasiado en mi espada", "Terco en mis decisiones"},
		"mago":     {"Arrogante intelectual", "Descuidado con lo mundane"},
		"clerigo":  {"Dogmático", "Juzgo demasiado rápido"},
		"picaro":   {"Codicioso", "No confío en nadie"},
	}

	personality.Flaws = []string{"Tengo un secreto que guardar", "Puedo ser imprudente"}
	if flaws, ok := flawMap[class]; ok {
		personality.Flaws = append(personality.Flaws, flaws...)
	}

	return personality
}

func (s *CharacterService) generateSpellsForClass(class string, level int) []domain.Spell {
	spells := []domain.Spell{}

	// Cantrips (level 0)
	cantrips := map[string][]string{
		"mago":       {"Fire Bolt", "Prestidigitation", "Mage Hand", "Detect Magic"},
		"clerigo":    {"Sacred Flame", "Guidance", "Thaumaturgy", "Light"},
		"druida":     {"Druidcraft", "Guidance", "Produce Flame", "Shillelagh"},
		"brujo":      {"Eldritch Blast", "Mage Hand", "Prestidigitation", "Minor Illusion"},
		"bardo":      {"Vicious Mockery", "Prestidigitation", "Mage Hand", "Minor Illusion"},
		"hechicero":  {"Fire Bolt", "Prestidigitation", "Mage Hand", "Ray of Frost"},
		"artifice":   {"Mage Hand", "Mending", "Guidance", "Spark"},
		"sangre":     {"Blood Boil", "Prestidigitation", "Mage Hand", "Chill Touch"},
	}

	if cantripList, ok := cantrips[class]; ok {
		for _, cantrip := range cantripList {
			spells = append(spells, domain.Spell{
				Name:  cantrip,
				Level: 0,
			})
		}
	}

	// Level 1+ spells based on character level
	if level >= 1 {
		level1Spells := map[string][]string{
			"mago":      {"Magic Missile", "Shield", "Detect Magic", "Sleep"},
			"clerigo":   {"Bless", "Cure Wounds", "Guiding Bolt", "Sanctuary"},
			"druida":    {"Entangle", "Faerie Fire", "Healing Word", "Thunderwave"},
			"brujo":     {"Hex", "Armor of Agathys", "Witch Bolt", "Hellish Rebuke"},
			"bardo":     {"Healing Word", "Dissonant Whispers", "Faerie Fire", "Cure Wounds"},
			"hechicero": {"Magic Missile", "Shield", "Burning Hands", "Mage Armor"},
		}

		if spellList, ok := level1Spells[class]; ok {
			numSpells := 2
			if level >= 3 {
				numSpells = 4
			}
			for i := 0; i < numSpells && i < len(spellList); i++ {
				spells = append(spells, domain.Spell{
					Name:     spellList[i],
					Level:    1,
					Prepared: true,
				})
			}
		}
	}

	return spells
}

func (s *CharacterService) isSpellcastingClass(class string) bool {
	spellcastingClasses := []string{"mago", "clerigo", "druida", "brujo", "bardo", "hechicero", "artifice", "sangre"}
	for _, c := range spellcastingClasses {
		if strings.EqualFold(class, c) {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
