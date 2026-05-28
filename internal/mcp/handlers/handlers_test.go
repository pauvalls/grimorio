package handlers

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/services"
)

func setupTestHandlers() (*CampaignHandlers, *CharacterHandlers, *QuestHandlers) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	canonRepo := repository.NewMemoryCanonRepository()
	monsterRepo := repository.NewMemoryMonsterRepository()

	campaignService := services.NewCampaignService(
		campaignRepo, actRepo, charRepo, npcRepo, questRepo, canonRepo, monsterRepo,
		"/tmp/test", "",
	)
	characterService := services.NewCharacterService(charRepo)
	questService := services.NewQuestService(questRepo)

	return NewCampaignHandlers(campaignService),
		NewCharacterHandlers(characterService),
		NewQuestHandlers(questService)
}

func newToolRequest(name string, args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	}
}

func TestHandleCreateCampaign(t *testing.T) {
	handlers, _, _ := setupTestHandlers()
	handler := handlers.HandleCreateCampaign()

	args := map[string]any{
		"name":    "test-campaign",
		"title":   "Test Campaign",
		"setting": "Forgotten Realms",
	}

	result, err := handler(context.Background(), newToolRequest("create_campaign", args))
	if err != nil {
		t.Fatalf("HandleCreateCampaign() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleCreateCampaign() returned error result: %v", result.Content)
	}
}

func TestHandleCreateCampaign_InvalidArgs(t *testing.T) {
	handlers, _, _ := setupTestHandlers()
	handler := handlers.HandleCreateCampaign()

	result, err := handler(context.Background(), newToolRequest("create_campaign", nil))
	if err != nil {
		t.Fatalf("HandleCreateCampaign() error: %v", err)
	}

	if !result.IsError {
		t.Errorf("HandleCreateCampaign() should return error for invalid args")
	}
}

func TestHandleSaveAreas(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	createHandler := handlers.HandleCreateCampaign()
	if _, err := createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "areas-test",
	})); err != nil {
		t.Fatal(err)
	}

	// Content must be 700+ words for v3.1.0 validation
	chapterContent := `## Cómo Dirigir Este Capítulo

### 1. Preparación Previa a la Sesión

**Materiales Necesarios:**
- **Mapas:** Mapa del pueblo de Phandalin, plano del molino, mapa de la cueva
- **Handouts:** Carta de Gundren, folleto del Carretel Roto
- **NPCs con Stats:** Iarno Albrek, Glasstaff, 4 guardias
- **Música/Ambientación:** Sonidos de bosque, música de taberna, tensión de combate

**Repaso de Consecuencias:**
Si los PJs aliaron con los Zhentarim en la sesión anterior, ahora tienen acceso a información privilegiada sobre el mercado negro. Si traicionaron a los miners, serán emboscados en el camino a Phandalin.

**Estado Inicial del Mundo:**
La Guardia de la Ciudad está en alerta máxima después del ataque al almacén. El mercado negro opera solo de noche. El templo ha cerrado sus puertas por temor a represalias.

### 2. Ritmo y Timing Sugerido

**Desglose por Escenas:**
- **Apertura (15-20 min):** Los PJs llegan al pueblo al anochecer. Lluvia ligera. Las calles están vacías excepto por la guardia.
- **Exploración/Investigación (30-45 min):** Los PJs investigan el lugar del crimen. Percepción DC 14 revela huellas, Investigación DC 16 encuentra el arma.
- **Encuentro Principal (45-60 min):** Combate con 4 goblins y 1 ogro en el puente. Alternativa: negociación con el líder goblin.
- **Resolución (15-20 min):** Los PJs reciben recompensa del alcalde. Una carta misteriosa llega con el siguiente gancho.

**Señales de Alerta:**
- **Si los jugadores están aburridos:** 1) Introducir un encuentro sorpresa, 2) Revelar información dramática, 3) Cambiar de escena abruptamente
- **Si la sesión va muy rápida:** 1) Encuentro con mercader ambulante, 2) Puzzle ambiental en el dungeon
- **Si la sesión va muy lenta:** 1) Descripciones extensas de ambientes menores, 2) Combates con minions, 3) Diálogos de NPCs secundarios

### 3. Ganchos de Personaje Obligatorios

**Por Background:**
- **Soldado:** Un antiguo camarada aparece pidiendo ayuda. O reconocés tácticas militares en el enemigo.
- **Criminal:** Contactos del bajo fondo te reconocen. Podés acceder a información privilegiada.
- **Sabio:** Textos antiguos que solo vos podés traducir. Conocimiento arcano revela secretos.
- **Noble:** Tu linaje te da acceso a eventos exclusivos. Otros nobles te tratan como igual.
- **Ermitaño:** Conocés rituales antiguos que otros ignoran. La naturaleza te responde.

**Por Clase:**
- **Guerrero:** Oportunidad: Combate singular con el campeón enemigo. Demostración de fuerza abre puertas.
- **Mago:** Oportunidad: Problema mágico que requiere identificación de hechizos o contramagia.
- **Pícaro:** Oportunidad: Trampas para desactivar, cerraduras para forzar, información para robar.
- **Clérigo:** Oportunidad: Muertos que necesitan ser guiados, maldiciones que requieren intervención divina.

**Momentos de Spotlight:**
Cada PJ debe tener al menos un momento de protagonismo. El pícaro desactiva la trampa que salva al grupo. El clérigo reconoce el símbolo sagrado. El guerrero sostiene la puerta mientras los demás escapan.

### 4. Manejo de Decisiones Críticas

**Puntos de Ramificación:**
1. **Decisión:** Los PJs ayudan a la facción A o B
   - **Consecuencia inmediata:** La facción elegida proporciona recursos
   - **Consecuencia a largo plazo:** La facción rechazada se vuelve enemiga en el Acto N+2
   - **Recovery path:** Si evitan decidir, ambas facciones los ignoran temporalmente

2. **Decisión:** Atacar de frente o infiltrarse
   - **Consecuencia inmediata:** Combate directo vs sigilo
   - **Consecuencia a largo plazo:** Alerta general o acceso sorpresa
   - **Recovery path:** Si son detectados infiltrándose, combate en desventaja

### 5. Ajustes por Tamaño de Grupo

| Grupo | Ajuste |
|-------|--------|
| **2-3 PJs** | Reducir enemigos a 2, CDs -2, aliado extra |
| **4 PJs** | Sin cambios — configuración por defecto |
| **5-6 PJs** | Enemigos extra, CDs +2, jefe con habilidades adicionales |

### 6. Tesoro y Recompensas

**Recompensas Principales:**
- 500 po divididos del tesoro del jefe
- Objeto mágico menor (poción de curación +2)
- Información crítica para el siguiente capítulo

**Recompensas Secundarias:**
- 100 po por completar quest secundaria
- Alianza con facción local
- Acceso a comerciante especial

### 7. Transición al Siguiente Capítulo

**Gancho Final:**
La carta encontrada menciona una ubicación en Neverwinter. Un NPC clave revela que el verdadero villano es alguien cercano. El objeto encontrado brilla cuando se acerca a cierto tipo de magia.

**Estado del Mundo Post-Capítulo:**
La facción aliada ahora controla el distrito sur. El mercado negro se ha mudado al puerto. La guardia ha aumentado patrullas nocturnas.`

	areasHandler := handlers.HandleSaveAreas()
	args := map[string]any{
		"campaign":       "areas-test",
		"chapter_number": float64(1),
		"title":          "The Beginning",
		"content":        chapterContent,
	}

	result, err := areasHandler(context.Background(), newToolRequest("save_areas", args))
	if err != nil {
		t.Fatalf("HandleSaveAreas() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleSaveAreas() returned error result: %v", result.Content)
	}
}

func TestHandleGenerateCharacter(t *testing.T) {
	_, charHandlers, _ := setupTestHandlers()

	// Create campaign first
	campaignRepo := repository.NewMemoryCampaignRepository()
	if err := campaignRepo.Create(&domain.Campaign{Name: "char-test", Title: "Char Test"}); err != nil {
		t.Fatal(err)
	}

	handler := charHandlers.HandleGenerateCharacter()
	args := map[string]any{
		"campaign":   "char-test",
		"name":       "Gandalf",
		"race":       "humano",
		"class":      "mago",
		"level":      float64(5),
		"background": "sabio",
		"alignment":  "LG",
	}

	result, err := handler(context.Background(), newToolRequest("generate_character", args))
	if err != nil {
		t.Fatalf("HandleGenerateCharacter() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleGenerateCharacter() returned error result: %v", result.Content)
	}
}

func TestHandleCreateQuest(t *testing.T) {
	_, _, questHandlers := setupTestHandlers()

	// Create campaign first
	campaignRepo := repository.NewMemoryCampaignRepository()
	if err := campaignRepo.Create(&domain.Campaign{Name: "quest-test", Title: "Quest Test"}); err != nil {
		t.Fatal(err)
	}

	handler := questHandlers.HandleCreateQuest()
	args := map[string]any{
		"campaign":    "quest-test",
		"quest_title": "Find the Sword",
		"quest_type":  "main",
		"hook":        "A stranger approaches...",
		"stakes":      "The kingdom's fate",
		"reward":      "1000 gold",
	}

	result, err := handler(context.Background(), newToolRequest("create_personal_quest", args))
	if err != nil {
		t.Fatalf("HandleCreateQuest() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleCreateQuest() returned error result: %v", result.Content)
	}
}

func TestHandleSaveLore(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	// Create campaign first
	createHandler := handlers.HandleCreateCampaign()
	if _, err := createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "lore-test",
	})); err != nil {
		t.Fatal(err)
	}

	handler := handlers.HandleSaveLore()
	args := map[string]any{
		"campaign": "lore-test",
		"content":  "# World History\n\nLong ago...",
	}

	result, err := handler(context.Background(), newToolRequest("save_lore", args))
	if err != nil {
		t.Fatalf("HandleSaveLore() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleSaveLore() returned error result: %v", result.Content)
	}
}

func TestHandleSaveLore_MissingCampaign(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	handler := handlers.HandleSaveLore()
	args := map[string]any{
		"content": "Some lore",
	}

	result, err := handler(context.Background(), newToolRequest("save_lore", args))
	if err != nil {
		t.Fatalf("HandleSaveLore() error: %v", err)
	}

	if !result.IsError {
		t.Errorf("HandleSaveLore() should return error for missing campaign")
	}
}

func TestHandleSaveNPCs(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	// Create campaign first
	createHandler := handlers.HandleCreateCampaign()
	if _, err := createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "npc-test",
	})); err != nil {
		t.Fatal(err)
	}

	handler := handlers.HandleSaveNPCs()
	args := map[string]any{
		"campaign": "npc-test",
		"content":  "# NPCs\n\n## Gandalf\nA wizard...",
	}

	result, err := handler(context.Background(), newToolRequest("save_npcs", args))
	if err != nil {
		t.Fatalf("HandleSaveNPCs() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleSaveNPCs() returned error result: %v", result.Content)
	}
}

func TestHandleSaveEncounters(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	// Create campaign first
	createHandler := handlers.HandleCreateCampaign()
	if _, err := createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "enc-test",
	})); err != nil {
		t.Fatal(err)
	}

	handler := handlers.HandleSaveEncounters()
	args := map[string]any{
		"campaign": "enc-test",
		"content":  "# Encounters\n\n## Ambush\nA bandit ambush...",
	}

	result, err := handler(context.Background(), newToolRequest("save_encounters", args))
	if err != nil {
		t.Fatalf("HandleSaveEncounters() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleSaveEncounters() returned error result: %v", result.Content)
	}
}

func TestHandleSaveBestiary(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	// Create campaign first
	createHandler := handlers.HandleCreateCampaign()
	if _, err := createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "best-test",
	})); err != nil {
		t.Fatal(err)
	}

	handler := handlers.HandleSaveBestiary()
	args := map[string]any{
		"campaign": "best-test",
		"content":  "# Bestiary\n\n## Goblin\nA small creature...",
	}

	result, err := handler(context.Background(), newToolRequest("save_bestiary", args))
	if err != nil {
		t.Fatalf("HandleSaveBestiary() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleSaveBestiary() returned error result: %v", result.Content)
	}
}

func TestHandleSaveMaps(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	// Create campaign first
	createHandler := handlers.HandleCreateCampaign()
	if _, err := createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "map-test",
	})); err != nil {
		t.Fatal(err)
	}

	handler := handlers.HandleSaveMaps()
	args := map[string]any{
		"campaign": "map-test",
		"content":  "# Maps\n\n## Dungeon\nA dark dungeon...",
	}

	result, err := handler(context.Background(), newToolRequest("save_maps", args))
	if err != nil {
		t.Fatalf("HandleSaveMaps() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleSaveMaps() returned error result: %v", result.Content)
	}
}

func TestHandleSaveIntroduction(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	// Create campaign first
	createHandler := handlers.HandleCreateCampaign()
	if _, err := createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "intro-test",
	})); err != nil {
		t.Fatal(err)
	}

	handler := handlers.HandleSaveIntroduction()
	args := map[string]any{
		"campaign": "intro-test",
		"content":  "# Introduction\n\n## Story Overview\n...",
	}

	result, err := handler(context.Background(), newToolRequest("save_introduction", args))
	if err != nil {
		t.Fatalf("HandleSaveIntroduction() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleSaveIntroduction() returned error result: %v", result.Content)
	}
}

func TestHandleSaveIntroduction_MissingCampaign(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	handler := handlers.HandleSaveIntroduction()
	args := map[string]any{
		"content": "Some content",
	}

	result, err := handler(context.Background(), newToolRequest("save_introduction", args))
	if err != nil {
		t.Fatalf("HandleSaveIntroduction() error: %v", err)
	}

	if !result.IsError {
		t.Errorf("HandleSaveIntroduction() should return error for missing campaign")
	}
}

func TestHandleSaveSettingGuide(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	// Create campaign first
	createHandler := handlers.HandleCreateCampaign()
	if _, err := createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "setting-test",
	})); err != nil {
		t.Fatal(err)
	}

	handler := handlers.HandleSaveSettingGuide()
	args := map[string]any{
		"campaign": "setting-test",
		"content":  "# Setting Guide\n\n## Geography\n...",
	}

	result, err := handler(context.Background(), newToolRequest("save_setting_guide", args))
	if err != nil {
		t.Fatalf("HandleSaveSettingGuide() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleSaveSettingGuide() returned error result: %v", result.Content)
	}
}

func TestHandleSaveSettingGuide_MissingCampaign(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	handler := handlers.HandleSaveSettingGuide()
	args := map[string]any{
		"content": "Some content",
	}

	result, err := handler(context.Background(), newToolRequest("save_setting_guide", args))
	if err != nil {
		t.Fatalf("HandleSaveSettingGuide() error: %v", err)
	}

	if !result.IsError {
		t.Errorf("HandleSaveSettingGuide() should return error for missing campaign")
	}
}

func TestHandleSaveAppendices(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	// Create campaign first
	createHandler := handlers.HandleCreateCampaign()
	if _, err := createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "appendix-test",
	})); err != nil {
		t.Fatal(err)
	}

	handler := handlers.HandleSaveAppendices()
	args := map[string]any{
		"campaign": "appendix-test",
		"content":  "# Appendices\n\n## Appendix A: Magic Items\n...",
	}

	result, err := handler(context.Background(), newToolRequest("save_appendices", args))
	if err != nil {
		t.Fatalf("HandleSaveAppendices() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleSaveAppendices() returned error result: %v", result.Content)
	}
}

func TestHandleSaveAppendices_MissingCampaign(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	handler := handlers.HandleSaveAppendices()
	args := map[string]any{
		"content": "Some content",
	}

	result, err := handler(context.Background(), newToolRequest("save_appendices", args))
	if err != nil {
		t.Fatalf("HandleSaveAppendices() error: %v", err)
	}

	if !result.IsError {
		t.Errorf("HandleSaveAppendices() should return error for missing campaign")
	}
}
