package validators

import (
	"testing"
)

func TestValidateIntegration(t *testing.T) {
	tests := []struct {
		name    string
		act     string
		bestiary string
		npcs    string
		wantErr bool
		checks  []string
	}{
		{
			name: "all references valid",
			act: `# Acto 1

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
- ← Área 1

**Secretos y Trampas:**
- **Detectar:** Percepción DC 12

**Desarrollo:**
- Si entran en combate: atacan.
`,
			bestiary: `## Goblin

*Pequeño humanoide, NE*

**Clase de Armadura:** 15
**Puntos de Golpe:** 7 (2d6)
**Desafío:** 1/4 (50 PX)
`,
			npcs: `## Eldrin

- **Raza/Clase:** Elfo Mago
- **Alineamiento:** NG
`,
			wantErr: false,
		},
		{
			name: "missing creature in bestiary",
			act: `# Acto 1

### Área 1: Test

> **Read-Aloud:** *Texto largo y descriptivo para alcanzar el mínimo de palabras requerido en esta área de prueba.*

**Descripción:** Descripción detallada con muchas palabras para cumplir el requisito de conteo de palabras mínimo y máximo en el validador de áreas.

**Criaturas:**
- 1 **Shadow Wraith**

**Tesoro:**
- **XP:** 100 XP

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
- 1 **Shadow Wraith**

**Tesoro:**
- **XP:** 100 XP

**Conexiones:**
- ← Área 1

**Secretos y Trampas:**
- **Detectar:** Percepción DC 12

**Desarrollo:**
- Si entran en combate: atacan.
`,
			bestiary: `## Goblin

*Pequeño humanoide, NE*
`,
			npcs:     "",
			wantErr:  true,
			checks:   []string{"creature_reference"},
		},
		{
			name: "xp budget imbalance",
			act: `# Acto 1

### Área 1: Test

> **Read-Aloud:** *Texto largo y descriptivo para alcanzar el mínimo de palabras requerido en esta área de prueba.*

**Descripción:** Descripción detallada con muchas palabras para cumplir el requisito de conteo de palabras mínimo y máximo en el validador de áreas.

**Criaturas:**
- 50 **Goblin**

**Tesoro:**
- **XP:** 5000 XP

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
- 50 **Goblin**

**Tesoro:**
- **XP:** 5000 XP

**Conexiones:**
- ← Área 1

**Secretos y Trampas:**
- **Detectar:** Percepción DC 12

**Desarrollo:**
- Si entran en combate: atacan.
`,
			bestiary: `## Goblin

*Pequeño humanoide, NE*
**Desafío:** 1/4 (50 PX)
`,
			npcs:     "",
			wantErr:  true,
			checks:   []string{"xp_budget"},
		},
		{
			name: "npc reference missing",
			act: `# Acto 1

### Área 1: Test

> **Read-Aloud:** *Texto largo y descriptivo para alcanzar el mínimo de palabras requerido en esta área de prueba.*

**Descripción:** Descripción detallada con muchas palabras para cumplir el requisito de conteo de palabras mínimo y máximo en el validador de áreas.

**NPCs:**
- **Noska Ur'gray** — Mercader

**Conexiones:**
- → Área 2

**Secretos y Trampas:**
- **Detectar:** Percepción DC 12

**Desarrollo:**
- Si hablan: vende información.

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
			bestiary: `## Goblin

*Pequeño humanoide, NE*
**Desafío:** 1/4 (50 PX)
`,
			npcs: `## Eldrin

- **Raza/Clase:** Elfo Mago
`,
			wantErr: true,
			checks:  []string{"npc_reference"},
		},
		{
			name: "treasure consistency check",
			act: `# Acto 1

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

**Tesoro:**
- **XP:** 50 XP

**Conexiones:**
- ← Área 1

**Secretos y Trampas:**
- **Detectar:** Percepción DC 12

**Desarrollo:**
- Si entran en combate: atacan.
`,
			bestiary: `## Goblin

*Pequeño humanoide, NE*
**Desafío:** 1/4 (50 PX)
`,
			npcs:     "",
			wantErr:  true,
			checks:   []string{"treasure"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateIntegration(tt.act, tt.bestiary, tt.npcs)
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

func TestCalculateXPBudget(t *testing.T) {
	act := `# Acto 1

### Área 1

**Criaturas:**
- 2 **Goblin**

**Tesoro:**
- **XP:** 100 XP

### Área 2

**Criaturas:**
- 1 **Orc**

**Tesoro:**
- **XP:** 100 XP
`
	bestiary := `## Goblin
**Desafío:** 1/4 (50 PX)
## Orc
**Desafío:** 1/2 (100 PX)
`

	xp := CalculateTotalXP(act, bestiary)
	// 2 goblins × 50 + 1 orc × 100 = 200
	if xp != 200 {
		t.Errorf("expected XP 200, got %d", xp)
	}

	xpPerPC := xp / 5 // default party size
	if xpPerPC != 40 {
		t.Errorf("expected XP per PC 40, got %d", xpPerPC)
	}
}
