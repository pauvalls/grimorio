package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mockHandoutRenderer struct{}

func (m *mockHandoutRenderer) GeneratePlayerMap(mapName string) (string, error) {
	return "# Mapa del Jugador\n\nVestíbulo visible", nil
}

func (m *mockHandoutRenderer) GenerateClueList() (string, error) {
	return "## Pistas\n\n- Sabéis que la llave está en la torre", nil
}

func (m *mockHandoutRenderer) GenerateNPCReference() (string, error) {
	return "## NPCs\n\n| NPC | Ubicación |\n|-----|-----------|\n| Eldrin | Área 3 |", nil
}

func (m *mockHandoutRenderer) GenerateSessionRecap() (string, error) {
	return "## Resumen\n\nExplorasteis 4 áreas.", nil
}

func TestGenerateHandouts_V2(t *testing.T) {
	tmpDir := t.TempDir()

	// Create narrative state
	state := `{"campaign_id":"test","revealed_clues":[{"description":"la llave está en la torre"}],"met_npcs":["Eldrin"],"last_session":{"session_num":3,"areas_visited":["Área 1"],"combats":1,"key_decisions":[]}}`
	_ = os.WriteFile(filepath.Join(tmpDir, "narrative_state.json"), []byte(state), 0644)

	// Create NPCs
	_ = os.MkdirAll(filepath.Join(tmpDir, "npcs"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "npcs", "npcs_and_factions.md"), []byte("### Eldrin\n- **Ubicación:** Área 3\n"), 0644)

	// Create a map
	_ = os.MkdirAll(filepath.Join(tmpDir, "maps"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "maps", "dungeon.md"), []byte("# Dungeon\n\n## Área 1\n- Puerta visible\n"), 0644)

	c := NewWithVersion(tmpDir, "wkhtmltopdf", 2)
	c.SetHandoutRenderer(&mockHandoutRenderer{})

	html, err := c.generateHandouts()
	if err != nil {
		t.Fatalf("generateHandouts() error: %v", err)
	}

	if html == "" {
		t.Fatal("generateHandouts() returned empty string")
	}

	if !strings.Contains(html, "Player Handouts") {
		t.Error("handouts should contain 'Player Handouts' heading")
	}
	if !strings.Contains(html, "Pistas") {
		t.Error("handouts should contain clues")
	}
	if !strings.Contains(html, "NPCs") {
		t.Error("handouts should contain NPC reference")
	}
	if !strings.Contains(html, "Resumen") {
		t.Error("handouts should contain session recap")
	}
}

func TestGenerateHandouts_V1(t *testing.T) {
	c := NewWithVersion("/tmp", "wkhtmltopdf", 1)
	html, err := c.generateHandouts()
	if err != nil {
		t.Fatalf("generateHandouts() v1 error: %v", err)
	}
	if html != "" {
		t.Error("v1 compiler should not generate handouts")
	}
}
