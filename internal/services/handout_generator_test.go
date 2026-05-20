package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandoutGenerator_GeneratePlayerMap(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a map with secrets
	mapContent := `# Mapa de la Mazmorra

## Área 1: Vestíbulo
- Puerta secreta al norte (DC 15)
- Trampa en el suelo (DC 14)
- Tesoro: 50 gp

## Área 2: Sala del Trono
- NPC: **Lord Blackthorn**
- Pista: La llave está en Área 3
`
	mapPath := filepath.Join(tmpDir, "maps", "dungeon.md")
	_ = os.MkdirAll(filepath.Dir(mapPath), 0755)
	_ = os.WriteFile(mapPath, []byte(mapContent), 0644)

	gen := NewHandoutGenerator(tmpDir)
	redacted, err := gen.GeneratePlayerMap("dungeon")
	if err != nil {
		t.Fatalf("GeneratePlayerMap() error: %v", err)
	}

	// Player map should NOT contain secrets
	if strings.Contains(redacted, "secreta") {
		t.Error("player map should not contain 'secreta'")
	}
	if strings.Contains(redacted, "Trampa") {
		t.Error("player map should not contain 'Trampa'")
	}
	if strings.Contains(redacted, "Tesoro") {
		t.Error("player map should not contain 'Tesoro'")
	}
	if strings.Contains(redacted, "Pista") {
		t.Error("player map should not contain 'Pista'")
	}

	// Player map SHOULD contain visible elements
	if !strings.Contains(redacted, "Vestíbulo") {
		t.Error("player map should contain 'Vestíbulo'")
	}
	if !strings.Contains(redacted, "Sala del Trono") {
		t.Error("player map should contain 'Sala del Trono'")
	}
	if !strings.Contains(redacted, "Lord Blackthorn") {
		t.Error("player map should contain visible NPC 'Lord Blackthorn'")
	}
}

func TestHandoutGenerator_GenerateClueList(t *testing.T) {
	tmpDir := t.TempDir()

	// Create narrative state with clues
	state := `{
		"campaign_id": "test",
		"revealed_clues": [
			{"id": "clue-1", "description": "La llave está en la torre", "session": 1},
			{"id": "clue-2", "description": "Lord Blackthorn es un vampiro", "session": 2}
		]
	}`
	statePath := filepath.Join(tmpDir, "canon", "narrative_state.json")
	_ = os.MkdirAll(filepath.Join(tmpDir, "canon"), 0755)
	_ = os.WriteFile(statePath, []byte(state), 0644)

	gen := NewHandoutGenerator(tmpDir)
	clues, err := gen.GenerateClueList()
	if err != nil {
		t.Fatalf("GenerateClueList() error: %v", err)
	}

	if !strings.Contains(clues, "Sabéis que") {
		t.Error("clue list should use second-person present tense 'Sabéis que'")
	}
	if !strings.Contains(clues, "La llave está en la torre") {
		t.Error("clue list should contain first clue")
	}
	if !strings.Contains(clues, "Lord Blackthorn es un vampiro") {
		t.Error("clue list should contain second clue")
	}
}

func TestHandoutGenerator_GenerateNPCReference(t *testing.T) {
	tmpDir := t.TempDir()

	// Create NPCs file
	npcs := `# NPCs y Facciones

## NPCs Principales

### Eldrin

- **Raza/Clase:** Elfo Mago
- **Alineamiento:** NG
- **Ubicación:** Área 3
- **Estadísticas de Combate:** CA 12, PG 18
- **Rol en la historia:** Aliado
- **Personalidad:** Valiente
- **Motivación:** Proteger la aldea
- **Secreto:** Es un espía
- **Involucramiento en Quests:** Quest: "La Llave Perdida" — informante
- **Conexiones:** Amigo de Thorn
- **Cita típica:** "Nunca retrocedo"

### Thorn

- **Raza/Clase:** Humano Guerrero
- **Alineamiento:** LG
- **Ubicación:** Área 5
- **Rol en la historia:** Guardián
`
	npcPath := filepath.Join(tmpDir, "npcs", "npcs_and_factions.md")
	_ = os.MkdirAll(filepath.Dir(npcPath), 0755)
	_ = os.WriteFile(npcPath, []byte(npcs), 0644)

	// Create narrative state with met NPCs
	state := `{
		"campaign_id": "test",
		"met_npcs": ["Eldrin"]
	}`
	statePath := filepath.Join(tmpDir, "canon", "narrative_state.json")
	_ = os.MkdirAll(filepath.Join(tmpDir, "canon"), 0755)
	_ = os.WriteFile(statePath, []byte(state), 0644)

	gen := NewHandoutGenerator(tmpDir)
	ref, err := gen.GenerateNPCReference()
	if err != nil {
		t.Fatalf("GenerateNPCReference() error: %v", err)
	}

	if !strings.Contains(ref, "Eldrin") {
		t.Error("NPC reference should contain met NPC 'Eldrin'")
	}
	if strings.Contains(ref, "Thorn") {
		t.Error("NPC reference should NOT contain unmet NPC 'Thorn'")
	}
	if !strings.Contains(ref, "Elfo Mago") {
		t.Error("NPC reference should contain race/class info")
	}
	if !strings.Contains(ref, "Área 3") {
		t.Error("NPC reference should contain location")
	}
}

func TestHandoutGenerator_GenerateSessionRecap(t *testing.T) {
	tmpDir := t.TempDir()

	state := `{
		"campaign_id": "test",
		"last_session": {
			"session_num": 3,
			"areas_visited": ["Área 1", "Área 2", "Área 3", "Área 4"],
			"combats": 2,
			"key_decisions": ["Dejaron escapar al villano"]
		}
	}`
	statePath := filepath.Join(tmpDir, "canon", "narrative_state.json")
	_ = os.MkdirAll(filepath.Join(tmpDir, "canon"), 0755)
	_ = os.WriteFile(statePath, []byte(state), 0644)

	gen := NewHandoutGenerator(tmpDir)
	recap, err := gen.GenerateSessionRecap()
	if err != nil {
		t.Fatalf("GenerateSessionRecap() error: %v", err)
	}

	if !strings.Contains(recap, "Sesión 3") {
		t.Error("recap should mention session number")
	}
	if !strings.Contains(recap, "4 áreas") {
		t.Error("recap should mention areas visited")
	}
	if !strings.Contains(recap, "2 combates") {
		t.Error("recap should mention combats")
	}
	if !strings.Contains(recap, "Dejaron escapar al villano") {
		t.Error("recap should mention key decisions")
	}
}
