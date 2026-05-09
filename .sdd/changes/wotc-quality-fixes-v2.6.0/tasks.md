# Tasks: WotC Quality Fixes v2.6.0

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~400-500 lines |
| 400-line budget risk | Medium-High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 (Phase 1-2: Introduction + Quests) → PR 2 (Phase 3-5: Characters + Validation + Docs) |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: Medium-High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Introduction + Quests generation & validation | PR 1 | Base: main; includes TASK-001 to TASK-005 |
| 2 | Characters + Validation + Documentation | PR 2 | Base: main (after PR 1 merges); includes TASK-006 to TASK-011 |

---

## Phase 1: Introduction

### Task 1.1: Generate introduction.md
**Type**: Content generation  
**File**: `campaigns/{campaign-name}/introduction.md`  
**Command**: `grimorio generate-introduction` (or equivalent content generation)

**Implementation:**
1. Ejecutar generación de `introduction.md` usando el template actualizado
2. Asegurar que el contenido incluya todas las secciones WotC requeridas:
   - Overview de la campaña
   - Adventure Background
   - Adventure Hooks
   - Running the Adventure
   - Adventure Flowchart reference
3. Verificar que el word count sea aproximadamente 5,000 palabras

**Acceptance:**
- [ ] Archivo `introduction.md` generado exitosamente
- [ ] Contiene todas las secciones WotC estándar
- [ ] Word count ≥ 5,000 palabras
- [ ] Sin errores de generación

---

### Task 1.2: Validate Introduction Word Count
**Type**: Validation  
**File**: `campaigns/{campaign-name}/introduction.md`  
**Command**: `wc -w campaigns/{campaign-name}/introduction.md`

**Implementation:**
1. Contar palabras del archivo generado:
   ```bash
   wc -w campaigns/{campaign-name}/introduction.md
   ```
2. Verificar que el count sea ≥ 5,000
3. Si es < 5,000, expandir contenido con:
   - Más detalle en Adventure Background
   - Hooks adicionales
   - Más ejemplos en Running the Adventure
4. Documentar el word count final

**Acceptance:**
- [ ] Word count verificado y documentado
- [ ] Count ≥ 5,000 palabras
- [ ] Si fue < 5,000, contenido expandido y re-validado

---

## Phase 2: Quests

### Task 2.1: Regenerate quests.md
**Type**: Content regeneration  
**File**: `campaigns/{campaign-name}/quests.md`  
**Command**: `grimorio generate-quests` (o regeneración manual)

**Implementation:**
1. Regenerar `quests.md` con estructura WotC completa
2. Asegurar que cada quest tenga:
   - Título descriptivo
   - Quest giver
   - Contexto/narrativa
   - Objetivos numerados (mínimo 3)
   - Recompensas explícitas
   - Estado (active/completed/failed)
3. Verificar consistencia con `narrative_state.json`

**Acceptance:**
- [ ] `quests.md` regenerado con formato WotC
- [ ] Cada quest tiene mínimo 3 objetivos
- [ ] Cada quest tiene al menos 1 reward especificado
- [ ] Quest givers identificados
- [ ] Estados de quests documentados

---

### Task 2.2: Regenerate quest_*.json files
**Type**: JSON regeneration  
**Files**: `campaigns/{campaign-name}/quest_*.json`  
**Command**: Generación programática o manual

**Implementation:**
1. Para cada quest en `quests.md`, crear/actualizar `quest_{id}.json` con:
   ```json
   {
     "id": "quest-001",
     "title": "Quest Title",
     "giver": "NPC Name",
     "objectives": [
       {"id": 1, "description": "...", "completed": false},
       {"id": 2, "description": "...", "completed": false},
       {"id": 3, "description": "...", "completed": false}
     ],
     "rewards": [
       {"type": "gold|item|xp|reputation", "value": "...", "amount": 0}
     ],
     "status": "active|completed|failed"
   }
   ```
2. Validar JSON syntax con `jq` o similar
3. Asegurar consistencia entre `quests.md` y `quest_*.json`

**Acceptance:**
- [ ] Todos los `quest_*.json` generados
- [ ] JSON válido (sin syntax errors)
- [ ] Mínimo 3 objectives por quest
- [ ] Mínimo 1 reward por quest
- [ ] Consistencia con `quests.md`

---

### Task 2.3: Validate Quests Structure
**Type**: Validation  
**Files**: `campaigns/{campaign-name}/quests.md`, `quest_*.json`  
**Command**: Validación manual o script custom

**Implementation:**
1. Iterar sobre todos los quests en `quests.md` y `quest_*.json`
2. Para cada quest, verificar:
   - objectives.length ≥ 3
   - rewards.length ≥ 1
   - Cada objective tiene descripción no vacía
   - Cada reward tiene type y value definidos
3. Reportar cualquier quest que no cumpla los requisitos
4. Fixear quests inválidos

**Acceptance:**
- [ ] 100% de quests tienen ≥ 3 objectives
- [ ] 100% de quests tienen ≥ 1 reward
- [ ] Todos los JSON files son válidos
- [ ] Reporte de validación documentado

---

## Phase 3: Characters

### Task 3.1: Regenerate characters.md
**Type**: Content regeneration  
**File**: `campaigns/{campaign-name}/characters.md`  
**Command**: `grimorio generate-characters` o regeneración manual

**Implementation:**
1. Regenerar `characters.md` con stats completos para cada personaje
2. Asegurar que cada character tenga:
   - Nombre y raza
   - Clase y nivel
   - Ability scores (STR, DEX, CON, INT, WIS, CHA)
   - HP (Hit Points)
   - AC (Armor Class)
   - Skills y proficiencies
   - Equipment
   - Background
3. Verificar que los stats estén en rangos válidos

**Acceptance:**
- [ ] `characters.md` regenerado con formato WotC
- [ ] Todos los characters tienen stats completos
- [ ] Ability scores en rango 8-18 (o justificadas por raza/class)
- [ ] HP ≥ 1 para todos los characters
- [ ] AC ≥ 1 para todos los characters

---

### Task 3.2: Validate Character Stats
**Type**: Validation  
**File**: `campaigns/{campaign-name}/characters.md`  
**Command**: Validación manual o script custom

**Implementation:**
1. Parsear `characters.md` para extraer stats de cada character
2. Para cada character, validar:
   - STR, DEX, CON, INT, WIS, CHA: rango 8-18 (allow racial bonuses >18)
   - HP ≥ 1
   - AC ≥ 1
3. Reportar cualquier stat fuera de rango
4. Fixear stats inválidos

**Acceptance:**
- [ ] 100% de ability scores en rango 8-18 (o justificadas)
- [ ] 100% de HP ≥ 1
- [ ] 100% de AC ≥ 1
- [ ] Reporte de validación documentado

---

## Phase 4: Validation

### Task 4.1: Execute grimorio-narrative-custodian
**Type**: Validation  
**Command**: `grimorio validate --all` o `grimorio-narrative-custodian`  
**Files**: Todos los archivos de la campaña

**Implementation:**
1. Ejecutar validación completa de la campaña:
   ```bash
   grimorio validate --all
   ```
2. Revisar output para errores o warnings
3. Si hay errores críticos:
   - Documentar cada error
   - Fixear según sugerencias del custodian
   - Re-ejecutar validación
4. Si no hay errores, documentar pass

**Acceptance:**
- [ ] Validación ejecutada exitosamente
- [ ] 0 critical errors
- [ ] Warnings documentados (si los hay)
- [ ] Todos los fixes aplicados

---

### Task 4.2: Re-compile PDF
**Type**: Artifact generation  
**Command**: `grimorio compile-pdf`  
**Output**: `campaigns/{campaign-name}/{campaign-name}.pdf`

**Implementation:**
1. Ejecutar compilación de PDF:
   ```bash
   grimorio compile-pdf {campaign-name}
   ```
2. Verificar que el PDF se genera sin errores
3. Validar que el PDF contiene:
   - Introduction completa
   - Todos los quests
   - Todos los characters
   - Áreas/encuentros
4. Verificar formatting y rendering

**Acceptance:**
- [ ] PDF generado exitosamente
- [ ] Sin errores de compilación
- [ ] Todas las secciones presentes
- [ ] Formatting correcto

---

## Phase 5: Documentation

### Task 5.1: Update README.md to v2.6.0
**Type**: Documentation  
**File**: `README.md`  
**Lines**: Update version references

**Implementation:**
1. Actualizar versión en README.md:
   - Cambiar referencias de v2.5.x a v2.6.0
   - Actualizar badge de versión (si existe)
   - Actualizar sección de features si hay cambios relevantes
2. Verificar que todos los links y referencias sean válidos
3. Commitear cambios con mensaje apropiado

**Acceptance:**
- [ ] README.md actualizado a v2.6.0
- [ ] Versión consistente en todo el documento
- [ ] Links válidos
- [ ] Changes commiteados

---

### Task 5.2: Add CHANGELOG.md entry for v2.6.0
**Type**: Documentation  
**File**: `CHANGELOG.md`  
**Lines**: Add new entry at top

**Implementation:**
1. Agregar nueva entrada en `CHANGELOG.md`:
   ```markdown
   ## [2.6.0] - 2026-05-09

   ### Fixed
   - WotC format quality improvements
   - Introduction word count validation (min 5,000 words)
   - Quest objectives and rewards validation (min 3 objectives, 1 reward)
   - Character stats validation (ability scores 8-18, HP≥1, AC≥1)

   ### Changed
   - Regenerated introduction.md with full WotC structure
   - Regenerated quests.md and quest_*.json with complete data
   - Regenerated characters.md with full stats
   - PDF compilation pipeline updated
   ```
2. Actualizar enlace de comparación de versiones (si aplica)
3. Verificar formato consistente con entradas anteriores

**Acceptance:**
- [ ] Entry v2.6.0 agregada en CHANGELOG.md
- [ ] Formato consistente con entradas anteriores
- [ ] Fecha correcta (2026-05-09)
- [ ] Cambios documentados apropiadamente

---

## Task Dependencies

```
Phase 1: TASK-001 → TASK-002
Phase 2: TASK-003 → TASK-004 → TASK-005
Phase 3: TASK-006 → TASK-007
Phase 4: TASK-008 → TASK-009
Phase 5: TASK-010 → TASK-011

Cross-phase:
TASK-002 (Intro valid) → No deps
TASK-005 (Quests valid) → No deps
TASK-007 (Chars valid) → No deps
TASK-008 (Custodian) → TASK-002, TASK-005, TASK-007 (all validations must pass first)
TASK-009 (PDF) → TASK-008 (custodian must pass)
TASK-010, TASK-011 (Docs) → TASK-009 (PDF must compile)
```

**Dependency Rationale:**
- Cada phase tiene validación interna que debe pasar antes de continuar
- Phase 4 (Validation) requiere que Phases 1-3 estén completas y validadas
- Phase 5 (Documentation) requiere que todo esté validado y PDF compilado
- TASK-010 y TASK-011 pueden hacerse en paralelo después de TASK-009

---

## Estimated Effort

| Phase | Tasks | Estimated Lines | Sessions |
|-------|-------|-----------------|----------|
| Phase 1 (Introduction) | 1.1, 1.2 | ~5,000 words | 1 session |
| Phase 2 (Quests) | 2.1, 2.2, 2.3 | ~200 lines JSON + MD | 1 session |
| Phase 3 (Characters) | 3.1, 3.2 | ~150 lines | 1 session |
| Phase 4 (Validation) | 4.1, 4.2 | ~0 lines (execution) | 0.5 session |
| Phase 5 (Docs) | 5.1, 5.2 | ~50 lines | 0.5 session |
| **Total** | **11 tasks** | **~5,400 lines/words** | **4 sessions** |

---

## Implementation Notes

### Word Count Strategy (TASK-001)
- Target: 5,000-6,000 words para margen de seguridad
- Expandir secciones existentes antes de agregar nuevas
- Usar ejemplos concretos para aumentar contenido útil

### JSON Validation (TASK-004, TASK-005)
- Usar `jq` para validar syntax: `jq . quest_*.json`
- Script custom para validar estructura (objectives ≥ 3, rewards ≥ 1)

### Character Stats (TASK-006, TASK-007)
- Ability scores: 8-18 base, +2 racial bonus permitido (max 20)
- HP: mínimo 1 + CON modifier por level
- AC: mínimo 10 + DEX modifier (sin armor)

### PDF Compilation (TASK-009)
- Requiere todas las validaciones previas passing
- Verificar que markdown renderiza correctamente antes de compilar

### Rollback Plan
- Git revert para cambios de contenido
- PDF es artifact generado, no requiere rollback
- CHANGELOG/README: editar para revertir versión

---

## Review Workload Forecast

- **Estimated changed lines**: ~400-500 lines (excluyendo contenido generado)
- **Files modified**: ~15-20 files (introduction, quests, characters, JSONs, docs)
- **400-line budget risk**: Medium-High ⚠️
- **Chained PRs recommended**: Yes
- **Delivery strategy**: ask-on-risk
- **Decision needed before apply**: Yes
- **Suggested work-unit PR split**: 
  - PR 1: Phase 1 + Phase 2 (Introduction + Quests)
  - PR 2: Phase 3 + Phase 4 + Phase 5 (Characters + Validation + Docs)
