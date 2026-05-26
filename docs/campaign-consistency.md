# Guía de Consistencia de Campaña

> Funcionalidades avanzadas para mantener coherencia narrativa en campañas largas.

---

## Tabla de Contenidos

1. [Resumen de Funcionalidades](#resumen-de-funcionalidades)
2. [P0: Consistencia de Estado](#p0-consistencia-de-estado)
3. [P1: Consecuencias Persistentes](#p1-consecuencias-persistentes)
4. [P2: Contenido Dinámico](#p2-contenido-dinámico)
5. [P3: Salud de Campaña](#p3-salud-de-campaña)
6. [Flujo de Trabajo Recomendado](#flujo-de-trabajo-recomendado)
7. [Referencia de Herramientas MCP](#referencia-de-herramientas-mcp)

---

## Resumen de Funcionalidades

| Fase | Funcionalidad | ¿Qué resuelve? | ¿Cuándo usar? |
|------|---------------|----------------|---------------|
| **P0** | Deduplicación de estado | Evita pistas y muertes duplicadas | Automático (siempre activo) |
| **P0** | Merge de quests/items | No pierde progreso al actualizar | Automático (siempre activo) |
| **P0** | Sync canon ↔ estado | NPCs muertos no "reviven" | Automático con `sync_to_canon=true` |
| **P0** | IDs estables | Referencias consistentes entre sesiones | Automático (siempre activo) |
| **P1** | Efectos diferidos persistentes | Consecuencias programadas no se pierden | Automático (siempre activo) |
| **P1** | Multi-sesión prep | Contexto de 3 sesiones atrás | Automático en `generate_session_prep` |
| **P1** | Escenarios enriquecidos | Prep con consecuencias y facciones | Automático con `with_scenarios=true` |
| **P2** | Tablas location-aware | Encuentros contextualizados por ubicación | Usar en `generate_random_tables` |
| **P2** | Generación dinámica de áreas | Contenido para lugares inesperados | Usar `generate_dynamic_area` |
| **P2** | Faction weighting | Encuentros según reputación | Automático en tablas context-aware |
| **P3** | Health checks | Detecta inconsistencias automáticamente | Usar `run_campaign_health` |
| **P3** | Context compression | Reduce payload >50% en campañas largas | Usar con `compressionEnabled=true` |
| **P3** | Rollback a sesión | Recuperar estado anterior | Usar `rollback_to_session` en errores |
| **P3** | Audit log | Historial de aprobaciones | Automático (siempre activo) |

---

## P0: Consistencia de Estado

### Deduplicación de Estado

**¿Qué hace?**
- `RevealedClues`: Deduplica por ID. Si la misma pista se revela dos veces, la segunda se ignora.
- `DeadNPCs`: Deduplica por `NPCID`. Un NPC solo puede morir una vez.
- `ActiveQuests` / `KeyItems`: Merge por ID en lugar de reemplazo destructivo.

**Ejemplo:**
```json
// Primera sesión: revelan pista "password-vault"
{
  "revealed_clues": [
    {"id": "password-vault", "description": "La contraseña es 'Sombra'"}
  ]
}

// Segunda sesión: vuelven a revelar la misma pista
{
  "revealed_clues": [
    {"id": "password-vault", "description": "La contraseña es 'Sombra'"}
  ]
}
// Resultado: NO se duplica, se mantiene una sola copia
```

**¿Cuándo usar?**
- **Siempre activo**. No requiere acción del DM.
- Asegurate de usar IDs consistentes en tus llamadas a `update_narrative_state`.

---

### Sync Canon ↔ Estado

**¿Qué hace?**
- Propaga muertes de NPCs desde `narrative_state` → `canon entities`
- Actualiza `CanonEntity.CanonState` a `dead` cuando un NPC muere
- Opcional, controlado por parámetro `sync_to_canon`

**Ejemplo:**
```json
// Matar un NPC y sincronizar con canon
{
  "campaign_id": "sunken-city",
  "session_num": 5,
  "dead_npcs": [
    {"npc_id": "lord-vampire", "name": "Lord Vampire", "cause": "Stake through heart"}
  ],
  "sync_to_canon": true
}
// Resultado: 
// 1. narrative_state.dead_npcs actualizado
// 2. canon entities["lord-vampire"].CanonState = "dead"
```

**¿Cuándo usar?**
- `sync_to_canon=true`: Cuando querés que el canon refleje el estado actual inmediatamente
- `sync_to_canon=false` (default): Cuando querés revisar cambios antes de aplicar al canon

**Recomendación:** Usar `true` en sesiones regulares, `false` en sesiones experimentales donde podrías hacer rollback.

---

### IDs Estables

**¿Qué hace?**
- Genera IDs determinísticos basados en SHA-256 del contenido
- Reemplaza contadores (`clue-1`, `clue-2`) que eran inestables
- Mismo contenido → mismo ID siempre

**Ejemplo:**
```json
// Antes (inestable):
"clue-1" → primera pista agregada
"clue-2" → segunda pista agregada
// Si borras clue-1 y agregás otra, la nueva sería "clue-2" (confuso)

// Ahora (estable):
"clue-a3f2b1c4" → hash del contenido
// Mismo contenido siempre genera "clue-a3f2b1c4"
```

**¿Cuándo usar?**
- **Automático**. El sistema genera IDs estables para clues, NPCs, quests, items.
- Si proveés tu propio ID, se usa ese (permite control manual).

---

## P1: Consecuencias Persistentes

### Efectos Diferidos Persistentes

**¿Qué hace?**
- Guarda efectos programados para sesiones futuras en `narrative_state.PendingEffects`
- Ejecuta automáticamente efectos cuando `ApplySession <= current_session`
- Respeta `IsRepeatable=false` (reglas no-repeatables solo disparan una vez)

**Ejemplo de regla de consecuencia:**
```json
{
  "id": "vampire-revenge",
  "trigger": "npc_death",
  "conditions": [
    {"type": "npc_alive", "params": {"npc_id": "vampire-spawn", "alive": false}}
  ],
  "effects": [
    {
      "type": "spawn_encounter",
      "description": "Vampire spawn attacks the party",
      "delay": "2"  // Aplica en 2 sesiones
    }
  ],
  "is_repeatable": false
}
```

**Flujo:**
1. Sesión 5: Matan al `vampire-spawn`
2. `evaluate_consequences` detecta la regla, crea `DelayedEffect` con `ApplySession=7`
3. Efecto se guarda en `narrative_state.PendingEffects`
4. Sesión 7: `generate_session_prep` incluye el efecto en `reminders`
5. DM ejecuta el encuentro de venganza

**¿Cuándo usar?**
- **Automático** al llamar `evaluate_consequences` después de `update_narrative_state`
- Revisar `session_prep.pending_effects` antes de cada sesión

---

### Multi-Sesión Previously On

**¿Qué hace?**
- Muestra las últimas 3 sesiones en lugar de solo la última
- Incluye contexto de arco narrativo
- Orden cronológico inverso (más reciente primero)

**Ejemplo:**
```
Previously On:

Arco: La Caída de la Corte de Vampiros

Sesión 7: La party se infiltró en el palacio y descubrió que el Rey está poseído.
Sesión 6: Combatieron contra los guardias vampiro y ganaron un aliado inesperado.
Sesión 5: Asesinaron al Lord Vampire, pero su hijo escapó con un juramento de venganza.
```

**¿Cuándo usar?**
- **Automático** en `generate_session_prep`
- Especialmente útil después de gaps de 2+ semanas entre sesiones

---

### Escenarios Enriquecidos

**¿Qué hace?**
- Genera `LikelyScenarios` priorizados:
  1. **Efectos pendientes** que vencen esta sesión
  2. **Decisiones sin resolver** de todas las sesiones
  3. **Cambios de facción** recientes
  4. **Quests activas** (último recurso)
- Cap en 7 escenarios para evitar sobrecarga

**Ejemplo:**
```json
{
  "likely_scenarios": [
    "⚠️ Venganza programada: Vampire spawn ataca (efecto diferido de sesión 5)",
    "📌 Decisión pendiente: La party perdonó al goblin chief - ¿volverá?",
    "🏰 Gremio de Ladrones: reputación -40 (hostil) - posible emboscada",
    "⚔️ Continuar quest 'El Amuleto Perdido' del Acto 2"
  ]
}
```

**¿Cuándo usar?**
- `generate_session_prep` con `with_scenarios=true`
- **Recomendado**: Siempre activar para mejor preparación

---

## P2: Contenido Dinámico

### Tablas Location-Aware

**¿Qué hace?**
- Filtra hechos del canon por ubicación con fuzzy matching
- Aplica modificadores de peso por reputación de facción (±80%)
- Excluye NPCs muertos automáticamente
- Boostea pistas reveladas (+3 peso)

**Ejemplo:**
```json
{
  "campaign_id": "sunken-city",
  "table_type": "encounter",
  "context": {
    "level_range": "5-7",
    "location_hint": "palace dungeon",
    "party_size": 4
  }
}
```

**Matching:**
- "palace" → matchea hechos con categoría "palace", "royal", "nobles"
- "dungeon" → matchea "dungeon", "cells", "underground"
- Facciones hostiles en zona → encuentros hostiles +50% peso
- NPCs muertos en área → excluidos de resultados

**¿Cuándo usar?**
- `generate_random_tables` con `location_hint` específico
- Cuando los jugadores van a zonas ya escritas del canon

---

### Generación Dinámica de Áreas

**¿Qué hace?**
- Genera áreas completas (3-5 features) on-demand
- 5 templates: wilderness, urban, dungeon, social, mixed
- Incluye: encuentros, tesoro, NPCs, boxed text, development branches
- Valida contra canon para evitar contradicciones

**Ejemplo:**
```json
{
  "campaign_id": "sunken-city",
  "location_description": "Los jugadores van a un templo abandonado en las afueras de la ciudad",
  "party_level": 5,
  "tone": "exploration",
  "auto_save": false
}
```

**Respuesta:**
```json
{
  "area": {
    "number": 7,
    "title": "Templo Abandonado de los Profundos",
    "features": [
      {"type": "hazard", "description": "Piso colapsante (DC 15 Dex)"},
      {"type": "clue", "description": "Símbolo del culto en el altar"},
      {"type": "treasure", "description": "Poción de curación en nicho secreto"}
    ],
    "encounters": [
      {"type": "combat", "cr": 5, "description": "2 Cultistas + 1 Acolito"}
    ],
    "boxed_text": "El templo se alza frente a vosotros...",
    "development_branches": [
      "SI investigan el altar → ENCUENTRAN mapa de túneles secretos",
      "SI ignoran el templo → PERSECUCIÓN de cultistas más tarde"
    ]
  },
  "validation": {"status": "pass", "warnings": []}
}
```

**¿Cuándo usar?**
- `generate_dynamic_area`: Jugadores van a lugar NO escrito
- `auto_save=false`: Para revisar antes de agregar al canon
- `auto_save=true`: Para agregar directamente (pasa por consistency gate)

---

### Faction Weighting

**¿Qué hace?**
- Modifica probabilidades de encuentros según reputación:
  - **Hostile (≤-30)**: Encuentros hostiles +50%, helpful -80%
  - **Unfriendly (-10 a -29)**: Hostiles +20%, helpful -40%
  - **Neutral (-10 a +30)**: Sin modificadores
  - **Friendly (+30 a +59)**: Hostiles -50%, helpful +40%
  - **Allied (≥+60)**: Hostiles -80%, helpful +60%

**Ejemplo:**
```json
// Facción "Thieves Guild" con reputación -40 (Hostile)
// Tabla de encuentros genera:
{
  "entries": [
    {"description": "Ladrones emboscan", "weight": 9},  // 1 → 9 (+80%)
    {"description": "Mercader ofrece ayuda", "weight": 1}  // 5 → 1 (-80%)
  ]
}
```

**¿Cuándo usar?**
- **Automático** cuando usás `location_hint` en `generate_random_tables`
- El sistema detecta facciones relevantes por ubicación

---

## P3: Salud de Campaña

### Health Checks

**¿Qué hace?**
- Ejecuta 5 reglas de validación automática:
  1. **Stale Quests**: Quests activas >10 sesiones → WARNING
  2. **Faction Contradictions**: Aliados con reputación hostil → CRITICAL
  3. **Orphaned Clues**: Pistas con prerequisitos no revelados → WARNING
  4. **Dead NPC Mismatch**: NPCs muertos en state pero vivos en canon → CRITICAL
  5. **McGuffin Drift**: Ubicación de McGuffin no coincide → CRITICAL

**Ejemplo:**
```json
{
  "campaign_id": "sunken-city"
}
```

**Respuesta:**
```json
{
  "report": {
    "overall_health": "fair",
    "findings": [
      {
        "severity": "CRITICAL",
        "rule": "faction_contradiction",
        "entity_id": "thieves-guild",
        "message": "Facción marcada como 'ally' en canon pero reputación -40 (hostile)"
      },
      {
        "severity": "WARNING",
        "rule": "stale_quest",
        "entity_id": "q-rescue-mission",
        "message": "Quest activa por 12 sesiones sin progreso"
      }
    ],
    "summary": {
      "critical": 1,
      "warning": 1,
      "info": 0
    }
  }
}
```

**¿Cuándo usar?**
- `run_campaign_health`: Antes de sesiones importantes
- **Recomendado**: Cada 5 sesiones o después de cambios mayores
- Pre-compilación de PDF para evitar inconsistencias

---

### Context Compression

**¿Qué hace?**
- Sesiones 1 a (actual-5): Resumen condensado en un solo bloque
- Últimas 5 sesiones: Detalladas como siempre
- Filtra NPCs/facciones por relevancia (ubicación actual, quests activas)
- Reduce payload >50% en campañas de 20+ sesiones

**Ejemplo:**
```json
{
  "campaign_id": "sunken-city",
  "session_num": 25,
  "compression_enabled": true,
  "compression_threshold": 5
}
```

**Sin compresión (25 sesiones):**
```
Session 1: [detalle completo]
Session 2: [detalle completo]
...
Session 25: [detalle completo]
// Total: ~500KB
```

**Con compresión:**
```
Arco 1 (Sesiones 1-20): La party descubrió el culto, asesinó al Lord Vampire,
y se ganó la enemistad del Gremio. 3 NPCs clave murieron, 2 quests completadas.

Sesión 21: [detalle]
Sesión 22: [detalle]
Sesión 23: [detalle]
Sesión 24: [detalle]
Sesión 25: [detalle]
// Total: ~200KB (-60%)
```

**¿Cuándo usar?**
- `compression_enabled=true`: Campañas de 10+ sesiones
- `compression_threshold=5`: Default, mostrar últimas 5 detalladas
- `compression_threshold=10`: Para contexto más amplio

---

### Rollback a Sesión

**¿Qué hace?**
- Restaura canon y narrative_state desde un checkpoint
- Checkpoints se crean automáticamente antes de cada `process_consistency_gate`
- Valida integridad con hash SHA256
- Lista sesiones "perdidas" (las posteriores al rollback)

**Ejemplo:**
```json
// Listar checkpoints disponibles
{
  "campaign_id": "sunken-city"
}
// Respuesta:
{
  "checkpoints": [
    {"session_num": 10, "created_at": "2026-05-20", "hash": "a3f2b1..."},
    {"session_num": 8, "created_at": "2026-05-15", "hash": "c7d4e9..."},
    {"session_num": 5, "created_at": "2026-05-10", "hash": "f1a2b3..."}
  ]
}

// Rollback a sesión 5
{
  "campaign_id": "sunken-city",
  "session_num": 5
}
// Respuesta:
{
  "status": "success",
  "restored_session": 5,
  "lost_sessions": [6, 7, 8, 9, 10],
  "warning": "Las sesiones 6-10 se perdieron. El audit log registra este rollback."
}
```

**¿Cuándo usar?**
- `rollback_to_session`: Error grave (ej: canon corrupto, decisión incorrecta)
- **Último recurso**: Siempre intentar fixes manuales primero
- El audit log registra el rollback para transparencia

---

### Audit Log

**¿Qué hace?**
- Registra todas las aprobaciones del consistency gate en JSONL
- Formato append-only (inmutable)
- Auto-purge después de 90 días (configurable)
- Campos: timestamp, campaignID, batchID, artifacts, decision, reason

**Ejemplo de entrada:**
```jsonl
{"timestamp":"2026-05-26T12:00:00Z","campaign_id":"sunken-city","batch_id":"batch-5","artifacts":["act-3-area-7","npc-vampire-lord"],"decision":"approved","reason":"Validated against canon, no contradictions"}
{"timestamp":"2026-05-26T12:05:00Z","campaign_id":"sunken-city","batch_id":"batch-6","artifacts":["quest-final"],"decision":"rejected","reason":"Quest reward 'Sword of Dawn' does not exist in canon"}
```

**¿Cuándo usar?**
- **Automático**: Cada aprobación/rechazo del gate se registra
- `get_audit_log`: Para auditoría o debugging
- **Útil**: Cuando necesitás rastrear cuándo se aprobó algo

---

## Flujo de Trabajo Recomendado

### Por Sesión

```
┌─────────────────────────────────────────────────────────────┐
│ ANTES DE LA SESIÓN                                          │
├─────────────────────────────────────────────────────────────┤
│ 1. generate_session_prep (with_scenarios=true)              │
│    → Revisar pending_effects, likely_scenarios              │
│ 2. dm_session_context (compression_enabled=true si 10+ ses) │
│    → Cargar contexto completo                               │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ DURANTE LA SESIÓN                                           │
├─────────────────────────────────────────────────────────────┤
│ 3. Si jugadores van a lugar NO escrito:                     │
│    → generate_dynamic_area (auto_save=false)                │
│    → Revisar, luego auto_save=true si OK                    │
│ 4. Si jugadores van a lugar ESCRITO:                        │
│    → generate_random_tables (location_hint=...)             │
│    → Encuentros contextualizados                            │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ DESPUÉS DE LA SESIÓN                                        │
├─────────────────────────────────────────────────────────────┤
│ 5. update_narrative_state (sync_to_canon=true)              │
│    → Clues, dead NPCs, decisions, completed quests          │
│ 6. evaluate_consequences                                     │
│    → Detectar efectos diferidos                             │
│ 7. update_faction_reputation (si cambia)                    │
│    → Ajustar reputaciones                                   │
│ 8. run_campaign_health (cada 5 sesiones)                    │
│    → Detectar inconsistencias                               │
└─────────────────────────────────────────────────────────────┘
```

### Cada 5 Sesiones

```
1. run_campaign_health
   → Revisar CRITICAL y WARNING findings
   → Fixear problemas antes de continuar

2. (Opcional) run_campaign_health --fix
   → Auto-fix para problemas simples (ej: sync dead NPCs)

3. Revisar audit log si hubo problemas
   → get_audit_log(days_back=30)
```

### Antes de Compilar PDF

```
1. run_campaign_health
   → Asegurar 0 CRITICAL findings

2. check_consistency (scope=full)
   → Validación completa de formato WotC

3. Si hay errores:
   → Fixear manualmente o rollback_to_session

4. compile_pdf
   → Generar PDF final
```

---

## Referencia de Herramientas MCP

### Gestión de Estado

| Herramienta | Parámetros | Cuándo usar |
|-------------|------------|-------------|
| `update_narrative_state` | `campaign_id`, `session_num`, `revealed_clues`, `dead_npcs`, `completed_quests`, `key_decisions`, `sync_to_canon` (bool, default false) | Final de cada sesión |
| `evaluate_consequences` | `campaign_id` | Después de `update_narrative_state` |
| `update_faction_reputation` | `campaign_id`, `faction_id`, `party_id`, `delta` (-100 a 100), `reason` | Cuando reputación cambia |

### Preparación de Sesión

| Herramienta | Parámetros | Cuándo usar |
|-------------|------------|-------------|
| `generate_session_prep` | `campaign_id`, `session_num`, `with_scenarios` (bool, default false) | Antes de cada sesión |
| `dm_session_context` | `campaign_id`, `session_num`, `include_prologue` (bool), `include_pdf_text` (bool), `compression_enabled` (bool, default false), `compression_threshold` (int, default 5) | Inicio de cada sesión |

### Contenido Dinámico

| Herramienta | Parámetros | Cuándo usar |
|-------------|------------|-------------|
| `generate_random_tables` | `campaign_id`, `table_type`, `context` (level_range, location_hint, party_size, setting_type) | Encuentros en zonas escritas |
| `generate_dynamic_area` | `campaign_id`, `location_description`, `party_level`, `tone` (combat/social/exploration), `auto_save` (bool, default false) | Jugadores van a lugar NO escrito |

### Salud de Campaña

| Herramienta | Parámetros | Cuándo usar |
|-------------|------------|-------------|
| `run_campaign_health` | `campaign_id` | Cada 5 sesiones, pre-PDF |
| `rollback_to_session` | `campaign_id`, `session_num` | Emergencia (canon corrupto) |
| `get_audit_log` | `campaign_id`, `days_back` (int, default 30) | Auditoría, debugging |
| `list_checkpoints` | `campaign_id` | Ver checkpoints disponibles |
| `process_consistency_gate` | `campaign_id`, `batch_id`, `proposals` (array) | Aprobar contenido en batch |

---

## Ejemplos de Flujos Completos

### Flujo: Jugadores Van a Lugar Inesperado

```
1. Jugadores: "Vamos al bosque prohibido" (NO está en el canon)

2. DM: generate_dynamic_area
   {
     "campaign_id": "sunken-city",
     "location_description": "Bosque prohibido al norte de la ciudad",
     "party_level": 5,
     "tone": "exploration",
     "auto_save": false
   }

3. Revisar área generada → OK

4. DM: process_consistency_gate (auto_save=true)
   → Agrega al canon

5. DM: generate_random_tables (location_hint="forbidden forest")
   → Encuentros contextualizados para la zona

6. Jugar sesión

7. DM: update_narrative_state
   {
     "current_location": "forbidden-forest",
     "revealed_clues": [...],
     "sync_to_canon": true
   }
```

### Flujo: Detectar y Fixear Inconsistencias

```
1. DM: run_campaign_health
   → CRITICAL: faction_contradiction (Thieves Guild es ally pero reputación -40)

2. DM revisa canon → NPC "Guild Master" marcado como "friendly" en canon

3. Opción A: Fix manual
   → Editar canon NPCs, cambiar a "hostile"
   → validate_canon para verificar

4. Opción B: Rollback
   → list_checkpoints
   → rollback_to_session(session_num=8) (antes de la contradicción)
   → Re-jugar sesiones 9-10 correctamente

5. DM: run_campaign_health
   → Verificar 0 CRITICAL findings
```

### Flujo: Campaña Larga (20+ Sesiones)

```
1. DM: generate_session_prep
   {
     "campaign_id": "sunken-city",
     "session_num": 21,
     "with_scenarios": true
   }
   → Previously On muestra sesiones 18, 19, 20 + arco condensado

2. DM: dm_session_context
   {
     "campaign_id": "sunken-city",
     "session_num": 21,
     "compression_enabled": true,
     "compression_threshold": 5
   }
   → Payload ~200KB en lugar de ~500KB

3. Jugar sesión normalmente

4. DM: update_narrative_state (sync_to_canon=true)
   → Estado sincronizado

5. DM: evaluate_consequences
   → Efectos diferidos aplicados

6. Cada 5 sesiones: run_campaign_health
   → Mantener campaña saludable
```

---

## Troubleshooting

### "Los jugadores se fueron a un lugar que no preparé"

**Solución:** `generate_dynamic_area`
```json
{
  "campaign_id": "...",
  "location_description": "Descripción del lugar inesperado",
  "party_level": 5,
  "tone": "exploration",
  "auto_save": false
}
```
Genera área completa en <2 segundos. Revisar y aprobar.

---

### "Hay inconsistencias en el canon"

**Solución:** `run_campaign_health`
```json
{"campaign_id": "..."}
```
Identifica problemas específicos. Fixear manualmente o `rollback_to_session`.

---

### "El payload de contexto es muy grande"

**Solución:** `dm_session_context` con compresión
```json
{
  "campaign_id": "...",
  "compression_enabled": true,
  "compression_threshold": 5
}
```
Reduce >50% el tamaño manteniendo contexto relevante.

---

### "Una consecuencia programada no se ejecutó"

**Diagnóstico:**
1. `run_campaign_health` → Verificar si hay warnings
2. Revisar `narrative_state.pending_effects` → ¿El efecto está ahí?
3. Verificar `ApplySession` → ¿Es <= current_session?

**Causas comunes:**
- `evaluate_consequences` no se llamó después de `update_narrative_state`
- Efecto con `IsRepeatable=false` ya se ejecutó
- Bug en el scheduler (revisar logs)

---

### "Necesito volver atrás, cometí un error"

**Solución:** `rollback_to_session`
```json
{
  "campaign_id": "...",
  "session_num": 10
}
```
**Advertencia:** Las sesiones 11+ se pierden. El audit log registra el rollback.

**Alternativa:** Fix manual sin rollback
- Editar `narrative_state.json` directamente
- `validate_canon` para verificar
- Menos destructivo pero más trabajo manual

---

## Mejores Prácticas

### ✅ Hacer

- Llamar `evaluate_consequences` después de CADA `update_narrative_state`
- Usar `sync_to_canon=true` en sesiones regulares
- Ejecutar `run_campaign_health` cada 5 sesiones
- Usar `compression_enabled=true` en campañas de 10+ sesiones
- Revisar `pending_effects` en session prep
- Usar `location_hint` en `generate_random_tables`
- Crear checkpoints antes de cambios mayores (`process_consistency_gate` lo hace auto)

### ❌ No Hacer

- Olvidar `evaluate_consequences` → efectos diferidos se pierden
- Usar `sync_to_canon=false` siempre → canon y state se desincronizan
- Ignorar CRITICAL findings en health reports
- No usar compresión en campañas largas → payload >500KB
- Hacer rollback sin revisar audit log primero
- Generar áreas dinámicas con `auto_save=true` sin revisar → puede crear contradicciones

---

## Recursos Adicionales

- **[DM Agent Guide](dm-agent-guide.md)** — Flujo completo de sesiones
- **[Session Tutorial](tutorials/session-tutorial.md)** — Primera sesión paso a paso
- **[MCP Tools](features/mcp-tools.md)** — Referencia completa de herramientas
- **[Developer Guide](developer-guide.md)** — Contribuir a Grimorio
