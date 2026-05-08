# Design: Migrate Grimorio v1.x to v2.0 — Area-Based Campaign Generation

## Technical Approach

Reemplazar el pipeline de generación de "escenas narrativas" (v1) por "áreas numeradas jugables" (v2) siguiendo el formato de módulos profesionales WotC (WDH, SKT). La arquitectura mantiene el servidor MCP existente; los cambios se concentran en:

1. **Agentes**: `grimorio-acts` → `grimorio-areas` (ya renombrado en agente actual); `grimorio-integrator` se fortalece con validaciones programáticas.
2. **Templates**: Todos los `.tmpl` se actualizan al formato técnico WotC (Read-Aloud → Features → Mechanics → Treasure → Connections → Secrets).
3. **Compiler v2**: TOC jerárquico (3 niveles), cross-references clickeables, stat blocks inline/dual, layout de 2 columnas, páginas de handouts.
4. **Handouts**: Nuevo servicio `HandoutGenerator` que produce mapas de jugador, pistas descubiertas y referencia rápida de NPCs.

## Architecture Decisions

| Decision | Choice | Alternatives | Rationale |
|----------|--------|--------------|-----------|
| Pipeline de generación | 5 fases secuenciales con gates de integración | Pipeline paralelo | El LLM necesita contexto acumulado; la integración secuencial detecta errores antes de propagarse |
| Áreas vs Escenas | Áreas numeradas (A1-A15) por acto | Mantener escenas narrativas | WotC usa áreas; reduce improvisación del DM del 70% al ~20% |
| Stat blocks | Inline para custom; referencia compacta para MM | Siempre inline | Reduce tamaño de PDF; MM es estándar en mesa |
| Compiler markup | HTML/CSS con wkhtmltopdf | WeasyPrint / Puppeteer | wkhtmltopdf ya está en dependencias; `--enable-local-file-access` soporta SVGs locales |
| Retrocompatibilidad | Flag `--compiler-version=1` + `grimorio-acts.md` legacy | Migración forzosa | Campañas v1.x existentes no deben romperse |
| Cross-references | Parseo regex en compilación + validación en integrador | AST markdown completo | Regex es suficiente para el dominio (nombres en negrita, "Área N"); AST sería overkill |

## Data Flow

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  grimorio-npc   │     │ grimorio-bestiary│     │grimorio-encounters│
│   (v2 delta)    │     │   (v2 delta)    │     │    (v2 delta)    │
└────────┬────────┘     └────────┬────────┘     └────────┬────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 ▼
                    ┌─────────────────────┐
                    │   lore.md (base)    │
                    └──────────┬──────────┘
                               ▼
                    ┌─────────────────────┐
                    │  grimorio-areas     │  ← NUEVO (reemplaza acts)
                    │  10-15 áreas/acto   │
                    └──────────┬──────────┘
                               ▼
                    ┌─────────────────────┐
                    │ grimorio-integrator │  ← FORTALECE (validación prog.)
                    │ Cross-ref, XP, DCs  │
                    └──────────┬──────────┘
                               ▼
                    ┌─────────────────────┐
                    │ grimorio-compiler   │  ← v2 (TOC, links, stat blocks)
                    │   + HandoutGenerator│
                    └──────────┬──────────┘
                               ▼
                         campaign.pdf
```

## Integration Points

| Component | Inputs | Outputs | Integration |
|-----------|--------|---------|-------------|
| `grimorio-areas` | lore.md, npcs.md, bestiary.md, encounters.md, maps.md | `acts/act_NN.md` (formato áreas) | Llama `save_act` y `validate_canon` |
| `grimorio-integrator` | Todos los `.md` de la campaña | Reporte de validación + correcciones | Usa `check_consistency` y `process_consistency_gate`; puede llamar `save_act` para auto-fix |
| `Compiler v2` | Todo el directorio de campaña | `campaign.html` → `campaign.pdf` | Nuevas secciones: `handouts/`, parseo de cross-references vía regex, TOC jerárquico con `<a href="#id">` |
| `HandoutGenerator` | `narrative_state.json`, `maps/`, `npcs/` | `handouts/player_map_*.svg`, `handouts/clues.md`, `handouts/npc_reference.md` | Inyectado en compiler como paso previo a la generación de HTML |

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `agents/grimorio-acts.md` | Delete | Reemplazado por `grimorio-areas.md` (ya existe con formato v2) |
| `agents/grimorio-areas.md` | Modify | Refinar prompts para forzar 150-200 palabras/área, DCs numéricos, conexiones bidireccionales |
| `agents/grimorio-integrator.md` | Modify | Agregar fase programática de validación (no solo manual) |
| `agents/grimorio-npc.md` | Modify | Agregar location (Área X), combat stats, quest involvement, secrets al formato |
| `agents/grimorio-bestiary.md` | Modify | Agregar role classification, encounter groups, source reference, tactics estructuradas |
| `agents/grimorio-encounters.md` | Modify | Agregar encounter templates (referencia por nombre), tactical map, conditions, alternative resolution, round-by-round |
| `internal/compiler/compiler.go` | Modify | TOC jerárquico 3 niveles; cross-reference links; inline stat blocks; 2-col layout; handout pages |
| `internal/compiler/templates/dnd-style.css` | Modify | Area number highlighting, stat block borders, read-aloud boxed style, handout page styles |
| `internal/compiler/templates/act.md.tmpl` | Modify | WotC format: Read-Aloud → Features → Mechanics → Treasure → Connections → Secrets |
| `internal/compiler/templates/encounter.md.tmpl` | Modify | Round-by-round, adjusted XP, encounter reference, tactical map, conditions |
| `internal/compiler/templates/npc.md.tmpl` | Modify | Alignment, location, stats, quest, secret fields |
| `internal/compiler/templates/monster.md.tmpl` | Modify | Role, encounter groups, source, structured tactics (Opening/Priorities/Retreat/Synergy) |
| `internal/services/handout.go` | Create | HandoutGenerator: player map redaction, clue list, NPC quick-reference |
| `internal/compiler/handouts.go` | Create | Integración de handouts en el pipeline de compilación |
| `cmd/grimorio/main.go` | Modify | Flag `--compiler-version={1\|2}` para retrocompatibilidad |
| `cmd/migrate-v1-to-v2/main.go` | Modify | Extender migrador para convertir `acts/` (escenas) → `acts/` (áreas numeradas) con best-effort |
| `internal/validators/area.go` | Create | Validador programático: cuenta de áreas, word count, DCs numéricos, conexiones bidireccionales |

## Risk Analysis

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| LLM genera "escenas" en vez de "áreas" | High | Prompt engineering con ejemplos WotC; validador `internal/validators/area.go` rechaza output sin áreas numeradas; retry loop en integrator |
| v1.x campaigns break | Medium | Flag `--compiler-version=1`; `grimorio-acts.md` legacy mode; migrador v1→v2 |
| PDF size increases >50% | Medium | MM references by default; inline stats solo para custom; compresión de imágenes base64 |
| Integrator becomes bottleneck | Medium | Caché de resultados de validación; `fast_mode` en consistency gate; override manual vía `DMOverrides` |
| Cross-reference links rotos en PDF | Low | Regex con lookahead para "Área N" y "**Nombre**"; fallback a texto plano si no se resuelve |

## Migration / Rollout

1. **Fase 1** (Foundation): lore, NPCs v2, bestiary v2, encounters v2 — verificar con `make test`
2. **Fase 2** (Areas): `grimorio-areas` genera áreas numeradas; validador programaático corre en CI
3. **Fase 3** (Integration): `grimorio-integrator` ejecuta cross-reference, balance, bidirectional checks
4. **Fase 4** (Visuals): Mapas, dividers, handouts
5. **Fase 5** (Compilation): Compiler v2 con TOC jerárquico y handout pages

**Rollback**: flag `--compiler-version=1` + restaurar `agents/grimorio-acts.md` desde git.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `internal/validators/area.go` — count areas, check DCs, bidirectional connections | Tablas de casos en `area_test.go` |
| Unit | `internal/compiler/` — TOC generation, cross-reference regex, stat block embedding | `compiler_test.go` con fixtures markdown |
| Unit | `internal/services/handout.go` — player map redaction, clue filtering | `handout_test.go` con mock campaign dirs |
| Integration | Pipeline completo: generar campaña de test → compilar PDF → assert estructura HTML | `TestCompileV2` en `compiler/` |
| Integration | Validador + integrator: campaña con referencias rotas → detectar y reportar | `TestIntegrationValidation` en `mcp/handlers/` |

## Go-Specific Considerations

- **Concurrency**: El compilador procesa archivos secuencialmente (I/O bound); no se requiere paralelismo. El integrator puede validar áreas en paralelo con `errgroup` si el volumen crece.
- **Interfaces**: `Compiler` ya tiene interfaz implícita; `HandoutGenerator` será una interfaz `HandoutRenderer` inyectada en `Compiler`.
- **Testing**: Usar `httptest` solo si se testea MCP handlers; para el compiler, usar `t.TempDir()` con fixtures markdown embebidos en `testdata/`.
- **Regex**: Compilar expresiones regulares de cross-reference una sola vez como `var` globales (ya existe el patrón en `compiler.go`).
- **Embed**: Nuevos templates de handout usarán `//go:embed` igual que los templates actuales.

## Open Questions

- [ ] ¿Se mantiene `acts/` como nombre de directorio o se migra a `areas/`? (Impacta en paths del compiler y migrador)
- [ ] ¿El handout de mapa de jugador requiere procesamiento SVG (eliminar capas) o es suficiente con un flag en el markdown del mapa?
