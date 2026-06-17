package validators

import (
	"testing"
)

func TestValidateNPCFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		checks  []string
	}{
		{
			name: "valid npc with all v2 fields (Spanish)",
			input: `# NPCs

## Test NPC

- **Raza/Clase:** Humano Guerrero
- **Alineamiento:** LG
- **Ubicación:** Área 3
- **Estadísticas de Combate:** CA 12, PG 18 (3d8+3), Espada corta +4 (1d6+2)
- **Rol en la historia:** Aliado
- **Personalidad:** Valiente
- **Motivación:** Proteger la aldea
- **Secreto:** Es un espía
- **Involucramiento en Quests:** Quest: "La Llave Perdida" — informante
- **Conexiones:** Amigo de Eldrin
- **Cita típica:** "Nunca retrocedo"
`,
			wantErr: false,
			checks:  []string{"alignment", "location", "combat_stats", "quest_involvement", "secret"},
		},
		{
			name: "valid npc with all v2 fields (English)",
			input: `# NPCs

## Test NPC

- **Race/Class:** Human Fighter
- **Alignment:** LG
- **Location:** Area 3
- **Combat Stats:** AC 12, HP 18 (3d8+3), Short sword +4 (1d6+2)
- **Role in story:** Ally
- **Personality:** Brave
- **Motivation:** Protect the village
- **Secret:** He is a spy
- **Quest Involvement:** Quest: "The Lost Key" — informant
- **Connections:** Friend of Eldrin
- **Typical quote:** "I never retreat"
`,
			wantErr: false,
			checks:  []string{"alignment", "location", "combat_stats", "quest_involvement", "secret"},
		},
		{
			name: "missing location",
			input: `# NPCs

## Test NPC

- **Raza/Clase:** Humano Guerrero
- **Alineamiento:** LG
- **Estadísticas de Combate:** CA 12, PG 18
- **Secreto:** Es un espía
`,
			wantErr: true,
			checks:  []string{"location"},
		},
		{
			name: "missing combat stats",
			input: `# NPCs

## Test NPC

- **Raza/Clase:** Humano Guerrero
- **Alineamiento:** LG
- **Ubicación:** Área 3
- **Secreto:** Es un espía
`,
			wantErr: true,
			checks:  []string{"combat_stats"},
		},
		{
			name: "missing quest involvement",
			input: `# NPCs

## Test NPC

- **Raza/Clase:** Humano Guerrero
- **Alineamiento:** LG
- **Ubicación:** Área 3
- **Estadísticas de Combate:** CA 12
- **Secreto:** Es un espía
`,
			wantErr: true,
			checks:  []string{"quest_involvement"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateNPCFormat(tt.input)
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

func TestValidateBestiaryFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		checks  []string
	}{
		{
			name: "valid creature with all v2 fields (Spanish)",
			input: `# Bestiario

## Test Monster

*Mediano no-muerto, NE*

- **Rol de combate:** skirmisher
- **Grupos de encuentro:** 2-3 con 1 líder
- **Fuente/Referencia:** Custom

**Clase de Armadura:** 12
**Puntos de Golpe:** 22 (5d8)
**Velocidad:** 30 pies

| FUE | DES | CON | INT | SAB | CAR |
|:---:|:---:|:---:|:---:|:---:|:---:|
| 10 (+0) | 14 (+2) | 12 (+1) | 6 (-2) | 10 (+0) | 6 (-2) |

**Desafío:** 1/4 (50 PX)

### Tácticas Estructuradas

- **Apertura:** Se esconde y espera.
- **Prioridades:** Atacar al más débil.
- **Sinergia:** Recibe órdenes del líder.
- **Retirada:** Huye al 25% HP.
`,
			wantErr: false,
			checks:  []string{"role", "encounter_groups", "source", "tactics"},
		},
		{
			name: "valid creature with all v2 fields (English)",
			input: `# Bestiary

## Test Monster

*Medium undead, NE*

- **Combat Role:** skirmisher
- **Encounter Groups:** 2-3 with 1 leader
- **Source/Reference:** Custom

**Armor Class:** 12
**Hit Points:** 22 (5d8)
**Speed:** 30 ft

| STR | DEX | CON | INT | WIS | CHA |
|:---:|:---:|:---:|:---:|:---:|:---:|
| 10 (+0) | 14 (+2) | 12 (+1) | 6 (-2) | 10 (+0) | 6 (-2) |

**Challenge:** 1/4 (50 XP)

### Structured Tactics

- **Opening:** Hides and waits.
- **Priorities:** Attack the weakest.
- **Synergy:** Takes orders from the leader.
- **Retreat:** Flees at 25% HP.
`,
			wantErr: false,
			checks:  []string{"role", "encounter_groups", "source", "tactics"},
		},
		{
			name: "missing role",
			input: `# Bestiario

## Test Monster

*Mediano no-muerto, NE*

**Clase de Armadura:** 12
**Puntos de Golpe:** 22
`,
			wantErr: true,
			checks:  []string{"role"},
		},
		{
			name: "missing structured tactics",
			input: `# Bestiario

## Test Monster

*Mediano no-muerto, NE*

- **Rol de combate:** brute
- **Grupos de encuentro:** 1
- **Fuente/Referencia:** MM p.234

**Clase de Armadura:** 12
`,
			wantErr: true,
			checks:  []string{"tactics"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateBestiaryFormat(tt.input)
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

func TestValidateEncounterFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		checks  []string
	}{
		{
			name: "valid encounter with all v2 fields (Spanish)",
			input: `# Encuentros

## Encuentro 1: Test

**Dificultad:** Medio
**XP Total:** 200
**Ambientación:** Bosque oscuro

### Mapa Táctico
- **Dimensiones:** 30 × 30 pies
- **Iluminación:** Luz tenue

### Condiciones y Efectos Ambientales
- **Condiciones iniciales:** Niebla densa

### Desarrollo Round-by-Round

#### Plantilla de Encuentro
Template: Ambush

#### Rondas
| Ronda | Enemigos | Eventos Ambientales | Condición de Victoria |
|-------|----------|---------------------|----------------------|
| 1 | 2 goblins | Niebla | Sobrevivir |

### Resolución Alternativa
- **Diplomacia:** Persuasión DC 12
- **Sigilo:** Sigilo DC 14
`,
			wantErr: false,
			checks:  []string{"tactical_map", "conditions", "round_by_round", "alternative_resolution", "template"},
		},
		{
			name: "valid encounter with all v2 fields (English)",
			input: `# Encounters

## Encounter 1: Test

**Difficulty:** Medium
**Total XP:** 200
**Setting:** Dark forest

### Tactical Map
- **Dimensions:** 30 × 30 ft
- **Lighting:** Dim light

### Conditions and Environmental Effects
- **Initial conditions:** Dense fog

### Round-by-Round Development

#### Encounter Template
Template: Ambush

#### Rounds
| Round | Enemies | Environmental Events | Victory Condition |
|-------|---------|---------------------|-------------------|
| 1 | 2 goblins | Fog | Survive |

### Alternative Resolution
- **Diplomacy:** Persuasion DC 12
- **Stealth:** Stealth DC 14
`,
			wantErr: false,
			checks:  []string{"tactical_map", "conditions", "round_by_round", "alternative_resolution", "template"},
		},
		{
			name: "missing round-by-round",
			input: `# Encuentros

## Encuentro 1: Test

**Dificultad:** Medio
**XP Total:** 200
`,
			wantErr: true,
			checks:  []string{"round_by_round"},
		},
		{
			name: "missing alternative resolution",
			input: `# Encuentros

## Encuentro 1: Test

**Dificultad:** Medio
**XP Total:** 200

### Desarrollo Round-by-Round
#### Plantilla de Encuentro
Template: Defense

#### Rondas
| Ronda | Enemigos |
|-------|----------|
| 1 | 2 goblins |
`,
			wantErr: true,
			checks:  []string{"alternative_resolution"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateEncounterFormat(tt.input)
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
