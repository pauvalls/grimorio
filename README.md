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
- **Multi-provider LLM** — Works with OpenAI, Anthropic, Groq, Ollama (via OpenCode / Claude Code)
- **D&D-styled PDF** — Professional layout with CSS Paged Media and wkhtmltopdf
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

Phase 3: Parallel subagents (delegate)
  ├─ Lore subagent: world, setting, conflict (MCP: get_template + save)
  ├─ Acts subagent: 3 acts with scenes (MCP: get_template + save_act)
  ├─ NPCs subagent: 5+ NPCs + factions (MCP: get_template + save_npcs)
  ├─ Bestiary subagent: 3-5 monsters (MCP: get_template + save_bestiary)
  ├─ Encounters subagent: balanced fights (MCP: get_template + save_encounters)
  └─ Maps subagent: scene descriptions (MCP: get_template + save_maps)

Phase 4: Compile PDF (MCP: compile_pdf)

Phase 5: Report generated files location
```

> **Important:** The main agent only asks questions and compiles the PDF. All content generation happens in parallel subagents — the main thread stays clean.

### Architecture

```
OpenCode / Claude Code
    │
    ├─ Agent grimorio-architect → Q&A + orchestrates generation
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
         ├─ generate_image   → DALL-E API images (optional, requires API key)
         └─ compile_pdf      → Generates D&D-styled PDF with embedded images
```

### Image Generation

Grimorio supports two modes of image generation:

#### Procedural SVG (Default — 100% local, free)

No API key needed. Generates maps and dividers on the fly:

| Tool | Purpose | Example |
|------|---------|---------|
| `generate_map` | Battle maps, dungeon layouts, city maps | `![Dungeon Map](assets/dungeon-map.svg)` |
| `generate_divider` | Decorative section separators | `![Divider](assets/ornate-divider.svg)` |

**Map styles:** `dungeon`, `landscape`, `city`

#### DALL-E API (Optional — requires OpenAI API key)

For cover art, NPC portraits, and monster illustrations:

```bash
# Set your API key
export OPENAI_API_KEY="sk-..."

# Or add to config
echo '{"dalle_api_key": "sk-..."}' > ~/.config/grimorio/config.json
```

| Tool | Purpose | Cost |
|------|---------|------|
| `generate_image` | Cover art, portraits, illustrations | ~$0.04-0.08/image (DALL-E 3) |

> **Tip:** OpenAI gives $5 free credit to new accounts (~60-120 images).

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
    │   └─ grimorio-architect.md         # Campaign designer agent
    └─ skills/
        └─ dnd-5e-srd/SKILL.md           # D&D 5e rules reference
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
      "tools": { "delegate": true, "delegation_list": true, "delegation_read": true, /* ... */ }
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
    ├── campaign.html
    └── campaign.pdf
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
- **Múltiples proveedores LLM** — Funciona con OpenAI, Anthropic, Groq, Ollama (vía OpenCode / Claude Code)
- **PDF estilo D&D** — Maquetación profesional con CSS Paged Media y wkhtmltopdf
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

Fase 3: Subagentes en paralelo (delegate)
  ├─ Subagente lore: mundo, ambientación, conflicto (MCP: get_template + save)
  ├─ Subagente actos: 3 actos con escenas (MCP: get_template + save_act)
  ├─ Subagente NPCs: 5+ NPCs + facciones (MCP: get_template + save_npcs)
  ├─ Subagente bestiario: 3-5 monstruos (MCP: get_template + save_bestiary)
  ├─ Subagente encuentros: combates balanceados (MCP: get_template + save_encounters)
  └─ Subagente mapas: descripciones de escenas (MCP: get_template + save_maps)

Fase 4: Compilar PDF (MCP: compile_pdf)

Fase 5: Mostrar ubicación de archivos generados
```

> **Importante:** El agente principal solo hace preguntas y compila el PDF. Toda la generación de contenido ocurre en subagentes paralelos — el hilo principal queda limpio.

### Arquitectura

```
OpenCode / Claude Code
    │
    ├─ Agente grimorio-architect → Q&A + orquesta la generación
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
         └─ compile_pdf      → Genera PDF estilo D&D
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
    │   └─ grimorio-architect.md         # Agente diseñador de campañas
    └─ skills/
        └─ dnd-5e-srd/SKILL.md           # Referencia de reglas D&D 5e
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
      "tools": { "delegate": true, "delegation_list": true, "delegation_read": true, /* ... */ }
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
    ├── campaign.html
    └── campaign.pdf
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

### Generación de Imágenes

Grimorio soporta dos modos de generación de imágenes:

#### SVG Procedural (Por defecto — 100% local, gratis)

Sin API key. Genera mapas y divisores al vuelo:

| Herramienta | Propósito | Ejemplo |
|------------|-----------|---------|
| `generate_map` | Mapas de batalla, mazmorras, ciudades | `![Mapa Mazmorra](assets/dungeon-map.svg)` |
| `generate_divider` | Divisores decorativos de sección | `![Divisor](assets/ornate-divider.svg)` |

**Estilos de mapa:** `dungeon`, `landscape`, `city`

#### DALL-E API (Opcional — requiere API key de OpenAI)

Para arte de portada, retratos de NPCs e ilustraciones:

```bash
# Configurar API key
export OPENAI_API_KEY="sk-..."

# O agregar al config
echo '{"dalle_api_key": "sk-..."}' > ~/.config/grimorio/config.json
```

| Herramienta | Propósito | Costo |
|------------|-----------|-------|
| `generate_image` | Portada, retratos, ilustraciones | ~$0.04-0.08/imagen (DALL-E 3) |

> **Tip:** OpenAI da $5 de crédito gratis a cuentas nuevas (~60-120 imágenes).

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
