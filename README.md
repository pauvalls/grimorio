<div align="center">

# 📜 Grimorio

**D&D 5e Campaign & One-shot Generator**

[![Version](https://img.shields.io/github/v/release/pauvalls/grimorio?label=version&style=for-the-badge&color=purple)](https://github.com/pauvalls/grimorio/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/pauvalls/grimorio/ci.yml?branch=main&style=for-the-badge&label=CI&color=blue)](https://github.com/pauvalls/grimorio/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/pauvalls/grimorio?style=for-the-badge&color=green)](LICENSE)

[🇬🇧 English](#english) · [🇪🇸 Español](#español)

</div>

---

<a name="english"></a>

## 🇬🇧 English

AI-powered D&D 5e campaign and one-shot generator. Turn a spark of an idea into a fully-formatted, print-ready PDF adventure book — complete with lore, NPCs, bestiary, encounters, and styled layouts inspired by official Wizards of the Coast manuals.

### Quick Start

**Fresh install:**
```bash
curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | bash
```

**Already installed? Update in place:**
```bash
curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | bash -s -- --update
```

Then type in your AI assistant chat:

```
/grimorio A sunken city where the nobles are aquatic vampires
```

### What You Get

- **Full Campaign PDF** — Lore, NPCs, bestiary, encounters, maps (WotC-styled layout)
- **Interactive Q&A** — Grimorio asks questions first, then generates
- **AI Images** — FREE cover art and illustrations (Pollinations.ai, no API key)
- **Procedural Maps** — SVG battle maps and decorative dividers
- **MCP Server** — Native integration with OpenCode and Claude Code
- **AI Dungeon Master** — `grimorio-dm` agent runs live D&D 5e sessions with narrative depth, strict information hiding, and canon compliance
- **100% Local** — No cloud dependencies required

### Documentation

| Topic | Description |
|-------|-------------|
| **[Getting Started](docs/getting-started.md)** | Installation, first campaign, brief template |
| **[Session Tutorial](docs/tutorials/session-tutorial.md)** | Run your first game (prep → play → post-session) |
| **[Character Creation](docs/tutorials/character-creation.md)** | Generate PCs, pre-gens, character worksheets |
| **[Session Generator](docs/tutorials/session-generator.md)** | Adapt sessions to specific characters |
| **[PDF Compiler](docs/features/pdf-compiler.md)** | Customize PDF output, CSS styles, sections |
| **[MCP Tools](docs/features/mcp-tools.md)** | Full tool reference (30+ tools) |
| **[Architecture](docs/features/architecture.md)** | How Grimorio works internally |
| **[DM Guide](docs/dm-guide.md)** | General advice for running games |
| **[DM Agent Guide](docs/dm-agent-guide.md)** | Running live sessions with the AI Dungeon Master |
| **[Developer Guide](docs/developer-guide.md)** | Contributing to Grimorio |

### What's New

See the [CHANGELOG](CHANGELOG.md) for the full list of changes by version.

**v4.0.0 — AI Dungeon Master (Latest)**
- **🤖 AI Dungeon Master** — `grimorio-dm` primary agent runs live D&D 5e sessions with narrative depth, strict information hiding, and canon compliance
- **📦 Campaign Context Aggregation** — `dm_session_context` MCP tool loads all campaign data (canon, narrative state, areas, NPCs, bestiary, prologue, factions) in a single payload
- **🎲 Dice Modes** — Auto (DM rolls), Manual (players roll), or Mixed (default) per session
- **📖 Game Modes** — Narrative (1-2 combats, social-first) or Tactical (3-5 combates, resource-heavy)
- **🎭 NPC Voices** — Each NPC has a unique `dialogue_voice` for distinct, immersive dialogue
- **🚫 Information Hiding** — Never reveals enemy HP, AC, or dice rolls; uses descriptive damage states instead

**Previous Highlights**
- **Automated Releases** — Push a `v*` tag and CI creates the release + auto-updates the changelog
- **Self-Update** — `install.sh --update` detects changes, rebuilds only what's needed, preserves your config
- **Narrative Prologue** — `grimorio_generate_prologue` generates 4-part WotC-style prologue with read-aloud boxes
- **V3 Service Complete** — All 12 TODOs resolved: Tactics, PlayerMap, Area, Handout, Milestone services with filesystem repositories
- **New MCP Tools** — `grimorio_generate_tactics`, `grimorio_get_tactics`, `grimorio_export_handout`, `grimorio_update_session_xp`, `grimorio_generate_area_by_number`, `grimorio_generate_player_map`
- **`make install` / `make update`** — Developer-friendly targets with version metadata and plugin sync
- **`grimorio version`** — Now shows real version (git describe), commit, build date, and Go version
- **16 Agent Prompts** — Extracted to `agents/` directory as filesystem source of truth with full MCP tool references
- **WotC Format Validation** — `check_consistency scope=full` validates developments, multiple solutions, character hooks, boxed text, and integration cross-references
- **16 Grimorio Skills** — WotC standards preserved for AI-assisted campaign generation
- **PDF Compiler Enhancements** — Auto-close unclosed divs, `page-break-inside: avoid` on all components, CSS for flowcharts and scene descriptions

---

<a name="español"></a>

## 🇪🇸 Español

Generador de campañas y one-shots de D&D 5e potenciado por IA. Convierte una chispa de idea en un libro de aventuras en PDF completamente formateado, listo para imprimir — con trasfondo, NPCs, bestiario, encuentros, y diseños con estilo inspirado en los manuales oficiales de Wizards of the Coast.

### Inicio Rápido

**Instalación desde cero:**
```bash
curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | bash
```

**¿Ya lo tenés instalado? Actualizalo:**
```bash
curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | bash -s -- --update
```

Luego escribe en el chat de tu asistente IA:

```
/grimorio Una ciudad hundida donde los nobles son vampiros acuáticos
```

### Qué Obtienes

- **PDF de Campaña Completo** — Trasfondo, NPCs, bestiario, encuentros, mapas (estilo WotC)
- **Q&A Interactivo** — Grimorio hace preguntas primero, luego genera
- **Imágenes IA** — Portada e ilustraciones GRATIS (Pollinations.ai, sin API key)
- **Mapas Procedimentales** — Mapas de batalla SVG y divisores decorativos
- **Servidor MCP** — Integración nativa con OpenCode y Claude Code
- **Dungeon Master IA** — Agente `grimorio-dm` ejecuta sesiones en vivo de D&D 5e con profundidad narrativa, ocultamiento de información y cumplimiento de canon
- **100% Local** — Sin dependencias de nube requeridas

### Documentación

| Tema | Descripción |
|------|-------------|
| **[Empezando](docs/getting-started.md)** | Instalación, primera campaña, plantilla de brief |
| **[Tutorial de Sesión](docs/tutorials/session-tutorial.md)** | Ejecuta tu primer juego (prep → jugar → post-sesión) |
| **[Creación de Personajes](docs/tutorials/character-creation.md)** | Genera PJs, pre-generados, hojas de trabajo |
| **[Generador de Sesiones](docs/tutorials/session-generator.md)** | Adapta sesiones a personajes específicos |
| **[Compilador PDF](docs/features/pdf-compiler.md)** | Personaliza salida PDF, estilos CSS, secciones |
| **[Herramientas MCP](docs/features/mcp-tools.md)** | Referencia completa de herramientas (30+ herramientas) |
| **[Arquitectura](docs/features/architecture.md)** | Cómo funciona Grimorio internamente |
| **[Guía de DM](docs/dm-guide.md)** | Consejos generales para dirigir juegos |
| **[Guía del Agente DM](docs/dm-agent-guide.md)** | Ejecutar sesiones en vivo con el Dungeon Master IA |
| **[Guía de Desarrollador](docs/developer-guide.md)** | Contribuir a Grimorio |

### Novedades

Ver el [CHANGELOG](CHANGELOG.md) para la lista completa de cambios por versión.

**v4.0.0 — Dungeon Master IA (Último)**
- **🤖 Dungeon Master IA** — Agente `grimorio-dm` ejecuta sesiones en vivo de D&D 5e con profundidad narrativa, ocultamiento de información y cumplimiento de canon
- **📦 Agregación de Contexto** — Herramienta MCP `dm_session_context` carga todos los datos de la campaña (canon, estado narrativo, áreas, NPCs, bestiario, prólogo, facciones) en un solo payload
- **🎲 Modos de Dados** — Automático (DM tira), Manual (jugadores tiran), o Mixto (default) por sesión
- **📖 Modos de Juego** — Narrativo (1-2 combates, social primero) o Táctico (3-5 combates, gestión de recursos)
- **🎭 Voces de NPCs** — Cada NPC tiene un `dialogue_voice` único para diálogos distintivos e inmersivos
- **🚫 Ocultamiento de Información** — Nunca revela HP, AC o tiradas de enemigos; usa estados de daño descriptivos

**Highlights Anteriores**
- **Releases Automatizados** — Pusheá un tag `v*` y CI crea el release + actualiza el changelog
- **Auto-Actualización** — `install.sh --update` detecta cambios, recompila solo lo necesario, preserva tu configuración
- **Prólogo Narrativo** — `grimorio_generate_prologue` genera prólogo 4 partes estilo WotC con boxed text
- **Servicios V3 Completos** — Los 12 TODOs resueltos: Tactics, PlayerMap, Area, Handout, Milestone con repositorios filesystem
- **Nuevas Herramientas MCP** — `grimorio_generate_tactics`, `grimorio_get_tactics`, `grimorio_export_handout`, `grimorio_update_session_xp`, `grimorio_generate_area_by_number`, `grimorio_generate_player_map`
- **`make install` / `make update`** — Targets para developers con metadata de versión y sync de plugins
- **`grimorio version`** — Muestra versión real (git describe), commit, fecha de build y versión de Go
- **16 Prompts de Agentes** — Extraídos a `agents/` como fuente de verdad con referencias completas a herramientas MCP
- **Validación de Formato WotC** — `check_consistency scope=full` valida developments, múltiples soluciones, ganchos de personaje, boxed text, e integración cross-reference
- **16 Skills de Grimorio** — Estándares WotC preservados para generación de campañas asistida por IA
- **Mejoras al Compilador PDF** — Auto-close de divs sin cierre, `page-break-inside: avoid` en todos los componentes, CSS para flowcharts y scene descriptions

---

## Story Brief Template / Plantilla de Brief de Historia

```
**Campaign Name / Nombre de Campaña:** (kebab-case, ej. "sunken-city")
**Setting / Ambientación:** (one-sentence premise / premisa de una oración)
**Tone / Tono:** (heroic/heroico, dark/oscuro, humorous/humorístico, political/político, horror/terror, mystery/misterio)
**Level Range / Rango de Nivel:** (1-3, 4-6, 7-10, 11-15, 16-20)
**Duration / Duración:** (one-shot, 3-5 sessions/sesiones, long campaign/campaña larga)
**Themes / Temas:** (comma-separated / separados por comas)
**Villain Type / Tipo de Villano:** (optional / opcional)
**McGuffin:** (optional / opcional)
```

**Example / Ejemplo:**
```
**Campaign Name:** sunken-city
**Setting:** A sunken city where nobles are aquatic vampires
**Tone:** Dark political intrigue
**Level Range:** 4-6
**Duration:** 3-5 sessions
**Themes:** Betrayal, redemption, cosmic horror
**Villain Type:** Vampire lord
**McGuffin:** Ancient artifact that controls water
```

---

## Campaign Structure / Estructura de Campaña

```
~/campaigns/
└── sunken-city/
    ├── campaign.pdf          # Final PDF / PDF final
    ├── campaign.html         # HTML version / versión HTML
    ├── lore.md               # World backstory / Trasfondo del mundo
    ├── areas/                 # Chapters with areas / Capítulos con áreas
    ├── npcs/                 # NPCs and factions / NPCs y facciones
    ├── bestiary/             # Monster stat blocks / Estadísticas de monstruos
    ├── encounters/           # Combat challenges / Desafíos de combate
    ├── maps/                 # Location descriptions / Descripciones de locaciones
    ├── quests/               # Quests and objectives / Misiones y objetivos
    ├── characters/           # Character sheets / Hojas de personaje
    ├── assets/               # Maps and AI art / Mapas y arte de IA
    ├── canon.json            # Canonical facts / Hechos canónicos
    └── narrative_state.json  # Session tracking / Seguimiento de sesiones
```

---

## Requirements / Requisitos

| Dependency / Dependencia | Auto-installed / Auto-instalado | Purpose / Propósito |
|------------|---------------|---------|
| Go 1.24+ | ✅ Yes / Sí | Build binary / Compilar binario |
| wkhtmltopdf | ✅ Yes / Sí | Compile PDF / Compilar PDF |
| Git | ❌ Must have / Debes tener | Clone repo / Clonar repo |

## Troubleshooting / Solución de Problemas

**Update fails with "git pull failed" / El update falla con "git pull failed":**

If the incremental update fails, the install directory may be corrupted. The script now auto-detects this and falls back to a full install. If it still fails, manually remove and reinstall:

```bash
rm -rf ~/.local/share/grimorio
curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | bash
```

**El update incremental falla:**

Si el directorio de instalación está corrupto, el script ahora lo detecta automáticamente y hace una instalación completa. Si aún falla, eliminalo manualmente y reinstalá:

```bash
rm -rf ~/.local/share/grimorio
curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | bash
```

---

## Architecture Overview / Resumen de Arquitectura

```
┌─────────────────────────────────────────────────────────────┐
│                   OpenCode / Claude Code                     │
│                                                              │
│  ┌──────────────┐  ┌──────────────────┐  ┌────────────────┐ │
│  │ /grimorio    │──│ grimorio-        │  │ grimorio-      │ │
│  │ Command      │  │ architect        │  │ artist         │ │
│  │ (Entry point)│  │ (Orchestrator)   │  │ (Images)       │ │
│  └──────────────┘  └──────────────────┘  └────────────────┘ │
│                              │                               │
│                              ▼                               │
│  Content Sub-agents (delegated):                             │
│  ├─ grimorio-lore       ├─ grimorio-npc                     │
│  ├─ grimorio-bestiary   ├─ grimorio-encounters              │
│  ├─ grimorio-maps       ├─ grimorio-areas (WotC format)     │
│  ├─ grimorio-quests     ├─ grimorio-characters              │
│  └─ grimorio-narrative-custodian (validation)               │
│                              │                               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   MCP Server (Go, stdio)                     │
│                                                              │
│  Content Tools (v1):    Narrative Coherence (v2.0):         │
│  ├─ create_campaign     ├─ generate_adventure_bible         │
│  ├─ get_template        ├─ validate_canon                   │
│  ├─ save_act/npcs/etc   ├─ update_narrative_state           │
│  └─ compile_pdf         └─ check_consistency                │
│                                                              │
│  Asset Tools:           Living World (v2.1):                │
│  ├─ generate_map (SVG)  ├─ update_faction_reputation        │
│  ├─ generate_divider    ├─ generate_random_tables           │
│  ├─ generate_image (AI) ├─ generate_handouts                │
│  └─ generate_images_batch └─ evaluate_consequences          │
│                                                              │
│  DM Experience (Phase 4):                                   │
│  ├─ generate_session_prep                                   │
│  ├─ generate_flowchart                                      │
│  ├─ dm_session_context  (v4.0)                              │
│  └─ grimorio-dm agent (v4.0)                                │
└─────────────────────────────────────────────────────────────┘
```

---

## License / Licencia

**Mozilla Public License 2.0** — See [LICENSE](LICENSE) for details.

Copyright (c) 2026 Pau Valls

---

## Contributing / Contribuir

See [Developer Guide](docs/developer-guide.md) for contribution guidelines.

Ver [Guía de Desarrollador](docs/developer-guide.md) para directrices de contribución.
