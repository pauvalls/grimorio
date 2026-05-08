# Tasks: Grimorio Narrative Quality Improvements

## Phase 1: Template Updates

### Task 1.1: Update areas.md.tmpl
**Type**: Template modification  
**File**: `internal/compiler/templates/areas.md.tmpl`  
**Lines**: Add after line 222 (after "Consecuencias de elecciones clave")

**Implementation:**
1. Add new section `### Puntos de Decisión` with table format containing:
   - Columnas: Decisión, Condición (IF), Consecuencia (THEN), Área/Acto Afectado
   - Al menos 3 filas de ejemplo por acto
2. Add new section `### Cambios de Estado del Mundo` con:
   - Lista de cambios permanentes (NPCs muertos, facciones alteradas, ubicaciones destruidas)
   - Referencias a `narrative_state.json` keys actualizadas
3. Incluir ejemplo con estructura IF-THEN explícita
4. Asegurar que el template total se mantenga bajo 300 líneas

**Acceptance:**
- [ ] Tabla de Puntos de Decisión presente con 4 columnas
- [ ] Sección de Cambios de Estado del Mundo presente
- [ ] Ejemplo IF-THEN provisto (ej: "IF PJs matan al NPC → ENTONCES facción se vuelve hostil")
- [ ] Template renderiza sin errores
- [ ] Template ≤ 300 líneas

---

### Task 1.2: Update npc.md.tmpl
**Type**: Template modification  
**File**: `internal/compiler/templates/npc.md.tmpl`  
**Lines**: Add after line 31 (after "Facción" en Conexiones)

**Implementation:**
1. Add new section `### Reputación de Facción` después de Conexiones
2. Add campo de score numérico (-100 a +100) con explicación de rangos:
   - -100 a -50: Hostil
   - -49 a -10: Desconfiado
   - -9 a +9: Neutral
   - +10 a +49: Amistoso
   - +50 a +100: Aliado
3. Add tabla de beneficios por tier (Rank 1/2/3) con:
   - Umbrales de reputación requeridos
   - Beneficios mecánicos concretos (descuentos, acceso, aliados)
4. Add reglas de propagación ally/enemy:
   - Si facción A es aliada de B y PJs ayudan a A → B gana +10 reputación
   - Si facción A es enemiga de B y PJs ayudan a A → B pierde -10 reputación
5. Incluir 3 ejemplos concretos de beneficios mecánicos:
   - Descuento del 10% en equipo (Rank 1, reputación +20)
   - Acceso a área restringida (Rank 2, reputación +50)
   - NPC aliado aparece en encuentro (Rank 3, reputación +80)

**Acceptance:**
- [ ] Campo de reputación numérica presente (-100 a +100)
- [ ] Tabla de 5 rangos con descripciones
- [ ] Tabla de beneficios con 3 tiers (Rank 1/2/3)
- [ ] Beneficios mecánicos especificados (descuentos, acceso, aliados)
- [ ] Reglas de propagación ally/enemy incluidas
- [ ] 3 ejemplos concretos provistos

---

### Task 1.3: Update introduction.md.tmpl
**Type**: Template modification  
**File**: `internal/compiler/templates/introduction.md.tmpl`  
**Lines**: Add after line 66 (after "Adventure Flowchart" section)

**Implementation:**
1. Add nueva sección `## Running the Adventure` después de Adventure Flowchart
2. Add subsección `### Manejo de Ramas Narrativas` con guía de 3 pasos:
   - Paso 1: Identificar puntos de decisión en cada acto
   - Paso 2: Trackear consecuencias en narrative_state.json
   - Paso 3: Propagar cambios a áreas/actos futuros
3. Add subsección `### Ejemplos de Ramificación` con 2 ejemplos:
   - Ejemplo 1: Consecuencia inmediata (IF PJs matan guardia → ENTONCES alarma se activa)
   - Ejemplo 2: Consecuencia diferida (IF PJs ignoran facción → ENTONCES facción pierde poder en Acto 3)
4. Add subsección `### Trackeo de Estado del Mundo` con checklist:
   - [ ] NPCs muertos registrados
   - [ ] Cambios de reputación de facciones
   - [ ] Objetos clave obtenidos
   - [ ] Pistas reveladas
   - [ ] Quests completadas/fallidas

**Acceptance:**
- [ ] Sección "Running the Adventure" presente
- [ ] 2 ejemplos de ramificación con estructura IF-THEN
- [ ] Checklist de trackeo de estado del mundo incluida
- [ ] Template ≤ 150 líneas

---

## Phase 2: Agent Updates

### Task 2.1: Update grimorio-areas.md
**Type**: Agent instruction modification  
**File**: `agents/grimorio-areas.md`  
**Lines**: Add to "REGLAS DE ORO" section (after line 278, before examples)

**Implementation:**
1. Add nueva Rule 9: "PUNTOS DE DECISIÓN OBLIGATORIOS" con:
   - Requerimiento de mínimo 3 puntos de decisión por acto
   - Cada punto debe tener consecuencia explícita documentada
   - Al menos 1 consecuencia debe propagar cross-área o cross-acto
2. Add ejemplo de formato IF-THEN:
   ```
   - **Decisión:** PJs deciden si matar o capturar al guardia
   - **IF** matan al guardia → **ENTONCES** alarma se activa, Área 5 se vuelve hostil
   - **IF** capturan al guardia → **ENTONCES** obtienen información, Área 5 permanece neutral
   ```
3. Add 4 nuevos ítems al "Checklist Pre-Guardado" (después del ítem 13):
   - ¿El acto tiene al menos 3 puntos de decisión?
   - ¿Cada punto de decisión tiene consecuencia explícita?
   - ¿Hay propagación cross-área documentada?
   - ¿Los cambios de estado del mundo están registrados?

**Acceptance:**
- [ ] Rule 9 presente con requerimiento de 3 puntos de decisión
- [ ] Ejemplo de formato IF-THEN incluido
- [ ] 4 nuevos ítems agregados al checklist
- [ ] Instrucciones claras y accionables

---

### Task 2.2: Update grimorio-narrative-custodian.md
**Type**: Agent instruction modification  
**File**: `agents/grimorio-narrative-custodian.md`  
**Lines**: Add to "Phase 2: Validate Content" (after Check 8, around line 89)

**Implementation:**
1. Add Check 9: **Decision Branch Completeness**
   - Verificar que cada decisión importante tenga al menos 2 ramas documentadas
   - Verificar que cada rama tenga consecuencia explícita
   - Si falta rama o consecuencia → ERROR con fix suggestion
2. Add Check 10: **Cross-Area Consequence Propagation**
   - Verificar que decisiones mayores (cambio de facción, muerte de NPC clave) propaguen a otras áreas/actos
   - Si no hay propagación documentada → ERROR con ejemplo de cómo propagar
3. Add Check 11: **World State Consistency**
   - Verificar que el estado del mundo sea consistente (NPCs muertos no aparecen vivos)
   - Verificar que reputaciones de facciones no sean contradictorias
   - Si hay inconsistencia → REJECT con fix específico
4. Add 4 nuevos Critical Issues a la lista de reject (after line 160):
   - Decision point without documented consequence
   - Cross-area propagation missing for major decision
   - World state inconsistency (dead NPC, reputation contradiction)
   - Faction benefit granted without meeting reputation threshold

**Acceptance:**
- [ ] Checks 9, 10, 11 presentes con lógica de validación
- [ ] 4 nuevos critical issues en lista de reject
- [ ] Reglas de validación específicas y accionables

---

## Phase 3: Install Script Update

### Task 3.1: Verify install.sh propagates changes
**Type**: Script verification  
**File**: `install.sh`  
**Lines**: Review lines 198-201 (agent file copying)

**Implementation:**
1. Verificar que install.sh copia archivos de agentes actualizados desde `agents/` al directorio del plugin
2. Verificar que no haya prompts hardcoded que override los agentes basados en archivos
3. Revisar líneas 198-201: confirmar que el loop `for agent_file in "$INSTALL_DIR/agents"/grimorio-*.md` incluye todos los agentes
4. (Opcional pero recomendado) Testear instalación en directorio fresco

**Acceptance:**
- [ ] install.sh copia grimorio-areas.md correctamente
- [ ] install.sh copia grimorio-narrative-custodian.md correctamente
- [ ] No hay prompts hardcoded que conflictúen con agentes basados en archivos
- [ ] Script no requiere modificaciones

---

## Phase 4: Testing & Documentation

### Task 4.1: Generate test campaign
**Type**: Integration testing  
**File**: Test campaign in `/tmp/grimorio-test/`

**Implementation:**
1. Generar campaña de test usando templates actualizados:
   ```bash
   cd /tmp && mkdir grimorio-test && cd grimorio-test
   grimorio init test-campaign
   grimorio generate-all
   ```
2. Verificar que las áreas tengan secciones de "Puntos de Decisión"
3. Verificar que los NPCs tengan scores de reputación y tablas de tier benefits
4. Verificar que introduction tenga sección "Running the Adventure"
5. Verificar que haya al menos 3 puntos de decisión por acto

**Acceptance:**
- [ ] Campaña de test generada exitosamente
- [ ] Puntos de Decisión presentes en todos los actos
- [ ] Tablas de reputación de facciones presentes en NPCs
- [ ] Running Guide presente en introduction
- [ ] Sin errores de renderizado de templates

---

### Task 4.2: Validate decision tree propagation
**Type**: Validation testing  
**File**: Test campaign analysis

**Implementation:**
1. Checkear que todos los puntos de decisión tengan estructura IF-THEN
2. Checkear que las consecuencias propaguen a áreas/actos futuros
3. Checkear que los cambios de estado del mundo estén documentados
4. Correr validación de narrative-custodian en la campaña de test:
   ```bash
   grimorio validate --all
   ```
5. Verificar que no haya critical issues relacionadas con decision branches

**Acceptance:**
- [ ] 100% de puntos de decisión tienen estructura IF-THEN
- [ ] Consecuencias cross-área documentadas
- [ ] Cambios de estado del mundo trackeados
- [ ] Validación de Narrative Custodian pasa sin critical issues

---

### Task 4.3: Update README (optional)
**Type**: Documentation  
**File**: `README.md`

**Implementation:**
1. Add sección "Narrative Quality Improvements" a la lista de features
2. Mencionar soporte de decision tree con estructura IF-THEN
3. Mencionar mecánicas de reputación de facciones con beneficios por tier
4. Mencionar trackeo de estado del mundo con narrative_state.json

**Acceptance:**
- [ ] README actualizado con nuevas features (opcional, puede diferirse)

---

## Task Dependencies

```
1.1 → 1.2 → 1.3 → 2.1 → 2.2 → 3.1 → 4.1 → 4.2
                                 ↘ 4.3 (optional)
```

**Dependency Rationale:**
- Templates (1.1-1.3) deben estar completos antes de actualizar agentes (2.1-2.2) porque los agentes referencian los templates
- Agentes deben estar actualizados antes de verificar install.sh (3.1) porque el script copia los agentes
- Install script debe estar verificado antes de testing (4.1-4.2) para asegurar que los cambios se propaguen
- README (4.3) es opcional y puede hacerse en paralelo o diferirse

---

## Estimated Effort

| Phase | Tasks | Estimated Lines | Sessions |
|-------|-------|-----------------|----------|
| Phase 1 (Templates) | 1.1, 1.2, 1.3 | ~120 lines | 1 session |
| Phase 2 (Agents) | 2.1, 2.2 | ~80 lines | 1 session |
| Phase 3 (Install) | 3.1 | ~0 lines (verification) | 0.5 session |
| Phase 4 (Testing) | 4.1, 4.2, [4.3] | ~50 lines (test docs) | 0.5 session |
| **Total** | **8 tasks** | **~250 lines** | **3 sessions** |

---

## Review Workload Forecast

- **Estimated changed lines**: ~250 lines
- **Files modified**: 5 files (3 templates, 2 agents, 1 script verification)
- **400-line budget risk**: Low ✅
- **Chained PRs recommended**: No (single PR is appropriate)
- **Decision needed before apply**: No

---

## Implementation Notes

### Template Line Budgets
- `areas.md.tmpl`: Current 235 lines → Target ≤ 300 lines (65 lines available)
- `npc.md.tmpl`: Current 103 lines → Target ≤ 180 lines (77 lines available)
- `introduction.md.tmpl`: Current 87 lines → Target ≤ 150 lines (63 lines available)

### Agent Line Budgets
- `grimorio-areas.md`: Current 280 lines → Adding ~40 lines for Rule 9 + checklist
- `grimorio-narrative-custodian.md`: Current 243 lines → Adding ~50 lines for Checks 9-11 + critical issues

### Testing Strategy
1. Generate minimal test campaign (1-shot, 1 act, 8-10 areas)
2. Manually verify new sections appear in generated markdown
3. Run `grimorio validate --all` to check narrative custodian rules
4. Verify no critical issues related to decision branches or reputation

### Rollback Plan
If issues are found post-implementation:
1. Templates: Revert to previous version from git
2. Agents: Revert to previous version from git
3. No database migrations or breaking changes → safe rollback
