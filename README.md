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

> 💡 **Built for OpenCode, adaptable to any system.** Grimorio is designed as an MCP server with native OpenCode integration (agents, skills, commands), but the underlying tools and PDF compiler work with any AI assistant or CLI workflow.

### Quick Start

**Linux / macOS:**
```bash
curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | sh
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/pauvalls/grimorio/main/install.ps1 | iex
```
> ℹ️ **Windows support** — PDF text extraction requires `poppler-utils`. Install inside WSL (`wsl --install` then `sudo apt install poppler-utils`) or skip the `include_pdf_text` flag. CI matrix validates Windows builds on every PR.

**Already installed? Update:**
```bash
grimorio update          # Update binary only
grimorio update skills   # Update skills
grimorio update agents   # Update agents
grimorio update commands # Update opencode.json config
grimorio update all      # Update everything at once
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

### How to Use

**1. Create a Campaign** (one-shot or full campaign)

In OpenCode, switch to the `grimorio-architect` agent and run:
```
/grimorio A sunken city where the nobles are aquatic vampires
```
The architect will ask you 6 questions (name, type, idea, level, tone, duration) and then generate the full campaign in batches:
- **Batch 1**: Lore, NPCs, bestiary, maps, setting guide, introduction
- **Batch 2**: Quests, encounters, characters, appendices
- **Batch 3**: Areas (chapter by chapter to avoid timeout)
- **Final**: Art, PDF compilation, validations

All files are saved to `~/campaigns/<campaign-name>/`.

**2. Run a Live Session** (DM assistant)

In OpenCode, switch to the `grimorio-dm` agent within your campaign directory:
```bash
cd ~/campaigns/sunken-city
grimorio-dm
```
The `grimorio-dm` agent loads the full campaign context (canon, areas, NPCs, bestiary) and acts as your AI Dungeon Master:
- Tracks session state (clues revealed, NPCs met, decisions made)
- Runs combat with descriptive damage states (hides enemy HP/AC)
- Maintains narrative coherence across sessions
- Generates session prep, handouts, and random tables

**🔊 Voice Narration (TTS) — Experimental**

> ⚠️ **Experimental Feature** — TTS is an optional add-on that requires manual setup. It works great once configured, but is not enabled by default.

Grimorio can narrate DM responses aloud using **Piper TTS** (local neural voice synthesis). The `grimorio-dm` agent will automatically narrate every narrative response when TTS is enabled.

**Quick Setup:**
```bash
# 1. Download Piper prebuilt binary
mkdir -p ~/.local/bin/piper
wget https://github.com/rhasspy/piper/releases/download/v1.2.0/piper_amd64.tar.gz -O /tmp/piper.tar.gz
tar -xzf /tmp/piper.tar.gz -C ~/.local/bin/piper

# 2. Download a voice (Spanish example)
mkdir -p ~/.local/share/piper
wget https://huggingface.co/rhasspy/piper-voices/resolve/main/es/es_ES/davefx/medium/es_ES-davefx-medium.onnx -P ~/.local/share/piper/
wget https://huggingface.co/rhasspy/piper-voices/resolve/main/es/es_ES/davefx/medium/es_ES-davefx-medium.onnx.json -P ~/.local/share/piper/

# 3. Create helper scripts
ln -s ~/.local/bin/piper/piper ~/.local/bin/piper
# Create tts-dm.sh and narrate scripts (see docs/tts-setup.md)

# 4. Set environment variables
export PIPER_MODEL_PATH="$HOME/.local/share/piper/es_ES-davefx-medium.onnx"
export PIPER_CONFIG_PATH="$HOME/.local/share/piper/es_ES-davefx-medium.onnx.json"
export PATH="$HOME/.local/bin:$PATH"

# 5. Apply to OpenCode
grimorio update commands
```

**How It Works in Sessions:**
1. Start a session with `grimorio-dm`
2. The agent asks: "🔊 TTS: Available. Activate? Yes/No"
3. If Yes, every narrative response is automatically narrated
4. The agent writes text → shows "🎙️ Narrando..." → executes TTS in background

**Features:**
- **Local & Free** — Runs on your machine, no API keys, no cloud
- **Auto-chunking** — Splits long text into ~150 char sentence chunks automatically
- **Background audio** — Uses `setsid` to detach from shell, survives timeouts
- **Multiple voices** — Browse [Piper Voices](https://huggingface.co/rhasspy/piper-voices/tree/main)

**See [docs/tts-experimental.md](docs/tts-experimental.md) for full setup, troubleshooting, and voice customization.**

**3. Update Content** (after rule changes or edits)

If you edit any markdown files manually, refresh the canon:
```bash
grimorio validate_canon --campaign sunken-city
grimorio check_consistency --campaign sunken-city
```

### Advanced Install

**Re-run installer** (if `grimorio update` fails):
```bash
curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | sh -s -- --update
```

**Developers** — Build from source:
```bash
git clone https://github.com/pauvalls/grimorio.git
cd grimorio
make install
```

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
| **[Campaign Consistency](docs/campaign-consistency.md)** | **NEW**: Health monitoring, rollback, dynamic content, persistent consequences |
| **[Developer Guide](docs/developer-guide.md)** | Contributing to Grimorio |

### What's New

See the [CHANGELOG](CHANGELOG.md) for the full list of changes by version.

**v4.1.0 — Campaign Consistency & Dynamic Content (Latest)**
- **📊 Campaign Health Monitoring** — Automated detection of stale quests, faction contradictions, dead NPC mismatches, and lore drift
- **🔄 Context Compression** — Rolling session summaries reduce DM payload >50% for 20+ session campaigns
- **🎯 Dynamic Location Content** — Location-aware random tables with faction reputation weighting (±80%)
- **🗺️ On-Demand Area Generation** — `generate_dynamic_area` creates contextual areas when players go off-map
- **⏳ Persistent Delayed Effects** — Consequence engine now persists scheduled effects across sessions
- **↩️ Rollback Capability** — Restore campaign state from checkpoints with SHA256 integrity validation
- **📝 Audit Logging** — Append-only JSONL log of all consistency gate approvals (90-day auto-purge)
- **🔍 Multi-Session Prep** — `Previously On` shows last 3 sessions + arc context
- **🎭 Enriched Scenarios** — Session prep prioritizes pending effects, unresolved decisions, faction changes

See **[Campaign Consistency Guide](docs/campaign-consistency.md)** for complete usage reference.

**v4.0.10 — Update Commands & Campaign Template (Latest)**
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

> 💡 **Creado para OpenCode, adaptable a cualquier sistema.** Grimorio está diseñado como servidor MCP con integración nativa para OpenCode (agents, skills, commands), pero las herramientas subyacentes y el compilador de PDF funcionan con cualquier asistente de IA o flujo de trabajo CLI.

### Inicio Rápido

**Linux / macOS:**
```bash
curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | sh
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/pauvalls/grimorio/main/install.ps1 | iex
```
> ℹ️ **Soporte de Windows** — la extracción de texto PDF requiere `poppler-utils`. Instalalo dentro de WSL (`wsl --install` y luego `sudo apt install poppler-utils`) u omití el flag `include_pdf_text`. La matriz de CI valida builds de Windows en cada PR.

**¿Ya lo tienes instalado? Actualízalo:**
```bash
grimorio update          # Actualizar solo el binario
grimorio update skills   # Actualizar skills
grimorio update agents   # Actualizar agents
grimorio update commands # Actualizar config de opencode.json
grimorio update all      # Actualizar todo de una
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

### Cómo Usar

**1. Crear una Campaña** (one-shot o campaña completa)

En OpenCode, cambia al agente `grimorio-architect` y ejecuta:
```
/grimorio Una ciudad hundida donde los nobles son vampiros acuáticos
```
El architect te hará 6 preguntas (nombre, tipo, idea, nivel, tono, duración) y luego generará la campaña completa en batches:
- **Batch 1**: Trasfondo, NPCs, bestiario, mapas, guía de setting, introducción
- **Batch 2**: Quests, encuentros, personajes, apéndices
- **Batch 3**: Áreas (capítulo por capítulo para evitar timeout)
- **Final**: Arte, compilación de PDF, validaciones

Todos los archivos se guardan en `~/campaigns/<nombre-campaña>/`.

**2. Ejecutar una Sesión en Vivo** (asistente de DM)

En OpenCode, cambia al agente `grimorio-dm` dentro del directorio de tu campaña:
```bash
cd ~/campaigns/ciudad-hundida
grimorio-dm
```
El agente `grimorio-dm` carga el contexto completo de la campaña (canon, áreas, NPCs, bestiario) y actúa como tu Dungeon Master IA:
- Rastrea el estado de la sesión (pistas reveladas, NPCs conocidos, decisiones tomadas)
- Ejecuta combate con estados de daño descriptivos (oculta HP/AC de enemigos)
- Mantiene coherencia narrativa entre sesiones
- Genera preparación de sesión, handouts y tablas aleatorias

**🔊 Narración por Voz (TTS) — Experimental**

> ⚠️ **Característica Experimental** — TTS es un complemento opcional que requiere configuración manual. Funciona muy bien una vez configurado, pero no está habilitado por defecto.

Grimorio puede narrar las respuestas del DM en voz alta usando **Piper TTS** (síntesis de voz neuronal local). El agente `grimorio-dm` narrará automáticamente cada respuesta narrativa cuando TTS está activado.

**Configuración Rápida:**
```bash
# 1. Descargar el binario precompilado de Piper
mkdir -p ~/.local/bin/piper
wget https://github.com/rhasspy/piper/releases/download/v1.2.0/piper_amd64.tar.gz -O /tmp/piper.tar.gz
tar -xzf /tmp/piper.tar.gz -C ~/.local/bin/piper

# 2. Descargar una voz (ejemplo en español)
mkdir -p ~/.local/share/piper
wget https://huggingface.co/rhasspy/piper-voices/resolve/main/es/es_ES/davefx/medium/es_ES-davefx-medium.onnx -P ~/.local/share/piper/
wget https://huggingface.co/rhasspy/piper-voices/resolve/main/es/es_ES/davefx/medium/es_ES-davefx-medium.onnx.json -P ~/.local/share/piper/

# 3. Crear scripts auxiliares
ln -s ~/.local/bin/piper/piper ~/.local/bin/piper
# Crear scripts tts-dm.sh y narrate (ver docs/tts-setup.md)

# 4. Configurar variables de entorno
export PIPER_MODEL_PATH="$HOME/.local/share/piper/es_ES-davefx-medium.onnx"
export PIPER_CONFIG_PATH="$HOME/.local/share/piper/es_ES-davefx-medium.onnx.json"
export PATH="$HOME/.local/bin:$PATH"

# 5. Aplicar a OpenCode
grimorio update commands
```

**Cómo Funciona en las Sesiones:**
1. Inicia una sesión con `grimorio-dm`
2. El agente pregunta: "🔊 TTS: Disponible. ¿Activar? Sí/No"
3. Si Sí, cada respuesta narrativa se narra automáticamente
4. El agente escribe texto → muestra "🎙️ Narrando..." → ejecuta TTS en background

**Características:**
- **Local y Gratis** — Corre en tu máquina, sin API keys, sin nube
- **División automática** — Divide texto largo en chunks de ~150 caracteres por oración
- **Audio en background** — Usa `setsid` para desprender del shell, sobrevive timeouts
- **Múltiples voces** — Explora [Piper Voices](https://huggingface.co/rhasspy/piper-voices/tree/main)

**Ver [docs/tts-experimental.md](docs/tts-experimental.md) para configuración completa, solución de problemas y personalización de voces.**

**3. Actualizar Contenido** (después de cambios manuales o ediciones)

Si editas archivos markdown manualmente, refresca el canon:
```bash
grimorio validate_canon --campaign ciudad-hundida
grimorio check_consistency --campaign ciudad-hundida
```

### Instalación Avanzada

**Re-ejecutar instalador** (si `grimorio update` falla):
```bash
curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | sh -s -- --update
```

**Developers** — Compilar desde fuente:
```bash
git clone https://github.com/pauvalls/grimorio.git
cd grimorio
make install
```

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
| **[Consistencia de Campaña](docs/campaign-consistency.md)** | **NUEVO**: Salud, rollback, contenido dinámico, consecuencias persistentes |
| **[Guía de Desarrollador](docs/developer-guide.md)** | Contribuir a Grimorio |

### Novedades

Ver el [CHANGELOG](CHANGELOG.md) para la lista completa de cambios por versión.

**v4.1.0 — Consistencia de Campaña y Contenido Dinámico (Último)**
- **📊 Monitoreo de Salud** — Detección automática de quests estancadas, contradicciones de facciones, NPCs muertos inconsistentes, drift de lore
- **🔄 Compresión de Contexto** — Resúmenes de sesiones reducen payload >50% en campañas de 20+ sesiones
- **🎯 Contenido por Ubicación** — Tablas aleatorias location-aware con weighting por reputación (±80%)
- **🗺️ Generación On-Demand** — `generate_dynamic_area` crea áreas contextuales cuando jugadores van fuera del mapa
- **⏳ Efectos Diferidos Persistentes** — Consecuencias programadas persisten entre sesiones
- **↩️ Rollback** — Restaurar estado desde checkpoints con validación SHA256
- **📝 Audit Log** — Log JSONL append-only de aprobaciones del gate (auto-purge 90 días)
- **🔍 Multi-Sesión Prep** — `Previously On` muestra últimas 3 sesiones + contexto de arco
- **🎭 Escenarios Enriquecidos** — Prep prioriza efectos pendientes, decisiones sin resolver, cambios de facción

Ver **[Guía de Consistencia de Campaña](docs/campaign-consistency.md)** para referencia completa de uso.

**v4.0.10 — Update Commands y Template de Campaña (Último)**
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
| Go 1.24+ | ❌ Only for source build / Solo para compilar desde fuente | Build binary / Compilar binario |
| Chrome/Chromium/Edge | ⚠️ Must install / Debes instalarlo | Compile PDF (headless) / Compilar PDF |
| Git | ❌ Must have / Debes tener | Clone repo / Clonar repo |

> 💡 Grimorio auto-detects your PDF engine. It prefers **Chromium/Chrome headless** (modern, maintained) and falls back to **wkhtmltopdf** (legacy) if needed.

### PDF Engine — Install by Platform / Instalar por Plataforma

**Linux (Arch / CachyOS / Manjaro):**
```bash
sudo pacman -S chromium
# or google-chrome-stable from AUR
```

**Linux (Debian/Ubuntu):**
```bash
sudo apt-get install chromium-browser
# or download Google Chrome from google.com/chrome
```

**macOS:**
```bash
brew install --cask google-chrome
# or: brew install --cask chromium
```

**Windows:**
```powershell
winget install Google.Chrome
# or: choco install googlechrome
# Edge is already installed on Windows 10/11 and works too
```

**Legacy (wkhtmltopdf) — not recommended for new installs:**
```bash
# Only if you cannot install Chrome/Chromium
sudo pacman -S wkhtmltopdf        # Arch
sudo apt-get install wkhtmltopdf  # Debian/Ubuntu
brew install --cask wkhtmltopdf   # macOS
```

## Troubleshooting / Solución de Problemas

### grimorio: command not found / grimorio: comando no encontrado

The binary is installed to `~/.local/bin/` on Linux/macOS or `%LOCALAPPDATA%\Grimorio\` on Windows. If the command is not found, add the directory to your PATH.

**Linux / macOS:**
```bash
export PATH="$HOME/.local/bin:$PATH"
```
To make it permanent, add the above line to `~/.bashrc`, `~/.zshrc`, or your shell's config file.

**Windows (PowerShell):**
```powershell
$env:Path += ";$env:LOCALAPPDATA\Grimorio"
```

### Permission denied / Permiso denegado

If you get `Permission denied` when running `grimorio`, ensure the binary is executable:

**Linux / macOS:**
```bash
chmod +x ~/.local/bin/grimorio
```

### Windows Execution Policy / Política de Ejecución de Windows

If PowerShell shows an execution policy error, run:
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Update fails / El update falla

If `grimorio update` fails, try re-running the installer:
```bash
curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | sh -s -- --update
```

Or on Windows:
```powershell
irm https://raw.githubusercontent.com/pauvalls/grimorio/main/install.ps1 | iex
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

## Reference Campaigns / Campañas de Referencia

**📚 [`examples/la-hoja-de-vlad/`](examples/la-hoja-de-vlad/)** — A complete 3-act gothic-political campaign (level 1–5) generated end-to-end with Grimorio: 12 NPCs, 28 files, a WotC-style compiled PDF, and 17 consistency rules passing. Use it as:

- A learning artifact: read the [annotated walkthrough](docs/walkthroughs/la-hoja-de-vlad.md) to see how lore, NPCs, chapters, and handouts fit together.
- A regression baseline: the validation engine's `wotc_*` rules are tested against this campaign.
- A starting point: copy the directory to `~/campaigns/` and remix.

> 🩸 **La Hoja de Vlad** is the campaign in this repo. It's the canonical answer to "what does a finished Grimorio campaign look like?".

---

## Campañas de Referencia

**📚 [`examples/la-hoja-de-vlad/`](examples/la-hoja-de-vlad/)** — Una campaña gótico-política completa de 3 actos (nivel 1-5) generada end-to-end con Grimorio: 12 NPCs, 28 archivos, un PDF compilado estilo WotC y 17 reglas de consistencia pasando. Usala como:

- Artefacto de aprendizaje: leé el [walkthrough anotado](docs/walkthroughs/la-hoja-de-vlad.md) para ver cómo encajan lore, NPCs, capítulos y handouts.
- Baseline de regresión: las reglas `wotc_*` del motor de validación se prueban contra esta campaña.
- Punto de partida: copiá el directorio a `~/campaigns/` y remixalo.

> 🩸 **La Hoja de Vlad** es la campaña de este repo. Es la respuesta canónica a "¿cómo se ve una campaña de Grimorio terminada?".

---

## License / Licencia

**Mozilla Public License 2.0** — See [LICENSE](LICENSE) for details.

Copyright (c) 2026 Pau Valls

---

## Contributing / Contribuir

See [Developer Guide](docs/developer-guide.md) for contribution guidelines.

Ver [Guía de Desarrollador](docs/developer-guide.md) para directrices de contribución.
