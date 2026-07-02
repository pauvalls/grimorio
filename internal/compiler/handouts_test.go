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
	_ = os.MkdirAll(filepath.Join(tmpDir, "canon"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "canon", "narrative_state.json"), []byte(state), 0644)

	// Create NPCs
	_ = os.MkdirAll(filepath.Join(tmpDir, "npcs"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "npcs", "npcs_and_factions.md"), []byte("### Eldrin\n- **Ubicación:** Área 3\n"), 0644)

	// Create a map
	_ = os.MkdirAll(filepath.Join(tmpDir, "maps"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "maps", "dungeon.md"), []byte("# Dungeon\n\n## Área 1\n- Puerta visible\n"), 0644)

	c := NewWithVersion(tmpDir, "", 2)
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
	c := NewWithVersion("/tmp", "", 1)
	html, err := c.generateHandouts()
	if err != nil {
		t.Fatalf("generateHandouts() v1 error: %v", err)
	}
	if html != "" {
		t.Error("v1 compiler should not generate handouts")
	}
}

// TestGenerateHandouts_NilRendererReturnsEmpty asserts REQ-2.3: when
// SetHandoutRenderer was never called (the renderer is nil), the
// function returns the empty string — no <div class="handout-page">,
// no "Player Handouts" h1, no placeholder HTML. The rendered PDF
// has no page break for the handouts section.
func TestGenerateHandouts_NilRendererReturnsEmpty(t *testing.T) {
	c := New(t.TempDir(), "")
	// Intentionally do NOT call c.SetHandoutRenderer.
	out, err := c.generateHandouts()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "" {
		t.Errorf("nil renderer should return empty string, got: %.200q", out)
	}
	if strings.Contains(out, "handout-page") {
		t.Errorf("nil renderer output should not contain 'handout-page' marker, got: %.200q", out)
	}
	if strings.Contains(out, "Player Handouts") {
		t.Errorf("nil renderer output should not contain 'Player Handouts' heading, got: %.200q", out)
	}
}

// stubHandoutRenderer is a minimal HandoutRenderer that returns the
// values pre-set on the struct. Used by TestGenerateHandouts_StubRendererPreserved
// to verify the positive path (a wired renderer is still called and
// its output reaches the final HTML).
type stubHandoutRenderer struct {
	clues string
	npcs  string
	recap string
	map_  string
}

func (s *stubHandoutRenderer) GenerateClueList() (string, error) {
	return s.clues, nil
}
func (s *stubHandoutRenderer) GenerateNPCReference() (string, error) {
	return s.npcs, nil
}
func (s *stubHandoutRenderer) GenerateSessionRecap() (string, error) {
	return s.recap, nil
}
func (s *stubHandoutRenderer) GeneratePlayerMap(name string) (string, error) {
	return s.map_, nil
}

// TestGenerateHandouts_StubRendererPreserved asserts the positive
// path: a wired renderer is called, its output appears in the
// generated HTML inside a <div class="handout-page">. This is the
// companion to TestGenerateHandouts_NilRendererReturnsEmpty — the
// regression check for "the early-return does not skip wired renderers".
func TestGenerateHandouts_StubRendererPreserved(t *testing.T) {
	c := New(t.TempDir(), "")
	c.SetHandoutRenderer(&stubHandoutRenderer{
		clues: "Test clue list",
		npcs:  "Test NPC reference",
	})
	out, err := c.generateHandouts()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, `<div class="handout-page"`) {
		t.Errorf("stub renderer output should contain handout-page div, got: %.200q", out)
	}
	if !strings.Contains(out, "Test clue list") {
		t.Errorf("stub renderer output missing 'Test clue list', got:\n%s", out)
	}
	if !strings.Contains(out, "Test NPC reference") {
		t.Errorf("stub renderer output missing 'Test NPC reference', got:\n%s", out)
	}
}
