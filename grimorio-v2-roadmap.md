# Grimorio v2.0 Roadmap: Sistema Correcto de Agentes y Estructura

## Estado Actual vs Objetivo

### Estado Actual (v1.x)
```
Usuario → grimorio-lore → grimorio-npc → grimorio-bestiary → grimorio-encounters → grimorio-maps → grimorio-acts → grimorio-cartographer → COMPILAR PDF
```

**Problemas:**
- Los actos generan "escenas narrativas" (3-5 por acto, 350 palabras cada una)
- Sin integración entre componentes (criaturas en actos no verificadas en bestiary)
- Templates son "guiones creativos", no "estándares técnicos"
- El compilador genera PDF plano sin referencias cruzadas
- Densidad técnica insuficiente (CDs sin números, tesoro inconsistente)

### Objetivo (v2.0)
```
FASE 1: Fundación
  Usuario → grimorio-lore → grimorio-npc → grimorio-bestiary → grimorio-maps
  
FASE 2: Áreas
  grimorio-areas → grimorio-encounters → grimorio-quests
  
FASE 3: Integración (NUEVO - OBLIGATORIO)
  grimorio-integrator → validación canon → cross-reference check
  
FASE 4: Visuales
  grimorio-cartographer → grimorio-artist → grimorio-dividers
  
FASE 5: Compilación
  grimorio-compiler v2 → PDF con índice jerárquico + referencias cruzadas
```

**Resultado:**
- 10-15 áreas numeradas por acto (180 palabras cada una)
- 100% de criaturas verificadas contra bestiary
- Tesoro con XP y formato estándar en cada área con criaturas
- CDs numéricos específicos (DC 10/12/14/15/18/20)
- PDF con navegación, índice jerárquico, y stat blocks inline

---

## Nueva Arquitectura de Agentes

### 1. grimorio-lore (Mejorado)
**Cambios:**
- Añadir "Adventure Background" estructurado (timeline de eventos)
- Incluir "Quest Hooks" (3-5 ganchos iniciales concretos)
- Definir "Key Locations" (lista de todas las localizaciones principales)
- Especificar "Faction Relations" (matriz de relaciones entre facciones)

**Output:**
```markdown
# Adventure Background
## Timeline
- 50 years ago: Lord Blackthorn makes pact
- 3 weeks ago: Tilly dies
- 1 week ago: Whispers reach village
- Yesterday: Reward posted

## Key Locations
1. Blackthorn Manor (3 floors + cellar)
2. Village of Thornwick
3. The Ritual Chamber

## Quest Hooks
1. Reward notice (50 gp)
2. Missing Brother Aldric
3. Visions of crying child

## Faction Relations
| Faction | Status | Goal |
|---------|--------|------|
| Villagers | Scared | End curse |
| Entity | Hostile | Spread influence |
```

### 2. grimorio-npc (Mejorado)
**Cambios:**
- Todos los NPCs DEBEN tener alineamiento (LG/CG/NE/etc)
- Todos los NPCs DEBEN tener stats para combate (o "non-combatant")
- Añadir "Location" (dónde se encuentran inicialmente)
- Añadir "Quest Involvement" (en qué quests participan)
- Añadir "Secrets" (1-2 secretos relevantes)
- Formato estándar WotC: "*LG male Illuskan human innkeeper*"

**Output:**
```markdown
### Brother Aldric
*LG male Chondathan human cleric*

**Location:** Trapped in collapsed wing (Area C4)
**Role:** Wandering cleric investigating haunting
**Motivation:** Banish darkness, rescue trapped souls
**Secret:** Knows the entity's true name but can't speak it
**Quest:** "Find Brother Aldric" - missing person hook
**Stats:** Use **Priest** (MM p. 348) with these changes:
- HP: 27 (5d8+5)
- Has **Turn Undead**
- Carries **Aldric's Journal**
```

### 3. grimorio-bestiary (Mejorado)
**Cambios:**
- Todos los stat blocks deben seguir formato WotC estándar
- Incluir "Tactics" (cómo lucha)
- Incluir "Role" (skirmisher, tank, controller, etc.)
- Incluir "Encounter Groups" (qué criaturas van juntas)
- Referencia clara: "(MM p. 279)" o "(custom)"

**Output:**
```markdown
## Murmuring Specter
*Medium undead, chaotic evil*

**Armor Class** 12
**Hit Points** 22 (5d8)
**Speed** 0 ft., fly 40 ft. (hover)

| STR | DEX | CON | INT | WIS | CHA |
|:---:|:---:|:---:|:---:|:---:|:---:|
| 1 (-5) | 14 (+2) | 11 (+0) | 10 (+0) | 10 (+0) | 15 (+2) |

**Damage Resistances** acid, cold, fire, lightning, thunder; bludgeoning, piercing, and slashing from nonmagical attacks
**Damage Immunities** necrotic, poison
**Condition Immunities** charmed, exhaustion, frightened, grappled, paralyzed, petrified, poisoned, prone, restrained, unconscious
**Senses** darkvision 60 ft., passive Perception 10
**Languages** Common
**Challenge** 1/4 (50 XP)

***Incorporeal Movement.*** The specter can move through...

### Tactics
Drifts through walls to ambush isolated PCs...

### Encounter Groups
- Easy: 1 specter
- Medium: 2 specters + 1 shadow
- Hard: 3 specters + 1 wraith
```

### 4. grimorio-areas (NUEVO - Reemplaza grimorio-acts)
**Propósito:** Generar áreas numeradas jugables, no escenas narrativas

**Input:**
- canon.json
- lore.md (con Key Locations y Quest Hooks)
- npcs.md (con locations iniciales)
- bestiary.md (con encounter groups)

**Output:**
- 10-15 áreas numeradas por acto
- Cada área: 150-200 palabras, formato estándar WotC
- Cross-references: a NPCs, bestiary, otras áreas

**Formato OBLIGATORIO:**
```markdown
### [Código]. [Nombre Descriptivo]

[2-3 líneas de contexto: quién está aquí y por qué]

This area has the following features:
- [Feature físico 1]
- [Feature físico 2]
- [Feature físico 3]

***[Mecánica/NPC 1].*** [Descripción con CDs específicos]

***[Mecánica/NPC 2].*** [Descripción con comportamiento]

***Treasure.*** [Ubicación específica]: [Cantidad] gp, [Objetos con valor]

***Secret Door/Trap.*** [DC específico] [mecanismo] [consecuencia]
```

### 5. grimorio-encounters (Mejorado)
**Cambios:**
- Generar encuentros como "plantillas" que las áreas referencian
- Cada encuentro DEBE tener: XP total, dificultad, desarrollo por rondas
- Incluir "Tactical Map" (descripción del terreno)
- Incluir "Conditions" (iluminación, terreno difícil, cobertura)

**Output:**
```markdown
## Encounter: Nursery Ambush
**Difficulty:** Hard
**XP Total:** 450 (90 XP per PJ)
**Location:** Area N2 (Nursery)

### Creatures
- 1 **Maren Voss** (use Wraith MM p. 302, HP: 45)
- 2 **Ghost Children** (use Will-o'-Wisp MM p. 301, no Consume Life)

### Terrain
- Dim light (candles)
- Broken cradle provides half cover
- 20x20 room

### Development by Round
**Round 1:** Maren screams, children swarm
**Round 2:** Maren heals 5 HP when child damages PC
**Round 3:** If Maren below 20 HP, children try to flee

### Alternative Resolution
- **Return Locket:** If PCs found Maren's Locket (Area M3) and place it in cradle, Maren stops attacking
- **Calm Emotions:** Spell gives advantage on Persuasion DC 16
```

### 6. grimorio-integrator (NUEVO - OBLIGATORIO)
**Propósito:** Verificar coherencia y completitud antes de compilación

**Fases:**
1. **Cross-Reference Check**
   - Verificar que cada criatura en áreas exista en bestiary
   - Verificar que cada NPC en áreas exista en npcs
   - Verificar que cada área referenciada exista
   - Verificar que cada item/key/objeto tenga ubicación definida

2. **Balance Audit**
   - Calcular XP total por acto
   - Verificar curva de dificultad
   - Identificar áreas sin tesoro que deberían tenerlo

3. **Consistency Check**
   - Verificar que NPCs no aparezcan en dos lugares a la vez
   - Verificar que keys/items no estén en múltiples lugares
   - Verificar conexiones bidireccionales entre áreas

4. **Completeness Check**
   - Cada área tiene al menos una criatura/NPC/mecánica
   - Cada área con combate tiene tesoro
   - Todos los CDs son numéricos
   - Todos los nombres están en bold

5. **Generate Summaries**
   - Tabla de NPCs por área
   - Tabla de tesoros por área
   - Mapa de conexiones
   - Lista de pistas/items clave

**Output:**
```json
{
  "status": "approved|warning|rejected",
  "checks": {
    "cross_references": "23/25 passed",
    "balance": "XP on target",
    "consistency": "1 warning: NPC appears in 2 areas",
    "completeness": "3 areas missing treasure"
  },
  "fixes": [
    "Added missing Specter to bestiary.md",
    "Fixed Area 7 connection to Area 3",
    "Added treasure to Areas 2, 5, 9"
  ]
}
```

### 7. grimorio-cartographer (Mejorado)
**Cambios:**
- Generar mapas con GRID de 5 pies
- Numerar áreas en el mapa (coincidiendo con áreas del texto)
- Generar versión DM (con secretos) y versión Player (sin secretos)
- Incluir compás (norte)
- Escala visible

### 8. grimorio-compiler v2 (Mejorado)
**Cambios:**
- Índice jerárquico con links
- Números de área destacados visualmente
- Referencias cruzadas clickeables ("see Area 3")
- Stat blocks inline para criaturas únicas
- Tabla de NPCs al inicio de cada acto
- Handouts automáticos:
  - Mapa del jugador
  - Lista de pistas encontradas
  - Resumen de NPCs conocidos

---

## Pipeline de Generación v2.0

### Fase 1: Fundación (Independiente)
```
[Usuario] → grimorio-lore → lore.md
                ↓
         grimorio-npc → npcs/npcs_and_factions.md
                ↓
         grimorio-bestiary → bestiary/bestiary.md
                ↓
         grimorio-maps → maps/maps.md (descripciones)
```

### Fase 2: Áreas (Depende de Fase 1)
```
[Usuario] → grimorio-areas → acts/act_01.md (10-15 áreas numeradas)
                ↓
         grimorio-encounters → encounters/encounters.md
                ↓
         grimorio-quests → quests/quests.md
```

### Fase 3: Integración (OBLIGATORIA)
```
[Usuario] → grimorio-integrator → validación + correcciones
                ↓
         grimorio-validate-canon → aprobación
                ↓
         grimorio-check-consistency → reporte
```

### Fase 4: Visuales (Depende de Fase 2-3)
```
[Usuario] → grimorio-cartographer → assets/*.svg (mapas con grid)
                ↓
         grimorio-artist → assets/*.webp (ilustraciones)
                ↓
         grimorio-dividers → assets/divider-*.svg
```

### Fase 5: Compilación
```
[Usuario] → grimorio-compiler-v2 → campaign.pdf
                ↓
         [PDF con: índice jerárquico, links, stat blocks, handouts]
```

---

## Cambios en Templates

### Template de Área (NUEVO)
```markdown
### {{NUMBER}}. {{NAME}}

{{2-3 lines of context}}

This area has the following features:
- {{Feature 1}}
- {{Feature 2}}
- {{Feature 3}}

***{{Feature/NPC Name}}.*** {{Description with specific DCs}}

***Treasure.*** {{Specific location}}: {{Amount}} gp, {{Items with value}}

***Secret {{Door/Trap}}.*** {{DC}} {{mechanism}} {{consequence}}
```

### Template de NPC (ACTUALIZADO)
```markdown
### {{NAME}}
*{{Alignment}} {{Gender}} {{Race}} {{Class}}*

**Location:** {{Initial location (Area code)}}
**Role:** {{Function in story}}
**Motivation:** {{What they want}}
**Secret:** {{Hidden information}}
**Quest:** {{Involved in which quest}}

**Stats:** {{Reference or inline stat block}}

**Quote:** "{{Characteristic line}}"
```

### Template de Bestiary (ACTUALIZADO)
```markdown
## {{NAME}}
*{{Size}} {{Type}}, {{Alignment}}*

**Armor Class** {{AC}}
**Hit Points** {{HP}}
**Speed** {{Speed}}

| STR | DEX | CON | INT | WIS | CHA |
|:---:|:---:|:---:|:---:|:---:|:---:|
| {{STR}} | {{DEX}} | {{CON}} | {{INT}} | {{WIS}} | {{CHA}} |

**Challenge** {{CR}} ({{XP}} XP)

***{{Trait}}.*** {{Description}}

### Tactics
{{How it fights}}

### Encounter Groups
- {{Difficulty}}: {{Creature combination}}
```

---

## Métricas de Éxito

### Métricas de Calidad del Contenido

| Métrica | v1.x Actual | v2.0 Objetivo | WDH Referencia |
|---------|-------------|---------------|----------------|
| Áreas por acto | 3-5 | 10-15 | 12-36 |
| Palabras por área | 350 | 150-200 | 150-250 |
| % áreas con mecánicas | 50% | 90%+ | 83% |
| % áreas con tesoro | 30% | 70%+ | 50% |
| CDs específicos | 40% | 100% | 100% |
| Referencias cruzadas | 10% | 80%+ | 90% |
| NPCs con stats | 0% | 60%+ | 70% |
| Tesoro con valor | 20% | 100% | 100% |

### Métricas del PDF

| Métrica | v1.x Actual | v2.0 Objetivo |
|---------|-------------|---------------|
| Índice jerárquico | No | Sí (3 niveles) |
| Referencias clickeables | No | Sí |
| Stat blocks inline | No | Sí |
| Mapas con grid | No | Sí |
| Versión DM/Player | No | Sí |
| Handouts automáticos | No | 3 tipos |
| Números de área visibles | No | Sí (destacados) |

### Métricas de Usabilidad

| Métrica | v1.x Actual | v2.0 Objetivo |
|---------|-------------|---------------|
| Tiempo de prep por sesión | 45-60 min | 15-20 min |
| Elementos que DM debe inventar | 70% | 10% |
| Búsqueda de stats durante juego | Alta | Mínima |
| Navegación durante juego | Difícil | Fácil |

---

## Timeline de Implementación

### Fase 1: Diseño (1 semana)
- [ ] Definir formatos finales de templates
- [ ] Diseñar arquitectura del integrador
- [ ] Especificar formato de stat blocks inline
- [ ] Diseñar sistema de referencias cruzadas

### Fase 2: Agentes (2 semanas)
- [ ] Reescribir grimorio-acts → grimorio-areas
- [ ] Crear grimorio-integrator
- [ ] Mejorar grimorio-npc (stats, alineamiento, location)
- [ ] Mejorar grimorio-bestiary (tácticas, encounter groups)
- [ ] Mejorar grimorio-encounters (desarrollo por rondas)

### Fase 3: Templates (1 semana)
- [ ] Crear template de área (nuevo)
- [ ] Actualizar template de NPC
- [ ] Actualizar template de bestiary
- [ ] Actualizar template de encounter
- [ ] Crear template de quest

### Fase 4: Compilador (2 semanas)
- [ ] Implementar índice jerárquico
- [ ] Implementar referencias cruzadas
- [ ] Implementar stat blocks inline
- [ ] Generar handouts automáticos
- [ ] Mejorar CSS para áreas numeradas

### Fase 5: Testing (1 semana)
- [ ] Generar campaña completa de prueba
- [ ] Comparar contra WDH métrica por métrica
- [ ] Testear con DM real (sesión de 3 horas)
- [ ] Iterar según feedback

### Fase 6: Release (3 días)
- [ ] Documentación de agentes
- [ ] Guía de migración v1→v2
- [ ] Release notes
- [ ] Anuncio

**Total: 7-8 semanas**

---

## Riesgos y Mitigaciones

### Riesgo 1: Complejidad del Integrador
**Problema:** El integrador puede volverse demasiado complejo y lento
**Mitigación:**
- Implementar por fases (primero cross-reference, luego balance)
- Usar caching de validaciones
- Permitir "override manual" para casos edge

### Riesgo 2: Calidad del Output de Áreas
**Problema:** El LLM puede seguir generando "escenas" en vez de "áreas"
**Mitigación:**
- Prompt engineering agresivo con ejemplos WotC
- Validación automática de formato
- Reject + retry si no sigue el patrón

### Riesgo 3: Tamaño del PDF
**Problema:** Más áreas = PDF más grande y lento
**Mitigación:**
- Compresión de imágenes
- Lazy loading de stat blocks (referencias al MM)
- Opción de "light PDF" (sin stat blocks inline)

### Riesgo 4: Backward Compatibility
**Problema:** Campañas v1.x no funcionarán con v2.0
**Mitigación:**
- Mantener modo "legacy" en compilador
- Script de migración automática
- Documentación clara de breaking changes

---

## Decisiones Técnicas

### Decisión 1: ¿Reemplazar o mantener grimorio-acts?
**Decisión:** REEMPLAZAR con grimorio-areas
**Razonamiento:** El concepto de "acto" es demasiado abstracto. "Área" es concreto y jugable.

### Decisión 2: ¿Stat blocks inline o referencias?
**Decisión:** REFERENCIAS por defecto, INLINE opcional
**Razonamiento:**
- Referencias mantienen el PDF ligero
- Inline es mejor para criaturas únicas del módulo
- El integrador decide cuáles poner inline

### Decisión 3: ¿Read-aloud en todas las áreas o solo algunas?
**Decisión:** SOLO en áreas importantes (20-30%)
**Razonamiento:**
- WotC lo usa así
- Evita fatiga del DM
- Las áareas comunes no necesitan read-aloud

### Decisión 4: ¿Cuánto tesoro por área?
**Decisión:** Tesoro en 70% de áreas con criaturas/NPCs
**Razonamiento:**
- WDH tiene 50% pero es urbano (menos combate)
- Dungeons típicos tienen más tesoro
- 70% es un buen balance

### Decisión 5: ¿Cómo manejar las "escenas sociales"?
**Decisión:** Tratarlas como "áreas sin combate"
**Razonamiento:**
- Una taberna es un "lugar" aunque no haya combate
- Debe tener NPCs, pistas, y mecánicas sociales
- El formato es el mismo, solo cambian las mecánicas

---

## Success Criteria

La v2.0 será considerada exitosa si:

1. **Un DM puede preparar una sesión en 15 minutos**
   - Leer el acto (10-15 áreas)
   - Entender las conexiones
   - Tener stats y tesoro listos

2. **Un DM puede dirigir sin improvisar mecánicas**
   - Todos los CDs están definidos
   - Todos los tesoros tienen valor
   - Todos los NPCs tienen comportamiento

3. **El PDF es navegable sin esfuerzo**
   - Índice jerárquico funcional
   - Links entre áreas
   - Búsqueda rápida

4. **La calidad es comparable a aventuras oficiales**
   - Métricas dentro del 20% de WDH
   - Feedback de DMs positivo
   - Reutilizable para múltiples campañas

---

## Archivos Relacionados

- `/tmp/opencode/analisis-calidad-pdf.md` — Análisis inicial
- `/tmp/opencode/analisis-detallado-wdh-vs-grimorio.md` — Análisis detallado
- `/home/pau/Grimorio/agents/grimorio-acts.md` — Agente mejorado
- `/home/pau/Grimorio/agents/grimorio-integrator.md` — Nuevo agente
- `/home/pau/Waterdeep_ Dragon Heist.md` — Referencia WotC (8004 líneas)
