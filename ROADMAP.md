# Grimorio MCP — Roadmap de Mejoras

> **Versión:** 2.0.0  
> **Fecha:** Mayo 2026  
> **Enfoque:** TDD-first, arquitectura limpia, evolución orgánica de campañas

---

## 📋 Estado Actual (v2.0.0)

El MCP actual tiene **17 herramientas** organizadas en 4 categorías:

| Categoría | Tools |
|-----------|-------|
| **Estructura** | `create_campaign`, `save_act`, `save_npcs`, `save_bestiary`, `save_encounters`, `save_maps` |
| **Assets** | `generate_map` (SVG), `generate_divider` (SVG), `generate_image`, `generate_images_batch` |
| **Output** | `compile_pdf`, `get_template` |
| **Coherencia Narrativa** | `generate_adventure_bible`, `validate_canon`, `update_narrative_state`, `check_consistency`, `process_consistency_gate` |

### Cambios Recientes Completados (v2.0.0)

- ✅ **Subsistema de Coherencia Narrativa** — Canon, validación, estado, y gates
- ✅ **Motor de Validación** — 10 reglas (NPC deaths, lore, entities, timeline, quests, etc.)
- ✅ **Consistency Gate** — Batch validation con approve/reject/retry
- ✅ **Arquitectura limpia** — Domain/services/repository separados
- ✅ **Tests** — 82.6% coverage en servicios, strict TDD
- ✅ **PDF compilation reordenado** — Lore → Acts → Apéndices
- ✅ **Template `act.md.tmpl` rediseñado** — Estilo Out of the Abyss
- ✅ **README actualizado** — Documentación completa EN+ES

---

## ✅ FASE 0: Fundamentos — Testing & Arquitectura

> **Estado:** COMPLETADA  
> **Entregable:** Base sólida con TDD, arquitectura limpia, y 80%+ coverage

### 0.1 Testing Infrastructure ✅

- [x] **Suite de testing con mocks**
  - Builders para `mcp.CallToolRequest`
  - Asserts custom para respuestas MCP
  - Fixtures de campañas de ejemplo
- [x] **Test helpers reutilizables**
  - `NewTestServer()` — servidor MCP aislado
  - `MakeRequest(tool, args)` — helper para invocar tools
- [x] **Coverage tracking**
  - Objetivo: 80% mínimo en servicios ✅ (82.6% alcanzado)

### 0.2 Refactor a Arquitectura Limpia ✅

```
internal/
├── mcp/
│   ├── server.go              ← Solo wiring de tools
│   └── handlers/              ← Thin adapters
│       ├── campaign.go
│       ├── canon.go           ← Narrative coherence handlers
│       └── ...
├── domain/                    ← Entidades de negocio ✅
│   ├── canon.go
│   ├── narrative_state.go
│   ├── validation.go
│   └── gate.go
├── services/                  ← Lógica de negocio ✅
│   ├── canon_service.go
│   ├── narrative_state_service.go
│   ├── validation_engine.go
│   └── consistency_gate.go
├── repository/                ← Persistencia ✅
│   ├── filesystem_canon.go
│   └── memory_canon.go
└── ...
```

### 0.3 Validación & Error Handling ✅

- [x] **Esquemas de validación por tool**
  - Validar tipos (string, number, bool)
  - Validar rangos (rooms: 2-10, level: 1-20)
- [x] **Error handling consistente**
  - Mensajes claros para el LLM
  - Diferenciar user error vs system error

---

## ✅ FASE 1: Coherencia Narrativa — Canon & Validación

> **Estado:** COMPLETADA  
> **Entregable:** Sistema de canon, validación, y tracking de estado

### 1.1 Domain Models ✅

- [x] `CanonDocument` — Facts, entities, rules, timeline
- [x] `NarrativeState` — Quests, clues, deaths, decisions
- [x] `ValidationResult` — Violations con severity y fix suggestions
- [x] `GateResult` — Batch validation con approve/reject/retry

### 1.2 Servicios Implementados ✅

- [x] `CanonService` — Initialize, load, save, validate canon
- [x] `NarrativeStateService` — Load, save, update state
- [x] `ValidationEngine` — 10 reglas de validación:
  - `npc_death_state` — NPCs muertos no pueden aparecer vivos
  - `entity_existence` — Entidades referenciadas deben existir
  - `world_rule_violation` — No violar reglas del mundo
  - `timeline_order` — Eventos en orden cronológico
  - `quest_reward_existence` — Recompensas deben existir
  - `level_encounter_balance` — Encuentros balanceados por nivel
  - `location_existence` — Localizaciones deben existir
  - `timeline_consistency` — Consistencia temporal
  - `prerequisite_clue_check` — Pistas requeridas reveladas
  - `faction_reputation_gate` — Reputación con facciones (placeholder)
- [x] `ConsistencyGateService` — ProcessBatch, GetGateStatus, ResetGate

### 1.3 MCP Tools ✅

- [x] `generate_adventure_bible` — Crea canon.json
- [x] `validate_canon` — Valida propuestas individuales
- [x] `update_narrative_state` — Actualiza estado post-sesión
- [x] `check_consistency` — Validación completa de campaña
- [x] `process_consistency_gate` — Gate de validación por lotes

### 1.4 Repositories ✅

- [x] Dual repository pattern (filesystem + in-memory)
- [x] `FilesystemCanonRepository`
- [x] `MemoryCanonRepository`
- [x] `FilesystemNarrativeStateRepository`
- [x] `MemoryNarrativeStateRepository`

### 1.5 Migración ✅

- [x] `migrate-v1-to-v2` — Convierte campañas existentes
- [x] Crea canon.json y narrative_state.json
- [x] Backup automático en `.v1-backup/`

---

## 🎯 FASE 2: Personajes & Fichas (PCs)

> **Duración estimada:** 2 sprints  
> **Objetivo:** Sistema completo de fichas de personaje jugador  
> **Entregable:** Crear, leer, actualizar personajes con stats, inventario, y relaciones

### 2.1 Modelo de Personaje

```go
type Character struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    CampaignID  string            `json:"campaign_id"`
    Race        string            `json:"race"`
    Class       string            `json:"class"`
    Level       int               `json:"level"`
    Background  string            `json:"background"`
    Alignment   string            `json:"alignment"`
    Stats       Stats             `json:"stats"`
    HP          HP                `json:"hp"`
    AC          int               `json:"ac"`
    Proficiency int               `json:"proficiency_bonus"`
    Skills      map[string]bool   `json:"skills"`
    Inventory   []Item            `json:"inventory"`
    Features    []Feature         `json:"features"`
    Personality Personality       `json:"personality"`
    Relationships []Relationship  `json:"relationships"`
    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
}
```

### 2.2 Tools Nuevas

**`generate_character`**
```json
{
  "campaign": "string (required, kebab-case)",
  "name": "string (required)",
  "race": "string (humano, elfo, enano, mediano, semielfo, semiorco, gnomo, tiefling, draconido)",
  "class": "string (guerrero, mago, clérigo, pícaro, paladín, bárbaro, ranger, brujo, druida, monje, bardo, hechicero)",
  "level": "number (1-20, default: 1)",
  "background": "string (soldado, acólito, criminal, sabio, noble, artesano, marinero, ermitaño, charlatán, héroe del pueblo)",
  "alignment": "string (LG, NG, CG, LN, N, CN, LE, NE, CE)"
}
```

**`save_character_sheet`**
- Guarda ficha en formato dual: JSON (estructurado) + Markdown (visual)
- Persistencia en `campaigns/{name}/characters/{character_name}.json`
- Template visual para visualización

**`get_character`**
- Obtener ficha completa por nombre
- Soporte para formato JSON o Markdown

**`list_characters`**
- Listar todos los personajes de una campaña
- Filtros: por clase, nivel, facción, estado (vivo, muerto, MIA)

**`update_character`**
- Actualizar stats, HP, inventario, nivel
- Tracking de cambios (audit log)

### 2.3 Templates de Ficha

- [ ] Template Markdown para visualización
- [ ] Secciones: Stats, Skills, Inventory, Features, Backstory
- [ ] Estilo acorde al template general de campaña

### 2.4 Integración con NPCs

- [ ] **Relaciones PC-NPC**: Vincular personajes con NPCs existentes
- [ ] **Facciones**: A qué facción pertenece cada personaje
- [ ] **Sistema de reputación**: Cómo los NPCs ven a cada PC

---

## 🎬 FASE 3: Integración Narrativa Avanzada

> **Duración estimada:** 2 sprints  
> **Objetivo:** Personajes vivos dentro de la narrativa, misiones personales  
> **Entregable:** Quests personales, tracking de estado, hooks narrativos

### 3.1 Sistema de Quests

```go
type Quest struct {
    ID            string       `json:"id"`
    CampaignID    string       `json:"campaign_id"`
    Title         string       `json:"title"`
    Type          QuestType    `json:"type"`           // personal, principal, secundaria
    Status        QuestStatus  `json:"status"`         // active, completed, failed, on_hold
    Hook          string       `json:"hook"`           // Cómo se introduce
    Stakes        string       `json:"stakes"`         // Qué se juega
    Reward        Reward       `json:"reward"`         // Recompensa
    CharacterID   *string      `json:"character_id"`   // Null si es quest grupal
    RelatedNPCs   []string     `json:"related_npcs"`   // NPCs involucrados
    RelatedActs   []string     `json:"related_acts"`   // Actos donde aparece
    ProgressNotes []Note       `json:"progress_notes"` // Notas de sesión
    CreatedAt     time.Time    `json:"created_at"`
    UpdatedAt     time.Time    `json:"updated_at"`
}
```

### 3.2 Tools Nuevas

**`create_personal_quest`**
```json
{
  "campaign": "string (required)",
  "character_name": "string (required)",
  "quest_title": "string (required)",
  "quest_type": "string (redención, venganza, descubrimiento, protección, redención, ascensión, supervivencia)",
  "hook": "string (cómo se introduce en la trama)",
  "stakes": "string (qué se juega el personaje)",
  "reward": "string (recompensa personal: conocimiento, reconciliación, poder, paz, etc.)"
}
```

**`create_group_quest`**
- Similar a personal quest pero sin `character_name`
- Stakes grupales, recompensas compartidas

**`update_quest_status`**
```json
{
  "campaign": "string (required)",
  "quest_id": "string (required)",
  "status": "string (active, completed, failed, on_hold)",
  "notes": "string (notas de progreso)"
}
```

**`list_quests`**
- Listar todas las quests de una campaña
- Filtros: por status, tipo, personaje asociado

### 3.3 Weaving Narrativo

**`weave_character_arc`**
- Lee el estado actual de la campaña (todos los acts, quests, relaciones)
- Identifica oportunidades para insertar desarrollo de personajes
- Sugiere puntos de inflexión narrativos
- Genera "session hooks" para el DM

**`generate_session_hooks`**
- Basado en estado actual, genera 3-5 hooks para la próxima sesión
- Considera: quests activas, relaciones tensas, threads sin resolver

### 3.4 Track de Relaciones

```go
type Relationship struct {
    From      string    `json:"from"`       // ID del personaje/NPC
    To        string    `json:"to"`         // ID del personaje/NPC
    Type      string    `json:"type"`       // ally, enemy, neutral, complicated, mentor, student
    Strength  int       `json:"strength"`   // -10 a +10
    History   []Event   `json:"history"`    // Eventos que definieron la relación
    UpdatedAt time.Time `json:"updated_at"`
}
```

**`update_relationship`**
- Modificar relación entre dos entidades
- Registrar evento que causó el cambio

---

## 📖 FASE 4: Evolución de Campaña

> **Duración estimada:** 2 sprints  
> **Objetivo:** Campañas que crecen orgánicamente, no solo acumulan acts  
> **Entregable:** Lectura de estado, generación de arcos, continuidad narrativa

### 4.1 Lectura de Estado

**`read_campaign_state`**
```json
{
  "campaign": "string (required)",
  "depth": "string (summary, detailed, full)"
}
```

**Respuesta estructurada:**
```json
{
  "campaign_name": "string",
  "acts": [
    {
      "number": 1,
      "title": "string",
      "summary": "string",
      "key_events": ["string"],
      "npcs_introduced": ["string"],
      "locations_visited": ["string"]
    }
  ],
  "characters": [
    {
      "name": "string",
      "level": 1,
      "class": "string",
      "status": "alive"
    }
  ],
  "active_quests": [...],
  "completed_quests": [...],
  "relationships": [...],
  "threads_unresolved": [...],
  "timeline": [...]
}
```

### 4.2 Generación de Arcos

**`generate_next_arc`**
```json
{
  "campaign": "string (required)",
  "arc_theme": "string (requerido, ej: 'traición', 'redención', 'guerra', 'descubrimiento')",
  "stakes_escalation": "string (local, regional, global, cosmic)",
  "num_acts": "number (2-5, default: 3)",
  "new_npcs": "number (cuántos NPCs nuevos sugerir)",
  "new_locations": "number (cuántas locations nuevas)",
  "challenges": ["string"]
}
```

**Lógica interna:**
1. Leer todo el contenido existente de la campaña
2. Identificar threads sin resolver (quests incompletas, NPCs misteriosos, profecías)
3. Escalar stakes según fase de campaña
4. Generar outline del nuevo arco (3-5 acts)
5. Sugerir nuevos NPCs con motivaciones conectadas
6. Sugerir nuevas locations relevantes
7. Proponer encounters clave
8. Verificar consistencia con lore existente

### 4.3 Continuidad & Consistencia

- [x] **`check_consistency`** ✅ IMPLEMENTADO EN v2.0
- [ ] **`update_timeline`**
  - Mantener línea temporal de eventos de campaña
  - Registrar cuándo ocurrió cada evento importante
  - Calcular tiempo transcurrido entre sesiones

### 4.4 Session Preparation

**`prep_session`**
```json
{
  "campaign": "string (required)",
  "session_number": "number",
  "focus": "string (opcional, ej: 'combate', 'roleplay', 'exploración')"
}
```

**Genera:**
- "Previously on..." resumen para jugadores
- Lista de posibles escenarios basado en estado actual
- NPCs relevantes para tener a mano
- Encounters preparados adaptativos
- Roll tables contextualizadas

---

## 🛠️ FASE 5: Optimización & Polish

> **Duración estimada:** 2 sprints  
> **Objetivo:** Performance, DX, y extensibilidad  
> **Entregable:** Sistema robusto, rápido, y extensible

### 5.1 Performance

- [ ] **Caching**
  - Cache de templates (no leer disco cada vez)
  - Cache de estado de campaña (invalidación por escritura)
  - Cache de imágenes generadas
- [ ] **Lazy loading**
  - No cargar todo el contenido de campaña en memoria
  - Paginación en `list_*` tools
- [ ] **Background jobs**
  - Generación de imágenes en goroutines (ya parcialmente implementado)
  - Compilación de PDF async
  - Pre-warm de caches

### 5.2 Developer Experience

- [ ] **Mejores mensajes de error**
  - Contexto rico para el LLM (qué falló, por qué, cómo arreglarlo)
  - Sugerencias de corrección
  - Ejemplos de uso correcto
- [ ] **Logging estructurado**
  - `slog` con niveles (debug, info, warn, error)
  - Correlation IDs para tracing
  - Métricas de uso de tools
- [ ] **Health checks**
  - Endpoint de health para monitoreo
  - Métricas de tiempo de respuesta por tool

### 5.3 Extensibilidad

- [ ] **Sistema de plugins**
  - Tools dinámicas registrables
  - Hooks para custom processing
- [ ] **Templates custom**
  - Templates definidos por usuario
  - Override de templates por campaña
- [ ] **Import/Export**
  - Formato FoundryVTT
  - Formato Roll20
  - JSON genérico

### 5.4 Documentación

- [ ] **Docs OpenAPI-style**
  - Descripción rica de cada tool
  - Ejemplos de requests/responses
  - Schemas de objetos
- [ ] **Guía de prompts sugeridos**
  - Cómo pedirle al LLM que use cada tool
  - Workflows comunes (ej: "crear campaña completa")
- [ ] **Best practices**
  - Cómo estructurar una campaña
  - Cuándo usar cada tool
  - Consejos de narrativa

---

## 🚀 FASE 6: Features Avanzadas (Backlog)

> **Estado:** Backlog, priorizable según feedback  
> **Objetivo:** Features que elevan la experiencia a nivel profesional

### 6.1 Combat & Encounter Builder

- [ ] **Encounter calculator**: Balance automático por nivel de party
- [ ] **Initiative tracker**: Estado de combate en tiempo real
- [ ] **Dynamic difficulty**: Ajuste on-the-fly según performance
- [ ] **Tactical map overlay**: Combinar mapas SVG con posiciones

### 6.2 World Building Avanzado

- [ ] **Gazetteer**: Enciclopedia del mundo con búsqueda
- [ ] **Faction simulator**: Facciones que evolucionan entre sesiones
- [ ] **Economy tracker**: Economía de lugares, inflación, recursos
- [ ] **Random tables contextuales**: Tablas que entienden el setting

### 6.3 Multi-campaign

- [ ] **Shared universe**: Personajes que cruzan campañas
- [ ] **Timeline global**: Eventos que afectan múltiples mesas
- [ ] **Cross-campaign references**: NPCs que aparecen en varias campañas

### 6.4 Integraciones

- [ ] **Discord bot**: Notificaciones, rolls, recordatorios
- [ ] **Calendar integration**: Scheduling de sesiones
- [ ] **Cloud sync**: Campañas en la nube (S3, GCS)
- [ ] **Web UI**: Dashboard visual para gestión

### 6.5 AI Avanzada

- [ ] **Narrative memory**: La AI recuerda detalles de sesiones pasadas
- [ ] **Voice generation**: Descripciones narradas
- [ ] **Procedural content**: Generación de dungeons, quests, NPCs con IA
- [ ] **Sentiment analysis**: Detectar cuándo los jugadores se aburren o se emocionan

---

## 📊 Cronograma Actualizado

| Fase | Estado | Duración | Focus | Deliverable Principal |
|------|--------|----------|-------|---------------------|
| **Fase 0** | ✅ COMPLETADA | 2 sprints | Tests + Arquitectura | 82.6% coverage, arquitectura limpia |
| **Fase 1** | ✅ COMPLETADA | 2 sprints | Coherencia Narrativa | Canon, validación, gates, estado |
| **Fase 2** | 🔄 PENDIENTE | 2 sprints | Personajes & Fichas | CRUD de personajes, templates |
| **Fase 3** | 🔄 PENDIENTE | 2 sprints | Integración Narrativa | Quests, relaciones, hooks |
| **Fase 4** | 🔄 PENDIENTE | 2 sprints | Evolución de Campaña | Lectura de estado, generación de arcos |
| **Fase 5** | 🔄 PENDIENTE | 2 sprints | Polish & Performance | Cache, logging, docs |
| **Fase 6** | 📋 BACKLOG | - | Avanzado | Combat, integraciones, AI |

**MVP funcional completo:** ~8-10 semanas (Fases 0-3)  
**Producción-ready:** ~12-14 semanas (Fases 0-5)

---

## 🧪 Estrategia de Testing

### TDD Cycle
```
1. RED    → Escribir test que falla
2. GREEN  → Implementar mínimo para pasar
3. REFACTOR → Limpiar y optimizar
4. REPEAT → Siguiente test
```

### Pirámide de Tests

```
         /\
        /  \     E2E Tests (5%)
       /    \    - Flujo completo de campaña
      /------\   
     /        \  Integration Tests (25%)
    /          \ - Handlers MCP con requests reales
   /------------\
  /              \ Unit Tests (70%)
 /                \- Servicios, dominio, validators
/__________________\
```

### Cobertura Actual (v2.0.0)

| Paquete | Cobertura | Estado |
|---------|-----------|--------|
| `internal/domain` | 93.1% | ✅ |
| `internal/services` | 82.6% | ✅ |
| `internal/repository` | - | Pendiente |
| `internal/mcp/handlers` | 64.3% | ⚠️ |
| `internal/compiler` | - | Pendiente |

---

## 🏗️ Principios de Arquitectura

### 1. Separation of Concerns
- **MCP Layer**: Solo adapters, traducen requests MCP a llamadas de servicio
- **Service Layer**: Lógica de negocio, orquestación
- **Domain Layer**: Entidades, value objects, reglas de negocio
- **Repository Layer**: Persistencia, abstracción de filesystem

### 2. Dependency Inversion
```go
// Interfaces definidas por los consumidores
type CampaignRepository interface {
    Create(campaign *domain.Campaign) error
    Read(name string) (*domain.Campaign, error)
    Update(campaign *domain.Campaign) error
    Delete(name string) error
    List() ([]string, error)
}

// Implementaciones en repository/
type FilesystemCampaignRepository struct { ... }
```

### 3. Interface Segregation
- Cada servicio tiene una interfaz enfocada
- No interfaces gigantes tipo "God interface"

### 4. TDD-First
- Ningún código de producción sin test que lo respalde
- Tests son documentación ejecutable
- Refactor seguro gracias a tests

---

## 📁 Estructura de Archivos Actual

```
grimorio/
├── cmd/
│   ├── grimorio/                    # Entry point (stdio MCP server)
│   └── migrate-v1-to-v2/            # Migration tool v1→v2
├── internal/
│   ├── mcp/
│   │   ├── server.go                # MCP tool definitions + handlers
│   │   └── handlers/
│   │       ├── campaign.go
│   │       ├── canon.go             # Narrative coherence handlers
│   │       ├── canon_test.go
│   │       ├── canon_gate_test.go
│   │       └── ...
│   ├── domain/                      # Domain models
│   │   ├── canon.go
│   │   ├── narrative_state.go
│   │   ├── validation.go
│   │   ├── gate.go
│   │   └── *_test.go
│   ├── services/                    # Business logic
│   │   ├── canon_service.go
│   │   ├── narrative_state_service.go
│   │   ├── validation_engine.go
│   │   ├── consistency_gate.go
│   │   └── *_test.go
│   ├── repository/                  # Persistence layer
│   │   ├── filesystem_canon.go
│   │   ├── memory_canon.go
│   │   └── *_test.go
│   ├── compiler/                    # Markdown → HTML → PDF pipeline
│   ├── svg/                         # Procedural SVG generator
│   ├── image/                       # Image provider abstraction
│   └── config/                      # Configuration management
├── campaigns/                       # Output directory
├── .claude-plugin/                  # Plugin manifest
├── commands/                        # Slash command definitions
├── agents/                          # Agent definitions
├── skills/                          # D&D 5e SRD skill
├── go.mod
├── go.sum
├── install.sh
├── CHANGELOG.md
├── README.md
└── ROADMAP.md
```

---

## 🎯 Métricas de Éxito

### Fase 0 ✅
- [x] 80%+ test coverage
- [x] Todos los handlers existentes tienen tests
- [x] Arquitectura limpia implementada

### Fase 1 ✅
- [x] Sistema de canon funcional (canon.json)
- [x] Motor de validación con 10 reglas
- [x] Consistency gate con batch processing
- [x] Estado narrativo trackeable

### Fase 2 (Próxima)
- [ ] CRUD completo de personajes
- [ ] Fichas visuales en Markdown
- [ ] Tests de integración para cada tool nueva

### Fase 3
- [ ] Sistema de quests funcional
- [ ] Relaciones trackables
- [ ] Hooks de sesión generados automáticamente

### Fase 4
- [ ] Estado de campaña legible completo
- [ ] Generación de arcos coherentes

### Fase 5
- [ ] Tiempo de respuesta < 500ms para lecturas
- [ ] Documentación completa
- [ ] Sistema de plugins funcional

---

## 📝 Notas de Implementación

### Convenciones de Nombres
- **Tools MCP**: `snake_case` (ej: `create_campaign`)
- **Funciones Go**: `PascalCase` para exportadas, `camelCase` para privadas
- **Archivos**: `snake_case.go`
- **Paquetes**: lowercase, cortos, descriptivos

### Manejo de Errores
- Siempre retornar errores, nunca panic
- Usar `fmt.Errorf("contexto: %w", err)` para wrapping
- Errores de dominio con tipos custom
- Mensajes de error claros para LLM

### Persistencia
- JSON para datos estructurados (characters, quests, canon, state)
- Markdown para contenido narrativo (acts, npcs)
- SVG/PNG para assets
- README.md como manifest de campaña

### Compatibilidad
- Mantener backward compatibility en tools existentes
- Versionar cambios breaking
- Deprecación gradual, no remoción abrupta

---

## 🤝 Contribución

### Proceso de Desarrollo
1. **Issue first**: Todo cambio empieza con un issue
2. **Branch por feature**: `feature/fase-X-nombre`
3. **PR con tests**: Sin tests, no hay merge
4. **Code review**: Mínimo 1 aprobación
5. **CI verde**: Todos los checks deben pasar

### Commits
- Conventional commits: `feat:`, `fix:`, `test:`, `refactor:`, `docs:`
- Descripción clara del "por qué", no solo el "qué"

---

**Última actualización:** Mayo 2026  
**Próxima revisión:** Al completar Fase 2
