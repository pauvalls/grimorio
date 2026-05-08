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

