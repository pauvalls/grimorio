# Change Proposal: WotC-Quality Adventure Generation Improvements

**Date:** 2026-05-08  
**Author:** SDD Propose Phase  
**Status:** Draft  
**Priority:** High  

---

## 1. Intent

### What We're Doing

Elevating GRIMORIO's adventure generation to match **Wizards of the Coast professional standards** by implementing systematic character hooks, branching decision paths, complete quest structures, and DM guidance that rivals published WotC adventures.

### Why This Matters

Current GRIMORIO output has **critical gaps** that prevent it from reaching professional quality:

1. **Character hooks exist but are NOT automatic** — The `hooks.go` service exists but is never called in the architect flow. Hooks are manually written or skipped entirely.

2. **Area "Developments" are weak** — Currently descriptions, not decision branches. WotC adventures use IF-THEN structures with explicit triggers, outcomes, and recovery paths.

3. **Quests are empty shells** — No objectives, rewards, or completion triggers. The `create_personal_quest` MCP tool accepts these fields but agents don't populate them consistently.

4. **No multiple solution paths** — Areas assume one approach (usually combat). WotC adventures provide stealth DC, social DC, and combat options for every major obstacle.

5. **NPCs lack complete stat blocks** — NPCs are described narratively but stat blocks live in bestiary, not linked. DMs must flip between files mid-session.

6. **No evolving central hub** — WotC adventures like *Waterdeep: Dragon Heist* feature hubs that change based on player decisions. GRIMORIO has no mechanism for this.

7. **Missing "Running the Adventure" guidance** — Each WotC chapter opens with 150-400 words of DM prep: pacing, tone, key decisions, when to improvise vs. follow script. GRIMORIO templates have this but it's not enforced.

**Business Impact:** Without these improvements, GRIMORIO remains a **draft generator**, not a **professional adventure creator**. DMs cannot run generated content without significant rewriting.

---

## 2. Scope

### In Scope

| Component | Change | Files Affected |
|-----------|--------|----------------|
| **Architect Flow** | Add automatic character hook generation in Phase 4 | `agents/grimorio-architect.md` |
| **MCP Tools** | Add `generate_character_hooks` tool | `internal/mcp/handlers/hooks.go` (new), `internal/mcp/server.go` |
| **Validators** | Add WotC-style rules for developments, hooks, quests, solutions | `internal/validators/area.go`, `internal/services/validation_engine.go` |
| **Templates** | Update areas template with WotC examples and required sections | `internal/compiler/templates/areas.md.tmpl` |
| **Quest Agent** | Enforce quest completeness (objectives, rewards, triggers) | `agents/grimorio-quests.md` |
| **Validation Rules** | Add 8 new WotC quality rules | `internal/services/validation_engine.go` |

### Out of Scope

- Changes to existing campaign storage format (canon.json schema remains v2.0)
- Image generation workflow (Phases 6-8 of architect)
- PDF compilation process
- Existing NPC/Bestiary generation agents (only stat block linking changes)
- Random table generation (separate enhancement track)

### Constraints

- Must maintain backward compatibility with existing campaigns
- Cannot break existing MCP tool signatures
- Must work within current delegate/subagent architecture
- Validation rules must be checkable programmatically (no subjective "quality" scores)

---

## 3. Approach

### Phase 1: Character Hook Integration

**Current State:**
```markdown
// hooks.go exists but is NEVER called in grimorio-architect.md
// Phase 4 generates quests, encounters, characters — but no hooks
```

**Target State:**

Add hook generation to Phase 4 of `grimorio-architect.md`:

```markdown
### Phase 4: Batch 2 — Contenido + Lore (PARALLEL)

**0. Character Hooks — Agent: grimorio-hooks** (NEW)
```
delegate(agent="grimorio-hooks", prompt="Generate CHARACTER HOOKS for campaign '{campaign_name}' at {campaign_path}.\n\nRead ALL character sheets from characters/*.md and canon.json.\n\nGenerate 2-3 hooks per character, tied to their background and class.\nUse generate_character_hooks MCP tool for each hook.")
```

**1. Lore — Agent: grimorio-lore**
...
```

**New MCP Tool:** `generate_character_hooks`

```go
// internal/mcp/handlers/hooks.go
func (h *HookHandlers) HandleGenerateCharacterHooks() server.ToolHandlerFunc {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        args := parseArgs(request)
        campaign := args["campaign"]
        characterName := args["character_name"] // optional: specific character
        count := args["count"] // default: 2
        
        hooks, warnings, err := h.service.GenerateHooks(ctx, campaign, characterName, count)
        if err != nil {
            return mcp.NewToolResultError(err.Error()), nil
        }
        
        // Return structured hooks for validation
        return mcp.NewToolResultText(marshalHooks(hooks, warnings)), nil
    }
}
```

**Validator Rule:** `character_hook_presence`
- **Check:** Each area must have 2-3 character hooks
- **Validation:** Regex match for `**Ganchos de Personaje**` section with minimum 2 bullet points
- **Severity:** error if missing, warning if <2 hooks

---

### Phase 2: WotC-Style Developments (Decision Branches)

**Current State:**
```markdown
**Desarrollo:**
- Si entran sigilosos: Qué pasa
- Si entran en combate: Cómo reaccionan las criaturas
```

**Problem:** These are descriptions, not decision trees. No triggers, no recovery paths.

**Target State:**

```markdown
**Desarrollos** (mínimo 3 ramas con recovery paths obligatorios)

1. **Los PJs investigan las pistas:** Descubren información crucial sobre el villano.
   - **Trigger:** Investigación DC 15 exitosa O gasto de recurso de clase
   - **Outcome:** Revelan clue-003, ventaja en Área 5
   - **Recovery:** Si no investigan, aliado NPC proporciona información más tarde (costo: reputación -5)

2. **Los PJs confrontan al NPC:** Se desencadena combate o negociación crítica.
   - **Trigger:** Acción hostil O fallo en Persuasión DC 18
   - **Outcome:** Combate con 3 guardias O NPC huye (aparece en Acto N+1)
   - **Recovery:** Si evitan confrontación, NPC los persigue en Acto N+1 con refuerzos

3. **Los PJs ignoran el área:** Pierden oportunidad de obtener ventaja.
   - **Trigger:** Decisión explícita de saltar el área
   - **Outcome:** Pierden tesoro y pista secundaria
   - **Recovery:** Consecuencia diferida al Acto N+1: área más difícil (+2 DC a todas las tiradas)
```

**Validator Rule:** `development_branch_structure`
- **Check:** Each area has 3-5 development branches with IF-THEN structure
- **Validation:** Parse for numbered branches, each containing Trigger/Outcome/Recovery
- **Severity:** error if <3 branches or missing recovery path

---

### Phase 3: Quest Completeness

**Current State:**
```markdown
// grimorio-quests.md creates quests but fields are optional
// Many generated quests lack objectives, rewards, or completion triggers
```

**Target State:**

Update `agents/grimorio-quests.md` to enforce structure:

```markdown
## Estructura de cada Misión (OBLIGATORIA)

Cada misión usa `create_personal_quest` con estos campos **REQUIRED**:

- **campaign**: Nombre de la campaña
- **quest_title**: Título atractivo
- **quest_type**: redencion / venganza / descubrimiento / proteccion
- **character_name**: Para quién es esta misión
- **hook**: Cómo se introduce (1-2 párrafos)
- **objectives**: Array de 2-3 objetivos concretos (ej: ["Derrotar al villano", "Rescatar al aliado"])
- **completion_trigger**: Condición de completado (ej: "Villain defeated AND item retrieved")
- **stakes**: Qué se pierde si falla
- **reward**: Array de recompensas (ej: ["500 gp", "Espada mágica +1", "Aliado: NPC X"])
- **xp_reward**: XP total (calculated: objetivo_principal × 100 + secundarios × 50)
```

**New Validator Rule:** `quest_completeness`
- **Check:** Quest has objectives (2-3), completion_trigger, rewards (array), XP
- **Validation:** Parse quest markdown for required sections
- **Severity:** error if any field missing

---

### Phase 4: Multiple Solution Paths

**Current State:**
```markdown
// Areas assume single solution path (usually combat)
// No explicit stealth or social options with DCs
```

**Target State:**

Add to area template:

```markdown
**Múltiples Soluciones** (mínimo 3 paths por obstáculo principal)

| Path | Habilidad | DC | Éxito | Fallo | Recovery |
|------|-----------|-----|-------|-------|----------|
| **Sigilo** | Sigilo | DC 15 | Pasan desapercibidos | Alerta temprana | Encuentro más difícil (+1 CR) |
| **Social** | Persuasión/Engaño | DC 18 | NPC proporciona información | NPC hostil | Ruta alternativa por combate |
| **Combate** | — | — | Derrotan guardias | Pierden recursos | NPC huye, aparece más tarde |
| **Creativo** | Herramientas de ladrón/Hechizo | DC 17 | Bypassean obstáculo | Gasto de recurso sin beneficio | Vuelven al path principal |
```

**Validator Rule:** `multiple_solution_paths`
- **Check:** Each major obstacle has 3+ solution paths with numeric DCs
- **Validation:** Parse for solution table or explicit path list with DCs
- **Severity:** error if <3 paths or any DC is non-numeric (e.g., "DC alto")

---

### Phase 5: NPC Stat Block Integration

**Current State:**
```markdown
// NPCs described in npcs/npcs_and_factions.md
// Stat blocks in bestiary/bestiary.md
// No cross-reference — DM must manually link them
```

**Target State:**

Update NPC template to include stat block reference:

```markdown
## NPC: Nombre del NPC

**Rol:** [Quest giver / Villain / Ally / Merchant]  
**Facción:** [Nombre de facción]  
**Ubicación:** [Área inicial]  
**Estado:** [alive/dead/missing]  

![Retrato](assets/npc-nombre.png)

### Descripción Narrativa
[2-3 párrafos: apariencia, personalidad, motivación, secreto]

### Stat Block Reference
**Usar stat block:** [Nombre exacto del bestiary.md]  
**CR:** [X]  
**XP:** [XXX]  

**Modificaciones al stat block:**
- [Cambio 1: ej: "Añade hechizo Curar Heridas 3/día"]
- [Cambio 2: ej: "Cambia alineamiento a Legal Neutral"]

### Información que Proporciona
- [Pista 1] — [Condición: ej: "Si Persuasión DC 15"]
- [Pista 2] — [Condición]

### Quests Relacionadas
- [Quest ID 1] — [Rol: giver/target/reward]
- [Quest ID 2] — [Rol]

### Muerte/Consecuencias
**SI muere en Acto N, ENTONCES:**
- [Consecuencia 1: ej: "Facción X declara hostilidad"]
- [Consecuencia 2: ej: "Quest Y falla automáticamente"]
```

**Validator Rule:** `npc_stat_block_link`
- **Check:** Every NPC references a stat block from bestiary.md
- **Validation:** Cross-reference NPC "Usar stat block" field against bestiary entity names
- **Severity:** error if stat block not found

---

### Phase 6: Evolving Central Hub

**Current State:**
```markdown
// No hub evolution mechanism
// Locations are static across acts
```

**Target State:**

Add to area template (already present but not enforced):

```markdown
**Evolución del Hub** (obligatorio para áreas hub: ciudades, bases, campamentos)

### Estado Inicial
[Descripción del hub al llegar los PJs: ej: "Taberna tranquila, NPCs desconfiados"]

### Puntos de Cambio
[Eventos que alteran el hub: ej: "Facción A gana poder", "Ataque enemigo"]

### Estados Posibles (2-3 variaciones según decisiones)

| Condición | Estado Resultante | Impacto en PJs |
|-----------|-------------------|----------------|
| **SI** ayudan a facción A, **ENTONCES** | Taberna se vuelve base de operaciones | Descuentos en equipos, información gratis |
| **SI** permanecen neutrales, **ENTONCES** | Hub neutral pero menos útil | Precios normales, información limitada |
| **SI** apoyan facción B, **ENTONCES** | Taberna cerrada para los PJs | Deben buscar otro hub, +1 día de viaje |

### Tracking de Estado
**Variable de estado:** `hub_state`  
**Valores posibles:** `friendly`, `neutral`, `hostile`  
**Actualizado por:** Decisiones en Áreas 3, 7, 12  
```

**Validator Rule:** `hub_evolution_presence`
- **Check:** Hub areas (identified by "hub" in name or type) have evolution section
- **Validation:** Parse for "Evolución del Hub" section with 2+ states
- **Severity:** warning if missing (not all areas are hubs)

---

### Phase 7: "Running the Adventure" Guidance

**Current State:**
```markdown
// Template has "Cómo Dirigir esta Escena" but agents don't consistently fill it
// No word count enforcement
```

**Target State:**

Update areas template and add validator:

```markdown
**Cómo Dirigir esta Escena** (5 subsecciones obligatorias, 150-400 palabras total)

1. **Preparación:** (2-3 bullets)
   - [Qué necesita el DM antes de la sesión]
   - [NPCs listos con stats y motivaciones]
   - [Mapas o tokens para posicionamiento]

2. **Ritmo Sugerido:** (timing estimado)
   - [Exploración inicial: X min]
   - [Interacción/Combate: Y min]
   - [Resolución/Consecuencias: Z min]

3. **Señales de los Jugadores:**
   - **Enganchados:** [hacen preguntas, toman notas, debaten estrategias]
   - **Aburridos:** [miran celulares, conversaciones fuera de juego, poca participación]

4. **Cuándo Improvisar:** (zonas flexibles)
   - [Diálogos de NPCs secundarios]
   - [Descripción de ambientes menores]
   - [Recompensas creativas por soluciones ingeniosas]

5. **Cuándo ceñirse al Guión:** (elementos críticos)
   - [Muertes de NPCs clave]
   - [Revelación de pistas principales]
   - [Consecuencias de decisiones mayores]
```

**Validator Rule:** `running_guidance_completeness`
- **Check:** Each area has all 5 subsections, 150-400 words total
- **Validation:** Word count + section presence check
- **Severity:** error if <150 words or missing any subsection

---

## 4. Implementation Plan

### Week 1: Foundation
- [ ] Create `internal/mcp/handlers/hooks.go` with `generate_character_hooks` tool
- [ ] Register tool in `internal/mcp/server.go`
- [ ] Create `agents/grimorio-hooks.md` agent
- [ ] Update `agents/grimorio-architect.md` Phase 4 to call hooks agent

### Week 2: Validators
- [ ] Add `ValidateDevelopments()` to `internal/validators/area.go`
- [ ] Add `ValidateQuestCompleteness()` to `internal/services/validation_engine.go`
- [ ] Add `ValidateMultipleSolutions()` to `internal/validators/area.go`
- [ ] Add `ValidateRunningGuidance()` to `internal/validators/area.go`
- [ ] Write unit tests for all new validators

### Week 3: Templates & Agents
- [ ] Update `internal/compiler/templates/areas.md.tmpl` with WotC examples
- [ ] Update `agents/grimorio-quests.md` to enforce quest structure
- [ ] Update `agents/grimorio-areas.md` to require new sections
- [ ] Update `internal/compiler/templates/npc.md.tmpl` for stat block linking

### Week 4: Integration & Testing
- [ ] End-to-end test: Generate full campaign with new validators
- [ ] Fix any validation false positives/negatives
- [ ] Update documentation in README.md
- [ ] Run existing test suite to ensure no regressions

---

## 5. Risks and Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| **Validators too strict** — Reject valid content | Medium | High | Start with warnings, escalate to errors after 2 weeks of tuning |
| **Agent prompt bloat** — Prompts become too long | Medium | Medium | Use template variables, extract common sections to shared includes |
| **Backward compatibility** — Breaks existing campaigns | Low | High | Version gate new validators (only apply to campaigns created after date X) |
| **Performance** — Validation slows generation | Low | Medium | Cache validation results, run validators in parallel where possible |
| **Agent non-compliance** — Agents ignore new requirements | Medium | High | Add validation to consistency gate — content rejected if missing required sections |

---

## 6. Success Criteria

### Quantitative Metrics

| Metric | Current | Target | Measurement |
|--------|---------|--------|-------------|
| **Character hooks per area** | 0-1 (manual) | 2-3 (automatic) | Count hooks in generated areas |
| **Development branches per area** | 1-2 (descriptions) | 3-5 (IF-THEN) | Parse for numbered branches with Trigger/Outcome/Recovery |
| **Quests with complete structure** | ~30% | 100% | Validate objectives, rewards, triggers present |
| **Solution paths per obstacle** | 1 (combat) | 3+ (stealth/social/combat) | Count paths with numeric DCs |
| **NPCs with stat block links** | ~20% | 100% | Cross-reference NPC files with bestiary |
| **Hub areas with evolution** | 0% | 100% (for hubs) | Check hub areas for evolution section |
| **Running guidance word count** | ~50 words | 150-400 words | Word count per area guidance section |

### Qualitative Metrics

- [ ] **DM Playtest:** A DM can run a generated adventure without rewriting content
- [ ] **WotC Comparison:** Generated output matches structure of published WotC adventures (e.g., *Curse of Strahd*, *Dragon Heist*)
- [ ] **Validator Coverage:** All 8 new WotC rules are enforced programmatically
- [ ] **Agent Compliance:** All agents produce content that passes validation on first attempt (no rejection loops)

### Definition of Done

- [ ] All 7 phases implemented and tested
- [ ] All 8 new validation rules passing unit tests
- [ ] End-to-end campaign generation succeeds with 100% validation pass rate
- [ ] Documentation updated (README.md, agent docs)
- [ ] No regressions in existing test suite
- [ ] Sample campaign generated and reviewed against WotC standards

---

## 7. Appendix: WotC Reference Examples

### Example: Character Hook (from *Curse of Strahd*)

```markdown
**Ganchos de Personaje:**

- **Para personaje con trasfondo religioso:** Reconocés los símbolos tallados en las paredes como pertenecientes a un culto antiguo que creías extinto. Tu conocimiento religioso te da ventaja en identificar sus rituales.

- **Para personaje con motivación de venganza:** Una de las figuras encapuchadas lleva un amuleto idéntico al que usaba quien destruyó tu hogar. Este es tu momento de confrontar el pasado.
```

### Example: Development Branch (from *Dragon Heist*)

```markdown
**Desarrollo 1: Los PJs investigan las pistas**

- **Trigger:** Investigación DC 15 exitosa
- **Outcome:** Descubren que el mapa apunta al Distrito de los Castillos
- **Recovery:** Si fallan, Volothamp Geddarm les da la información (pero cobra 50 gp)

**Desarrollo 2: Los PJs confrontan al NPC**

- **Trigger:** Acción hostil o Persuasión DC 18 fallida
- **Outcome:** Combate con 3 guardias (stat block: Guard, MM p. 347)
- **Recovery:** Si los PJs huyen, los guardias no persiguen (pero alertan a la zona)
```

### Example: Multiple Solution Paths (from *Tomb of Annihilation*)

```markdown
**Obstáculo: Puerta sellada mágicamente**

| Path | Habilidad | DC | Éxito | Fallo |
|------|-----------|-----|-------|-------|
| **Sigilo** | Sigilo | DC 17 | Pasan por túnel lateral | Alerta a los guardias |
| **Social** | Persuasión | DC 20 | Guardias abren la puerta | Guardias atacan |
| **Combate** | — | — | Derrotan guardias | Pierden recursos, alertan zona |
| **Mágico** | Arcano | DC 18 | Disipan el sello | Pierden slot de hechizo, puerta permanece |
```

---

## 8. Related Documents

- `agents/grimorio-architect.md` — Main orchestration flow (Phase 4 modification)
- `agents/grimorio-quests.md` — Quest generation (structure enforcement)
- `internal/validators/area.go` — Area validation rules (new checks)
- `internal/services/validation_engine.go` — Quest/NPC validation (new rules)
- `internal/compiler/templates/areas.md.tmpl` — Area template (WotC examples)
- `internal/services/hooks.go` — Hook generation service (already exists, needs integration)

---

## 9. Next Steps

1. **Review this proposal** — Stakeholder approval required before proceeding
2. **Create technical design** — Detailed implementation specs for each phase
3. **Break down into tasks** — Granular task list for implementation
4. **Implement Phase 1** — Character hook integration (lowest risk, highest value)
5. **Iterate** — Test, validate, refine before moving to next phase

---

**Approval Required:** This proposal must be reviewed and approved before proceeding to the Design phase.

**Reviewers:** 
- [ ] Product Owner
- [ ] Technical Lead
- [ ] DM Playtesters (for qualitative validation)
