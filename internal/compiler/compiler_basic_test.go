package compiler

import (
	"testing"
)

func TestNew(t *testing.T) {
	c := New("/tmp/campaign", "weasyprint")
	if c == nil {
		t.Fatal("New() returned nil")
	}
	if c.CampaignDir != "/tmp/campaign" {
		t.Errorf("CampaignDir = %s, want /tmp/campaign", c.CampaignDir)
	}
	if c.PDFEngine != "weasyprint" {
		t.Errorf("PDFEngine = %s, want weasyprint", c.PDFEngine)
	}
}

func TestNew_DefaultEngine(t *testing.T) {
	c := New("/tmp/campaign", "")
	// Should auto-detect an available engine, not hardcode to wkhtmltopdf
	if c.PDFEngine == "" {
		t.Error("Default PDFEngine should not be empty")
	}
	if c.CompilerVersion != 2 {
		t.Errorf("Default CompilerVersion = %d, want 2", c.CompilerVersion)
	}
}

func TestNewWithVersion(t *testing.T) {
	c := NewWithVersion("/tmp/campaign", "", 1)
	if c.CompilerVersion != 1 {
		t.Errorf("CompilerVersion = %d, want 1", c.CompilerVersion)
	}
	c2 := NewWithVersion("/tmp/campaign", "", 2)
	if c2.CompilerVersion != 2 {
		t.Errorf("CompilerVersion = %d, want 2", c2.CompilerVersion)
	}
	c3 := NewWithVersion("/tmp/campaign", "", 99)
	if c3.CompilerVersion != 2 {
		t.Errorf("Invalid version should default to 2, got %d", c3.CompilerVersion)
	}
}

func TestGetTemplate(t *testing.T) {
	tests := []struct {
		name     string
		tmplType string
		wantErr  bool
	}{
		// "areas" template was removed in v5.0.2 WU7 — grimorio-chapters
		// is the only chapter skill; the legacy `areas` template and
		// the `save_areas` MCP tool are gone.
		{"npc", "npc", false},
		{"monster", "monster", false},
		{"encounter", "encounter", false},
		{"map", "map", false},
		{"lore", "lore", false},
		{"introduction", "introduction", false},
		{"setting-guide", "setting-guide", false},
		{"appendices", "appendices", false},
		{"chapter", "chapter", false},
		{"unknown", "unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := GetTemplate(tt.tmplType)
			if tt.wantErr {
				if err == nil {
					t.Error("GetTemplate() expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("GetTemplate() error: %v", err)
			}
			if tmpl == "" {
				t.Error("GetTemplate() returned empty template")
			}
		})
	}
}
