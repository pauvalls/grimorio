package validators

import (
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLang string
		wantErr  bool
	}{
		{
			name: "pure Spanish markers",
			input: `### Área 1: Entrada
>> **Texto para Leer** *Ves una cueva oscura.*
**Si los PJs entran:** 
- **Consecuencia inmediata:** Los goblins atacan.
- **Recuperación:** Pueden huir.`,
			wantLang: "es",
			wantErr:  false,
		},
		{
			name: "pure English markers",
			input: `### Area 1: Entrance
>> **Read-Aloud Text** *You see a dark cave.*
**If the PCs enter:**
- **Consequence immediate:** The goblins attack.
- **Recovery:** They can flee.`,
			wantLang: "en",
			wantErr:  false,
		},
		{
			name: "mixed Spanish and English rejected",
			input: `### Área 1: Entrada
>> **Texto para Leer** *Ves una cueva.*
**If the PCs enter:**
- **Consequence immediate:** Goblins attack.`,
			wantLang: "mixed",
			wantErr:  true,
		},
		{
			name:     "no markers defaults to es",
			input:    `# Chapter 1\n\nSome content without markers.`,
			wantLang: "es",
			wantErr:  false,
		},
		{
			name: "English NPC fields",
			input: `## Thorin
- **Alignment:** LG
- **Location:** Town square
- **Combat Stats:** AC 15
- **Secret:** He is a spy
- **Quest Involvement:** Main quest`,
			wantLang: "en",
			wantErr:  false,
		},
		{
			name: "Spanish NPC fields",
			input: `## Thorin
- **Alineamiento:** LG
- **Ubicación:** Plaza del pueblo
- **Estadísticas de Combate:** CA 15
- **Secreto:** Es un espía
- **Involucramiento en Quests:** Quest principal`,
			wantLang: "es",
			wantErr:  false,
		},
		{
			name: "mixed NPC fields rejected",
			input: `## Thorin
- **Alineamiento:** LG
- **Location:** Town square
- **Secreto:** Es un espía`,
			wantLang: "mixed",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lang, err := DetectLanguage(tt.input)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for input %q, got nil", tt.name)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for input %q: %v", tt.name, err)
			}
			if lang != tt.wantLang {
				t.Errorf("DetectLanguage() = %q, want %q", lang, tt.wantLang)
			}
		})
	}
}

func TestBilingualPattern(t *testing.T) {
	tests := []struct {
		name    string
		es      string
		en      string
		input   string
		wantMatch bool
	}{
		{
			name:      "matches Spanish variant",
			es:        `Texto para Leer`,
			en:        `Read-Aloud Text`,
			input:     `>> **Texto para Leer**`,
			wantMatch: true,
		},
		{
			name:      "matches English variant",
			es:        `Texto para Leer`,
			en:        `Read-Aloud Text`,
			input:     `>> **Read-Aloud Text**`,
			wantMatch: true,
		},
		{
			name:      "no match for unrelated text",
			es:        `Texto para Leer`,
			en:        `Read-Aloud Text`,
			input:     `>> **Something else**`,
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := BilingualPattern(tt.es, tt.en)
			got := p.MatchString(tt.input)
			if got != tt.wantMatch {
				t.Errorf("BilingualPattern(%q, %q).MatchString(%q) = %v, want %v",
					tt.es, tt.en, tt.input, got, tt.wantMatch)
			}
		})
	}
}
