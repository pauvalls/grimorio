# Save Commands - Formato de Markdown Esperado

## Overview

Los comandos `save_npcs` y `save_bestiary` ahora:
1. ✅ Escriben archivos markdown
2. ✅ Actualizan `canon.json` con las entidades
3. ✅ Crean archivos JSON individuales para los repositorios
4. ✅ Hacen rollback si el parseo falla

---

## Save NPCs

### Formato Esperado

```markdown
## NPCs Principales

### Nombre del NPC
*Alineamiento Raza Clase*

- **Ubicación:** Lugar donde se encuentra
- **Faction:** Nombre de la facción
- **Estadísticas:** AC 18, HP 38

#### Descripción Opcional

Párrafo con más detalles del NPC.

### Otro NPC
*Legal Neutral Humano Guerrero*

- **Ubicación:** Cuartel
- **Faction:** Guardia
```

### Ejemplo Completo

```markdown
## NPCs Principales

### Madre Superiora Elara Voss
*Legal Bueno Humana Clériga de Pelor*

- **Ubicación:** Catedral de la Aurora Eterna
- **Faction:** Orden de la Plata
- **Estadísticas:** AC 18, HP 38

**Rol en la historia:** Aliada conflictiva y fuente de curación mágica.

#### Apariencia Física

Elara mide 1,72 metros, con una complexión delgada pero firme. Sus ojos son de un verde pálido.

#### Personalidad

**Mannerisms:** Se toca constantemente el símbolo sagrado cuando está nerviosa.

**Motivations:** Quiere salvar a su hermano Cael, escondido en las criptas.

## Facciones

### Orden de la Plata
- **Líder:** Elara Voss
- **Objetivo:** Proteger la ciudad de la Peste Gris
- **Aliados:** Guardia de la Ciudad
- **Enemigos:** Hijos del Cenizo

### Guardia de la Ciudad
- **Líder:** Capitán Aldric Thorne
- **Objetivo:** Mantener el orden
```

### Qué se Extrae

| Campo | Cómo se extrae |
|-------|----------------|
| `name` | Heading `### Nombre` |
| `role` | Primera línea después del nombre (ej: "*Legal Bueno...*") |
| `faction` | `- **Faction:** Nombre` |
| `description` | `- **Ubicación:**` o primer párrafo |
| `stats.AC` | `- **Estadísticas:** AC 18` |
| `stats.HP` | `- **Estadísticas:** HP 38` |

### Facciones

Las facciones se detectan automáticamente bajo la sección `## Facciones`.

---

## Save Bestiary

### Formato Esperado

```markdown
# Nombre del Monstruo
*CR 1/4, Tipo/Tamaño*

- **AC:** 15
- **HP:** 33
- **Type:** undead / humanoid / beast

### Description

Descripción del monstruo.

### Tactics

Tácticas de combate.

# Otro Monstruo
*CR 5, Humanoide Mediano*

- **AC:** 17
- **HP:** 90
```

### Ejemplo Completo

```markdown
# Cenizo Recién Convertido
*CR 1/4, No-muerto Mediano*

- **AC:** 15
- **HP:** 33
- **Type:** Undead

### Descripción

Los Cenizos son víctimas de la Peste Gris que han perdido su humanidad.

### Tácticas

Atacan en grupo, buscando rodear a sus presas.

# Vex Terrow - El Harúspice
*CR 5, Humanoide Mediano (humano)*

- **AC:** 17
- **HP:** 90
- **Type:** Humanoid

### Descripción

El líder de los Hijos del Cenizo.

### Acciones

- **Ataque de Peste:** +6 al ataque, 1d6+3 daño necrótico
```

### Qué se Extrae

| Campo | Cómo se extrae |
|-------|----------------|
| `name` | Heading `# Nombre` |
| `CR` | `*CR 1/4` o `*CR 5` |
| `type` | `Type: undead` o detectado por palabras clave |
| `size` | `No-muerto Mediano` → `Medium` |
| `stats.AC` | `- **AC:** 15` |
| `stats.HP` | `- **HP:** 33` |

---

## Errores Comunes

### ❌ No se detecta el NPC

```markdown
NPCs:
- Thorin
```

**Problema:** Falta heading `###`

**Solución:**
```markdown
### Thorin Ironforge
*Legal Good Dwarf Fighter*
```

### ❌ No se extrae la faction

```markdown
### Elara
- Ubicación: Catedral
```

**Problema:** Falta `- **Faction:**`

**Solución:**
```markdown
### Elara
- **Faction:** Orden de la Plata
```

### ❌ Bestiary no detecta monstruos

```markdown
## Monstruos
Cenizo - CR 1/4
```

**Problema:** Heading `##` en vez de `#`

**Solución:**
```markdown
# Cenizo
*CR 1/4, No-muerto*
```

---

## Testing

Después de ejecutar `save_npcs` o `save_bestiary`:

```bash
# Verificar canon.json
cat ~/campaigns/<campaign>/canon/canon.json | jq '.entities[] | select(.type=="npc")'

# Verificar JSON files
ls -la ~/campaigns/<campaign>/npcs/*.json
ls -la ~/campaigns/<campaign>/monsters/*.json

# Verificar en dm_session_context
dm_session_context(campaign_id="<campaign>", session_num=1)
# → payload.npcs y payload.bestiary NO deberían estar vacíos
```

---

## Rollback Automático

Si el parseo falla:
- ❌ El archivo markdown **NO** se escribe
- ❌ canon.json **NO** se actualiza
- ✅ Se retorna un error descriptivo

Ejemplo de error:
```
failed to parse NPCs from markdown: no NPCs or factions found in markdown - 
expected format: ## Name followed by - **Name** — description
```

---

## Session 0 Flow Recomendado

1. **Generar campaña:** `grimorio_generate_adventure_bible(...)`
2. **Generar capítulos:** `save_areas(...)` para cada capítulo
3. **Guardar NPCs:** `save_npcs(campaign="...", content="...")`
4. **Guardar Bestiary:** `save_bestiary(campaign="...", content="...")`
5. **Verificar:** `dm_session_context(campaign_id="...", session_num=1)`
6. **Jugar:** Los NPCs y monstruos ya están cargados ✅

---

## Archivos Afectados

| Archivo | Qué contiene |
|---------|--------------|
| `npcs/npcs_and_factions.md` | Markdown fuente (source of truth) |
| `npcs/*.json` | JSON files para `FilesystemNPCRepository` |
| `bestiary/bestiary.md` | Markdown fuente |
| `bestiary/*.json` | JSON files para `FilesystemMonsterRepository` |
| `canon/canon.json` | Índice de todas las entidades |

---

## Notas de Implementación

- **Atomicidad:** Se usa temp file + `os.Rename()` para escrituras atómicas
- **Upsert:** Las entidades se actualizan si ya existen (por ID)
- **ID Generation:** Los IDs se generan con `sanitizeID(nombre)` → "Thorin Ironforge" → "thorin-ironforge"
- **Rollback:** Si cualquier paso falla, se elimina el temp file y no se escribe nada
