package validators

import (
	"strings"
	"testing"
)

func TestValidateAreaMarkdown(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		checks  []string
	}{
		{
			name: "valid chapter with 10 areas",
			input: generateValidAct(10, 300),
			wantErr: false,
		},
		{
			name: "valid chapter with 7 areas (minimum)",
			input: generateValidAct(7, 300),
			wantErr: false,
		},
		{
			name: "too few areas (below 7)",
			input: generateValidAct(5, 300),
			wantErr: true,
			checks:  []string{"area_count"},
		},
		{
			name: "too many areas (above 15)",
			input: generateValidAct(20, 300),
			wantErr: true,
			checks:  []string{"area_count"},
		},
		{
			name: "area too short (below 150)",
			input: generateValidAct(10, 80),
			wantErr: true,
			checks:  []string{"word_count"},
		},
		{
			name: "area too long (above 600)",
			input: generateValidAct(10, 700),
			wantErr: true,
			checks:  []string{"word_count"},
		},
		{
			name: "area at 450 words passes (WotC range)",
			input: generateValidAct(10, 450),
			wantErr: false,
		},
		{
			name: "non-numeric DC",
			input: `# Acto 1

### Área 1: Test

> **Read-Aloud:** *Texto.*

**Descripción:**
- **Percepción DC alto:** algo

**Criaturas:**
- 1 **Goblin**

**Tesoro:**
- **XP:** 50 XP

**Conexiones:**
- → Área 2

**Secretos y Trampas:**
- **Detectar:** Percepción DC alto

**Desarrollo:**
- Si entran en combate: atacan.

### Área 2: Test 2

> **Read-Aloud:** *Texto.*

**Descripción:** Descripción con suficientes palabras para pasar el conteo de palabras sin problemas. Tiene que ser larga.

**Criaturas:**
- 1 **Goblin**

**Tesoro:**
- **XP:** 50 XP

**Conexiones:**
- ← Área 1

**Secretos y Trampas:**
- **Detectar:** Percepción DC 12

**Desarrollo:**
- Si entran en combate: atacan.
`,
			wantErr: true,
			checks:  []string{"numeric_dc"},
		},
		{
			name: "one-way connection",
			input: `# Acto 1

### Área 1: Test

> **Read-Aloud:** *Texto largo y descriptivo para alcanzar el mínimo de palabras requerido en esta área de prueba.*

**Descripción:** Descripción detallada con muchas palabras para cumplir el requisito de conteo de palabras mínimo y máximo en el validador de áreas.

**Criaturas:**
- 1 **Goblin**

**Tesoro:**
- **XP:** 50 XP

**Conexiones:**
- → Área 2

**Secretos y Trampas:**
- **Detectar:** Percepción DC 12

**Desarrollo:**
- Si entran en combate: atacan.

### Área 2: Test 2

> **Read-Aloud:** *Texto largo y descriptivo para alcanzar el mínimo de palabras requerido en esta área de prueba.*

**Descripción:** Descripción detallada con muchas palabras para cumplir el requisito de conteo de palabras mínimo y máximo en el validador de áreas.

**Criaturas:**
- 1 **Goblin**

**Tesoro:**
- **XP:** 50 XP

**Conexiones:**
- → Área 3

**Secretos y Trampas:**
- **Detectar:** Percepción DC 12

**Desarrollo:**
- Si entran en combate: atacan.
`,
			wantErr: true,
			checks:  []string{"bidirectional"},
		},
		{
			name: "missing treasure with creatures",
			input: `# Acto 1

### Área 1: Test

> **Read-Aloud:** *Texto largo y descriptivo para alcanzar el mínimo de palabras requerido en esta área de prueba.*

**Descripción:** Descripción detallada con muchas palabras para cumplir el requisito de conteo de palabras mínimo y máximo en el validador de áreas.

**Criaturas:**
- 1 **Goblin**

**Conexiones:**
- → Área 2

**Secretos y Trampas:**
- **Detectar:** Percepción DC 12

**Desarrollo:**
- Si entran en combate: atacan.

### Área 2: Test 2

> **Read-Aloud:** *Texto largo y descriptivo para alcanzar el mínimo de palabras requerido en esta área de prueba.*

**Descripción:** Descripción detallada con muchas palabras para cumplir el requisito de conteo de palabras mínimo y máximo en el validador de áreas.

**Criaturas:**
- 1 **Goblin**

**Conexiones:**
- ← Área 1

**Secretos y Trampas:**
- **Detectar:** Percepción DC 12

**Desarrollo:**
- Si entran en combate: atacan.
`,
			wantErr: true,
			checks:  []string{"treasure"},
		},
		{
			name: "empty area",
			input: `# Acto 1

### Área 1: Empty

> **Read-Aloud:** *Texto largo y descriptivo para alcanzar el mínimo de palabras requerido en esta área de prueba que está vacía.*

**Descripción:** Descripción detallada con muchas palabras para cumplir el requisito de conteo de palabras mínimo y máximo en el validador de áreas.

**Conexiones:**
- → Área 2

**Desarrollo:**
- Nada pasa.

### Área 2: Test 2

> **Read-Aloud:** *Texto largo y descriptivo para alcanzar el mínimo de palabras requerido en esta área de prueba.*

**Descripción:** Descripción detallada con muchas palabras para cumplir el requisito de conteo de palabras mínimo y máximo en el validador de áreas.

**Criaturas:**
- 1 **Goblin**

**Tesoro:**
- **XP:** 50 XP

**Conexiones:**
- ← Área 1

**Secretos y Trampas:**
- **Detectar:** Percepción DC 12

**Desarrollo:**
- Si entran en combate: atacan.
`,
			wantErr: true,
			checks:  []string{"interactive_element"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateAreaMarkdown(tt.input)
			if tt.wantErr && result.Valid {
				t.Errorf("expected validation to fail, but it passed. Errors: %v", result.Errors)
			}
			if !tt.wantErr && !result.Valid {
				t.Errorf("expected validation to pass, but it failed. Errors: %v", result.Errors)
			}
			for _, check := range tt.checks {
				found := false
				for _, err := range result.Errors {
					if err.Field == check {
						found = true
						break
					}
				}
				if tt.wantErr && !found {
					t.Errorf("expected error for field %q, but not found. Errors: %v", check, result.Errors)
				}
				if !tt.wantErr && found {
					t.Errorf("unexpected error for field %q. Errors: %v", check, result.Errors)
				}
			}
		})
	}
}

func generateValidAct(areaCount, wordsPerArea int) string {
	var sb strings.Builder
	sb.WriteString("# Chapter 1: Test\n\n")
	sb.WriteString("> **Level:** 1-2 | **Duration:** 3-4 hours | **Tone:** Dark\n")
	sb.WriteString("> **Objective:** Find the key\n")
	sb.WriteString("> **Failure:** The villain escapes\n\n")

	// Calculate filler words needed per area.
	// Fixed overhead per area: ~70 words (headings, sections, etc.)
	// We split the remainder between Read-Aloud and Description.
	overhead := 70
	fillerCount := (wordsPerArea - overhead) / 2
	if fillerCount < 10 {
		fillerCount = 10
	}
	fillerWords := strings.Repeat("word ", fillerCount)

	for i := 1; i <= areaCount; i++ {
		next := i + 1
		prev := i - 1
		if next > areaCount {
			next = 1
		}
		if prev < 1 {
			prev = areaCount
		}

		sb.WriteString("### Area ")
		sb.WriteString(itoa(i))
		sb.WriteString(": Test ")
		sb.WriteString(itoa(i))
		sb.WriteString("\n\n")
		sb.WriteString("> **Read-Aloud Text:** *")
		sb.WriteString(fillerWords)
		sb.WriteString("*\n\n")
		sb.WriteString("**Description:** ")
		sb.WriteString(fillerWords)
		sb.WriteString("\n\n")
		sb.WriteString("**Creatures:**\n- 1 **Goblin**\n\n")
		sb.WriteString("**Treasure:**\n- **XP:** 50 XP\n- **Coin:** 10 gp\n\n")
		sb.WriteString("**Connections:**\n")
		sb.WriteString("- → Area ")
		sb.WriteString(itoa(next))
		sb.WriteString("\n")
		sb.WriteString("- ← Area ")
		sb.WriteString(itoa(prev))
		sb.WriteString("\n\n")
		sb.WriteString("**Secrets and Traps:**\n- **Detect:** Perception DC 12\n\n")
		sb.WriteString("**Development:**\n- If they enter combat: they attack.\n\n")
	}

	return sb.String()
}

func TestValidateAreaMarkdown_LetteredAreas(t *testing.T) {
	// Generate 7 lettered areas with enough words each
	var sb strings.Builder
	sb.WriteString("# Chapter 1: Urban Investigation\n\n")

	letters := []string{"E1", "E2", "E3", "E4", "E5", "E6", "E7"}
	names := []string{"Tavern Entrance", "Main Hall", "Kitchen", "Cellar", "Back Alley", "Rooftop", "Secret Room"}
	filler := strings.Repeat("word ", 100)

	for i, letter := range letters {
		next := letters[(i+1)%len(letters)]
		prev := letters[(i-1+len(letters))%len(letters)]

		sb.WriteString("### Area " + letter + ": " + names[i] + "\n\n")
		sb.WriteString("> **Read-Aloud Text:** *" + filler + "*\n\n")
		sb.WriteString("**Description:** " + filler + "\n\n")
		sb.WriteString("**Creatures:**\n- 1 **Guard**\n\n")
		sb.WriteString("**Treasure:**\n- **XP:** 25 XP\n\n")
		sb.WriteString("**Connections:**\n")
		sb.WriteString("- → Area " + next + "\n")
		sb.WriteString("- ← Area " + prev + "\n\n")
		sb.WriteString("**Secrets and Traps:**\n- **Detect:** Perception DC 12\n\n")
		sb.WriteString("**Development:**\n- If the PCs investigate: they find clues.\n\n")
	}

	result := ValidateAreaMarkdown(sb.String())
	if !result.Valid {
		t.Errorf("expected lettered areas to pass validation, got errors: %v", result.Errors)
	}
}

func TestValidateChapterWordCount(t *testing.T) {
	tests := []struct {
		name      string
		wordCount int
		wantErr   bool
	}{
		{"below minimum (2000)", 2000, true},
		{"at minimum (3000)", 3000, false},
		{"within range (6000)", 6000, false},
		{"at maximum (16000)", 16000, false},
		{"above maximum (17000)", 17000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate text with approximately the desired word count
			words := strings.Repeat("word ", tt.wordCount)
			result := ValidateChapterWordCount(words)
			if tt.wantErr && result.Valid {
				t.Errorf("expected validation to fail for %d words", tt.wordCount)
			}
			if !tt.wantErr && !result.Valid {
				t.Errorf("expected validation to pass for %d words, got errors: %v", tt.wordCount, result.Errors)
			}
		})
	}
}


