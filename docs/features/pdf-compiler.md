# PDF Compiler / Compilador PDF

## Overview / Descripción General

Grimorio's PDF compiler converts your campaign markdown files into a professional, print-ready PDF with D&D-styled layouts inspired by Wizards of the Coast manuals.

El compilador PDF de Grimorio convierte tus archivos markdown de campaña en un PDF profesional, listo para imprimir, con diseños estilo D&D inspirados en los manuales de Wizards of the Coast.

---

## Features / Características

### What's Included / Qué Está Incluido

| Feature / Característica | Description / Descripción |
|--------------------------|--------------------------|
| **Two-Column Layout** | Professional D&D magazine-style layout / Diseño profesional tipo revista D&D |
| **WotC-Style CSS** | Fonts (Cinzel, Lora), colors, borders inspired by official modules / Fuentes, colores, bordes inspirados en módulos oficiales |
| **Table of Contents** | Hierarchical TOC with page references / TOC jerárquico con referencias de página |
| **Cover Page** | Full-page cover with AI-generated art / Portada de página completa con arte de IA |
| **Stat Blocks** | Professional monster/NPC stat block formatting / Formato profesional de estadísticas |
| **Read-Aloud Text** | Boxed text sections for DM narration / Secciones de texto enmarcado para narración del DM |
| **DM Sidebars** | Tips, secrets, and guidance (DM-only) / Consejos, secretos y guía (solo DM) |
| **Shock Points** | 3-level horror/cosmic horror tracking system / Sistema de seguimiento de horror de 3 niveles |
| **Character Sheets** | Full character appendix with stats and backstory / Apéndice completo de personajes con estadísticas y trasfondo |
| **Quests** | Quest appendix with objectives and decision trees / Apéndice de misiones con objetivos y árboles de decisión |
| **Image Embedding** | SVG maps and AI art embedded inline / Mapas SVG y arte de IA incrustados en línea |
| **Handouts** | Player-facing and DM-only versions / Versiones para jugadores y solo DM |

---

## Usage / Uso

### Compile Campaign / Compilar Campaña

```bash
/grimorio compile_pdf campaign="sunken-city"
```

Or with custom title:

O con título personalizado:

```bash
/grimorio compile_pdf campaign="sunken-city" title="La Piedra de la Universidad"
```

### Compiler Version / Versión del Compilador

```bash
# Version 2 (default) — Hierarchical TOC, cross-refs, handouts
/grimorio compile_pdf campaign="sunken-city"

# Version 1 — Legacy flat structure
/grimorio compile_pdf campaign="sunken-city" compiler_version=1
```

---

## PDF Structure / Estructura del PDF

### Compilation Order / Orden de Compilación

```
1. Cover Page / Portada
   └─ Title, subtitle, cover art / Título, subtítulo, arte de portada

2. Table of Contents / Índice
   └─ Hierarchical with page numbers / Jerárquico con números de página

3. Session Zero / Sesión Cero
   └─ Safety tools, house rules, character creation / Herramientas de seguridad, reglas de casa, creación de personajes

4. Introduction / Introducción
   └─ Campaign overview, tone, summary / Resumen de campaña, tono, resumen

5. Lore y Ambientación / Lore and Setting
   └─ World backstory, history, conflicts / Trasfondo del mundo, historia, conflictos

6. Chapters (Areas) / Capítulos (Áreas)
   ├─ Act 1: 10-15 numbered areas / Acto 1: 10-15 áreas numeradas
   ├─ Act 2: 10-15 numbered areas / Acto 2: 10-15 áreas numeradas
   └─ Act 3: 10-15 numbered areas / Acto 3: 10-15 áreas numeradas

7. Setting Guide (DM-Only) / Guía de Ambientación (Solo DM)
   └─ Spoilers, secrets, DM guidance / Spoilers, secretos, guía para DM

8. Appendices / Apéndices
   ├─ Apéndice A: NPCs y Facciones / NPCs and Factions
   ├─ Apéndice B: Bestiario / Bestiary
   ├─ Apéndice C: Encuentros / Encounters
   ├─ Apéndice D: Mapas de Referencia / Reference Maps
   ├─ Apéndice E: Faction Tracker (if applicable) / Rastreador de Facciones
   ├─ Apéndice F: Adventure Roster (if applicable) / Lista de Aventuras
   ├─ Apéndice G: Character Sheets / Hojas de Personaje
   └─ Apéndice H: Quests / Misiones

9. Handouts (if generated) / Handouts (si se generaron)
   └─ Player-facing and DM-only versions / Versiones para jugadores y solo DM
```

---

## CSS Styles / Estilos CSS

### WotC-Style Classes / Clases Estilo WotC

#### DM Sidebars / Barras Laterales de DM

```css
.dm-sidebar
```

**Usage / Uso:**
```markdown
> ##### DM Only: Running the Scene
> 
> **Prep:** Read areas 5-7 before the session.
> **Pacing:** Start with tension, build to climax.
> **Player Signals:** If they hesitate, drop a clue.
```

**Appearance / Apariencia:**
- Red left border (4px) / Borde izquierdo rojo (4px)
- "DM Only" label / Etiqueta "Solo DM"
- Background: light parchment / Fondo: pergamino claro

#### Stat Blocks v2 / Bloques de Estadísticas v2

```css
.stat-block-v2
```

**Features / Características:**
- Gradient background / Fondo degradado
- Enhanced shadow / Sombra mejorada
- Professional borders / Bordes profesionales

#### Shock Points / Puntos de Shock

```css
.shock-point-mild
.shock-point-moderate
.shock-point-intense
```

**Usage / Uso:**
```markdown
<div class="shock-point-mild">
**Mild Shock (1 PS):** Witnessing a corpse
</div>

<div class="shock-point-moderate">
**Moderate Shock (2 PS):** Seeing the egg hatch
</div>

<div class="shock-point-intense">
**Intense Shock (3 PS):** Direct contact with the Dark Witness
</div>
```

#### Session Prep Cards / Tarjetas de Preparación de Sesión

```css
.session-prep-card
```

**Usage / Uso:**
```markdown
<div class="session-prep-card">
## Session 1: The Heist
**Focus:** University infiltration
**Characters:** Raika, Samuel
</div>
```

#### Character Worksheet / Hoja de Trabajo de Personaje

```css
.character-worksheet
```

**Features / Características:**
- Fillable fields / Campos completables
- Backstory prompts / Prompts de trasfondo
- Personality selectors / Selectores de personalidad

#### Encounter Recommendations / Recomendaciones de Encuentros

```css
.encounter-recommendation
```

**Features / Características:**
- CR badges / Insignias de CR
- Type indicators / Indicadores de tipo
- Party-level adjustments / Ajustes de nivel de grupo

---

## Customization / Personalización

### Add Custom CSS / Añadir CSS Personalizado

Edit `/internal/compiler/templates/dnd-style.css`:

```css
/* Custom sidebar style / Estilo de sidebar personalizado */
.my-custom-sidebar {
  background: #f0e6d3;
  border-left: 4px solid #8b0000;
  padding: 8px 12px;
  margin: 10px 0;
}
```

### Change Fonts / Cambiar Fuentes

Edit the CSS `@import` statement:

Edita la declaración `@import` del CSS:

```css
@import url('https://fonts.googleapis.com/css2?family=Cinzel:wght@400;700&family=Lora:ital,wght@0,400;0,700;1,400&display=swap');
```

### Adjust Colors / Ajustar Colores

Main color variables / Variables de color principales:

```css
:root {
  --primary-color: #5a3d2b;    /* Brown / Marrón */
  --accent-color: #c9ad6a;     /* Gold / Dorado */
  --text-color: #1a1a1a;       /* Dark gray / Gris oscuro */
  --background: #f5f0e6;       /* Parchment / Pergamino */
}
```

---

## Troubleshooting / Solución de Problemas

### "Images not appearing in PDF" / "Las imágenes no aparecen en el PDF"

**Cause / Causa:** wkhtmltopdf can't access local files / wkhtmltopdf no puede acceder a archivos locales

**Solution / Solución:**
```bash
# Recompile with --enable-local-file-access
/grimorio compile_pdf campaign="sunken-city" --enable-local-file-access
```

### "PDF too large" / "PDF muy grande"

**Cause / Causa:** Too many high-resolution images / Demasiadas imágenes de alta resolución

**Solution / Solución:**
- Reduce image count / Reduce cantidad de imágenes
- Use SVG instead of PNG for maps / Usa SVG en lugar de PNG para mapas
- Compress AI art / Comprime arte de IA

### "Characters/Quests not in PDF" / "Personajes/Misiones no están en el PDF"

**Cause / Causa:** Sections not included in compiler / Secciones no incluidas en el compilador

**Solution / Solución:**
Make sure your `compiler.go` includes these sections:

Asegúrate de que tu `compiler.go` incluya estas secciones:

```go
{"Apéndice G: Character Sheets", filepath.Join(c.CampaignDir, "characters"), true},
{"Apéndice H: Quests", filepath.Join(c.CampaignDir, "quests"), true},
```

### "CSS not applying" / "El CSS no se aplica"

**Cause / Causa:** CSS file not embedded / Archivo CSS no incrustado

**Solution / Solución:**
Check that `dnd-style.css` is in `/internal/compiler/templates/`:

Revisa que `dnd-style.css` esté en `/internal/compiler/templates/`:

```bash
ls /internal/compiler/templates/dnd-style.css
```

---

## Examples / Ejemplos

### Example: Campaign with All Features / Ejemplo: Campaña con Todas las Características

```bash
# Generate full campaign / Genera campaña completa
/grimorio "A university where students study forbidden magic"

# Generate characters / Genera personajes
/grimorio generate_character campaign="forbidden-university" name="Raika" race="elfo" class="druida" level=4
/grimorio generate_character campaign="forbidden-university" name="Samuel" race="humano" class="mago" level=4

# Generate session prep / Genera preparación de sesión
/grimorio generate_session_prep campaign="forbidden-university" session_num=1 characters="Raika,Samuel"

# Compile PDF / Compila PDF
/grimorio compile_pdf campaign="forbidden-university"
```

**Output / Salida:**
- `campaign.pdf` (2-5 MB, 50-150 pages) / (2-5 MB, 50-150 páginas)
- Includes: Characters, Quests, Session Prep, WotC-style CSS / Incluye: Personajes, Misiones, Prep de Sesión, CSS estilo WotC

### Example: One-Shot with Pre-Gens / Ejemplo: One-Shot con Pre-generados

```bash
# Generate one-shot / Genera one-shot
/grimorio "A heist to steal an arcane egg from a university"

# Generate 4 pre-generated characters / Genera 4 personajes pre-generados
/grimorio generate_character campaign="heist" name="Rook" race="humano" class="picaro" level=4
/grimorio generate_character campaign="heist" name="Sera" race="tiefling" class="brujo" level=4
/grimorio generate_character campaign="heist" name="Gromm" race="enano" class="guerrero" level=4
/grimorio generate_character campaign="heist" name="Lyra" race="elfo" class="hechicero" level=4

# Compile PDF / Compila PDF
/grimorio compile_pdf campaign="heist"
```

**Output / Salida:**
- `campaign.pdf` (1-3 MB, 30-60 pages) / (1-3 MB, 30-60 páginas)
- Includes: 4 character sheets, 1 quest, 1 act / Incluye: 4 hojas de personaje, 1 misión, 1 acto

---

## Performance / Rendimiento

### Compilation Times / Tiempos de Compilación

| Campaign Size / Tamaño de Campaña | Pages / Páginas | Time / Tiempo |
|----------------------------------|-----------------|---------------|
| One-shot / One-shot | 30-60 | 5-10 seconds / segundos |
| Short campaign / Campaña corta | 60-100 | 10-20 seconds / segundos |
| Full campaign / Campaña completa | 100-200 | 20-40 seconds / segundos |
| Epic campaign / Campaña épica | 200+ | 40-60 seconds / segundos |

### PDF Size / Tamaño de PDF

| Content / Contenido | Size / Tamaño |
|---------------------|---------------|
| Text only / Solo texto | 500 KB - 1 MB |
| With SVG maps / Con mapas SVG | 1-2 MB |
| With AI art (5-10 images) / Con arte de IA (5-10 imágenes) | 2-5 MB |
| With AI art (20+ images) / Con arte de IA (20+ imágenes) | 5-10 MB |

---

## Next Steps / Próximos Pasos

- **[Session Tutorial](tutorials/session-tutorial.md)** — Run your first session / Ejecuta tu primera sesión
- **[Character Creation](tutorials/character-creation.md)** — Generate PCs / Genera PJs
- **[Session Generator](tutorials/session-generator.md)** — Adapt sessions to characters / Adapta sesiones a personajes
- **[MCP Tools](mcp-tools.md)** — Full tool reference / Referencia completa de herramientas
