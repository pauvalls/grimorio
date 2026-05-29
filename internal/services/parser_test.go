package services

import (
	"fmt"
	"testing"
)

func TestEntityParser_ParseChapter(t *testing.T) {
	parser := NewEntityParser()

	sampleChapter := `# Capítulo 1: El Comienzo

## Apertura Narrativa

Los héroes llegan al pueblo.

## NPCs en este Capítulo

### Aldeano Mayor
*Neutral humano*

Un viejo lider del pueblo.

**Estadísticas:** AC 12, HP 15

### Mercader Misterioso
*Neutral bueno humano*

Vende items mágicos raros.

## Encuentros

### Encuentro 1: Emboscada
*Dificultad: Medium*

Los bandidos atacan desde las sombras.

**Monstruos:**
- 3x Bandido
- 1x Líder Bandido

**Recompensas:**
- 100 XP
- 50 gold

## Áreas

### Área 1: Entrada del Pueblo

> Los jugadores ven el pueblo desde la colina.

La entrada está custodiada por guardias.

**Características:**
- **Muralla:** Madera reforzada

**Tesoro:**
- Monedas de cobre (10 gp)

**Desarrollo:** Los guardias permiten el paso si los PJs son amistosos.

**Ganchos:**
- El mercader les ofrece una quest

### Área 2: La Taberna

> Humo y risas salen de la taberna.

Un lugar acogedor para descansar.

**Características:**
- **Barra:** El tabernero conoce rumores

**Tesoro:**
- Poción de curación (50 gp)

**Desarrollo:** El tabernero menciona la cueva misteriosa.

## Consecuencias y Transición

El pueblo queda a salvo.

**Gancho al siguiente capítulo:** Una carta llega con noticias oscuras.
`

	result, err := parser.ParseChapter(sampleChapter, "test-campaign", 1)
	if err != nil {
		t.Fatalf("ParseChapter() unexpected error: %v", err)
	}

	// Should extract 2 areas
	if len(result.Areas) != 2 {
		t.Errorf("ParseChapter() areas = %d, want 2", len(result.Areas))
	}

	// Check first area
	if len(result.Areas) > 0 {
		area1 := result.Areas[0]
		if area1.Title != "Entrada del Pueblo" {
			t.Errorf("Area 1 title = %q, want %q", area1.Title, "Entrada del Pueblo")
		}
		if area1.AreaNumber != 1 {
			t.Errorf("Area 1 number = %d, want 1", area1.AreaNumber)
		}
	}

	// Should extract 2 inline NPCs
	if len(result.NPCs) != 2 {
		t.Errorf("ParseChapter() NPCs = %d, want 2", len(result.NPCs))
	}

	// Check first NPC
	if len(result.NPCs) > 0 {
		npc1 := result.NPCs[0]
		if npc1.Name != "Aldeano Mayor" {
			t.Errorf("NPC 1 name = %q, want %q", npc1.Name, "Aldeano Mayor")
		}
		if npc1.Role != "chapter-inline" {
			t.Errorf("NPC 1 role = %q, want %q", npc1.Role, "chapter-inline")
		}
	}

	// Should extract 1 encounter
	if len(result.Encounters) != 1 {
		t.Errorf("ParseChapter() encounters = %d, want 1", len(result.Encounters))
	}

	// Check encounter
	if len(result.Encounters) > 0 {
		enc1 := result.Encounters[0]
		if enc1.Name != "Emboscada" {
			t.Errorf("Encounter 1 name = %q, want %q", enc1.Name, "Emboscada")
		}
		if enc1.Difficulty != "medium" {
			t.Errorf("Encounter 1 difficulty = %q, want %q", enc1.Difficulty, "medium")
		}
	}
}

func TestEntityParser_ParseChapter_NoAreas(t *testing.T) {
	parser := NewEntityParser()

	content := `# Capítulo 1: El Comienzo

## Apertura Narrativa

Los héroes llegan.

## Consecuencias

Fin del capítulo.
`

	result, err := parser.ParseChapter(content, "test-campaign", 1)
	if err == nil {
		t.Fatal("ParseChapter() expected error for no areas, got nil")
	}
	if result != nil {
		t.Error("ParseChapter() expected nil result for error case")
	}
}

func TestEntityParser_ParseChapter_MultipleAreas(t *testing.T) {
	parser := NewEntityParser()

	var content string
	content = "# Capítulo 1\n\n## Áreas\n\n"
	for i := 1; i <= 12; i++ {
		content += fmt.Sprintf("### Área %d: Lugar %d\n\nDescripción del área.\n\n", i, i)
	}

	result, err := parser.ParseChapter(content, "test-campaign", 1)
	if err != nil {
		t.Fatalf("ParseChapter() unexpected error: %v", err)
	}

	if len(result.Areas) != 12 {
		t.Errorf("ParseChapter() areas = %d, want 12", len(result.Areas))
	}

	// Verify sequential numbering
	for i, area := range result.Areas {
		expectedNum := i + 1
		if area.AreaNumber != expectedNum {
			t.Errorf("Area %d number = %d, want %d", i, area.AreaNumber, expectedNum)
		}
	}
}

func TestEntityParser_ParseChapter_AreaValidation(t *testing.T) {
	parser := NewEntityParser()

	// Test with 16 areas (should still parse, validation is separate)
	var content string
	content = "# Capítulo 1\n\n## Áreas\n\n"
	for i := 1; i <= 16; i++ {
		content += fmt.Sprintf("### Área %d: Lugar %d\n\nDescripción.\n\n", i, i)
	}

	result, err := parser.ParseChapter(content, "test-campaign", 1)
	if err != nil {
		t.Fatalf("ParseChapter() unexpected error: %v", err)
	}

	if len(result.Areas) != 16 {
		t.Errorf("ParseChapter() areas = %d, want 16", len(result.Areas))
	}
}
