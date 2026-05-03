<div align="center">

# 📜 Grimorio

**D&D One-shot & Campaign Generator**

[English](#english) · [Español](#español)

</div>

---

<a name="english"></a>

## 🇬🇧 English

AI-powered D&D 5e campaign and one-shot generator. Turn a spark of an idea into a fully-formatted, print-ready PDF adventure book — complete with lore, NPCs, bestiary, encounters, and styled layouts inspired by official Wizards of the Coast manuals.

### Features

- **Full campaign generation** — Lore, acts, NPCs, monsters, encounters, and maps
- **Interactive Q&A flow** — `/grimorio` asks questions first, then generates via parallel subagents
- **Image generation** — Procedural SVG maps (free) + AI cover art/portraits via Pollinations.ai (FREE, no API key) or DALL-E 3 (optional)
- **Multi-provider LLM** — Works with OpenAI, Anthropic, Groq, Ollama (via OpenCode / Claude Code)
- **D&D-styled PDF** — Professional layout with CSS Paged Media and wkhtmltopdf, with embedded images (maps, AI art, portraits)
- **MCP Server** — Native integration with OpenCode and Claude Code as MCP tools
- **Zero cloud dependencies** — Runs 100% locally, no servers required

### Quick Install

```bash
curl -sSL https://raw.githubusercontent.com/pauvalls/Grimorio/main/install.sh | bash
```

**What the installer does:**

| Step | Claude Code | OpenCode |
|------|-------------|----------|
| Go 1.23+ | Installs if missing | Installs if missing |
| wkhtmltopdf | Installs if missing | Installs if missing |
| Build binary | `~/.local/bin/grimorio` | `~/.local/bin/grimorio` |
| Plugin files | `~/.claude/plugins/grimorio/` | `~/.config/opencode/plugins/grimorio/` |
| `.mcp.json` | Uses `${CLAUDE_PLUGIN_ROOT}` | Uses absolute path |
| MCP server config | Auto-discovered via `.mcp.json` | Added to `opencode.json` mcp section |
| Agent config | Auto-discovered from plugin | Added to `opencode.json` agent section |
| `/grimorio` command | Auto-discovered from plugin | Added to `opencode.json` command section |
| Shell PATH | `~/.local/bin` added | `~/.local/bin` added |

### Requirements

| Dependency | Auto-installed | Purpose |
|------------|---------------|---------|
| Go 1.23+ | ✅ Yes | Build the MCP server binary |
| wkhtmltopdf | ✅ Yes | Compile HTML to PDF |
| Git | ❌ Must have | Clone the repository |

### Usage

#### OpenCode / Claude Code

Once installed, type in the chat:

```
/grimorio A sunken city where the nobles are aquatic vampires
```

The `grimorio-architect` agent will:

```
Phase 1: Interactive Q&A
  ├─ Campaign name? (kebab-case)
  ├─ One-shot or full campaign?
  ├─ Player level? (1-20)
  ├─ Desired tone? (heroic, dark, humorous, political intrigue)
  └─ Duration? (one-shot, 3-5 sessions, long campaign)

Phase 2: Create campaign structure (MCP: create_campaign)

Phase 3: Launch Orchestrator (single delegate)
  └─ grimorio-orchestrator coordinates internally:
      ├─ grimorio-cartographer: cover art, battle maps, portraits (MCP: generate_image + generate_map)
      ├─ Lore subagent: world, setting, conflict (MCP: save)
      ├─ NPCs subagent: 5+ NPCs + factions (MCP: save_npcs)
      ├─ Bestiary subagent: 3-5 monsters (MCP: save_bestiary)
      ├─ Encounters subagent: balanced fights (MCP: save_encounters)
      ├─ Maps subagent: scene descriptions (MCP: save_maps)
      └─ Acts subagent: 3 acts with scenes (MCP: save_act)

Phase 4: Compile PDF (MCP: compile_pdf) — embeds all images

Phase 5: Report generated files location
```

> **Important:** The main agent only asks questions and launches the orchestrator with a single `delegate` call. The orchestrator handles all subagent coordination internally — the main thread does zero polling.

### Architecture

```
OpenCode / Claude Code
    │
    ├─ Agent grimorio-architect → Q&A + single delegate to orchestrator
    ├─ Agent grimorio-orchestrator → Coordinates all subagents internally
    ├─ Agent grimorio-cartographer → Battle maps, SVGs, images (subagent)
    ├─ Command /grimorio        → Triggers the workflow above
    └─ Skill dnd-5e-srd         → D&D 5e rules context
         │
         ▼
    MCP Server (Go, stdio)
         │
         ├─ create_campaign  → Directory structure
         ├─ get_template     → Structured content templates
         ├─ save_act         → Saves acts as Markdown
         ├─ save_npcs        → Saves characters and factions
         ├─ save_bestiary    → Saves stat blocks
         ├─ save_encounters  → Saves encounters
         ├─ save_maps        → Saves scenes
         ├─ generate_map     → Procedural SVG battle maps (100% local)
         ├─ generate_divider → Decorative SVG section dividers (100% local)
          ├─ generate_image   → AI images via Pollinations.ai (FREE) or DALL-E (optional)
         └─ compile_pdf      → Generates D&D-styled PDF with embedded images
```

### Image Generation

Grimorio supports three modes of image generation:

#### AI Images (Default — FREE via Pollinations.ai)

No API key needed. Generates cover art, NPC portraits, and illustrations using Pollinations.ai (FLUX model):

| Tool | Purpose | Cost |
|------|---------|------|
| `generate_image` | Cover art, portraits, illustrations | **FREE** |

#### Procedural SVG (100% local, free)

No API key needed. Generates maps and dividers on the fly:

| Tool | Purpose | Example |
|------|---------|---------|
| `generate_map` | Battle maps, dungeon layouts, city maps | `![Dungeon Map](assets/dungeon-map.svg)` |
| `generate_divider` | Decorative section separators | `![Divider](assets/ornate-divider.svg)` |

**Map styles:** `dungeon`, `landscape`, `city`

#### DALL-E API (Optional — higher quality, paid)

For premium image quality, switch to DALL-E 3:

```bash
# Set your API key
export OPENAI_API_KEY="sk-..."

# Or add to config
echo '{"image_provider": "dalle", "dalle_api_key": "sk-..."}' > ~/.config/grimorio/config.json
```

| Tool | Purpose | Cost |
|------|---------|------|
| `generate_image` | Cover art, portraits, illustrations | ~$0.04-0.08/image (DALL-E 3) |

> **Tip:** OpenAI gives $5 free credit to new accounts (~60-120 images).

**PDF Image Embedding:**

All images (SVG maps, AI-generated PNGs, dividers) are automatically embedded into the PDF:
- Images referenced in Markdown with `![alt](assets/file.png)` appear inline
- All images in `assets/` are included in a "Campaign Visuals" gallery at the end
- SVGs are embedded as vector graphics, PNGs as base64

### Plugin Structure

```
~/.config/opencode/plugins/grimorio/    (OpenCode)
~/.claude/plugins/grimorio/             (Claude Code)
    │
    ├─ grimorio                          # MCP server binary
    ├─ .mcp.json                         # MCP server config (per-tool)
    ├─ .claude-plugin/plugin.json        # Plugin manifest
    ├─ commands/
    │   └─ grimorio.md                   # /grimorio slash command
    ├─ agents/
    │   ├─ grimorio-architect.md         # Campaign designer agent (Q&A + single delegate)
    │   ├─ grimorio-orchestrator.md      # Internal coordinator (handles all subagents)
    │   └─ grimorio-cartographer.md      # Maps, SVGs, images subagent
    └─ skills/
        └─ dnd-5e-srd/SKILL.md           # D&D 5e rules reference

Source code structure:
    │
    ├─ cmd/grimorio/                     # Entry point (stdio MCP server)
    ├─ internal/
    │   ├── mcp/server.go                # MCP tool definitions + handlers
    │   ├── compiler/compiler.go         # Markdown → HTML → PDF pipeline
    │   ├── svg/svg.go                   # Procedural SVG generator (maps, dividers)
    │   ├── image/                       # Image provider abstraction (Pollinations.ai, DALL-E)
    │   └── config/config.go             # Configuration management
    └─ internal/compiler/templates/      # Embedded CSS + Markdown templates
```

### OpenCode Configuration

The installer automatically adds the following to `~/.config/opencode/opencode.json`:

```jsonc
{
  "mcp": {
    "grimorio": {
      "command": ["/home/user/.config/opencode/plugins/grimorio/grimorio"],
      "type": "local",
      "enabled": true
    }
  },
  "agent": {
    "grimorio-architect": {
      "mode": "primary",
      "tools": { "delegate": true /* ... */ }
    }
  },
  "command": {
    "grimorio": {
      "agent": "grimorio-architect",
      "subtask": false,
      "template": "Generate a D&D 5e campaign..."
    }
  }
}
```

### Output Structure

Every generated campaign lives in its own directory:

```
~/campaigns/
└── sunken-city/
    ├── README.md
    ├── lore.md
    ├── acts/
    │   ├── act_1_the_descent.md
    │   ├── act_2_the_feast.md
    │   └── act_3_the_ritual.md
    ├── npcs/
    │   └── npcs_and_factions.md
    ├── bestiary/
    │   └── bestiary.md
    ├── encounters/
    │   └── encounters.md
    ├── maps/
    │   └── maps_and_scenes.md
    ├── assets/
    │   ├── dungeon-map.svg          ← Procedural battle maps
    │   ├── ornate-divider.svg       ← Decorative section dividers
    │   └── cover-art.png            ← AI generated images (Pollinations.ai or DALL-E)
    ├── campaign.html
    └── campaign.pdf                 ← Final PDF with embedded images
```

### Available Templates

The MCP server exposes structured templates for each content type:

| Template   | Description                              |
|------------|------------------------------------------|
| `act`      | Act/chapter structure with scenes        |
| `npc`      | NPC sheet with motivation and secret     |
| `monster`  | Full D&D 5e stat block                   |
| `encounter`| Encounter with difficulty balancing      |
| `map`      | Scene description with areas             |
| `lore`     | World-building and conflicts             |

### MCP Tools

All tools available through the MCP server:

| Tool | Type | Description |
|------|------|-------------|
| `create_campaign` | File | Creates campaign directory structure |
| `get_template` | Template | Returns structured Markdown template |
| `save_act` | File | Saves act as Markdown file |
| `save_npcs` | File | Saves NPCs and factions |
| `save_bestiary` | File | Saves monster stat blocks |
| `save_encounters` | File | Saves combat encounters |
| `save_maps` | File | Saves scene descriptions |
| `generate_map` | SVG | Procedural battle map generator (free) |
| `generate_divider` | SVG | Decorative section dividers (free) |
| `generate_image` | AI | Image generation via Pollinations.ai (FREE) or DALL-E (optional) |
| `compile_pdf` | PDF | Compiles all content into styled PDF |

### Development

#### Manual Build

```bash
go build -o grimorio ./cmd/grimorio
```

#### Run MCP Server (standalone)

```bash
./grimorio
```

The server runs in stdio mode (stdin/stdout) for MCP communication.

---

<a name="español"></a>

## 🇪🇸 Español

Generador de campañas y one-shots de D&D 5e impulsado por IA. Convierte una idea en un libro de aventura completo con formato profesional, listo para imprimir — incluyendo lore, NPCs, bestiario, encuentros y maquetación inspirada en los manuales oficiales de Wizards of the Coast.

### Características

- **Generación completa de campañas** — Lore, actos, NPCs, monstruos, encuentros y mapas
- **Flujo interactivo Q&A** — `/grimorio` pregunta primero, después genera con subagentes en paralelo
- **Generación de imágenes** — Mapas SVG procedurales (gratis) + portadas/retratos IA vía Pollinations.ai (GRATIS, sin API key) o DALL-E 3 (opcional)
- **Múltiples proveedores LLM** — Funciona con OpenAI, Anthropic, Groq, Ollama (vía OpenCode / Claude Code)
- **PDF estilo D&D** — Maquetación profesional con CSS Paged Media y wkhtmltopdf, con imágenes embebidas (mapas, arte IA, retratos)
- **Servidor MCP** — Integración nativa con OpenCode y Claude Code como herramientas MCP
- **Sin dependencias de nube** — Funciona 100% local, sin servidores

### Instalación Rápida

```bash
curl -sSL https://raw.githubusercontent.com/pauvalls/Grimorio/main/install.sh | bash
```

**Qué hace el instalador:**

| Paso | Claude Code | OpenCode |
|------|-------------|----------|
| Go 1.23+ | Instala si falta | Instala si falta |
| wkhtmltopdf | Instala si falta | Instala si falta |
| Compilar binario | `~/.local/bin/grimorio` | `~/.local/bin/grimorio` |
| Archivos del plugin | `~/.claude/plugins/grimorio/` | `~/.config/opencode/plugins/grimorio/` |
| `.mcp.json` | Usa `${CLAUDE_PLUGIN_ROOT}` | Usa ruta absoluta |
| Config MCP | Auto-detectado vía `.mcp.json` | Agregado a `opencode.json` sección mcp |
| Config agente | Auto-detectado del plugin | Agregado a `opencode.json` sección agent |
| Comando `/grimorio` | Auto-detectado del plugin | Agregado a `opencode.json` sección command |
| Shell PATH | `~/.local/bin` agregado | `~/.local/bin` agregado |

### Requisitos

| Dependencia | Auto-instalada | Propósito |
|------------|---------------|-----------|
| Go 1.23+ | ✅ Sí | Compilar el binario del servidor MCP |
| wkhtmltopdf | ✅ Sí | Compilar HTML a PDF |
| Git | ❌ Requerido | Clonar el repositorio |

### Uso

#### OpenCode / Claude Code

Una vez instalado, escribí en el chat:

```
/grimorio Una ciudad sumergida donde los nobles son vampiros acuáticos
```

El agente `grimorio-architect` va a:

```
Fase 1: Preguntas interactivas
  ├─ ¿Nombre de la campaña? (kebab-case)
  ├─ ¿One-shot o campaña completa?
  ├─ ¿Nivel de los jugadores? (1-20)
  ├─ ¿Tono deseado? (heroico, oscuro, humorístico, intriga política)
  └─ ¿Duración? (one-shot, 3-5 sesiones, campaña larga)

Fase 2: Crear estructura (MCP: create_campaign)

Fase 3: Lanzar Orquestador (delegate único)
  └─ grimorio-orchestrator coordina internamente:
      ├─ grimorio-cartographer: portada, mapas de batalla, retratos (MCP: generate_image + generate_map)
      ├─ Subagente lore: mundo, ambientación, conflicto (MCP: save)
      ├─ Subagente NPCs: 5+ NPCs + facciones (MCP: save_npcs)
      ├─ Subagente bestiario: 3-5 monstruos (MCP: save_bestiary)
      ├─ Subagente encuentros: combates balanceados (MCP: save_encounters)
      ├─ Subagente mapas: descripciones de escenas (MCP: save_maps)
      └─ Subagente actos: 3 actos con escenas (MCP: save_act)

Fase 4: Compilar PDF (MCP: compile_pdf) — embebe todas las imágenes

Fase 5: Mostrar ubicación de archivos generados
```

> **Importante:** El agente principal solo hace preguntas y lanza el orquestador con un único `delegate`. El orquestador coordina todos los subagentes internamente — el hilo principal no hace polling.

### Arquitectura

```
OpenCode / Claude Code
    │
    ├─ Agente grimorio-architect → Q&A + delegate único al orquestador
    ├─ Agente grimorio-orchestrator → Coordina todos los subagentes internamente
    ├─ Agente grimorio-cartographer → Mapas, SVGs, imágenes (subagente)
    ├─ Comando /grimorio         → Dispara el flujo de arriba
    └─ Skill dnd-5e-srd          → Contexto de reglas D&D 5e
         │
         ▼
    Servidor MCP (Go, stdio)
         │
         ├─ create_campaign  → Estructura de carpetas
         ├─ get_template     → Templates de contenido estructurado
         ├─ save_act         → Guarda actos en Markdown
         ├─ save_npcs        → Guarda personajes y facciones
         ├─ save_bestiary    → Guarda stat blocks
         ├─ save_encounters  → Guarda encuentros
         ├─ save_maps        → Guarda escenas
         ├─ generate_map     → Mapas SVG procedurales (100% local)
         ├─ generate_divider → Divisores decorativos SVG (100% local)
          ├─ generate_image   → Imágenes IA vía Pollinations.ai (GRATIS) o DALL-E (opcional)
         └─ compile_pdf      → Genera PDF estilo D&D con imágenes embebidas
```

### Estructura del Plugin

```
~/.config/opencode/plugins/grimorio/    (OpenCode)
~/.claude/plugins/grimorio/             (Claude Code)
    │
    ├─ grimorio                          # Binario del servidor MCP
    ├─ .mcp.json                         # Config del servidor MCP (por herramienta)
    ├─ .claude-plugin/plugin.json        # Manifiesto del plugin
    ├─ commands/
    │   └─ grimorio.md                   # Comando slash /grimorio
    ├─ agents/
    │   ├─ grimorio-architect.md         # Agente diseñador (Q&A + delegate único)
    │   ├─ grimorio-orchestrator.md      # Coordinador interno (maneja todos los subagentes)
    │   └─ grimorio-cartographer.md      # Subagente mapas, SVGs, imágenes
    └─ skills/
        └─ dnd-5e-srd/SKILL.md           # Referencia de reglas D&D 5e

Estructura del código fuente:
    │
    ├─ cmd/grimorio/                     # Punto de entrada (servidor MCP stdio)
    ├─ internal/
    │   ├── mcp/server.go                # Definiciones de herramientas MCP + handlers
    │   ├── compiler/compiler.go         # Pipeline Markdown → HTML → PDF
    │   ├── svg/svg.go                   # Generador SVG procedural (mapas, divisores)
    │   ├── image/                       # Abstracción de proveedores de imagen (Pollinations.ai, DALL-E)
    │   └── config/config.go             # Gestión de configuración
    └─ internal/compiler/templates/      # CSS + templates Markdown embebidos
```

### Configuración de OpenCode

El instalador agrega automáticamente lo siguiente a `~/.config/opencode/opencode.json`:

```jsonc
{
  "mcp": {
    "grimorio": {
      "command": ["/home/user/.config/opencode/plugins/grimorio/grimorio"],
      "type": "local",
      "enabled": true
    }
  },
  "agent": {
    "grimorio-architect": {
      "mode": "primary",
      "tools": { "delegate": true /* ... */ }
    }
  },
  "command": {
    "grimorio": {
      "agent": "grimorio-architect",
      "subtask": false,
      "template": "Generate a D&D 5e campaign..."
    }
  }
}
```

### Estructura de Salida

Cada campaña generada vive en su propio directorio:

```
~/campaigns/
└── ciudad-sumergida/
    ├── README.md
    ├── lore.md
    ├── acts/
    │   ├── act_1_el_descenso.md
    │   ├── act_2_el_festin.md
    │   └── act_3_el_ritual.md
    ├── npcs/
    │   └── npcs_and_factions.md
    ├── bestiary/
    │   └── bestiary.md
    ├── encounters/
    │   └── encounters.md
    ├── maps/
    │   └── maps_and_scenes.md
    ├── assets/
    │   ├── dungeon-map.svg          ← Mapas de batalla procedurales
    │   ├── ornate-divider.svg       ← Divisores decorativos
    │   └── cover-art.png            ← Imágenes IA generadas (Pollinations.ai o DALL-E)
    ├── campaign.html
    └── campaign.pdf                 ← PDF final con imágenes embebidas
```

### Templates Disponibles

El servidor MCP expone templates estructurados para cada tipo de contenido:

| Template    | Descripción                                |
|-------------|--------------------------------------------|
| `act`       | Estructura de acto/capítulo con escenas    |
| `npc`       | Ficha de NPC con motivación y secreto      |
| `monster`   | Stat block completo D&D 5e                 |
| `encounter` | Encuentro con balance de dificultad        |
| `map`       | Descripción de escena con áreas            |
| `lore`      | Ambientación y conflictos                  |

### Herramientas MCP

Todas las herramientas disponibles a través del servidor MCP:

| Herramienta | Tipo | Descripción |
|------------|------|-------------|
| `create_campaign` | Archivo | Crea estructura de directorios |
| `get_template` | Template | Devuelve template Markdown estructurado |
| `save_act` | Archivo | Guarda acto como archivo Markdown |
| `save_npcs` | Archivo | Guarda NPCs y facciones |
| `save_bestiary` | Archivo | Guarda stat blocks de monstruos |
| `save_encounters` | Archivo | Guarda encuentros de combate |
| `save_maps` | Archivo | Guarda descripciones de escenas |
| `generate_map` | SVG | Generador de mapas procedurales (gratis) |
| `generate_divider` | SVG | Divisores decorativos (gratis) |
| `generate_image` | IA | Generación de imágenes vía Pollinations.ai (GRATIS) o DALL-E (opcional) |
| `compile_pdf` | PDF | Compila todo en PDF estilizado |

### Generación de Imágenes

Grimorio soporta tres modos de generación de imágenes:

#### Imágenes IA (Por defecto — GRATIS vía Pollinations.ai)

Sin API key. Genera portadas, retratos de NPCs e ilustraciones usando Pollinations.ai (modelo FLUX):

| Herramienta | Propósito | Costo |
|------------|-----------|-------|
| `generate_image` | Portada, retratos, ilustraciones | **GRATIS** |

#### SVG Procedural (100% local, gratis)

Sin API key. Genera mapas y divisores al vuelo:

| Herramienta | Propósito | Ejemplo |
|------------|-----------|---------|
| `generate_map` | Mapas de batalla, mazmorras, ciudades | `![Mapa Mazmorra](assets/dungeon-map.svg)` |
| `generate_divider` | Divisores decorativos de sección | `![Divisor](assets/ornate-divider.svg)` |

**Estilos de mapa:** `dungeon`, `landscape`, `city`

#### DALL-E API (Opcional — mayor calidad, de pago)

Para calidad premium, cambia a DALL-E 3:

```bash
# Configurar API key
export OPENAI_API_KEY="sk-..."

# O agregar al config
echo '{"image_provider": "dalle", "dalle_api_key": "sk-..."}' > ~/.config/grimorio/config.json
```

| Herramienta | Propósito | Costo |
|------------|-----------|-------|
| `generate_image` | Portada, retratos, ilustraciones | ~$0.04-0.08/imagen (DALL-E 3) |

> **Tip:** OpenAI da $5 de crédito gratis a cuentas nuevas (~60-120 imágenes).

**Imágenes en el PDF:**

Todas las imágenes (mapas SVG, PNGs generados por IA, divisores) se embeben automáticamente en el PDF:
- Las imágenes referenciadas en Markdown con `![alt](assets/archivo.png)` aparecen inline
- Todas las imágenes en `assets/` se incluyen en una galería "Visuales de la Campaña" al final
- Los SVGs se embeben como gráficos vectoriales, los PNGs como base64

### Desarrollo

#### Compilar manualmente

```bash
go build -o grimorio ./cmd/grimorio
```

#### Ejecutar servidor MCP (standalone)

```bash
./grimorio
```

El servidor corre en modo stdio (stdin/stdout) para comunicación MCP.

---

## License

MIT
