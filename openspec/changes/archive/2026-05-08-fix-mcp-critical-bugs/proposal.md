# Proposal: Fix MCP Critical Bugs

## Intent

Arreglar 7 bugs críticos del servidor MCP descubiertos durante la generación de la campaña "la-llorona". Incluyen validaciones contradictorias, concurrencia insegura, timeouts faltantes, herramientas incompletas y degradación de performance.

## Scope

### In Scope
- Handler `update_narrative_state`: permitir `session_num=0` para auto-incremento
- `AssetService`: reemplazar mutex global con rate limiter por campaña
- Repositories filesystem: agregar `sync.RWMutex` a lecturas/escrituras
- `Compiler.htmlToPDF`: agregar `context.WithTimeout` de 60s
- Nuevo handler `grimorio_save_characters` + servicio Save()
- `SessionPrepService`: crear estado inicial si no existe (como hace Update)
- `FlowchartService.buildNodes`: reducir complejidad algorítmica

### Out of Scope
- Refactor mayor de arquitectura
- Cambios en lógica de negocio de generación de contenido
- Nuevas capabilities de MCP fuera de personajes

## Capabilities

### New Capabilities
- `character-save`: Guardar personajes manuales via MCP (`grimorio_save_characters`)

### Modified Capabilities
- `narrative-state-update`: Cambiar validación de `session_num` para soportar 0
- `session-prep`: Auto-crear estado inicial cuando no existe
- `asset-generation`: Rate limiting por campaña en lugar de global
- `campaign-compilation`: Timeout de 60s en generación de PDF
- `flowchart-generation`: Optimización de buildNodes O(n²) → O(n)

## Approach

Fixes incrementales con tests:
1. **Handlers**: Relajar validación en `canon.go:146` para `session_num >= 0`
2. **Services**: Reemplazar `sync.Mutex` global con `map[string]*sync.Mutex` por campaña en `AssetService`
3. **Repositories**: Agregar `mu sync.RWMutex` a todos los filesystem repos, lock en Save/Load/Delete/List
4. **Compiler**: Wrap `exec.Command` con `context.WithTimeout(60s)`
5. **Characters**: Agregar `SaveCharacter()` en `CharacterService` + handler MCP
6. **SessionPrep**: Mirror del comportamiento de `NarrativeStateService.Update()` - crear estado inicial
7. **Flowchart**: Reestructurar `buildNodes` para evitar loops anidados sobre relationships

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/mcp/handlers/canon.go` | Modified | Validación session_num |
| `internal/mcp/handlers/character.go` | Modified | Nuevo handler save_characters |
| `internal/services/asset_service.go` | Modified | Rate limiter por campaña |
| `internal/services/session_prep.go` | Modified | Auto-crear estado inicial |
| `internal/services/flowchart.go` | Modified | Optimizar buildNodes |
| `internal/compiler/compiler.go` | Modified | Timeout wkhtmltopdf |
| `internal/repository/filesystem_*.go` | Modified | Agregar sync.RWMutex |
| `internal/services/character_service.go` | Modified | Método SaveCharacter |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Cambio en validación afecta clients existentes | Low | `session_num=0` es comportamiento documentado; clients con valores >0 no se afectan |
| Mutex en repos introduce deadlock | Low | Usar `defer mu.Unlock()`, patrones estándar Go |
| Timeout de 60s insuficiente para PDFs grandes | Low | Monitorear; ajustar a 120s si es necesario |
| Race conditions en tests existentes | Med | Correr con `-race` en CI |

## Rollback Plan

Cada fix es independiente. Si algo falla:
1. Revertir commit individual del fix problemático
2. Los repos filesystem con mutex son backward-compatible (sin cambio de API)
3. El handler de personajes es aditivo (no afecta funcionalidad existente)

## Dependencies

- Ninguna externa. Todos los cambios son internos del proyecto.

## Success Criteria

- [ ] `update_narrative_state` con `session_num=0` auto-incrementa correctamente
- [ ] 5 llamadas concurrentes a `generate_image` terminan en <10s (vs 15s+ actuales)
- [ ] Tests con `-race` pasan sin errores en repository layer
- [ ] `compile_pdf` termina con error controlado si `wkhtmltopdf` tarda >60s
- [ ] `grimorio_save_characters` persiste personajes manuales correctamente
- [ ] `generate_session_prep` funciona en campaña sin estado previo
- [ ] Coverage mantiene o mejora: mcp/handlers >61.7%, services >82.4%, repository >59.8%
