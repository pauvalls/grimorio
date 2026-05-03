# Grimorio MCP — Roadmap de Mejoras

> **Versión:** 1.0.0  
> **Fecha:** Mayo 2026  
> **Enfoque:** TDD-first, arquitectura limpia, evolución orgánica de campañas

---

## 📋 Estado Actual (Baseline)

El MCP actual tiene **12 herramientas** organizadas en 3 categorías:

| Categoría | Tools |
|-----------|-------|
| **Estructura** | `create_campaign`, `save_act`, `save_npcs`, `save_bestiary`, `save_encounters`, `save_maps` |
| **Assets** | `generate_map` (SVG), `generate_divider` (SVG), `generate_image`, `generate_images_batch` |
| **Output** | `compile_pdf`, `get_template` |

### Problemas Identificados
- ❌ Cero tests — 0% coverage
- ❌ Handlers monolíticos — lógica de negocio mezclada con MCP
- ❌ Sin validación estructurada de inputs
- ❌ Sin lectura de campañas existentes (solo escritura)
- ❌ Sin persistencia estructurada de datos (solo markdown flat)
- ❌ Sin concepto de personajes jugadores (PCs)
- ❌ Sin tracking de estado narrativo
- ❌ Sin evolución orgánica de campañas

---

## 🎯 FASE 0: Fundamentos — Testing & Arquitectura

> **Duración estimada:** 2 sprints  
> **Objetivo:** Base sólida con TDD, arquitectura limpia, y 80%+ coverage  
> **Entregable:** Toda la funcionalidad actual, pero con tests y arquitectura de producción

### 0.1 Testing Infrastructure

- [ ] **Suite de testing con mocks**
  - Builders para `mcp.CallToolRequest`
  - Asserts custom para respuestas MCP
  - Fixtures de campañas de ejemplo
- [ ] **Test helpers reutilizables**
  - `NewTestServer()` — servidor MCP aislado
  - `MakeRequest(tool, args)` — helper para invocar tools
  - `AssertSuccess(result)` / `AssertError(result, expected)`
- [ ] **Coverage tracking**
  - Objetivo: 80% mínimo en handlers
  - Reporte en CI

### 0.2 Refactor a Arquitectura Limpia

```
internal/
├── mcp/
│   ├── server.go              ← Solo wiring de tools
│   ├── server_test.go         ← Tests de integración
│   └── handlers/              ← Thin adapters
│       ├── campaign.go
│       ├── act.go
│       ├── npc.go
│       ├── bestiary.go
│       ├── encounter.go
│       ├── map.go
│       ├── asset.go
│       └── pdf.go
├── domain/                    ← Entidades de negocio
│   ├── campaign.go
│   ├── character.go
│   ├── quest.go
│   └── template.go
├── services/                  ← Lógica de negocio
│   ├── campaign_service.go
│   ├── character_service.go
│   ├── quest_service.go
│   └── asset_service.go
├── repository/                ← Persistencia
│   ├── filesystem.go
│   └── interfaces.go
└── validators/                ← Validación de inputs
    └── campaign.go
```

**Tareas:**
- [ ] Extraer interfaces de servicios
- [ ] Crear capa de repositorio (filesystem abstraction)
- [ ] Implementar inyección de dependencias
- [ ] Separar handlers MCP de lógica de negocio

### 0.3 Validación & Error Handling

- [ ] **Esquemas de validación por tool**
  - Validar tipos (string, number, bool)
  - Validar rangos (rooms: 2-10, level: 1-20)
  - Validar formatos (kebab-case para nombres)
- [ ] **Error handling consistente**
  - Códigos de error estructurados
  - Mensajes claros para el LLM
  - Diferenciar user error vs system error
- [ ] **Sanitización robusta**
  - Mejorar `sanitize()` (actualmente solo reemplaza no-alfanuméricos)
  - Validar paths (path traversal protection)
  - Normalizar nombres de archivo

### 0.4 CI/CD Pipeline

- [ ] **GitHub Actions**
  - Tests en PR
  - Lint con `golangci-lint`
  - Coverage report con `codecov`
- [ ] **Pre-commit hooks**
  - `go fmt`
  - `go vet`
  - Tests rápidos (< 30s)

---

## 🎭 FASE 1: Personajes & Fichas (PCs)

> **Duración estimada:** 2 sprints  
> **Objetivo:** Sistema completo de fichas de personaje jugador  
> **Entregable:** Crear, leer, actualizar personajes con stats, inventario, y relaciones

### 1.1 Modelo de Personaje

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

### 1.2 Tools Nuevas

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

### 1.3 Templates de Ficha

- [ ] Template Markdown para visualización
- [ ] Secciones: Stats, Skills, Inventory, Features, Backstory
- [ ] Estilo acorde al template general de campaña

### 1.4 Integración con NPCs

- [ ] **Relaciones PC-NPC**: Vincular personajes con NPCs existentes
- [ ] **Facciones**: A qué facción pertenece cada personaje
- [ ] **Sistema de reputación**: Cómo los NPCs ven a cada PC

---

## 🎬 FASE 2: Integración Narrativa

> **Duración estimada:** 2 sprints  
> **Objetivo:** Personajes vivos dentro de la narrativa, misiones personales  
> **Entregable:** Quests personales, tracking de estado, hooks narrativos

### 2.1 Sistema de Quests

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

### 2.2 Tools Nuevas

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

### 2.3 Weaving Narrativo

**`weave_character_arc`**
- Lee el estado actual de la campaña (todos los acts, quests, relaciones)
- Identifica oportunidades para insertar desarrollo de personajes
- Sugiere puntos de inflexión narrativos
- Genera "session hooks" para el DM

**`generate_session_hooks`**
- Basado en estado actual, genera 3-5 hooks para la próxima sesión
- Considera: quests activas, relaciones tensas, threads sin resolver

### 2.4 Track de Relaciones

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

## 📖 FASE 3: Evolución de Campaña

> **Duración estimada:** 2 sprints  
> **Objetivo:** Campañas que crecen orgánicamente, no solo acumulan acts  
> **Entregable:** Lectura de estado, generación de arcos, continuidad narrativa

### 3.1 Lectura de Estado

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

### 3.2 Generación de Arcos

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

### 3.3 Continuidad & Consistencia

**`check_consistency`**
- Comparar nuevo contenido propuesto contra lore existente
- Detectar contradicciones (NPC muerto que reaparece, timeline incorrecta)
- Sugerir correcciones

**`update_timeline`**
- Mantener línea temporal de eventos de campaña
- Registrar cuándo ocurrió cada evento importante
- Calcular tiempo transcurrido entre sesiones

### 3.4 Session Preparation

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

## 🛠️ FASE 4: Optimización & Polish

> **Duración estimada:** 2 sprints  
> **Objetivo:** Performance, DX, y extensibilidad  
> **Entregable:** Sistema robusto, rápido, y extensible

### 4.1 Performance

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

### 4.2 Developer Experience

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

### 4.3 Extensibilidad

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

### 4.4 Documentación

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

## 🚀 FASE 5: Features Avanzadas (Backlog)

> **Estado:** Backlog, priorizable según feedback  
> **Objetivo:** Features que elevan la experiencia a nivel profesional

### 5.1 Combat & Encounter Builder

- [ ] **Encounter calculator**: Balance automático por nivel de party
- [ ] **Initiative tracker**: Estado de combate en tiempo real
- [ ] **Dynamic difficulty**: Ajuste on-the-fly según performance
- [ ] **Tactical map overlay**: Combinar mapas SVG con posiciones

### 5.2 World Building Avanzado

- [ ] **Gazetteer**: Enciclopedia del mundo con búsqueda
- [ ] **Faction simulator**: Facciones que evolucionan entre sesiones
- [ ] **Economy tracker**: Economía de lugares, inflación, recursos
- [ ] **Random tables contextuales**: Tablas que entienden el setting

### 5.3 Multi-campaign

- [ ] **Shared universe**: Personajes que cruzan campañas
- [ ] **Timeline global**: Eventos que afectan múltiples mesas
- [ ] **Cross-campaign references**: NPCs que aparecen en varias campañas

### 5.4 Integraciones

- [ ] **Discord bot**: Notificaciones, rolls, recordatorios
- [ ] **Calendar integration**: Scheduling de sesiones
- [ ] **Cloud sync**: Campañas en la nube (S3, GCS)
- [ ] **Web UI**: Dashboard visual para gestión

### 5.5 AI Avanzada

- [ ] **Narrative memory**: La AI recuerda detalles de sesiones pasadas
- [ ] **Voice generation**: Descripciones narradas
- [ ] **Procedural content**: Generación de dungeons, quests, NPCs con IA
- [ ] **Sentiment analysis**: Detectar cuándo los jugadores se aburren o se emocionan

---

## 📊 Cronograma Sugerido

| Fase | Duración | Focus | Deliverable Principal |
|------|----------|-------|---------------------|
| **Fase 0** | 2 sprints | Tests + Arquitectura | 80%+ coverage, arquitectura limpia |
| **Fase 1** | 2 sprints | Personajes & Fichas | CRUD de personajes, templates |
| **Fase 2** | 2 sprints | Integración Narrativa | Quests, relaciones, hooks |
| **Fase 3** | 2 sprints | Evolución de Campaña | Lectura de estado, generación de arcos |
| **Fase 4** | 2 sprints | Polish & Performance | Cache, logging, docs |
| **Fase 5** | Backlog | Avanzado | Combat, integraciones, AI |

**MVP funcional completo:** ~8-10 semanas (Fases 0-3)  
**Producción-ready:** ~12-14 semanas (Fases 0-4)

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

### Estructura de Tests

```
internal/
├── mcp/
│   ├── server_test.go              ← Tests de wiring
│   └── handlers/
│       ├── campaign_test.go
│       ├── character_test.go
│       └── quest_test.go
├── services/
│   ├── campaign_service_test.go
│   ├── character_service_test.go
│   └── quest_service_test.go
├── domain/
│   ├── character_test.go
│   └── quest_test.go
└── repository/
    └── filesystem_test.go
```

### Tipos de Tests

- **Unit tests**: Lógica de negocio pura, sin dependencias externas
- **Integration tests**: Handlers MCP con servidor real, filesystem temporal
- **E2E tests**: Flujo completo desde creación de campaña hasta PDF

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

## 📁 Estructura de Archivos Propuesta

```
grimorio/
├── cmd/
│   └── grimorio/
│       └── main.go
├── internal/
│   ├── mcp/
│   │   ├── server.go
│   │   ├── handlers/
│   │   │   ├── campaign.go
│   │   │   ├── act.go
│   │   │   ├── character.go
│   │   │   ├── quest.go
│   │   │   ├── npc.go
│   │   │   ├── bestiary.go
│   │   │   ├── encounter.go
│   │   │   ├── map.go
│   │   │   ├── asset.go
│   │   │   └── pdf.go
│   │   └── handlers_test.go
│   ├── domain/
│   │   ├── campaign.go
│   │   ├── character.go
│   │   ├── quest.go
│   │   ├── npc.go
│   │   └── template.go
│   ├── services/
│   │   ├── campaign_service.go
│   │   ├── character_service.go
│   │   ├── quest_service.go
│   │   └── asset_service.go
│   ├── repository/
│   │   ├── interfaces.go
│   │   ├── filesystem.go
│   │   └── memory.go          ← Para tests
│   ├── validators/
│   │   └── validators.go
│   ├── compiler/
│   │   └── compiler.go
│   ├── svg/
│   │   └── svg.go
│   ├── image/
│   │   └── image.go
│   └── config/
│       └── config.go
├── campaigns/                  ← Output de campañas
├── templates/                  ← Templates Markdown/CSS
├── tests/
│   ├── fixtures/
│   │   └── campaigns/
│   └── integration/
├── docs/
│   ├── api.md
│   └── examples.md
├── .github/
│   └── workflows/
│       └── ci.yml
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── ROADMAP.md                  ← Este archivo
```

---

## 🎯 Métricas de Éxito

### Fase 0
- [ ] 80%+ test coverage
- [ ] Todos los handlers existentes tienen tests
- [ ] Build pasa en CI
- [ ] Sin warnings de `golangci-lint`

### Fase 1
- [ ] CRUD completo de personajes
- [ ] Fichas visuales en Markdown
- [ ] Tests de integración para cada tool nueva

### Fase 2
- [ ] Sistema de quests funcional
- [ ] Relaciones trackables
- [ ] Hooks de sesión generados automáticamente

### Fase 3
- [ ] Estado de campaña legible completo
- [ ] Generación de arcos coherentes
- [ ] Check de consistencia funcional

### Fase 4
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
- JSON para datos estructurados (characters, quests)
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
**Próxima revisión:** Al completar Fase 0
