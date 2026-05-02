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
- **Multi-provider LLM** — Works with OpenAI, Anthropic, Groq, Ollama (via OpenCode)
- **D&D-styled PDF** — Professional layout with CSS Paged Media and wkhtmltopdf
- **MCP Server** — Native integration with OpenCode as MCP tools
- **Zero cloud dependencies** — Runs 100% locally, no servers required

### Quick Install

```bash
curl -sSL https://raw.githubusercontent.com/paupena/grimorio/main/install.sh | bash
```

### Usage with OpenCode

Once installed, simply type in the chat:

```
/grimorio A sunken city where the nobles are aquatic vampires
```

The `grimorio-architect` agent will ask you about player level, tone, and duration, then generate the entire adventure step by step.

### Architecture

```
OpenCode (Claude)
    │
    ├─ Agent grimorio-architect → Generates lore, plot, stats
    ├─ Command /grimorio        → Orchestrates the full workflow
    └─ Skill dnd-5e-srd         → D&D 5e rules context
         │
         ▼
    MCP Server (Go, stdio)
         │
         ├─ create_campaign  → Directory structure
         ├─ save_act         → Saves acts as Markdown
         ├─ save_npcs        → Saves characters
         ├─ save_bestiary    → Saves stat blocks
         ├─ save_encounters  → Saves encounters
         ├─ save_maps        → Saves scenes
         └─ compile_pdf      → Generates D&D-styled PDF
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

#### Requirements

- Go 1.23+
- wkhtmltopdf

#### Build

```bash
go build -o grimorio ./cmd/grimorio
```

#### Run MCP Server

```bash
./grimorio
```

The server runs in stdio mode (stdin/stdout) for OpenCode communication.

---

<a name="español"></a>

## 🇪🇸 Español

Generador de campañas y one-shots de D&D 5e impulsado por IA. Convierte una idea en un libro de aventura completo con formato profesional, listo para imprimir — incluyendo lore, NPCs, bestiario, encuentros y maquetación inspirada en los manuales oficiales de Wizards of the Coast.

### Características

- **Generación completa de campañas** — Lore, actos, NPCs, monstruos, encuentros y mapas
- **Múltiples proveedores LLM** — Funciona con OpenAI, Anthropic, Groq, Ollama (vía OpenCode)
- **PDF estilo D&D** — Maquetación profesional con CSS Paged Media y wkhtmltopdf
- **Servidor MCP** — Integración nativa con OpenCode como herramientas MCP
- **Sin dependencias de nube** — Funciona 100% local, sin servidores

### Instalación Rápida

```bash
curl -sSL https://raw.githubusercontent.com/paupena/grimorio/main/install.sh | bash
```

### Uso con OpenCode

Una vez instalado, simplemente escribe en el chat:

```
/grimorio Una ciudad sumergida donde los nobles son vampiros acuáticos
```

El agente `grimorio-architect` te preguntará el nivel de los jugadores, el tono y la duración, luego generará toda la aventura paso a paso.

### Arquitectura

```
OpenCode (Claude)
    │
    ├─ Agente grimorio-architect → Genera lore, trama, stats
    ├─ Comando /grimorio         → Orquesta todo el flujo
    └─ Skill dnd-5e-srd          → Contexto de reglas D&D 5e
         │
         ▼
    Servidor MCP (Go, stdio)
         │
         ├─ create_campaign  → Estructura de carpetas
         ├─ save_act         → Guarda actos en Markdown
         ├─ save_npcs        → Guarda personajes
         ├─ save_bestiary    → Guarda stat blocks
         ├─ save_encounters  → Guarda encuentros
         ├─ save_maps        → Guarda escenas
         └─ compile_pdf      → Genera PDF estilo D&D
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

### Desarrollo

#### Requisitos

- Go 1.23+
- wkhtmltopdf

#### Compilar

```bash
go build -o grimorio ./cmd/grimorio
```

#### Ejecutar Servidor MCP

```bash
./grimorio
```

El servidor corre en modo stdio (stdin/stdout) para comunicación con OpenCode.

---

## License

MIT
