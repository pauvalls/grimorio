# SDD Solutions — Problemas Comunes y Soluciones

Este archivo documenta problemas frecuentes y sus soluciones usando el workflow SDD.

---

## Índice

1. [Agentes no delegan correctamente](#agentes-no-delegan-correctamente)
2. [Listar clases, razas, backgrounds disponibles](#listar-clases-razas-backgrounds-disponibles)
3. [Validar contenido WotC](#validar-contenido-wotc)
4. [Generar NPCs con estándares WotC](#generar-npcs-con-estándares-wotc)
5. [Crear áreas con Developments](#crear-áreas-con-developments)
6. [Verificar stat blocks de NPCs](#verificar-stat-blocks-de-npcs)
7. [Generar character hooks automáticamente](#generar-character-hooks-automáticamente)
8. [Compilar PDF con imágenes](#compilar-pdf-con-imágenes)

---

## Agentes no delegan correctamente

### Problema
El grimorio-architect (u otro agent) ejecuta tareas directamente en lugar de delegar a sub-agentes especializados.

### Síntomas
- El agent usa MCP tools directamente (save_npcs, save_bestiary) en lugar de delegate
- Contenido generado no cumple estándares WotC
- El agent reporta "completado" pero faltan fases

### Solución SDD

1. Verificar configuración del agent:
```bash
cat agents/grimorio-architect.md | grep -A20 "Delegation Strategy"
cat ~/.config/opencode/opencode.json | grep -A30 '"grimorio-architect"'
```

2. Forzar delegación explícita en el prompt:
```
CRITICAL: Delegation Required

Sos el grimorio-architect. Tu trabajo es ORQUESTAR, no ejecutar.

DELEGA ESTO:
- NPCs → delegate(agent="grimorio-npc", ...)
- Areas → delegate(agent="grimorio-areas", ...)
- Quests → delegate(agent="grimorio-quests", ...)
- Bestiary → delegate(agent="grimorio-bestiary", ...)
- Encounters → delegate(agent="grimorio-encounters", ...)
- Lore → delegate(agent="grimorio-lore", ...)
- Validation → delegate(agent="grimorio-narrative-custodian", ...)

NO USAR:
- save_npcs() directamente
- save_bestiary() directamente
- save_areas() directamente

REPORTÁ PROGRESO después de cada delegación.
```

### Fix Permanente

1. Actualizar agents/grimorio-architect.md:
   - Agregar sección "CRITICAL: Delegation Strategy" al principio
   - Incluir DO's y DON'Ts explícitos
   - Mostrar patrón de delegación con ejemplos

2. Actualizar ~/.config/opencode/opencode.json:
   - Agregar instrucción de delegación en el prompt del agent
   - Actualizar description a "ORCHESTRATOR ONLY"

3. Commit y push:
```bash
git add agents/grimorio-architect.md
git commit -m "fix: Add explicit delegation strategy"
git push
```

---

## Listar Clases, Razas, Backgrounds Disponibles

### Problema
Necesitás saber qué opciones de personajes están disponibles para generar NPCs o PCs coherentes.

### Solución SDD

#### Comando: Listar Clases
```bash
# Clases en español
grep -i "class:" agents/grimorio-characters.md | sort | uniq

# Clases en inglés (PHB)
# Barbarian, Bard, Cleric, Druid, Fighter, Monk, Paladin, Ranger, Rogue, Sorcerer, Warlock, Wizard
```

#### Comando: Listar Razas
```bash
# Razas comunes en D&D 5e
# Human (Humano), Elf (Elfo), Dwarf (Enano), Halfling (Mediano), 
# Tiefling, Dragonborn, Gnome, Half-Orc (Semi-orco), Half-Elf (Semi-elfo)
```

#### Comando: Listar Backgrounds
```bash
# Backgrounds del PHB
# Acolyte, Charlatan, Criminal, Entertainer, Folk Hero, Guild Artisan, 
# Hermit, Noble, Outlander, Sage, Sailor, Soldier, Urchin
```

---

## Validar Contenido WotC

### Problema
El contenido generado no cumple con los estándares WotC (Developments, Character Hooks, Multiple Solutions, etc.)

### Solución SDD

#### Validadores Disponibles

```bash
# Checkear si los validadores están implementados
grep -n "ValidateDevelopments|ValidateMultipleSolutions|ValidateCharacterHooks|ValidateBoxedText" internal/validators/*.go
```

#### Checklist WotC

**Para Áreas:**
- [ ] Developments: 3-5 ramas con IF-THEN
- [ ] Character Hooks: 2-3 hooks por área
- [ ] Multiple Solutions: 2+ paths (stealth/social/combat)
- [ ] Boxed Text: 100-600 palabras, 2da persona, presente
- [ ] DCs: Numéricos (nunca "alto/bajo")
- [ ] Tesoro: Con XP explícito

**Para NPCs:**
- [ ] Apariencia: 3-5 párrafos (150-250 palabras)
- [ ] Personalidad: 2-3 párrafos (100-150 palabras)
- [ ] Voz: 1 párrafo (50-80 palabras)
- [ ] Secretos: 3-5 (1 trivial, 2 importantes, 1 de campaña)
- [ ] Plot Hooks: 2-3 hooks
- [ ] Diálogo: 3-5 líneas read-aloud
- [ ] Stat Block: En bestiary.md

#### Comando de Validación

```bash
# Validar área específica
grimorio_validate_canon(
  campaign_id="mi-campana",
  proposal_id="act-1-area-5",
  proposal_type="act",
  content="[contenido del área]"
)

# Check de consistencia completo
grimorio_check_consistency(
  campaign_id="mi-campana",
  scope="full"
)
```

---

## Generar NPCs con Estándares WotC

### Problema
Los NPCs generados tienen descripciones muy cortas (100-200 palabras) en lugar del estándar WotC (500-800 palabras).

### Solución SDD

#### Prompt para grimorio-npc

Generá NPCs para la campaña "{campaign}" siguiendo estándares WotC:

**REQUISITOS POR NPC PRINCIPAL:**

1. Apariencia Física (3-5 párrafos, 150-250 palabras)
   - Altura y complexión
   - Rostro (ojos, nariz, boca, expresión)
   - Cabello (color, estilo)
   - Vestimenta (ropa, accesorios, símbolos)
   - Características distintivas (cicatrices, tatuajes)

2. Personalidad (2-3 párrafos, 100-150 palabras)
   - Mannerisms (gestos, tics, hábitos)
   - Speech patterns (cómo habla, vocabulario)
   - Motivations (qué lo impulsa, metas, miedos)

3. Voz (1 párrafo, 50-80 palabras)
   - Tono (grave, agudo, ronco)
   - Accent (regional, social)
   - Catchphrases (frases típicas)

4. Secretos (3-5 secretos)
   - 1 trivial (flavor)
   - 2 importantes (quest-relevant)
   - 1 de campaña (plot-altering)

5. Plot Hooks (2-3 hooks)
   - Por qué interactúa con los PJs
   - Cómo puede ayudar/obstaculizar
   - Conexión con trama principal

6. Diálogo Read-Aloud (3-5 líneas)
   - Saludo/apertura
   - Información/reacción
   - Cierre/llamada a la acción

**VALIDACIÓN:**
- Total: 500-800 palabras para NPCs principales
- Stat block en bestiary.md
- "Ver bestiary.md: [Nombre]" en npcs.md

---

## Crear Áreas con Developments

### Problema
Las áreas generadas no tienen la sección Developments con 3-5 ramas de decisión.

### Solución SDD

#### Template de Developments

```markdown
### Developments

**Si los PJs [acción específica]:**
- **Consecuencia inmediata:** [qué pasa ahora]
- **Consecuencia futura:** [qué pasa en área X o acto N]
- **Recuperación:** [cómo continuar si falla]

**Si los PJs [otra acción]:**
- **Consecuencia inmediata:** [...]
- **Consecuencia futura:** [...]
- **Recuperación:** [...]

**Si los PJs [tercera acción]:**
- **Consecuencia inmediata:** [...]
- **Consecuencia futura:** [...]
- **Recuperación:** [...]
```

---

## Verificar Stat Blocks de NPCs

### Problema
NPCs mencionados en npcs.md no tienen stat block en bestiary.md.

### Solución SDD

#### Comando de Verificación

```bash
# Extraer nombres de NPCs de npcs.md
grep "^### " npcs/npcs_and_factions.md | sed 's/### //'

# Extraer nombres de bestiary.md
grep "^### " bestiary/bestiary.md | sed 's/### //'

# Verificar cruces (debería mostrar NPCs sin stat block)
comm -23 <(grep "^### " npcs/npcs_and_factions.md | sort) \
         <(grep "^### " bestiary/bestiary.md | sort)
```

---

## Generar Character Hooks Automáticamente

### Problema
Las áreas no tienen character hooks personalizados para cada PJ.

### Solución SDD

#### MCP Tool

```bash
# Generar hooks para todos los PCs
grimorio_generate_character_hooks(
  campaign="mi-campana"
)
```

---

## Compilar PDF con Imágenes

### Problema
El PDF se compila sin imágenes o con referencias rotas.

### Solución SDD

#### Checklist Pre-Compilación

- [ ] Todas las imágenes en assets/
- [ ] Referencias en markdown: ![alt](assets/filename.png)
- [ ] Cover art: ![Cover](assets/cover-*.png) en README.md
- [ ] NPCs: ![NPC Name](assets/npc-*.png) en npcs.md
- [ ] Scenes: ![Scene](assets/scene-*.png) en acts/*.md
- [ ] Monsters: ![Monster](assets/monster-*.png) en bestiary.md

#### Comando

```bash
# Compilar PDF
grimorio_compile_pdf(
  campaign="mi-campana",
  title="Mi Campaña"
)

# Verificar imágenes generadas
ls -la {campaign_path}/assets/*.png

# Verificar referencias en markdown
grep -r "assets/.*\.png" {campaign_path}/*.md {campaign_path}/**/*.md
```

---

## Comandos Útiles

### Listar Contenido de Campaña

```bash
# Contar palabras por archivo
wc -w campaigns/{campaign}/**/*.md

# Contar áreas por acto
grep -c "^### Área" campaigns/{campaign}/acts/*.md

# Listar NPCs
grep "^### " campaigns/{campaign}/npcs/npcs_and_factions.md

# Listar quests
ls -1 campaigns/{campaign}/quests/*.md
```

### Validar Estructura

```bash
# Checkear bidireccionalidad de conexiones
grep "→ Área" campaigns/{campaign}/acts/*.md | sort

# Checkear DCs numéricos
grep -i "DC [0-9]" campaigns/{campaign}/acts/*.md | head -10

# Checkear XP en tesoros
grep -i "XP" campaigns/{campaign}/acts/*.md | head -10
```

### Debug de Delegación

```bash
# Ver logs de delegación
delegation_list

# Leer resultado de delegación
delegation_read({delegation_id})

# Checkear si agent está usando delegate
grep "delegate(" agents/grimorio-architect.md | wc -l
```

---

## Word Count Comparison

### WotC Adventures (Referencia)

| Aventura | Palabras | Tipo |
|----------|----------|------|
| Lost Mine of Phandelver | 37,408 | One-shot (1-5) |
| Waterdeep: Dragon Heist | 138,910 | Campaña (1-10) |
| Curse of Strahd | 152,336 | Campaña (1-10) |

**Promedio Campaña**: ~140,000 palabras  
**Promedio One-shot**: ~40,000 palabras

### GRIMORIO Output Esperado

- **Campaña 3 actos**: 80,000-120,000 palabras
- **One-shot**: 25,000-40,000 palabras
- **NPC principal**: 500-800 palabras
- **Área individual**: 150-200 palabras
- **Boxed text**: 100-600 palabras

---

## Referencias

- **WotC Standards**: openspec/changes/grimorio-100-wotc-quality/archive-report.md
- **Agent Instructions**: agents/grimorio-*.md
- **Validators**: internal/validators/*.go
- **MCP Tools**: internal/mcp/handlers/*.go

---

**Última actualización**: 2026-05-08  
**Versión**: 1.0
