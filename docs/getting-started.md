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
curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | sh
```

This installs:
- The Grimorio binary in `~/.grimorio/grimorio`, symlinked to `~/.local/bin/grimorio`
- The OpenCode integration: `mcp.servers.grimorio`, skills, agents and the `/grimorio` command
- Nothing else. It checks for a PDF engine but never installs one.

## Requirements

| Dependency | Needed for | Notes |
|------------|-----------|-------|
| Python 3 | Install and update | Merges `opencode.json` without corrupting it |
| Chrome / Chromium | PDF compilation | Preferred engine |
| wkhtmltopdf | PDF compilation | Fallback only, if Chrome is not available |
| Go 1.24+ | Building from source | Not needed for a normal install |

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

- **[Session Tutorial](tutorials/session-tutorial.md)** — Run your first session
- **[Character Creation](tutorials/character-creation.md)** — Generate PCs
- **[Session Generator](tutorials/session-generator.md)** — Adapt sessions to your party
- **[PDF Compiler](features/pdf-compiler.md)** — Customize PDF output
- **[MCP Tools](features/mcp-tools.md)** — Full tool reference
- **[Architecture](features/architecture.md)** — How Grimorio works

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

In addition to PDF, the `export_campaign` MCP tool can export your campaign as:

- **Markdown** — concatenated in canonical file order, good for git or Obsidian
- **EPUB 3** — a valid e-reader format (OPF, NCX, XHTML)

There is no `grimorio export_campaign` CLI command: ask the agent instead, for example
"export my-campaign as markdown". PDF remains the default and most-styled output.

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
curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | sh
```

Esto instala:
- El binario de Grimorio en `~/.grimorio/grimorio`, enlazado en `~/.local/bin/grimorio`
- La integración con OpenCode: `mcp.servers.grimorio`, skills, agents y el comando `/grimorio`
- Nada más. Verifica si hay un motor de PDF, pero no lo instala.

## Requisitos

| Dependencia | Para qué | Notas |
|-------------|----------|-------|
| Python 3 | Instalar y actualizar | Fusiona `opencode.json` sin corromperlo |
| Chrome / Chromium | Compilar el PDF | Motor preferido |
| wkhtmltopdf | Compilar el PDF | Solo como respaldo si no hay Chrome |
| Go 1.24+ | Compilar desde el código fuente | No hace falta para una instalación normal |

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

- **[Tutorial de Sesión](tutorials/session-tutorial.md)** — Ejecuta tu primera sesión
- **[Creación de Personajes](tutorials/character-creation.md)** — Genera PJs
- **[Generador de Sesiones](tutorials/session-generator.md)** — Adapta sesiones a tu grupo
- **[Compilador PDF](features/pdf-compiler.md)** — Personaliza el PDF
- **[Herramientas MCP](features/mcp-tools.md)** — Referencia completa de herramientas
- **[Arquitectura](features/architecture.md)** — Cómo funciona Grimorio

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

Además de PDF, la tool MCP `export_campaign` puede exportar tu campaña como:

- **Markdown** — concatenado en orden canónico de archivos, ideal para git u Obsidian
- **EPUB 3** — formato válido de e-reader (OPF, NCX, XHTML)

No existe un comando CLI `grimorio export_campaign`: hay que pedírselo al agente, por ejemplo
"exportá mi-campana como markdown". PDF sigue siendo el default y el más estilizado.

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
