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
			name: "valid act with 10 areas",
			input: generateValidAct(10, 180),
			wantErr: false,
		},
		{
			name: "too few areas",
			input: generateValidAct(5, 180),
			wantErr: true,
			checks:  []string{"area_count"},
		},
		{
			name: "too many areas",
			input: generateValidAct(20, 180),
			wantErr: true,
			checks:  []string{"area_count"},
		},
		{
			name: "area too short",
			input: generateValidAct(10, 80),
			wantErr: true,
			checks:  []string{"word_count"},
		},
		{
			name: "area too long",
			input: generateValidAct(10, 350),
			wantErr: true,
			checks:  []string{"word_count"},
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
	sb.WriteString("# Acto 1: Test\n\n")
	sb.WriteString("> **Nivel:** 1-2 | **Duración:** 3-4 horas | **Tono:** Oscuro\n")
	sb.WriteString("> **Objetivo:** Encontrar la llave\n")
	sb.WriteString("> **Fallo:** El villano escapa\n\n")

	// Calculate filler words needed per area.
	// Fixed overhead per area: ~70 words (headings, sections, etc.)
	// We split the remainder between Read-Aloud and Description.
	overhead := 70
	fillerCount := (wordsPerArea - overhead) / 2
	if fillerCount < 10 {
		fillerCount = 10
	}
	fillerWords := strings.Repeat("palabra ", fillerCount)

	for i := 1; i <= areaCount; i++ {
		next := i + 1
		prev := i - 1
		if next > areaCount {
			next = 1
		}
		if prev < 1 {
			prev = areaCount
		}

		sb.WriteString("### Área ")
		sb.WriteString(itoa(i))
		sb.WriteString(": Test ")
		sb.WriteString(itoa(i))
		sb.WriteString("\n\n")
		sb.WriteString("> **Read-Aloud:** *")
		sb.WriteString(fillerWords)
		sb.WriteString("*\n\n")
		sb.WriteString("**Descripción:** ")
		sb.WriteString(fillerWords)
		sb.WriteString("\n\n")
		sb.WriteString("**Criaturas:**\n- 1 **Goblin**\n\n")
		sb.WriteString("**Tesoro:**\n- **XP:** 50 XP\n- **Moneda:** 10 gp\n\n")
		sb.WriteString("**Conexiones:**\n")
		sb.WriteString("- → Área ")
		sb.WriteString(itoa(next))
		sb.WriteString("\n")
		sb.WriteString("- ← Área ")
		sb.WriteString(itoa(prev))
		sb.WriteString("\n\n")
		sb.WriteString("**Secretos y Trampas:**\n- **Detectar:** Percepción DC 12\n\n")
		sb.WriteString("**Desarrollo:**\n- Si entran en combate: atacan.\n\n")
	}

	return sb.String()
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	var result []byte
	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}
	return string(result)
}
