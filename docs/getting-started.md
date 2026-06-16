---
title: Getting Started / Empezando
lang: en/es
---

<div class="lang-selector">
<a href="#english">English</a> | <a href="#espanol">Español</a>
</div>

---

<a name="english"></a>

# Getting Started with Grimorio

## Quick Install

```bash
curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | bash
```

This installs:
- Go 1.24+ (if missing)
- Chrome/Chromium/Edge or wkhtmltopdf (for PDF compilation)
- Grimorio binary to `~/.local/bin/grimorio`
- Plugin files to your AI assistant's plugin directory

## Requirements

| Dependency | Auto-installed | Purpose |
|------------|---------------|---------|
| Go 1.24+ | ✅ Yes | Build the MCP server binary |
| Chrome/Chromium | ⚠️ Check install | Compile HTML to PDF (preferred) |
| wkhtmltopdf | ⚠️ Fallback | Compile HTML to PDF (legacy) |
| Git | ❌ Must have | Clone the repository |

## Your First Campaign

Once installed, type in your AI assistant chat:

```
/grimorio A sunken city where the nobles are aquatic vampires
```

The `grimorio-architect` agent will first ask you to pick your language (English or Spanish), then ask 6 more questions:
0. **Language** (English or Spanish, default English)
1. Campaign name (kebab-case, e.g., "sunken-city")
2. One-shot or full campaign?
3. Player level range?
4. Desired tone?
5. Duration?

Then it generates everything automatically:
- Lore, NPCs, bestiary, encounters, maps
- Acts with numbered areas (WotC format)
- AI-generated cover art and illustrations (cached automatically)
- Procedural battle maps (SVG)
- Professional PDF (D&D-styled layout)

### Campaign Templates

You can also pre-fill the answers with a template: Urban Fantasy, Gothic Horror, Maritime Adventure, Dungeon Crawl, or Political Intrigue. The template sets sensible defaults for tone, level range, and game mode.

## Story Brief Template

```
**Campaign Name:** (kebab-case)
**Setting:** (one-sentence premise)
**Tone:** (heroic, dark, humorous, political, horror, mystery)
**Level Range:** (1-3, 4-6, 7-10, 11-15, 16-20)
**Duration:** (one-shot, 3-5 sessions, long campaign)
**Themes:** (comma-separated)
**Villain Type:** (optional)
**McGuffin:** (optional)
```

## Next Steps

- **[Session Tutorial](session-tutorial.md)** — Run your first session
- **[Character Creation](character-creation.md)** — Generate PCs
- **[Session Generator](session-generator.md)** — Adapt sessions to your party
- **[PDF Compiler](../features/pdf-compiler.md)** — Customize PDF output
- **[MCP Tools](../features/mcp-tools.md)** — Full tool reference
- **[Architecture](../features/architecture.md)** — How Grimorio works

## Validate your campaign

Before you compile to PDF, run the consistency gate:

```bash
grimorio validate --scope=all my-campaign
```

The CLI runs 17 rules (lore integrity, narrative-state parity, faction
reputation, WotC format) and prints a PASS/WARN/FAIL summary. Exit
code is 0 when clean, 1 when there are errors, 2 on usage problems.

```bash
# Just the WotC rules
grimorio validate --scope=wotc my-campaign

# Machine-readable output for CI / scripts
grimorio validate --scope=all --json my-campaign > report.json
```

For a real-world example of a full campaign going through the gate,
read the **[La Hoja de Vlad walkthrough](walkthroughs/la-hoja-de-vlad.md)**.

## Export to Other Formats

In addition to PDF, you can export your campaign to:

```bash
# Concatenated Markdown (canonical file order)
grimorio export_campaign --campaign my-campaign --format=markdown

# EPUB 3 (valid e-reader format with OPF, NCX, XHTML)
grimorio export_campaign --campaign my-campaign --format=epub
```

PDF remains the default and most-styled output. Markdown is great for sharing in version control or editing in Obsidian. EPUB is for e-readers and mobile reading.

## Check Campaign Health

Get a 0-100 score across six axes of campaign quality:

```
Use the campaign_health_dashboard MCP tool from any agent
```

The dashboard measures:

| Axis | What it measures |
|------|------------------|
| **Overall Health** | Weighted average of all axes |
| **Canon Completeness** | Are all referenced entities defined? |
| **Narrative Coherence** | Are timeline + decisions consistent? |
| **Faction Balance** | Are factions distributed sensibly? |
| **WotC Compliance** | Do areas/hooks/box text meet WotC format? |
| **Hook Coverage** | Are there enough character hooks per area? |

## Image Generation Cache

Generated images are cached automatically at `~/.cache/grimorio/images/` using a SHA-256 key derived from prompt + model + dimensions + provider. Re-running the same prompt returns instantly from the cache. The MCP `force_regenerate` parameter bypasses the cache when you need a fresh result.

## Campaign Storage

Campaigns are stored in `~/campaigns/` by default:

```
~/campaigns/
└── sunken-city/
    ├── campaign.pdf          # Final PDF
    ├── campaign.html         # HTML version
    ├── lore.md               # World backstory
    ├── areas/                 # Chapters with areas
    ├── npcs/                 # NPCs and factions
    ├── bestiary/             # Monster stat blocks
    ├── characters/           # Character sheets
    └── assets/               # Maps and AI art
```

To change the default location:

```bash
export CAMPAIGN_ROOT="/path/to/your/campaigns"
```

---

<a name="espanol"></a>

# Empezando con Grimorio

## Instalación Rápida

```bash
curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | bash
```

Esto instala:
- Go 1.24+ (si falta)
- Chrome/Chromium/Edge o wkhtmltopdf (para compilar PDFs)
- Binario de Grimorio en `~/.local/bin/grimorio`
- Archivos del plugin en el directorio de plugins de tu asistente IA

## Requisitos

| Dependencia | Auto-instalado | Propósito |
|------------|---------------|---------|
| Go 1.24+ | ✅ Sí | Compilar el binario del servidor MCP |
| Chrome/Chromium | ⚠️ Verificar instalación | Compilar HTML a PDF (preferido) |
| wkhtmltopdf | ⚠️ Fallback | Compilar HTML a PDF (legacy) |
| Git | ❌ Debes tenerlo | Clonar el repositorio |

## Tu Primera Campaña

Una vez instalado, escribe en el chat de tu asistente IA:

```
/grimorio Una ciudad hundida donde los nobles son vampiros acuáticos
```

El agente `grimorio-architect` primero te preguntará el idioma (inglés o español), luego te hará 6 preguntas más:
0. **Idioma** (inglés o español, default inglés)
1. Nombre de campaña (kebab-case, ej. "ciudad-hundida")
2. ¿One-shot o campaña completa?
3. ¿Rango de nivel de jugadores?
4. ¿Tono deseado?
5. ¿Duración?

Luego genera todo automáticamente:
- Trasfondo, NPCs, bestiario, encuentros, mapas
- Actos con áreas numeradas (formato WotC)
- Portada e ilustraciones generadas por IA (cacheadas automáticamente)
- Mapas de batalla procedimentales (SVG)
- PDF profesional (estilo D&D)

### Templates de Campaña

También podés pre-rellenar las respuestas con un template: Urban Fantasy, Gothic Horror, Maritime Adventure, Dungeon Crawl o Political Intrigue. El template setea defaults razonables para tono, rango de nivel y modo de juego.

## Plantilla de Brief de Historia

```
**Nombre de Campaña:** (kebab-case)
**Ambientación:** (premisa de una oración)
**Tono:** (heroico, oscuro, humorístico, político, terror, misterio)
**Rango de Nivel:** (1-3, 4-6, 7-10, 11-15, 16-20)
**Duración:** (one-shot, 3-5 sesiones, campaña larga)
**Temas:** (separados por comas)
**Tipo de Villano:** (opcional)
**McGuffin:** (opcional)
```

## Próximos Pasos

- **[Tutorial de Sesión](session-tutorial.md)** — Ejecuta tu primera sesión
- **[Creación de Personajes](character-creation.md)** — Genera PJs
- **[Generador de Sesiones](session-generator.md)** — Adapta sesiones a tu grupo
- **[Compilador PDF](../features/pdf-compiler.md)** — Personaliza el PDF
- **[Herramientas MCP](../features/mcp-tools.md)** — Referencia completa de herramientas
- **[Arquitectura](../features/architecture.md)** — Cómo funciona Grimorio

## Validá tu Campaña

Antes de compilar a PDF, corré el consistency gate:

```bash
grimorio validate --scope=all mi-campana
```

El CLI corre 17 reglas (lore integrity, narrative-state parity, faction
reputation, formato WotC) e imprime un resumen PASS/WARN/FAIL. El exit
code es 0 cuando está limpio, 1 cuando hay errores, 2 en problemas de
uso.

```bash
# Solo las reglas WotC
grimorio validate --scope=wotc mi-campana

# Output machine-readable para CI / scripts
grimorio validate --scope=all --json mi-campana > report.json
```

Para un ejemplo real de una campaña completa pasando por el gate, leé
el **[walkthrough de La Hoja de Vlad](walkthroughs/la-hoja-de-vlad.md)**.

## Exportá a Otros Formatos

Además de PDF, podés exportar tu campaña a:

```bash
# Markdown concatenado (orden canónico de archivos)
grimorio export_campaign --campaign mi-campana --format=markdown

# EPUB 3 (formato válido de e-reader con OPF, NCX, XHTML)
grimorio export_campaign --campaign mi-campana --format=epub
```

PDF sigue siendo el default y el más estilizado. Markdown es ideal para compartir en control de versiones o editar en Obsidian. EPUB es para e-readers y lectura móvil.

## Chequeá la Salud de Campaña

Obtené un puntaje 0-100 en seis ejes de calidad de campaña:

```
Usá la tool MCP campaign_health_dashboard desde cualquier agente
```

El dashboard mide:

| Eje | Qué mide |
|-----|----------|
| **Overall Health** | Promedio ponderado de todos los ejes |
| **Canon Completeness** | ¿Están definidas todas las entidades referenciadas? |
| **Narrative Coherence** | ¿Son consistentes timeline + decisiones? |
| **Faction Balance** | ¿Están las facciones distribuidas razonablemente? |
| **WotC Compliance** | ¿Cumplen las áreas/hooks/box text el formato WotC? |
| **Hook Coverage** | ¿Hay suficientes hooks de personaje por área? |

## Caché de Generación de Imágenes

Las imágenes generadas se cachean automáticamente en `~/.cache/grimorio/images/` usando una clave SHA-256 derivada de prompt + modelo + dimensiones + provider. Re-ejecutar el mismo prompt devuelve instantáneamente desde la caché. El parámetro MCP `force_regenerate` bypasea la caché cuando necesitás un resultado fresco.

## Almacenamiento de Campañas

Las campañas se guardan en `~/campaigns/` por defecto:

```
~/campaigns/
└── ciudad-hundida/
    ├── campaign.pdf          # PDF final
    ├── campaign.html         # Versión HTML
    ├── lore.md               # Trasfondo del mundo
    ├── areas/                 # Capítulos con áreas
    ├── npcs/                 # NPCs y facciones
    ├── bestiary/             # Estadísticas de monstruos
    ├── characters/           # Hojas de personaje
    └── assets/               # Mapas y arte de IA
```

Para cambiar la ubicación por defecto:

```bash
export CAMPAIGN_ROOT="/path/to/your/campaigns"
```
