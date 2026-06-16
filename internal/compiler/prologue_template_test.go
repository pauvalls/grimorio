package compiler_test

import (
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/pauvalls/grimorio/internal/compiler"
	"github.com/pauvalls/grimorio/internal/domain"
)

func TestPrologueTemplate_RendersFourParts(t *testing.T) {
	tmplStr, err := compiler.GetTemplate("prologue")
	if err != nil {
		t.Fatalf("GetTemplate('prologue') error: %v", err)
	}

	tmpl, err := template.New("prologue").Parse(tmplStr)
	if err != nil {
		t.Fatalf("Failed to parse prologue template: %v", err)
	}

	prologue := &domain.Prologue{
		CampaignID: "test-campaign",
		Tone:       "heroic",
		Parts: []domain.ProloguePart{
			{Order: 1, Title: "Gancho Narrativo", Content: "Un gran destino aguarda.", IsReadAloud: true},
			{Order: 2, Title: "Trasfondo", Content: "El mundo se encuentra en peligro.", IsReadAloud: false},
			{Order: 3, Title: "Conexiones", Content: "NPCs: El Sabio, La Reina.", IsReadAloud: false},
			{Order: 4, Title: "El Camino por Delante", Content: "La aventura comienza ahora.", IsReadAloud: true},
		},
		GeneratedAt: time.Now(),
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, prologue); err != nil {
		t.Fatalf("Failed to execute prologue template: %v", err)
	}

	output := buf.String()

	// Verify the output contains the prologue wrapper
	if !strings.Contains(output, `<div class="prologue">`) {
		t.Errorf("Output should contain prologue wrapper div")
	}
	if !strings.Contains(output, `</div>`) {
		t.Errorf("Output should close the prologue wrapper")
	}

	// Verify heading
	if !strings.Contains(output, `<h2 id="sec-prologue">`) {
		t.Errorf("Output should contain prologue heading")
	}
	if !strings.Contains(output, "Prologue") {
		t.Errorf("Output should contain 'Prologue' (English default; i18n-english-default)")
	}

	// Verify all 4 parts are rendered
	for i := 1; i <= 4; i++ {
		class := "prologue-part-" + string(rune('0'+i))
		if !strings.Contains(output, class) {
			t.Errorf("Output should contain class '%s' for part %d", class, i)
		}
	}

	// Verify part titles
	expectedTitles := []string{"Gancho Narrativo", "Trasfondo", "Conexiones", "El Camino por Delante"}
	for _, title := range expectedTitles {
		if !strings.Contains(output, title) {
			t.Errorf("Output should contain title '%s'", title)
		}
	}

	// Verify read-aloud markers
	if !strings.Contains(output, `<div class="read-aloud">`) {
		t.Errorf("Output should contain read-aloud div for read-aloud parts")
	}

	// Verify non-read-aloud uses <p> tag
	if !strings.Contains(output, `<p>El mundo se encuentra en peligro.</p>`) {
		t.Errorf("Output should contain <p> tag for non-read-aloud content")
	}

	// Verify read-aloud parts don't use <p> tag
	if strings.Contains(output, `<p>Un gran destino aguarda.</p>`) {
		t.Errorf("Read-aloud content should not be wrapped in <p> tag")
	}
}

func TestPrologueTemplate_RendersWithMinimalData(t *testing.T) {
	tmplStr, err := compiler.GetTemplate("prologue")
	if err != nil {
		t.Fatalf("GetTemplate('prologue') error: %v", err)
	}

	tmpl, err := template.New("prologue").Parse(tmplStr)
	if err != nil {
		t.Fatalf("Failed to parse prologue template: %v", err)
	}

	// Test with empty parts (should still render without error)
	prologue := &domain.Prologue{
		CampaignID:  "test",
		Tone:        "heroic",
		Parts:       []domain.ProloguePart{},
		GeneratedAt: time.Now(),
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, prologue); err != nil {
		t.Fatalf("Failed to execute prologue template with empty parts: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `<div class="prologue">`) {
		t.Errorf("Output should contain prologue wrapper even with empty parts")
	}
}

func TestPrologueTemplate_AllPartsRendered(t *testing.T) {
	tmplStr, err := compiler.GetTemplate("prologue")
	if err != nil {
		t.Fatalf("GetTemplate('prologue') error: %v", err)
	}

	tmpl, err := template.New("prologue").Parse(tmplStr)
	if err != nil {
		t.Fatalf("Failed to parse prologue template: %v", err)
	}

	prologue := &domain.Prologue{
		CampaignID: "test",
		Tone:       "heroic",
		Parts: []domain.ProloguePart{
			{Order: 1, Title: "Hook", Content: "Hook text", IsReadAloud: true},
			{Order: 2, Title: "Context", Content: "Context text", IsReadAloud: false},
			{Order: 3, Title: "Connections", Content: "Connections text", IsReadAloud: false},
			{Order: 4, Title: "Road Ahead", Content: "Road ahead text", IsReadAloud: true},
		},
		GeneratedAt: time.Now(),
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, prologue); err != nil {
		t.Fatalf("Failed to execute prologue template: %v", err)
	}

	output := buf.String()

	// Count occurrences of each part class
	for i := 1; i <= 4; i++ {
		class := "prologue-part-1"
		switch i {
		case 2:
			class = "prologue-part-2"
		case 3:
			class = "prologue-part-3"
		case 4:
			class = "prologue-part-4"
		}
		count := strings.Count(output, class)
		if count != 1 {
			t.Errorf("Class '%s' appears %d times, expected exactly 1", class, count)
		}
	}
}
