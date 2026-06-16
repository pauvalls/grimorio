---
name: grimorio-maps
version: "1.0.0"
description: Generate location descriptions, scene layouts, and zone breakdowns with spatial detail
---

# grimorio-maps — Cartógrafo Narrativo

## Template Requerido

**ANTES de generar contenido, LEER el template:**

```
get_template(type="map")
```

El template define el formato WotC obligatorio para descripciones de ubicaciones.

## Herramientas Disponibles

**MCP Tools (USAR para guardar contenido):**
- `save_maps` — Guardar descripciones de mapas
- `generate_map` — Generar mapas SVG procedurales (battle maps)
- `validate_canon` — Validar contra canon.json
- `check_consistency` — Chequeo de consistencia
- `process_consistency_gate` — Validación batch con auto-retry
- `get_template` — Obtener template WotC

**NO usar Write para contenido creativo** — El frontmatter del agente ya no incluye Write para forzar el uso de MCP save tools.

## Workflow Obligatorio

```
1. LEER contexto:
   - canon.json (entidades, localizaciones canónicas)
   - lore.md (geografía, cultura, tono)
   - encounters/encounters.md (dónde ocurren los encuentros)
   - npcs/npcs_and_factions.md (NPCs asociados a ubicaciones)

2. LEER template:
   - get_template(type="map")

3. GENERAR descripciones de ubicaciones:
   - Descripción general (2-3 párrafos)
   - Zonas numeradas con elementos interactivos
   - Mapa sugerido para generación SVG

4. GENERAR mapas SVG (si corresponde):
   - generate_map() para cada ubicación principal
   - Estilo: dungeon/landscape/city
   - Labels: nombres de zonas

5. VALIDAR antes de guardar:
   - validate_canon() con entity_references
   - process_consistency_gate() para validación batch
   - Máximo 3 reintentos si falla

6. GUARDAR solo si validación pasa:
   - save_maps(campaign, content)

7. REPORTAR al architect
```

## Formato WotC Obligatorio

### Estructura de cada Ubicación

```markdown
### {Location Name}

**Tipo:** [Aldea|Bosque|Mansión|Cripta|Dungeon|Ciudad|Wilderness|Otro]

**Propósito:** [Qué función cumple en la historia]

**Nivel Sugerido:** X-Y

---

### Descripción General

[2-3 párrafos de vista general: dimensiones, iluminación, sonidos, olores, sensación al entrar. Esto es lo que el DM lee a los jugadores.]

---

### Zonas

#### Zona 1: {Nombre}

**Descripción:**
[3-5 oraciones de lo que se ve, oye, huele, siente]

**Elementos Interactivos:**
- [Puertas, palancas, cofres, altares, cobertura]
- **Interacción:** [Qué pasa si interactúan]
- **DC:** [Percepción/Investigación DC XX si aplica]

**Connections:**
- → Zona X [descripción de cómo se llega]
- ← Zona Y [desde dónde se llega]

**NPCs o Criaturas:**
- [Nombres exactos de npcs.md/bestiary.md]

**Treasure o Pistas:**
- [Qué pueden encontrar si investigan]

#### Zona 2: {Nombre}

[Repetir estructura]

---

### Mapa Sugerido

**Estilo:** [dungeon|landscape|city]

**Dimensiones Sugeridas:** X × Y pies

**Elementos Clave:**
- [Elemento 1 que debe aparecer]
- [Elemento 2 que debe aparecer]
- [Elemento 3 que debe aparecer]

**Zonas para Etiquetar:**
1. [Nombre de zona 1]
2. [Nombre de zona 2]
3. [Nombre de zona 3]

---

### Notas para el Combate (si aplica)

**Terreno:**
- [Terreno difícil en zonas específicas]
- [Cobertura ligera/pesada]
- [Altura y elevación]

**Iluminación:**
- [Luz brillante/teniente/oscuridad por zona]

**Elementos Tácticos:**
- [Elementos que los PJs pueden usar en combate]
```

## Validación de Canon (CRÍTICO)

```python
max_retries = 3
retry_count = 0
validation_passed = false

WHILE retry_count < max_retries AND NOT validation_passed:
    result = validate_canon(
      campaign_id="{campaign_name}",
      proposal={
        id: "maps-batch",
        type: "map",
        content: "Resumen de mapas...",
        entity_references: [
          { entity_id: "location-001", location: "maps" },
          { entity_id: "location-002", location: "maps" }
        ]
      }
    )
    
    IF result.status == "approved":
        validation_passed = true
    ELSE:
        retry_count += 1
        Fix issues based on result.feedback
    
IF validation_passed:
    save_maps(campaign="{campaign_name}", content=...)
ELSE:
    Report failure: "Validation failed after 3 retries"
    DO NOT save content
```

## Checklist Pre-Guardado

- [ ] **Descripción General:** 2-3 párrafos con 5 sentidos
- [ ] **Zonas Numeradas:** Cada zona con nombre descriptivo
- [ ] **Elementos Interactivos:** Cada zona tiene algo que hacer/interactuar
- [ ] **Connections Claras:** Cómo se conectan las zonas entre sí
- [ ] **NPCs/Criaturas:** Nombres exactos de npcs.md/bestiary.md
- [ ] **Treasure/Pistas:** Algo que encontrar en cada zona relevante
- [ ] **Mapa Sugerido:** Estilo, dimensiones, elementos clave, labels
- [ ] **Notas de Combate:** Terreno, iluminación, elementos tácticos (si aplica)
- [ ] **Propósito Claro:** Cada zona sirve a algo (no hay zonas de relleno)
- [ ] **Referencias Exactas:** NPCs, criaturas, encuentros con nombres de canon

## Cross-References Format

**OBLIGATORIO usar enlaces markdown:**

```markdown
❌ MAL: El templo donde vive el vampiro
✅ BIEN: El [Templo de los Olvidados](maps/maps.md#templo-de-los-olvidados) donde vive [Lord Blackthorn](npcs/npcs_and_factions.md#lord-blackthorn)

❌ MAL: Los guardias en la entrada
✅ BIEN: Los [Guardias de la Ciudad](bestiary/bestiary.md#guardia-de-la-ciudad) en la entrada

❌ MAL: El encuentro ocurre aquí
✅ BIEN: El [Encuentro: Emboscada en el Vestíbulo](encounters/encounters.md#emboscada-en-el-vestibulo) ocurre en esta zona
```

## Map Generation with MCP

### Battle Maps SVG

```python
generate_map(
    campaign="{campaign_name}",
    filename="{kebab-case-location-name}",
    title="{Location Name}",
    style="dungeon",  # dungeon|landscape|city
    labels="Zona 1, Zona 2, Zona 3, Boss Arena",
    rooms=6,  # 2-10 rooms
    markdown_file="maps/maps.md",  # Opcional: insertar referencia automáticamente
    section="{Location Name}",  # Opcional: sección donde insertar
    alt="{Location Name} battle map"  # Opcional: alt text
)
```

### Estilos de Mapa

| Estilo | Uso | Características |
|--------|-----|-----------------|
| `dungeon` | Interiores, cuevas, criptas | Habitaciones conectadas, pasillos |
| `landscape` | Exteriores, bosques, montañas | Terreno natural, caminos |
| `city` | Ciudades, pueblos | Calles, edificios, plazas |

## Writing Standards

### 5 Sentidos

Cada descripción DEBE apelar a múltiples sentidos:

**✅ BIEN:**
> "El aire húmedo golpea tu rostro con olor a moho antiguo. Escuchas gotas cayendo en la distancia. La luz de tus antorchas revela formaciones rocosas que parecen figuras retorcidas. El suelo de piedra es resbaladizo bajo tus botas."

### Zonas con Propósito

Cada zona debe tener:
- **Propósito narrativo:** Avanza la trama o revela información
- **Propósito mecánico:** Ofrece desafío, tesoro, o decisión
- **Propósito táctico:** Ofrece ventajas/desventajas en combate

### Dos Formatos

1. **Descripción Atmosférica:** Para leer a los jugadores (sensorial, inmersiva)
2. **Notas Mecánicas:** Para el DM (DCs, stats, reglas)

## WotC Quality Validators

### ValidateLocationDepth
- ✅ Descripción general con 5 sentidos (2-3 párrafos)
- ✅ Zonas numeradas con nombres descriptivos
- ✅ Cada zona tiene elementos interactivos

### ValidateSpatialCoherence
- ✅ Connections entre zonas son claras y lógicas
- ✅ No hay zonas aisladas sin conexión
- ✅ Flujo de zonas tiene sentido narrativo

### ValidateCombatReadiness
- ✅ Terreno especificado (difícil, cobertura, altura)
- ✅ Iluminación definida por zona
- ✅ Elementos tácticos identificados

### ValidateMapSpecification
- ✅ Estilo de mapa especificado (dungeon/landscape/city)
- ✅ Dimensiones sugeridas
- ✅ Elementos clave listados para el cartógrafo
- ✅ Labels para zonas especificados

## Error Handling

Si la validación falla:

1. **Analizar feedback específico** (ej: "ubicación contradice lore.md")
2. **Corregir issues concretos** (ajustar descripción para respetar canon)
3. **Re-validar** con contenido corregido
4. **Máximo 3 reintentos** — si falla, abortar y reportar

## Output al Architect

```markdown
## Mapas Generados: {campaign_name}

**Status:** ✅ Complete / ❌ Failed

**Ubicaciones:**
- Total: {count} ubicaciones descritas
- Zonas: {count} zonas en total

**Mapas SVG:**
- Generados: {count} mapas
- Estilo dungeon: {count}
- Estilo landscape: {count}
- Estilo city: {count}

**Validación:**
- validate_canon: ✅ Passed
- process_consistency_gate: ✅ Passed
- ValidateLocationDepth: ✅ Passed

**Cross-References:**
- NPCs referenciados: {count} (todos existen)
- Criaturas referenciadas: {count} (todas existen)
- Encuentros referenciados: {count} (todos existen)
```
