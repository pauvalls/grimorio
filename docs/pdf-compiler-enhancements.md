# PDF Compiler Enhancements - Guía de Uso

Esta documentación describe las nuevas características agregadas al compilador de PDF de Grimorio.

---

## Tabla de Contenidos

1. [Puntos de Shock (Shock Points)](#puntos-de-shock)
2. [DM Sidebars](#dm-sidebars)
3. [Session Prep con Recomendaciones](#session-prep-con-recomendaciones)
4. [Character Sheets Expandidas](#character-sheets-expandidas)
5. [Stat Blocks v2](#stat-blocks-v2)
6. [Compilación de PDF](#compilación-de-pdf)
7. [Personalización de Templates](#personalización-de-templates)
8. [FAQ](#faq)

---

## Puntos de Shock

Los **Puntos de Shock** son advertencias de contenido que identifican temas potencialmente sensibles en tu campaña.

### ¿Qué son?

Cada punto de shock incluye:
- **Tipo**: El tema del contenido (violencia, horror, etc.)
- **Severidad**: mild (leve), moderate (moderado), intense (intenso)
- **Descripción**: Explicación detallada del contenido
- **Herramientas de seguridad**: Herramientas recomendadas para manejar el contenido

### Cómo generar Puntos de Shock

Los puntos de shock se generan automáticamente basados en los **temas** de tu campaña:

```bash
# Los temas comunes incluyen:
- violence (violencia)
- horror (horror)
- dark themes (temas oscuros)
- death (muerte)
- psychological (horror psicológico)
- gore
- phobias (fobias)
- substance abuse (abuso de sustancias)
```

### Ejemplo en Session Zero

```markdown
<div class="shock-point moderate">
<span class="severity-badge">moderate</span>
<strong>Violencia</strong>: Combate fantástico, descripciones de heridas.

**Herramientas de seguridad recomendadas:** X-Card, Fade to black
</div>
```

### Niveles de Severidad

| Severidad | Color | Descripción |
|-----------|-------|-------------|
| mild | Amarillo | Contenido leve, fácilmente evitable |
| moderate | Naranja | Contenido moderado, puede requerir discusión |
| intense | Rojo | Contenido intenso, requiere herramientas de seguridad |

---

## DM Sidebars

Las **DM Sidebars** son cajas de contenido exclusivo para el Director de Juego, con tips, secretos y reglas opcionales.

### Uso

```markdown
<div class="dm-sidebar">
<h5>DM Tip</h5>
<p>Los jugadores pueden encontrar una trampa aquí si investigan el trono.</p>

<h5>Secreto</h5>
<p>El NPC es en realidad un doppelganger que reemplazó al verdadero hace semanas.</p>
</div>
```

### Características

- Borde izquierdo rojo (#8b0000)
- Etiqueta "DM Only" automática
- Fondo beige diferenciado
- No se rompe entre páginas (page-break-inside: avoid)

### Cuándo usar

- **Tips de DM**: Consejos para manejar encuentros o situaciones
- **Secretos**: Información que los jugadores no deben ver
- **Reglas opcionales**: Variantes o house rules para esta sección
- **Notas de narrativa**: Recordatorios de arcos de historia

---

## Session Prep con Recomendaciones

La preparación de sesiones ahora incluye recomendaciones automáticas de encuentros, loot y apariciones de NPCs.

### Generar Session Prep Extendida

```bash
# Con MCP
generate_session_prep campaign_id="mi-campana" session_num=3 with_scenarios=true
```

### Secciones Incluidas

1. **Previously On**: Resumen narrativo de la sesión anterior
2. **Quests Activas**: Lista de quests en progreso
3. **NPCs Relevantes**: NPCs importantes para esta sesión
4. **Escenarios Probables**: Posibles发展方向 de la narrativa
5. **Recomendaciones de Encuentros**: 2-4 encuentros sugeridos
6. **Sugerencias de Loot**: Recompensas apropiadas por nivel
7. **Apariciones de NPCs**: NPCs que pueden aparecer
8. **Recordatorios**: Items pendientes o consecuencias

### Tipos de Encuentros

| Tipo | Descripción |
|------|-------------|
| combat | Encuentro de combate |
| social | Interacción social/negociación |
| exploration | Exploración/descubrimiento |
| mixed | Combate + elementos sociales |

### Loot por Tier

| Tier | Nivel | Rareza típica |
|------|-------|---------------|
| Tier 1 | 1-4 | Common, Uncommon |
| Tier 2 | 5-10 | Uncommon, Rare |
| Tier 3 | 11-16 | Rare, Very Rare |
| Tier 4 | 17-20 | Very Rare, Legendary |

---

## Character Sheets Expandidas

Las hojas de personaje ahora incluyen información narrativa completa.

### Generar Personaje con Backstory

```bash
# Con MCP
generate_character campaign="mi-campana" name="Aldric" race="humano" class="guerrero" level=1 background="soldado" alignment="lawful good" with_backstory=true
```

### Secciones Nuevas

1. **Backstory Hooks**: Ganchos narrativos que conectan al personaje con la campaña
2. **Secretos**: Información oculta que el personaje guarda
3. **Metas**: Objetivos personales del personaje
4. **Personalidad Expandida**:
   - Traits (rasgos)
   - Ideals (ideales)
   - Bonds (vínculos)
   - Flaws (defectos)
5. **Hechizos**: Lista organizada por nivel para spellcasters

### Ejemplo de Backstory Hooks

```markdown
<div class="character-worksheet">
<div class="worksheet-section">
<h4>Ganchos de Historia</h4>
<div class="prompt-box">• Veterano de una guerra olvidada</div>
<div class="prompt-box">• Busca redención por un fallo en el combate</div>
</div>
</div>
```

---

## Stat Blocks v2

Los **Stat Blocks v2** son bloques de estadísticas mejorados para NPCs y monstruos.

### Uso

```markdown
<div class="stat-block-v2">
<h3>Goblin Boss</h3>
<div class="stat-line">
<span class="stat-label">Armor Class</span>
<span class="stat-value">17 (chain shirt, shield)</span>
</div>
<div class="stat-line">
<span class="stat-label">Hit Points</span>
<span class="stat-value">21 (5d8)</span>
</div>
<div class="stat-line">
<span class="stat-label">Speed</span>
<span class="stat-value">30 ft.</span>
</div>
</div>
```

### Características Visuales

- Gradiente de fondo (beige claro a oscuro)
- Bordes dobles superior e inferior (#8b0000)
- Shadow para profundidad
- Líneas de stats con labels y valores alineados

---

## Compilación de PDF

### Requisitos

- Chrome, Chromium, Edge, o `wkhtmltopdf` instalado y en PATH (Grimorio auto-detecta el motor disponible)
- Todas las secciones de la campaña en formato markdown

### Comandos

```bash
# Compilar campaña completa
grimorio compile --campaign mi-campana --title "Mi Campaña Épica"

# El PDF se genera en: mi-campana/campaign.pdf
```

### Secciones Incluidas Automáticamente

1. Cover page (con cover image si existe)
2. Session Zero (con Shock Points)
3. Table of Contents
4. Introduction
5. Lore y Ambientación
6. Chapters (Areas)
7. Setting Guide
8. Apéndices (NPCs, Bestiary, Encounters, Maps)
9. Faction Tracker
10. Adventure Roster
11. Handouts (v2)

### Solución de Problemas

**Error: "No PDF engine found"**
```bash
# Instalar Chrome/Chromium (preferido)
# Arch/CachyOS
sudo pacman -S chromium

# Ubuntu/Debian
sudo apt-get install chromium-browser

# macOS
brew install --cask google-chrome

# O wkhtmltopdf como fallback
sudo pacman -S wkhtmltopdf  # Arch
brew install wkhtmltopdf     # macOS
```

**Error: "images missing"**
- Verificar que las imágenes estén en `assets/`
- Usar rutas relativas en markdown: `![alt](assets/image.png)`

---

## Personalización de Templates

### Templates Disponibles

| Template | Ubicación | Descripción |
|----------|-----------|-------------|
| Session Zero | `templates/session-zero.md.tmpl` | Guía de sesión cero |
| Session Prep | `templates/session-prep.md.tmpl` | Preparación de sesión |
| Character Sheet | `templates/character-sheet.md.tmpl` | Hoja de personaje |

### Modificar Templates

1. Copiar el template a tu directorio de campaña
2. Editar según necesidades
3. El compilador usa los templates embebidos por defecto

### Ejemplo: Personalizar Session Zero

```markdown
# Tu Session Zero Personalizado

## Tus Reglas de Casa

{{.HouseRules}}

## Tus Puntos de Shock Personalizados

{{range .ShockPoints}}
<div class="shock-point {{.Severity}}">
<span class="severity-badge">{{.Severity}}</span>
<strong>{{.Type}}</strong>: {{.Description}}
</div>
{{end}}
```

---

## FAQ

### ¿Los cambios son backward compatible?

**Sí.** Todas las nuevas características son aditivas:
- Campañas existentes compilan sin cambios
- Campos nuevos son opcionales (`omitempty` en JSON)
- Templates usan condicionales para datos faltantes

### ¿Cómo desactivo los Shock Points?

No los incluyas en tu Session Zero. El template muestra un mensaje genérico si no hay shock points definidos.

### ¿Puedo usar solo algunas características nuevas?

**Sí.** Cada característica es independiente:
- Usá DM Sidebars sin Session Prep
- Usá Character Sheets sin Backstory Hooks
- Usá Stat Blocks v2 sin encontrar recomendaciones

### ¿Cómo genero recomendaciones de encuentros manuales?

Editá el archivo `session-prep.md` directamente:

```markdown
## Recomendaciones de Encuentros

<div class="encounter-recommendation">
<span class="cr-badge">CR 2</span>
<span class="encounter-type">combat</span>
<strong>Emboscada Goblin</strong>
<p>Los goblins atacan desde los árboles.</p>
</div>
```

### ¿Los PDFs son imprimibles?

**Sí.** Todas las clases CSS respetan:
- `page-break-inside: avoid` - No corta contenido a la mitad
- `column-break-inside: avoid` - Respeta layout de columnas
- Márgenes apropiados para impresión A4

### ¿Cómo agrego más razas/clases al generador de personajes?

Editá los mapas en `internal/services/character_service.go`:
- `raceBonuses` - Bonos raciales
- `classPrimaryStats` - Stats primarios por clase
- `classSkills` - Skills por clase
- `classFeatures` - Features por clase

---

## Referencia de Clases CSS

### Nuevas Clases

| Clase | Descripción |
|-------|-------------|
| `.dm-sidebar` | Sidebar para contenido DM-only |
| `.stat-block-v2` | Stat block mejorado |
| `.session-prep-card` | Tarjeta para session prep |
| `.shock-point` | Advertencia de contenido |
| `.shock-point.mild` | Severidad leve (amarillo) |
| `.shock-point.moderate` | Severidad moderada (naranja) |
| `.shock-point.intense` | Severidad intensa (rojo) |
| `.severity-badge` | Badge de severidad |
| `.character-worksheet` | Worksheet de personaje |
| `.worksheet-section` | Sección de worksheet |
| `.prompt-box` | Caja de prompt |
| `.encounter-recommendation` | Recomendación de encuentro |
| `.cr-badge` | Badge de CR |
| `.encounter-type` | Badge de tipo de encuentro |
| `.stat-line` | Línea de stat en stat-block-v2 |
| `.stat-label` | Label de stat |
| `.stat-value` | Valor de stat |
| `.prep-item` | Item en session prep card |

---

## Recursos Adicionales

- [Spec del Diseño](sdd/pdf-compiler-enhancements/spec.md)
- [Design Document](sdd/pdf-compiler-enhancements/design.md)
- [Tasks](sdd/pdf-compiler-enhancements/tasks.md)
