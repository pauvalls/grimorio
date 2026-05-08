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

---

## 🔧 Agentes No Usan Templates

### Problema
Los agentes (grimorio-areas, grimorio-npc, etc.) tienen acceso a `get_template` pero NO LO USAN antes de generar contenido.

### Síntomas
- Áreas sin formato WotC (falta Developments, Character Hooks, Boxed Text)
- NPCs sin estructura completa (falta Apariencia, Personalidad, Voz, Secretos)
- Quests sin objectives/rewards completos
- Contenido generado no sigue el template oficial

### Solución SDD

#### 1. Actualizar Agentes

Agregar instrucción CRÍTICA al principio de cada agent file:

```markdown
---

## CRITICAL: READ TEMPLATE FIRST

**BEFORE generating ANY content, you MUST:**

1. **Read the template** using `get_template` MCP tool:
   ```
   get_template(type="{template_type}")
   ```

2. **Study the template structure** - note all required sections

3. **Follow the template EXACTLY** - do not skip any sections

4. **Fill in all required fields** - use your specialized knowledge

**Template Types by Agent:**
- grimorio-areas → `get_template(type="areas")`
- grimorio-npc → `get_template(type="npc")`
- grimorio-bestiary → `get_template(type="monster")`
- grimorio-encounters → `get_template(type="encounter")`
- grimorio-maps → `get_template(type="map")`
- grimorio-lore → `get_template(type="lore")`

**DO NOT generate content without reading the template first.**

---
```

#### 2. Verificar en el Prompt de Delegación

Cuando delegués a un sub-agente, incluí la instrucción explícita:

```markdown
delegate(agent="grimorio-areas", prompt="Generate AREAS for campaign '{campaign}'.

**CRITICAL:** 
1. FIRST call get_template(type='areas') to read the WotC template
2. Study the template structure (Developments, Character Hooks, Boxed Text, etc.)
3. Generate areas FOLLOWING the template EXACTLY
4. Do NOT skip any sections

Read these files first:
- lore.md
- npcs/npcs_and_factions.md
- bestiary/bestiary.md
- quests/*.md

Then generate {act_count} acts with 10-15 numbered areas each.")
```

#### 3. Validar Post-Generación

```bash
# Checkear si las áreas tienen formato WotC
grep -c "### Developments" campaigns/{campaign}/areas/*.md
grep -c "### Character Hooks" campaigns/{campaign}/areas/*.md
grep -c ">> \*\*Texto para Leer\*\*" campaigns/{campaign}/areas/*.md

# Si es 0, el agente NO usó el template
# Si es >0, el agente SIGUIÓ el template
```

### Fix Permanente

1. **Actualizar CADA agent file** (`grimorio-areas.md`, `grimorio-npc.md`, etc.):
   - Agregar sección "CRITICAL: READ TEMPLATE FIRST" después de la línea 27
   - Incluir lista de template types por agente
   - Mostrar comando `get_template` explícito

2. **Actualizar prompts de delegación**:
   - Incluir instrucción "FIRST call get_template(...)" en cada delegate()
   - Verificar que el agente leyó el template antes de generar

3. **Validar post-generación**:
   - Usar grep para checkear secciones WotC
   - Si faltan, regenerar con instrucción explícita

### Comandos de Verificación

```bash
# Verificar que los agentes tienen la instrucción
grep -l "READ TEMPLATE FIRST" agents/grimorio-*.md | wc -l

# Verificar contenido generado
grep -c "### Developments" campaigns/{campaign}/areas/*.md
# Debería ser > 0 para cada área

grep -c "### Character Hooks" campaigns/{campaign}/areas/*.md
# Debería ser > 0 para cada área

grep -c ">>" campaigns/{campaign}/areas/*.md
# Debería ser > 0 (boxed text)
```

---

---

## 🛡️ Narrative Custodian No Revisa Formato WotC

### Problema
El `grimorio-narrative-custodian` solo valida **consistencia de canon** (entidades, timeline, mcguffins) pero **NO valida el formato WotC del contenido** (áreas, NPCs, quests).

### Síntomas
- `grimorio_check_consistency` reporta "fair" o "approved" pero las áreas no tienen Developments
- NPCs sin formato WotC (menos de 500 palabras, sin 6 secciones)
- Quests sin objectives/rewards completos
- El check automático pasa pero el contenido NO cumple estándares WotC

### Ejemplo Real

```bash
# Check de consistencia (solo valida canon)
grimorio_check_consistency(
  campaign_id="la-hoja-de-vlad",
  scope="full"
)

# Resultado:
# ✅ mcguffin_continuity: PASSED
# ✅ entity_not_found: PASSED
# ✅ npc_alive_check: PASSED
# ❌ Pero las áreas NO tienen Developments, Character Hooks, Boxed Text
```

### Causa Raíz

El `ValidationEngine` en `internal/services/validation_engine.go` tiene DOS tipos de validación:

1. **Canon Consistency** (automática) ✅
   - `entity_not_found`
   - `npc_alive_check`
   - `mcguffin_continuity`
   - `location_existence`
   - `timeline_consistency`

2. **WotC Format** (automática SOLO si proposal.Type es correcto) ⚠️
   - `wotc_developments`
   - `wotc_character_hooks`
   - `wotc_multiple_solutions`
   - `wotc_boxed_text`
   - `wotc_npc_word_count`

**Problema**: La validación WotC solo se ejecuta si el `proposal.Type` es `"act"`, `"npc"`, o `"quest"`. Si el custodian no recibe proposals con esos tipos, NO valida el formato.

### Solución SDD

#### 1. Validación Manual Post-Generación

Después de cada batch, el architect debe hacer **DOS checks**:

```bash
# Check 1: Canon Consistency (automático)
grimorio_check_consistency(
  campaign_id="{campaign}",
  scope="full"
)

# Check 2: WotC Format (manual con grep)
# Áreas
grep -c "### Developments" campaigns/{campaign}/areas/*.md
# Debería ser >= 3 por área (3-5 ramas)

grep -c "### Character Hooks" campaigns/{campaign}/areas/*.md
# Debería ser >= 2 por área (2-3 hooks)

grep -c ">> \*\*Texto para Leer\*\*" campaigns/{campaign}/areas/*.md
# Debería ser >= 1 por área (boxed text)

# NPCs
wc -w campaigns/{campaign}/npcs/*.md
# Debería ser 500-800 palabras por NPC principal

grep -c "### Secretos" campaigns/{campaign}/npcs/*.md
# Debería ser >= 3 por NPC (3-5 secretos)

# Quests
jq '.objectives | length' campaigns/{campaign}/quests/*.json
# Debería ser 2-4 objectives por quest

jq '.reward.type' campaigns/{campaign}/quests/*.json
# Debería tener reward.type definido
```

#### 2. Agregar Validación WotC al Custodian

Actualizar `grimorio-narrative-custodian.md` para incluir check de formato:

```markdown
## CRITICAL: Validate WotC Format (NOT Just Canon)

When validating batches, you MUST check BOTH:

### 1. Canon Consistency (existing checks)
- entity_not_found
- npc_alive_check
- mcguffin_continuity
- location_existence
- timeline_consistency

### 2. WotC Format (NEW - REQUIRED)

**For Areas (acts/*.md):**
- ✅ Developments: 3-5 ramas con IF-THEN
- ✅ Character Hooks: 2-3 hooks por área
- ✅ Multiple Solutions: 2+ paths (stealth/social/combat)
- ✅ Boxed Text: 100-600 palabras, 2da persona, presente
- ✅ Numeric DCs: Never "alto/bajo"

**For NPCs (npcs/*.md):**
- ✅ Word Count: 500-800 palabras (major), 200-400 (minor)
- ✅ 6 Sections: Appearance, Personality, Voice, Secrets, Hooks, Dialogue
- ✅ Stat Block Link: "Ver bestiary.md: [Name]"

**For Quests (quests/*.json):**
- ✅ Objectives: 2-4 objectives
- ✅ Rewards: type + description + value
- ✅ XP: Included in reward

### Validation Command

```bash
# Run BOTH checks
grimorio_check_consistency(campaign_id="{campaign}", scope="full")

# Then manually verify WotC format
grep -c "### Developments" campaigns/{campaign}/areas/*.md
grep -c "### Character Hooks" campaigns/{campaign}/areas/*.md
grep -c ">>" campaigns/{campaign}/areas/*.md
wc -w campaigns/{campaign}/npcs/*.md
```

### Reject Criteria

**REJECT batch if:**
- Canon consistency has errors/criticals
- OR WotC format validation fails (less than 80% compliance)

**Example:**
```
Batch 3 (Areas) Validation:
- Canon Consistency: ✅ PASSED
- WotC Format: ❌ FAILED
  - Developments: 0/12 areas (0%)
  - Character Hooks: 0/12 areas (0%)
  - Boxed Text: 0/12 areas (0%)
  
Decision: REJECT - Regenerate areas with WotC template
```
```

#### 3. Actualizar Validation Engine (Opcional - Code Fix)

Agregar validación WotC explícita en `validation_engine.go`:

```go
// After canon consistency checks, ALWAYS validate WotC format
func (e *ValidationEngine) CheckConsistency(ctx context.Context, campaignID string, scope domain.ConsistencyScope) (*domain.ConsistencyReport, error) {
    // ... existing canon checks ...
    
    // NEW: Load and validate all content files
    e.validateAllAreas(report, campaignID)
    e.validateAllNPCs(report, campaignID)
    e.validateAllQuests(report, campaignID)
    
    return report, nil
}

func (e *ValidationEngine) validateAllAreas(report *domain.ConsistencyReport, campaignID string) {
    // Load all area files
    // For each area, call validateWotCFormat()
    // Add failures to report.Issues
}
```

### Comandos de Verificación Completa

```bash
# Full validation script
cat > /tmp/validate_campaign.sh << 'SCRIPT'
#!/bin/bash
CAMPAIGN=$1

echo "=== Validating Campaign: $CAMPAIGN ==="
echo ""

echo "1. Canon Consistency..."
grimorio_check_consistency(campaign_id="$CAMPAIGN", scope="full")
echo ""

echo "2. WotC Format - Areas..."
echo "   Developments: $(grep -c '### Developments' campaigns/$CAMPAIGN/areas/*.md 2>/dev/null | awk -F: '{sum+=$2} END {print sum}')"
echo "   Character Hooks: $(grep -c '### Character Hooks' campaigns/$CAMPAIGN/areas/*.md 2>/dev/null | awk -F: '{sum+=$2} END {print sum}')"
echo "   Boxed Text: $(grep -c '>>' campaigns/$CAMPAIGN/areas/*.md 2>/dev/null | awk -F: '{sum+=$2} END {print sum}')"
echo ""

echo "3. WotC Format - NPCs..."
echo "   Total NPCs: $(ls -1 campaigns/$CAMPAIGN/npcs/*.md 2>/dev/null | wc -l)"
echo "   Avg Word Count: $(wc -w campaigns/$CAMPAIGN/npcs/*.md 2>/dev/null | tail -1 | awk '{print $1/NR}')"
echo ""

echo "4. WotC Format - Quests..."
echo "   Total Quests: $(ls -1 campaigns/$CAMPAIGN/quests/*.json 2>/dev/null | wc -l)"
echo "   Quests with Objectives: $(grep -l '"objectives"' campaigns/$CAMPAIGN/quests/*.json 2>/dev/null | wc -l)"
echo ""

echo "=== Validation Complete ==="
SCRIPT

chmod +x /tmp/validate_campaign.sh
/tmp/validate_campaign.sh la-hoja-de-vlad
```

### Checklist de Validación Completa

**El Narrative Custodian DEBE validar:**

| Tipo | Check | Herramienta | Threshold |
|------|-------|-------------|-----------|
| **Canon** | entity_not_found | `grimorio_check_consistency` | 0 errors |
| **Canon** | npc_alive_check | `grimorio_check_consistency` | 0 errors |
| **Canon** | mcguffin_continuity | `grimorio_check_consistency` | 0 errors |
| **WotC** | Developments | `grep -c "### Developments"` | ≥3 por área |
| **WotC** | Character Hooks | `grep -c "### Character Hooks"` | ≥2 por área |
| **WotC** | Boxed Text | `grep -c ">>"` | ≥1 por área |
| **WotC** | NPC Word Count | `wc -w` | 500-800 (major) |
| **WotC** | NPC Stat Links | `grep -c "Ver bestiary"` | 1 por NPC |
| **WotC** | Quest Objectives | `jq '.objectives \| length'` | 2-4 por quest |
| **WotC** | Quest Rewards | `jq '.reward.type'` | Defined |

---

## 📝 Resumen: Validación en Dos Capas

```
Capa 1: Canon Consistency (automática)
  ↓
grimorio_check_consistency(campaign_id, scope="full")
  - entity_not_found
  - npc_alive_check
  - mcguffin_continuity
  - location_existence
  - timeline_consistency

Capa 2: WotC Format (manual o automática con fix)
  ↓
grep commands + validation_engine.go fix
  - Developments (3-5 ramas)
  - Character Hooks (2-3 hooks)
  - Boxed Text (100-600 words)
  - NPC Word Count (500-800)
  - Quest Completeness (objectives, rewards)

AMBAS capas deben pasar para considerar la campaña "APPROVED".
```

---

**TL;DR**: El narrative custodian actual SOLO valida canon (entidades, timeline). Necesita validación MANUAL o CODE FIX para validar formato WotC (Developments, Character Hooks, Boxed Text, etc.). Agregar commands de verificación grep al workflow del architect.


---

## 📁 Paths Incorrectos: /home/pau/Grimorio/campaigns/ vs /home/pau/campaigns/

### Problema
Los agentes grimorio están guardando contenido en `/home/pau/Grimorio/campaigns/` cuando deberían usar `/home/pau/campaigns/`.

### Síntomas
- `save_areas` guarda en `/home/pau/Grimorio/campaigns/{campaign}/areas/`
- `save_npcs` guarda en `/home/pau/Grimorio/campaigns/{campaign}/npcs/`
- `save_bestiary` guarda en `/home/pau/Grimorio/campaigns/{campaign}/bestiary/`
- **Debería guardar en:** `/home/pau/campaigns/{campaign}/`

### Causa Raíz

**Working Directory Incorrecto:**
- El working directory del proyecto es `/home/pau/Grimorio/`
- Los agents usan paths relativos como `campaigns/{campaign}/areas/`
- MCP tools resuelven paths relativos desde el working directory
- Resultado: `/home/pau/Grimorio/campaigns/` en vez de `/home/pau/campaigns/`

**Ejemplo:**
```bash
# Working directory actual
pwd
# /home/pau/Grimorio/

# Agente ejecuta:
save_areas(campaign="la-hoja-de-vlad", content="...", chapter_number="1")

# MCP tool resuelve path relativo:
# /home/pau/Grimorio/campaigns/la-hoja-de-vlad/areas/chapter_01.md ❌

# Path correcto debería ser:
# /home/pau/campaigns/la-hoja-de-vlad/areas/chapter_01.md ✅
```

### Solución SDD

#### 1. Usar Paths Absolutos en Agentes (RECOMENDADO)

Actualizar TODOS los agents para que usen paths absolutos:

**Antes:**
```markdown
**PRIMERO** leé `{campaign_path}/canon.json`
```

**Después:**
```markdown
**PRIMERO** leé `/home/pau/campaigns/{campaign_name}/canon.json`
```

**Ejemplo completo para grimorio-areas.md:**
```markdown
## CRITICAL: Use Correct Campaign Path

**ALWAYS use absolute paths:**
- ✅ `/home/pau/campaigns/{campaign_name}/areas/`
- ❌ `campaigns/{campaign_name}/areas/` (resuelve a /home/pau/Grimorio/campaigns/)

**When calling MCP tools:**
```python
save_areas(
  campaign="la-hoja-de-vlad",
  content=content,
  chapter_number=1
)
# MCP tool internally resolves to:
# /home/pau/campaigns/la-hoja-de-vlad/areas/chapter_01.md ✅
```
```

#### 2. Configurar Variable de Entorno (ALTERNATIVA)

Agregar `.env` o `.opencode` en el root del proyecto:

```bash
# /home/pau/Grimorio/.env
GRIMORIO_CAMPAIGNS_ROOT=/home/pau/campaigns
```

Y actualizar los MCP tools para leer esta variable:

```go
// internal/mcp/grimorio.go
func get_campaigns_root() string {
    if env := os.Getenv("GRIMORIO_CAMPAIGNS_ROOT"); env != "" {
        return env
    }
    return "/home/pau/campaigns" // default
}
```

#### 3. Validación Pre-Save (DEFENSIVO)

Agregar validación en cada MCP tool antes de guardar:

```go
func (s *GrimorioService) SaveAreas(ctx context.Context, campaign string, content string, chapter int) error {
    // Validate campaign path exists
    expectedPath := fmt.Sprintf("/home/pau/campaigns/%s", campaign)
    if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
        return fmt.Errorf("campaign directory not found at %s - did you mean /home/pau/campaigns/ instead of /home/pau/Grimorio/campaigns/?", expectedPath)
    }
    
    // Proceed with save...
}
```

### Comandos de Verificación

```bash
# Verificar dónde están las campañas
ls -la /home/pau/campaigns/
ls -la /home/pau/Grimorio/campaigns/

# Si están en el lugar equivocado, mover:
mv /home/pau/Grimorio/campaigns/la-hoja-de-vlad /home/pau/campaigns/

# Verificar estructura correcta
tree /home/pau/campaigns/la-hoja-de-vlad/
# Debería mostrar:
# ├── areas/
# ├── npcs/
# ├── bestiary/
# ├── encounters/
# ├── characters/
# ├── quests/
# ├── lore.md
# └── canon.json
```

### Checklist de Paths Correctos

**Antes de ejecutar cualquier agente grimorio:**

| Check | Comando | Resultado Esperado |
|-------|---------|-------------------|
| Campaign existe | `ls /home/pau/campaigns/{campaign}/` | ✅ Directory exists |
| NO en Grimorio | `ls /home/pau/Grimorio/campaigns/{campaign}/` | ❌ No such file |
| Areas path | `ls /home/pau/campaigns/{campaign}/areas/` | ✅ Chapter files |
| NPCs path | `ls /home/pau/campaigns/{campaign}/npcs/` | ✅ npc_and_factions.md |
| Bestiary path | `ls /home/pau/campaigns/{campaign}/bestiary/` | ✅ bestiary.md |

### Migración de Datos (Si ya se generó en lugar equivocado)

```bash
# 1. Verificar qué hay en el lugar incorrecto
ls -la /home/pau/Grimorio/campaigns/

# 2. Mover al lugar correcto
mv /home/pau/Grimorio/campaigns/la-hoja-de-vlad /home/pau/campaigns/

# 3. Verificar estructura
tree /home/pau/campaigns/la-hoja-de-vlad/

# 4. Eliminar directorio vacío
rmdir /home/pau/Grimorio/campaigns/
```

### Prevención Futura

**En el prompt de cada sub-agente, agregar:**

```markdown
## CRITICAL: Campaign Path

**ALWAYS use:** `/home/pau/campaigns/{campaign_name}/`
**NEVER use:** `campaigns/{campaign_name}/` (relative path)

This is because the working directory is `/home/pau/Grimorio/` and relative paths will resolve incorrectly.
```

---

**TL;DR:** Working directory es `/home/pau/Grimorio/` pero las campañas viven en `/home/pau/campaigns/`. Usar SIEMPRE paths absolutos en agents y MCP tools. Si se generó en el lugar equivocado, mover manualmente y validar estructura.


---

## 📖 Contexto Narrativo: Pedir Historia Antes de Generar

### Problema
Los agentes grimorio generan contenido SIN pedir primero una descripción de la historia/trama que el DM quiere. Resultado: contenido genérico, desconectado de la visión del DM.

### Síntomas
- Áreas sin conexión con el arco narrativo principal
- NPCs sin relación con la trama específica
- Encuentros que no avanzan la historia que el DM quería contar
- El DM tiene que regenerar múltiples veces hasta que "encaja"
- **Falta contexto narrativo:** ¿qué historia quiere contar el DM?

### Ejemplo Real

**Sin contexto (INCORRECTO):**
```
User: "Generá áreas para la-hoja-de-vlad"
Agente: → Genera 12 áreas genéricas de vampiros
Result: Áreas técnicamente correctas pero sin conexión con la historia específica
```

**Con contexto (CORRECTO):**
```
User: "Generá áreas para la-hoja-de-vlad"
Agente: "¿Podés describir la historia/trama que querés contar? 
         - ¿Qué arco narrativo tiene Act 1, 2, 3?
         - ¿Qué eventos clave deben ocurrir?
         - ¿Qué tono buscás (terror político, acción, misterio)?
         - ¿Hay escenas específicas que querés incluir?"
User: "Act 1: Investigación en la corte, Act 2: Viaje a las montañas, Act 3: Confrontación final"
Agente: → Genera 12 áreas alineadas con esa estructura
Result: Áreas que cuentan LA historia que el DM quería
```

### Causa Raíz

**Los agentes asumen que con `canon.json` y `lore.md` es suficiente:**
- `canon.json` → Entidades y reglas del mundo
- `lore.md` → Historia del mundo, no la trama específica

**Falta:**
- `story_brief.md` o `trama.md` → La historia específica que el DM quiere contar
- Arco narrativo por actos
- Escenas clave que DEBEN ocurrir
- Tono y ritmo deseado

### Solución SDD

#### 1. Agregar Paso de "Story Brief" al Workflow

**Antes de generar CUALQUIER contenido (áreas, NPCs, encuentros):**

```markdown
## CRITICAL: Request Story Brief First

**BEFORE generating content, you MUST ask:**

> "Para generar contenido alineado con tu visión, necesito que me describas:
> 
> 1. **Arco narrativo por actos:**
>    - Act 1: ¿Qué establece la historia? (3-4 áreas)
>    - Act 2: ¿Qué complica la trama? (4-5 áreas)
>    - Act 3: ¿Cómo resuelve? (3-4 áreas)
> 
> 2. **Escenas clave que DEBEN ocurrir:**
>    - ¿Hay encuentros específicos que querés incluir?
>    - ¿Revelaciones importantes?
>    - ¿Momentos emocionales para los PCs?
> 
> 3. **Tono y ritmo:**
>    - ¿Terror político? ¿Acción constante? ¿Misterio?
>    - ¿Ritmo: lento (exploración) o rápido (combate)?
> 
> 4. **Personajes importantes:**
>    - ¿NPCs que deben aparecer sí o sí?
>    - ¿Villanos específicos?
>    - ¿Aliados clave?
> 
> 5. **Tema central:**
>    - ¿De qué trata LA HISTORIA (no el mundo)?
>    - Ej: 'Redención', 'Venganza', 'Sacrificio', 'Traición'"

**Wait for user response BEFORE proceeding.**
```

#### 2. Crear Archivo `story_brief.md` o `trama.md`

**Estructura del archivo:**

```markdown
# Story Brief: {Campaign Name}

## Arco Narrativo

### Act 1: [Nombre]
- **Objetivo:** Establecer conflicto principal
- **Áreas:** 3-4
- **Eventos clave:**
  1. Los PCs reciben la misión en la corte
  2. Descubren la primera pista (carta cifrada)
  3. Encuentro con el villano (sin saber que lo es)
- **Tono:** Misterio político, tensión social
- **Revelación:** Alguien en la corte está traicionando

### Act 2: [Nombre]
- **Objetivo:** Complicar la trama, subir apuestas
- **Áreas:** 4-5
- **Eventos clave:**
  1. Viaje a las montañas (encuentro aleatorio)
  2. Descubren el ritual en la cueva
  3. Traición de aliado
  4. Persecución
- **Tono:** Acción, peligro creciente
- **Revelación:** El villano quiere revivir al antiguo señor vampiro

### Act 3: [Nombre]
- **Objetivo:** Confrontación final
- **Áreas:** 3-4
- **Eventos clave:**
  1. Infiltración en la fortaleza
  2. Enfrentamiento con el villano
  3. Decisión moral (sacrificio o victoria pírrica)
- **Tono:** Épico, emocional
- **Resolución:** Los PCs deciden el destino del reino

## Personajes Clave

### NPCs que DEBEN aparecer:
- **Lord Volkov:** Villano principal, aparece en Act 1 y 3
- **Elena Corvus:** Aliada, traiciona en Act 2
- **El Cuervo:** Informante, aparece en Act 1 y 2

### Momentos para PCs:
- **PC 1 (Paladín):** Enfrentar su juramento vs. bien mayor
- **PC 2 (Mago):** Descubrir secreto de su linaje
- **PC 3 (Pícaro):** Redimirse por traición pasada

## Temas Centrales
- **Principal:** Traición y confianza
- **Secundario:** Sacrificio por el bien común
- **Terciario:** Poder corrompe

## Escenas Obligatorias
1. Baile en la corte donde alguien es envenenado (Act 1)
2. Cruce de puente colgante bajo ataque (Act 2)
3. Decisión: salvar al aliado o detener el ritual (Act 3)
```

#### 3. Actualizar Agents para Leer `story_brief.md`

**Para grimorio-areas.md:**

```markdown
## CRITICAL: Read Story Brief First

**BEFORE generating areas:**

1. **Check if story brief exists:**
   ```bash
   ls /home/pau/campaigns/{campaign}/story_brief.md
   ```

2. **If EXISTS:** Read it first
   ```markdown
   Read file: /home/pau/campaigns/{campaign}/story_brief.md
   ```
   
3. **If NOT EXISTS:** Ask user to provide it
   > "No encontré `story_brief.md`. Para generar áreas alineadas con tu visión, necesito que describas:
   > - Arco narrativo por actos
   > - Escenas clave que deben ocurrir
   > - Tono y ritmo deseado
   > 
   > ¿Querés que te genere una plantilla de story_brief.md para completar?"

4. **Use story brief to structure areas:**
   - Act 1 areas → Establish conflict (story brief Act 1)
   - Act 2 areas → Complicate plot (story brief Act 2)
   - Act 3 areas → Resolve (story brief Act 3)
```

**Para grimorio-npc.md:**

```markdown
## CRITICAL: Read Story Brief for NPC Roles

**BEFORE generating NPCs:**

1. **Read story_brief.md** to identify:
   - NPCs que DEBEN aparecer (villanos, aliados, informantes)
   - Roles narrativos (mentor, antagonista, aliado, traidor)
   - Arcos personales vinculados a PCs

2. **Map NPCs to story beats:**
   - Lord Volkov → Aparece en Act 1 (corte) y Act 3 (confrontación)
   - Elena Corvus → Aliada en Act 1-2, traiciona en Act 2
   - El Cuervo → Informante en Act 1 y 2

3. **Generate NPCs with narrative purpose:**
   - Each NPC should advance the story
   - Not just "random tavern keeper" but "tavern keeper who witnessed the murder"
```

**Para grimorio-encounters.md:**

```markdown
## CRITICAL: Read Story Brief for Encounter Design

**BEFORE generating encounters:**

1. **Read story_brief.md** to identify:
   - Encuentros obligatorios (baile envenenado, puente, ritual)
   - Momentos emocionales para PCs
   - Revelaciones que deben ocurrir

2. **Design encounters that advance the plot:**
   - Combat → Not just "fight goblins" but "fight mercenaries protecting the evidence"
   - Social → Not just "talk to noble" but "interrogate noble who knows the traitor"
   - Exploration → Not just "explore cave" but "find the ritual site before it's too late"

3. **Place encounters according to story arc:**
   - Act 1: Lower stakes, introduce threat
   - Act 2: Rising action, complications
   - Act 3: Climax, final confrontation
```

#### 4. Workflow Actualizado

```
1. User crea campaña → canon.json (reglas del mundo)
2. User escribe lore.md (historia del mundo)
3. User escribe story_brief.md (trama específica) ← NUEVO
4. Agente lee story_brief.md → genera áreas alineadas
5. Agente lee story_brief.md → genera NPCs con roles narrativos
6. Agente lee story_brief.md → genera encuentros que avanzan la trama
```

### Comandos de Verificación

```bash
# Verificar si existe story brief
ls /home/pau/campaigns/{campaign}/story_brief.md

# Si no existe, crear plantilla
cat > /home/pau/campaigns/{campaign}/story_brief.md << 'TEMPLATE'
# Story Brief: {Campaign Name}

## Arco Narrativo

### Act 1: [Nombre]
- **Objetivo:** 
- **Áreas:** 3-4
- **Eventos clave:**
  1. 
  2. 
  3. 
- **Tono:** 

### Act 2: [Nombre]
- **Objetivo:** 
- **Áreas:** 4-5
- **Eventos clave:**
  1. 
  2. 
  3. 
  4. 
- **Tono:** 

### Act 3: [Nombre]
- **Objetivo:** 
- **Áreas:** 3-4
- **Eventos clave:**
  1. 
  2. 
  3. 
- **Tono:** 

## Personajes Clave

### NPCs que DEBEN aparecer:
- 

### Momentos para PCs:
- 

## Temas Centrales
- **Principal:** 
- **Secundario:** 
- **Terciario:** 

## Escenas Obligatorias
1. 
2. 
3. 
TEMPLATE
```

### Checklist de Story Brief

**Antes de generar contenido, verificar:**

| Check | Pregunta | Estado |
|-------|----------|--------|
| Arco por actos | ¿Están definidos Act 1, 2, 3? | ☐ |
| Eventos clave | ¿Hay 3-4 eventos por acto? | ☐ |
| NPCs obligatorios | ¿Lista de NPCs que deben aparecer? | ☐ |
| Momentos para PCs | ¿Arcos personales definidos? | ☐ |
| Temas | ¿Tema principal + secundarios? | ☐ |
| Escenas obligatorias | ¿Escenas que SÍ o SÍ deben ocurrir? | ☐ |
| Tono por acto | ¿Tono especificado (terror, acción, misterio)? | ☐ |

### Ejemplo de Prompt Inicial

**Cuando el usuario pide generar contenido SIN story brief:**

```
⚠️ **No encontré story_brief.md para esta campaña**

Para generar contenido alineado con TU visión (no contenido genérico), necesito que me describas:

1. **Arco narrativo:** ¿Qué historia querés contar en Act 1, 2, 3?
2. **Escenas clave:** ¿Hay momentos específicos que DEBEN ocurrir?
3. **Personajes:** ¿NPCs que deben aparecer sí o sí?
4. **Tono:** ¿Terror político? ¿Acción? ¿Misterio?
5. **Tema:** ¿De qué trata la historia? (traición, redención, venganza, etc.)

**Opciones:**
A) Me describís la historia ahora y yo genero el story_brief.md
B) Te genero una plantilla vacía para que completes
C) Genero contenido genérico (no recomendado)

¿Qué preferís?
```

---

**TL;DR:** Los agentes NO deben generar contenido sin antes pedir/leer `story_brief.md`. Sin contexto narrativo, el contenido es genérico y desconectado de la visión del DM. Agregar paso obligatorio: "Pedir historia → Crear story_brief.md → Leer story_brief.md → Generar contenido alineado".


---

## ⏱️ Timeout de Sub-Agentes: Grimorio-Areas Necesita 30 Minutos

### Problema
Los sub-agentes `grimorio-areas` están fallando por timeout con 15 minutos. Generar 10-15 áreas con formato WotC completo (Developments, Character Hooks, Boxed Text, Multiple Solutions) requiere más tiempo.

### Síntomas
- Error: `Sub-agent timed out after 15 minutes`
- Generación de áreas se corta a la mitad
- Áreas incompletas o sin validación de canon
- **Pérdida de trabajo:** Hay que regenerar desde cero

### Causa Raíz

**Complejidad de generación de áreas:**
1. Leer `story_brief.md` (si existe)
2. Leer `lore.md` + `canon.json`
3. Leer template WotC (`get_template(type="areas")`)
4. Generar 10-15 áreas × 150-200 palabras c/u = 1500-3000 palabras
5. Cada área incluye:
   - Boxed Text (100-600 palabras)
   - Developments (3-5 ramas IF-THEN)
   - Character Hooks (2-3 hooks)
   - Multiple Solutions (2+ paths)
   - Stats, DCs, treasure, creatures
6. Validar con `validate_canon` por batch
7. Guardar con `save_areas`

**Tiempo estimado:**
- Lectura de contexto: 1-2 min
- Generación por área: 2-3 min × 12 áreas = 24-36 min
- Validación + guardado: 2-3 min
- **Total: 27-41 minutos**

**Timeout actual: 15 min ❌ (insuficiente)**
**Timeout recomendado: 30 min ✅ (margen seguro)**

### Solución SDD

#### 1. Configurar Timeout en Orchestrator

**En el prompt del orchestrator, al delegar a grimorio-areas:**

```markdown
## Sub-Agent Timeout Configuration

**For grimorio-areas:**
- **Timeout:** 30 minutes (1800000 ms)
- **Reason:** Complex WotC format generation (10-15 areas × 150-200 words each)

**Example delegation:**
```
delegate(
  agent: "grimorio-areas",
  prompt: "...",
  timeout: 1800000  // 30 minutes
)
```

**For other agents:**
- grimorio-npc: 20 minutes (1200000 ms)
- grimorio-bestiary: 20 minutes (1200000 ms)
- grimorio-encounters: 15 minutes (900000 ms)
- grimorio-lore: 15 minutes (900000 ms)
- grimorio-maps: 15 minutes (900000 ms)
- grimorio-quests: 15 minutes (900000 ms)
```

#### 2. Actualizar opencode.json o Config de Agentes

**Agregar configuración de timeout:**

```json
{
  "agents": {
    "grimorio-areas": {
      "timeout": 1800000,
      "model": "qwen3.5-plus",
      "description": "Generate WotC-style playable areas"
    },
    "grimorio-npc": {
      "timeout": 1200000,
      "model": "qwen3.5-plus",
      "description": "Generate NPCs and factions"
    },
    "grimorio-bestiary": {
      "timeout": 1200000,
      "model": "qwen3.5-plus",
      "description": "Generate monsters and stat blocks"
    }
  }
}
```

#### 3. Dividir en Batches Más Pequeños (ALTERNATIVA)

Si el timeout es un problema persistente, dividir la generación:

**En vez de:**
```
grimorio-areas: "Generá 12 áreas para la-hoja-de-vlad"
```

**Hacer:**
```
grimorio-areas: "Generá áreas 1-4 (Act 1) para la-hoja-de-vlad"
grimorio-areas: "Generá áreas 5-8 (Act 2) para la-hoja-de-vlad"
grimorio-areas: "Generá áreas 9-12 (Act 3) para la-hoja-de-vlad"
```

**Ventajas:**
- Cada batch: 4 áreas × 2-3 min = 8-12 min (dentro del timeout de 15 min)
- Más fácil validar y corregir errores por batch
- Progreso incremental visible

**Desventajas:**
- Más delegaciones manuales
- Riesgo de inconsistencia entre batches
- Requiere coordinación de chapter numbers

#### 4. Streaming Progress (PREVENTIVO)

Agregar indicación de progreso para detectar problemas temprano:

```markdown
## Progress Reporting

**While generating areas, report progress every 3 areas:**

```
Progress: 3/12 áreas completadas (25%)
- Área 1: ✅ Completa
- Área 2: ✅ Completa
- Área 3: ✅ Completa
- Tiempo estimado restante: 18 minutos
```

**If progress stalls:**
- Stop and restart with smaller batch
- Check if validation is blocking
- Adjust timeout if needed
```

### Comandos de Verificación

```bash
# Verificar timeout actual en config
cat opencode.json | jq '.agents["grimorio-areas"].timeout'

# Si no existe, agregar:
cat > /tmp/timeout_config.json << 'JSON'
{
  "agents": {
    "grimorio-areas": { "timeout": 1800000 },
    "grimorio-npc": { "timeout": 1200000 },
    "grimorio-bestiary": { "timeout": 1200000 }
  }
}
JSON

# Merge con opencode.json existente
jq -s '.[0] * .[1]' opencode.json /tmp/timeout_config.json > opencode.json.tmp
mv opencode.json.tmp opencode.json
```

### Timeout Recommendations por Agente

| Agente | Timeout | Razón |
|--------|---------|-------|
| **grimorio-areas** | 30 min | 10-15 áreas × WotC format complejo |
| **grimorio-npc** | 20 min | 8-12 NPCs × 6 secciones c/u |
| **grimorio-bestiary** | 20 min | 6-10 monstruos × stat blocks completos |
| **grimorio-encounters** | 15 min | 5-8 encuentros × setup/combate/aftermath |
| **grimorio-lore** | 15 min | 1 documento narrativo continuo |
| **grimorio-maps** | 15 min | 6-10 ubicaciones × descripciones |
| **grimorio-quests** | 15 min | 4-6 quests × objectives/rewards |

### Señales de Timeout Inminente

**Monitorear durante ejecución:**
- ❌ Sin progreso por 5+ minutos
- ❌ Mensajes de "still working" repetidos
- ❌ Generación se detiene a mitad de batch
- ❌ Error: `context deadline exceeded`

**Acciones preventivas:**
1. Dividir batch actual en 2 partes
2. Aumentar timeout a 45 min si es crítico
3. Reducir scope (menos áreas por batch)
4. Usar modelo más rápido (tradeoff: calidad)

### Ejemplo de Delegación con Timeout

**En el orchestrator:**

```markdown
## Delegating to grimorio-areas

```bash
delegate(
  agent: "grimorio-areas",
  prompt: "
    CRITICAL: Generate Act 1 areas (areas 1-4) for la-hoja-de-vlad.
    
    Context:
    - story_brief.md: Act 1 = Investigación en la corte
    - lore.md: Vampire political intrigue
    - canon.json: Key entities and rules
    
    Requirements:
    - 4 areas, 150-200 words each
    - WotC format: Developments, Character Hooks, Boxed Text, Multiple Solutions
    - Validate with validate_canon before saving
    
    Timeout: 30 minutes (1800000 ms)
  ",
  timeout: 1800000
)
```
```

### Manejo de Timeout Errors

**Si un sub-agente timeoutea:**

```markdown
## Timeout Recovery Protocol

1. **Check progress:**
   - ¿Cuántas áreas se completaron antes del timeout?
   - ¿Hay archivos guardados parcialmente?

2. **Decide strategy:**
   - **Option A:** Restart with same timeout (if <50% complete)
   - **Option B:** Increase timeout to 45 min (if >50% complete)
   - **Option C:** Split remaining work (if clear stopping point)

3. **Restart with adjusted parameters:**
   ```
   delegate(
     agent: "grimorio-areas",
     prompt: "Generate areas 5-8 (remaining from previous batch)",
     timeout: 1800000
   )
   ```

4. **Merge results:**
   - Verify chapter numbers are sequential
   - Check for duplicate areas
   - Validate consistency across batches
```

---

**TL;DR:** Timeout de 15 min es insuficiente para grimorio-areas (genera 10-15 áreas × WotC format = 27-41 min reales). Configurar timeout a 30 min (1800000 ms) para grimorio-areas. Alternativa: dividir en batches de 4 áreas (8-12 min c/u). Timeout recommendations: areas=30min, npc=20min, bestiary=20min, encounters/lore/maps/quests=15min.


---

## 📂 Estructura de Carpetas Incorrecta en Campañas

### Problema
Los agentes grimorio están guardando archivos en carpetas incorrectas dentro de la campaña. Ejemplo: `lore.md` se guarda en la raíz pero los otros agentes no lo encuentran porque asumen otra estructura.

### Síntomas
- `grimorio-areas` no encuentra `lore.md` (busca en `/lore/lore.md` pero está en `/lore.md`)
- `grimorio-npc` no encuentra `bestiary.md` (busca en `/bestiary/bestiary.md` pero está en otra ubicación)
- `grimorio-encounters` no encuentra `npcs_and_factions.md`
- **Error común:** `File not found: /home/pau/campaigns/{campaign}/lore/lore.md`
- **Realidad:** El archivo está en `/home/pau/campaigns/{campaign}/lore.md` (sin subcarpeta)

### Causa Raíz

**Inconsistencia en paths entre agentes:**

| Agente | Asume que lore está en | Asume que npcs está en | Realidad |
|--------|----------------------|----------------------|----------|
| grimorio-areas | `/lore/lore.md` ❌ | `/npcs/npcs_and_factions.md` ✅ | `lore.md` en raíz |
| grimorio-npc | `/lore.md` ✅ | `/npcs/npcs_and_factions.md` ✅ | Correcto |
| grimorio-bestiary | `/lore.md` ✅ | N/A | Correcto |
| grimorio-encounters | `/lore.md` ✅ | `/npcs/npcs_and_factions.md` ✅ | Correcto |
| grimorio-maps | `/canon.json` ✅ | `/npcs/npcs_and_factions.md` ✅ | Correcto |

**Problema:** Algunos agentes usan `/lore/lore.md` (subcarpeta) pero el archivo está en `/lore.md` (raíz).

### Estructura CORRECTA de Campañas

```
/home/pau/campaigns/{campaign_name}/
├── lore.md                    ✅ (archivo único en raíz)
├── canon.json                 ✅ (archivo único en raíz)
├── story_brief.md            ✅ (archivo único en raíz, opcional)
├── introduction.md           ✅ (archivo único en raíz, opcional)
├── README.md                 ✅ (archivo único en raíz, opcional)
│
├── areas/                    ✅ (subcarpeta para áreas numeradas)
│   ├── chapter_01_sombras_en_la_corte.md
│   ├── chapter_02_traiciones_y_alianzas.md
│   └── chapter_03_guerra_abierta.md
│
├── npcs/                     ✅ (subcarpeta para NPCs)
│   └── npcs_and_factions.md
│
├── bestiary/                 ✅ (subcarpeta para monstruos)
│   └── bestiary.md
│
├── encounters/               ✅ (subcarpeta para encuentros)
│   └── encounters.md
│
├── maps/                     ✅ (subcarpeta para mapas)
│   └── maps.md
│
├── quests/                   ✅ (subcarpeta para quests)
│   └── quests.md
│
├── characters/               ✅ (subcarpeta para PCs)
│   ├── dmitri_volkov.md
│   ├── elena_corvus.md
│   └── ...
│
└── appendices/               ✅ (subcarpeta para anexos)
    ├── items.md
    ├── monsters.md
    └── handouts.md
```

### Paths CORRECTOS por Agente

**Todos los agentes deben usar ESTA estructura:**

```markdown
## File Paths (CRITICAL)

**ALWAYS use these exact paths:**

| File | Path | Notes |
|------|------|-------|
| Lore | `/home/pau/campaigns/{campaign}/lore.md` | NOT `/lore/lore.md` |
| Canon | `/home/pau/campaigns/{campaign}/canon.json` | Root level |
| Story Brief | `/home/pau/campaigns/{campaign}/story_brief.md` | Optional |
| NPCs | `/home/pau/campaigns/{campaign}/npcs/npcs_and_factions.md` | In subfolder |
| Bestiary | `/home/pau/campaigns/{campaign}/bestiary/bestiary.md` | In subfolder |
| Encounters | `/home/pau/campaigns/{campaign}/encounters/encounters.md` | In subfolder |
| Maps | `/home/pau/campaigns/{campaign}/maps/maps.md` | In subfolder |
| Quests | `/home/pau/campaigns/{campaign}/quests/quests.md` | In subfolder |
| Characters | `/home/pau/campaigns/{campaign}/characters/` | In subfolder |
| Areas | `/home/pau/campaigns/{campaign}/areas/` | In subfolder |
```

### Solución SDD

#### 1. Actualizar TODOS los Agents con Paths Correctos

**Para grimorio-areas.md:**

```markdown
## CRITICAL: File Paths

**BEFORE reading any files, use these EXACT paths:**

```bash
# Root files (NO subfolder)
/home/pau/campaigns/{campaign}/lore.md
/home/pau/campaigns/{campaign}/canon.json
/home/pau/campaigns/{campaign}/story_brief.md

# Subfolder files
/home/pau/campaigns/{campaign}/npcs/npcs_and_factions.md
/home/pau/campaigns/{campaign}/bestiary/bestiary.md
/home/pau/campaigns/{campaign}/encounters/encounters.md
/home/pau/campaigns/{campaign}/maps/maps.md
/home/pau/campaigns/{campaign}/quests/quests.md
```

**DO NOT use:**
- ❌ `/home/pau/campaigns/{campaign}/lore/lore.md` (wrong!)
- ❌ `/home/pau/campaigns/{campaign}/npcs/npc.md` (wrong!)
- ❌ Relative paths like `lore.md` or `npcs/npcs.md`

**ALWAYS use absolute paths from `/home/pau/campaigns/{campaign}/`**
```

**Para grimorio-npc.md:**

```markdown
## CRITICAL: File Paths

**Read files in this order:**

1. `/home/pau/campaigns/{campaign}/canon.json` (root)
2. `/home/pau/campaigns/{campaign}/lore.md` (root, NOT lore/lore.md)
3. `/home/pau/campaigns/{campaign}/story_brief.md` (root, if exists)
4. `/home/pau/campaigns/{campaign}/bestiary/bestiary.md` (subfolder)
```

**Para grimorio-encounters.md:**

```markdown
## CRITICAL: File Paths

**Read files in this order:**

1. `/home/pau/campaigns/{campaign}/canon.json` (root)
2. `/home/pau/campaigns/{campaign}/lore.md` (root, NOT lore/lore.md)
3. `/home/pau/campaigns/{campaign}/npcs/npcs_and_factions.md` (subfolder)
4. `/home/pau/campaigns/{campaign}/bestiary/bestiary.md` (subfolder)
5. `/home/pau/campaigns/{campaign}/story_brief.md` (root, if exists)
```

#### 2. Validación Pre-Lectura

Agregar verificación de que los archivos existen antes de leer:

```markdown
## Pre-Flight Check

**BEFORE generating content, verify all required files exist:**

```bash
# Check root files
test -f /home/pau/campaigns/{campaign}/lore.md && echo "✅ lore.md" || echo "❌ lore.md MISSING"
test -f /home/pau/campaigns/{campaign}/canon.json && echo "✅ canon.json" || echo "❌ canon.json MISSING"

# Check subfolder files
test -f /home/pau/campaigns/{campaign}/npcs/npcs_and_factions.md && echo "✅ npcs" || echo "❌ npcs MISSING"
test -f /home/pau/campaigns/{campaign}/bestiary/bestiary.md && echo "✅ bestiary" || echo "❌ bestiary MISSING"
```

**If any file is missing:**
1. Stop generation
2. Report which file is missing
3. Ask user to generate that file first
4. DO NOT proceed with incomplete context
```

#### 3. Script de Verificación de Estructura

```bash
#!/bin/bash
# verify_campaign_structure.sh

CAMPAIGN=$1
BASE="/home/pau/campaigns/$CAMPAIGN"

echo "=== Verifying Campaign Structure: $CAMPAIGN ==="
echo ""

# Root files
echo "Root Files:"
test -f "$BASE/lore.md" && echo "  ✅ lore.md" || echo "  ❌ lore.md MISSING"
test -f "$BASE/canon.json" && echo "  ✅ canon.json" || echo "  ❌ canon.json MISSING"
test -f "$BASE/story_brief.md" && echo "  ✅ story_brief.md" || echo "  ℹ️  story_brief.md (optional)"

# Subfolders
echo ""
echo "Subfolders:"
for dir in areas npcs bestiary encounters maps quests characters; do
  test -d "$BASE/$dir" && echo "  ✅ $dir/" || echo "  ❌ $dir/ MISSING"
done

# Key files in subfolders
echo ""
echo "Key Files:"
test -f "$BASE/npcs/npcs_and_factions.md" && echo "  ✅ npcs/npcs_and_factions.md" || echo "  ❌ npcs/npcs_and_factions.md MISSING"
test -f "$BASE/bestiary/bestiary.md" && echo "  ✅ bestiary/bestiary.md" || echo "  ❌ bestiary/bestiary.md MISSING"
test -f "$BASE/encounters/encounters.md" && echo "  ✅ encounters/encounters.md" || echo "  ❌ encounters/encounters.md MISSING"
test -f "$BASE/maps/maps.md" && echo "  ✅ maps/maps.md" || echo "  ❌ maps/maps.md MISSING"

echo ""
echo "=== Verification Complete ==="
```

**Uso:**
```bash
chmod +x verify_campaign_structure.sh
./verify_campaign_structure.sh la-hoja-de-vlad
```

#### 4. Migración de Estructura Incorrecta

**Si la campaña tiene estructura incorrecta:**

```bash
# Estructura incorrecta (ejemplo)
/home/pau/campaigns/{campaign}/
├── lore/
│   └── lore.md          ❌ (debería estar en raíz)
├── npcs/
│   └── npcs_and_factions.md ✅
└── ...

# Migrar a estructura correcta
cd /home/pau/campaigns/{campaign}/

# Mover lore.md a raíz
mv lore/lore.md ./lore.md
rmdir lore/

# Mover canon.json a raíz (si está en subcarpeta)
mv canon/canon.json ./canon.json 2>/dev/null || true
rmdir canon/ 2>/dev/null || true

# Verificar nueva estructura
tree -L 2
```

### Checklist de Estructura

**Antes de generar contenido, verificar:**

| Check | Path | Estado |
|-------|------|--------|
| Lore en raíz | `/home/pau/campaigns/{camp}/lore.md` | ☐ |
| Canon en raíz | `/home/pau/campaigns/{camp}/canon.json` | ☐ |
| Story brief en raíz | `/home/pau/campaigns/{camp}/story_brief.md` | ☐ |
| NPCs en subcarpeta | `/home/pau/campaigns/{camp}/npcs/npcs_and_factions.md` | ☐ |
| Bestiary en subcarpeta | `/home/pau/campaigns/{camp}/bestiary/bestiary.md` | ☐ |
| Encounters en subcarpeta | `/home/pau/campaigns/{camp}/encounters/encounters.md` | ☐ |
| Maps en subcarpeta | `/home/pau/campaigns/{camp}/maps/maps.md` | ☐ |
| Quests en subcarpeta | `/home/pau/campaigns/{camp}/quests/quests.md` | ☐ |
| Characters en subcarpeta | `/home/pau/campaigns/{camp}/characters/` | ☐ |
| Areas en subcarpeta | `/home/pau/campaigns/{camp}/areas/` | ☐ |

### Error Común: "File not found: lore/lore.md"

**Cuando veas este error:**

```
Error: File not found: /home/pau/campaigns/la-hoja-de-vlad/lore/lore.md
```

**Solución:**

1. **Verificar dónde está realmente el archivo:**
   ```bash
   find /home/pau/campaigns/la-hoja-de-vlad/ -name "lore.md"
   # Resultado probable: /home/pau/campaigns/la-hoja-de-vlad/lore.md
   ```

2. **Si está en la raíz pero el agente busca en subcarpeta:**
   - El agente tiene un bug de path
   - Actualizar el agente para usar `/lore.md` en vez de `/lore/lore.md`

3. **Si está en subcarpeta pero debería estar en raíz:**
   ```bash
   mv /home/pau/campaigns/la-hoja-de-vlad/lore/lore.md /home/pau/campaigns/la-hoja-de-vlad/lore.md
   rmdir /home/pau/campaigns/la-hoja-de-vlad/lore/
   ```

### Tabla Resumen: Paths Correctos

| Tipo | Path Correcto | Path Incorrecto |
|------|--------------|-----------------|
| **Lore** | `/home/pau/campaigns/{camp}/lore.md` | ❌ `/lore/lore.md` |
| **Canon** | `/home/pau/campaigns/{camp}/canon.json` | ❌ `/canon/canon.json` |
| **Story Brief** | `/home/pau/campaigns/{camp}/story_brief.md` | ❌ `/story_brief/story_brief.md` |
| **NPCs** | `/home/pau/campaigns/{camp}/npcs/npcs_and_factions.md` | ❌ `/npcs.md` |
| **Bestiary** | `/home/pau/campaigns/{camp}/bestiary/bestiary.md` | ❌ `/bestiary.md` |
| **Encounters** | `/home/pau/campaigns/{camp}/encounters/encounters.md` | ❌ `/encounters.md` |
| **Maps** | `/home/pau/campaigns/{camp}/maps/maps.md` | ❌ `/maps.md` |
| **Quests** | `/home/pau/campaigns/{camp}/quests/quests.md` | ❌ `/quests.md` |
| **Characters** | `/home/pau/campaigns/{camp}/characters/*.md` | ❌ `/characters.json` |
| **Areas** | `/home/pau/campaigns/{camp}/areas/chapter_*.md` | ❌ `/areas.md` |

---

**TL;DR:** Lore y canon van en RAÍZ (`/lore.md`, `/canon.json`). NPCs, bestiary, encounters, maps, quests, characters, areas van en SUBCARPETAS (`/npcs/`, `/bestiary/`, etc.). Todos los agents deben usar paths absolutos desde `/home/pau/campaigns/{campaign}/`. Verificar estructura con script antes de generar.


---

## 🗂️ Estructura Real de Campañas vs Paths que Buscan los Agents

### Problema Documentado con Ejemplos Reales

**Campaña:** `la-hoja-de-vlad` en `/home/pau/campaigns/la-hoja-de-vlad/`

**Lo que buscan los agents ❌:**
```
/home/pau/campaigns/la-hoja-de-vlad/lore/lore.md          ❌ File not found
/home/pau/campaigns/la-hoja-de-vlad/npcs/npcs.md          ❌ File not found
/home/pau/campaigns/la-hoja-de-vlad/quests/quests.md      ❌ File not found
/home/pau/campaigns/la-hoja-de-vlad/characters/characters.md ❌ File not found
```

**Lo que REALMENTE existe ✅:**
```
/home/pau/campaigns/la-hoja-de-vlad/lore.md               ✅ (en raíz, no subcarpeta)
/home/pau/campaigns/la-hoja-de-vlad/npcs/npcs_and_factions.md ✅ (nombre completo)
/home/pau/campaigns/la-hoja-de-vlad/quests/quest_1778275935.json ⚠️ (JSON, no MD)
/home/pau/campaigns/la-hoja-de-vlad/characters/dmitri_volkov.json ⚠️ (múltiples .json)
/home/pau/campaigns/la-hoja-de-vlad/canon/canon.json      ⚠️ (en subcarpeta, debería estar en raíz)
```

### Estructura CORRECTA Confirmada

```
/home/pau/campaigns/{campaign_name}/
├── lore.md                    ✅ RAÍZ (no lore/lore.md)
├── canon.json                 ✅ RAÍZ (no canon/canon.json)
├── narrative_state.json       ✅ RAÍZ (no canon/narrative_state.json)
├── story_brief.md            ✅ RAÍZ (opcional)
├── README.md                 ✅ RAÍZ
├── session-zero.md           ✅ RAÍZ
│
├── acts/                     ✅ CARPETA para actos/capítulos
│   ├── chapter_01_sombras_en_la_corte.md
│   ├── chapter_02_traiciones_y_alianzas.md
│   └── chapter_03_la_revelaci_n_de_vlad.md
│
├── areas/                    ❌ OBSOLETA - eliminar
│
├── npcs/
│   └── npcs_and_factions.md  ✅ (NO npcs.md)
│
├── bestiary/
│   └── bestiary.md           ✅
│
├── encounters/
│   └── encounters.md         ✅
│
├── maps/
│   └── maps.md               ✅
│
├── quests/
│   └── quest_*.json          ⚠️ (JSONs individuales, no quests.md)
│
├── characters/
│   ├── dmitri_volkov.json    ⚠️ (JSONs individuales, no characters.md)
│   ├── elena_corvus.json
│   └── ...
│
└── assets/                   ✅ SVGs, imágenes, battle maps
    ├── callejon-dagas-rotas.svg
    ├── castillo-sombria.svg
    └── santuario-juramentos.svg
```

### Paths CORRECTOS por Tipo de Archivo

| Tipo | Path CORRECTO | Path INCORRECTO | Estado |
|------|--------------|-----------------|--------|
| **Lore** | `/home/pau/campaigns/{camp}/lore.md` | ❌ `/lore/lore.md` | File not found |
| **Canon** | `/home/pau/campaigns/{camp}/canon.json` | ❌ `/canon/canon.json` | File not found |
| **Narrative State** | `/home/pau/campaigns/{camp}/narrative_state.json` | ❌ `/canon/narrative_state.json` | File not found |
| **NPCs** | `/home/pau/campaigns/{camp}/npcs/npcs_and_factions.md` | ❌ `/npcs/npcs.md` | File not found |
| **Bestiary** | `/home/pau/campaigns/{camp}/bestiary/bestiary.md` | ❌ `/bestiary.md` | File not found |
| **Encounters** | `/home/pau/campaigns/{camp}/encounters/encounters.md` | ❌ `/encounters.md` | File not found |
| **Maps** | `/home/pau/campaigns/{camp}/maps/maps.md` | ❌ `/maps.md` | File not found |
| **Quests** | `/home/pau/campaigns/{camp}/quests/*.json` | ❌ `/quests/quests.md` | File not found |
| **Characters** | `/home/pau/campaigns/{camp}/characters/*.json` | ❌ `/characters/characters.md` | File not found |
| **Acts** | `/home/pau/campaigns/{camp}/acts/chapter_*.md` | ❌ `/areas/chapter_*.md` | Obsoleto |
| **Assets** | `/home/pau/campaigns/{camp}/assets/*.svg` | ✅ Correcto | OK |

### Solución SDD

#### 1. Eliminar Carpeta `areas/` Obsoleta

```bash
# Verificar que acts/ ya tiene el contenido
ls -la /home/pau/campaigns/{campaign}/acts/

# Si acts/ tiene los capítulos, eliminar areas/
rm -rf /home/pau/campaigns/{campaign}/areas/

# Verificar eliminación
ls -la /home/pau/campaigns/{campaign}/ | grep -E "acts|areas"
# Debería mostrar solo: acts/
```

#### 2. Mover `canon.json` y `narrative_state.json` a Raíz

```bash
# Mover archivos de canon/ a raíz
cd /home/pau/campaigns/{campaign}/
mv canon/canon.json ./canon.json
mv canon/narrative_state.json ./narrative_state.json

# Eliminar carpeta canon/ vacía
rmdir canon/

# Verificar
ls -la *.json
# Debería mostrar: canon.json, narrative_state.json, campaign.json
```

#### 3. Actualizar Agents con Paths Reales

**grimorio-areas.md (debería ser grimorio-acts.md):**

```markdown
## CRITICAL: File Paths (VERIFIED)

**ALWAYS use these EXACT paths:**

```bash
# Root files
/home/pau/campaigns/{campaign}/lore.md
/home/pau/campaigns/{campaign}/canon.json
/home/pau/campaigns/{campaign}/narrative_state.json
/home/pau/campaigns/{campaign}/story_brief.md

# Subfolder files
/home/pau/campaigns/{campaign}/npcs/npcs_and_factions.md
/home/pau/campaigns/{campaign}/bestiary/bestiary.md
/home/pau/campaigns/{campaign}/encounters/encounters.md
/home/pau/campaigns/{campaign}/maps/maps.md

# Output folder
/home/pau/campaigns/{campaign}/acts/chapter_{N}_{title}.md
```

**DO NOT use:**
- ❌ `/lore/lore.md` (should be `/lore.md`)
- ❌ `/canon/canon.json` (should be `/canon.json`)
- ❌ `/npcs/npcs.md` (should be `/npcs/npcs_and_factions.md`)
- ❌ `/areas/` folder (obsolete, use `/acts/`)
```

**grimorio-npc.md:**

```markdown
## CRITICAL: File Paths

**Read order:**
1. `/home/pau/campaigns/{campaign}/canon.json`
2. `/home/pau/campaigns/{campaign}/lore.md`
3. `/home/pau/campaigns/{campaign}/npcs/npcs_and_factions.md` (read existing, append new)
4. `/home/pau/campaigns/{campaign}/bestiary/bestiary.md`

**Write to:**
- `/home/pau/campaigns/{campaign}/npcs/npcs_and_factions.md`
```

**grimorio-quests.md:**

```markdown
## CRITICAL: File Paths

**Quests are stored as individual JSON files:**
- `/home/pau/campaigns/{campaign}/quests/quest_{timestamp}.json`

**DO NOT try to read/write quests.md (does not exist)**

**To list existing quests:**
```bash
ls /home/pau/campaigns/{campaign}/quests/*.json
```

**To create new quest:**
Use `create_personal_quest` MCP tool - it handles JSON creation automatically
```

**grimorio-generate-character.md:**

```markdown
## CRITICAL: File Paths

**Characters are stored as individual JSON files:**
- `/home/pau/campaigns/{campaign}/characters/{character_name}.json`

**DO NOT try to read/write characters.md (does not exist)**

**To list existing characters:**
```bash
ls /home/pau/campaigns/{campaign}/characters/*.json
```

**To create new character:**
Use `generate_character` or `save_characters` MCP tool - handles JSON format
```

#### 4. Script de Migración Completa

```bash
#!/bin/bash
# migrate_campaign_structure.sh

CAMPAIGN=$1
BASE="/home/pau/campaigns/$CAMPAIGN"

echo "=== Migrating Campaign Structure: $CAMPAIGN ==="
echo ""

# 1. Move canon files to root
if [ -d "$BASE/canon" ]; then
  echo "Moving canon files to root..."
  mv "$BASE/canon/canon.json" "$BASE/canon.json" 2>/dev/null || true
  mv "$BASE/canon/narrative_state.json" "$BASE/narrative_state.json" 2>/dev/null || true
  rmdir "$BASE/canon/" 2>/dev/null || true
  echo "  ✅ canon.json and narrative_state.json moved to root"
else
  echo "  ℹ️  canon/ folder not found (already migrated?)"
fi

# 2. Remove obsolete areas/ folder
if [ -d "$BASE/areas" ]; then
  echo "Removing obsolete areas/ folder..."
  # First check if acts/ exists and has content
  if [ -d "$BASE/acts" ] && [ "$(ls -A $BASE/acts)" ]; then
    rm -rf "$BASE/areas/"
    echo "  ✅ areas/ removed (acts/ has content)"
  else
    echo "  ⚠️  WARNING: acts/ is empty or missing. Review before deleting areas/"
  fi
else
  echo "  ℹ️  areas/ folder not found (already removed?)"
fi

# 3. Verify structure
echo ""
echo "=== Verifying New Structure ==="
echo "Root files:"
test -f "$BASE/lore.md" && echo "  ✅ lore.md" || echo "  ❌ lore.md MISSING"
test -f "$BASE/canon.json" && echo "  ✅ canon.json" || echo "  ❌ canon.json MISSING"
test -f "$BASE/narrative_state.json" && echo "  ✅ narrative_state.json" || echo "  ❌ narrative_state.json MISSING"

echo ""
echo "Folders:"
test -d "$BASE/acts" && echo "  ✅ acts/" || echo "  ❌ acts/ MISSING"
test -d "$BASE/areas" && echo "  ❌ areas/ STILL EXISTS (should be removed)" || echo "  ✅ areas/ removed"
test -d "$BASE/npcs" && echo "  ✅ npcs/" || echo "  ❌ npcs/ MISSING"
test -d "$BASE/bestiary" && echo "  ✅ bestiary/" || echo "  ❌ bestiary/ MISSING"

echo ""
echo "Key files:"
test -f "$BASE/npcs/npcs_and_factions.md" && echo "  ✅ npcs/npcs_and_factions.md" || echo "  ❌ npcs/npcs_and_factions.md MISSING"
test -f "$BASE/bestiary/bestiary.md" && echo "  ✅ bestiary/bestiary.md" || echo "  ❌ bestiary/bestiary.md MISSING"

echo ""
echo "=== Migration Complete ==="
```

**Uso:**
```bash
chmod +x migrate_campaign_structure.sh
./migrate_campaign_structure.sh la-hoja-de-vlad
```

#### 5. Pre-Flight Check para Agents

**Antes de generar contenido, los agents deben verificar:**

```markdown
## Pre-Flight File Check

**BEFORE generating, verify these files exist:**

```bash
REQUIRED_ROOT_FILES=(
  "/home/pau/campaigns/{campaign}/lore.md"
  "/home/pau/campaigns/{campaign}/canon.json"
  "/home/pau/campaigns/{campaign}/narrative_state.json"
)

REQUIRED_SUBFOLDERS=(
  "/home/pau/campaigns/{campaign}/npcs/npcs_and_factions.md"
  "/home/pau/campaigns/{campaign}/bestiary/bestiary.md"
  "/home/pau/campaigns/{campaign}/encounters/encounters.md"
  "/home/pau/campaigns/{campaign}/maps/maps.md"
)

# Check root files
for file in "${REQUIRED_ROOT_FILES[@]}"; do
  if [ ! -f "$file" ]; then
    echo "❌ ERROR: Required file not found: $file"
    echo "   Expected: lore.md, canon.json, narrative_state.json in ROOT"
    echo "   NOT in subfolders like lore/, canon/"
    exit 1
  fi
done

# Check subfolder files
for file in "${REQUIRED_SUBFOLDERS[@]}"; do
  if [ ! -f "$file" ]; then
    echo "❌ ERROR: Required file not found: $file"
    exit 1
  fi
done

echo "✅ All required files found"
```

**If check fails:**
1. Stop generation immediately
2. Report which file is missing
3. Suggest migration script
4. DO NOT proceed with incomplete context
```

### Checklist de Verificación

**Antes de ejecutar cualquier agente grimorio:**

| Check | Comando | Resultado Esperado |
|-------|---------|-------------------|
| Lore en raíz | `test -f /home/pau/campaigns/{camp}/lore.md` | ✅ true |
| Canon en raíz | `test -f /home/pau/campaigns/{camp}/canon.json` | ✅ true |
| Narrative state en raíz | `test -f /home/pau/campaigns/{camp}/narrative_state.json` | ✅ true |
| NPCs nombre correcto | `test -f /home/pau/campaigns/{camp}/npcs/npcs_and_factions.md` | ✅ true |
| acts/ existe | `test -d /home/pau/campaigns/{camp}/acts` | ✅ true |
| areas/ NO existe | `test -d /home/pau/campaigns/{camp}/areas` | ❌ false (debe estar eliminada) |
| canon/ NO existe | `test -d /home/pau/campaigns/{camp}/canon` | ❌ false (debe estar eliminada) |

### Errores Comunes y Soluciones

| Error | Causa | Solución |
|-------|-------|----------|
| `File not found: lore/lore.md` | Agente busca en subcarpeta | Mover a `/lore.md` (raíz) |
| `File not found: canon/canon.json` | Agente busca en subcarpeta | Mover a `/canon.json` (raíz) |
| `File not found: npcs/npcs.md` | Nombre incorrecto | Renombrar a `npcs_and_factions.md` |
| `File not found: quests/quests.md` | Quests son JSONs | Usar `list_quests` MCP tool |
| `File not found: characters/characters.md` | Characters son JSONs | Usar `list_characters` MCP tool |
| `areas/` y `acts/` coexisten | Migración incompleta | Eliminar `areas/`, mantener `acts/` |

### Migración de `areas/` a `acts/`

**Si tenés contenido en `areas/` que necesitás mover:**

```bash
# 1. Verificar contenido de areas/
ls -la /home/pau/campaigns/{campaign}/areas/

# 2. Si hay contenido útil, moverlo a acts/
mv /home/pau/campaigns/{campaign}/areas/chapter_*.md /home/pau/campaigns/{campaign}/acts/

# 3. Verificar que acts/ tiene todo
ls -la /home/pau/campaigns/{campaign}/acts/

# 4. Eliminar areas/
rm -rf /home/pau/campaigns/{campaign}/areas/

# 5. Confirmar
tree -L 2 /home/pau/campaigns/{campaign}/
```

### Resumen: Cambios Críticos

| Elemento | Antes ❌ | Después ✅ |
|----------|---------|-----------|
| Lore | `/lore/lore.md` | `/lore.md` |
| Canon | `/canon/canon.json` | `/canon.json` |
| Narrative State | `/canon/narrative_state.json` | `/narrative_state.json` |
| NPCs | `/npcs/npcs.md` | `/npcs/npcs_and_factions.md` |
| Acts Folder | `/areas/` | `/acts/` |
| Quests | `/quests/quests.md` | `/quests/*.json` (individual) |
| Characters | `/characters/characters.md` | `/characters/*.json` (individual) |

---

**TL;DR:** 
- **Raíz:** `lore.md`, `canon.json`, `narrative_state.json` (NO en subcarpetas)
- **NPCs:** `npcs/npcs_and_factions.md` (NO `npcs.md`)
- **Acts:** `/acts/` (NO `/areas/` - obsoleto)
- **Quests/Characters:** JSONs individuales (NO `.md` unificado)
- **Migración:** Mover canon a raíz, eliminar `areas/`, eliminar `canon/`

