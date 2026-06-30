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
The architect will first ask you to pick your language (English or Spanish), then ask 6 more questions (name, type, idea, level, tone, duration) and then generate the full campaign in batches:
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

**3. Validate & Update Content** (after rule changes or edits)

Validate your campaign without compiling the PDF:
```bash
grimorio validate sunken-city                 # All checks, human output
grimorio validate sunken-city --scope=wotc    # WotC format only
grimorio validate sunken-city --scope=structure # Structure only
grimorio validate sunken-city --json          # Machine-readable output
```

Refresh canon after manual edits:
```bash
grimorio validate_canon --campaign sunken-city
grimorio check_consistency --campaign sunken-city
```

**4. Export to Other Formats** (optional)

In addition to PDF, export your campaign to:
```bash
grimorio export_campaign --campaign sunken-city --format=pdf       # Default
grimorio export_campaign --campaign sunken-city --format=markdown  # Concatenated .md
grimorio export_campaign --campaign sunken-city --format=epub      # EPUB 3
```

**5. Check Campaign Health** (optional)

Get a 0-100 score across six axes of campaign quality:
```
Use the campaign_health_dashboard MCP tool from any agent.
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
| **[Getting Started](docs/getting-started.md)** | Installation, first campaign, brief template, `validate` CLI, exports |
| **[Session Tutorial](docs/tutorials/session-tutorial.md)** | Run your first game (prep → play → post-session) |
| **[Character Creation](docs/tutorials/character-creation.md)** | Generate PCs, pre-gens, character worksheets |
| **[Session Generator](docs/tutorials/session-generator.md)** | Adapt sessions to specific characters |
| **[PDF Compiler](docs/features/pdf-compiler.md)** | Customize PDF output, CSS styles, sections |
| **[MCP Tools](docs/features/mcp-tools.md)** | Full tool reference (40+ tools, including health, export, treasure) |
| **[Architecture](docs/features/architecture.md)** | How Grimorio works internally (image cache, V2/V3 repos, compiler strategy) |
| **[DM Guide](docs/dm-guide.md)** | General advice for running games |
| **[DM Agent Guide](docs/dm-agent-guide.md)** | Running live sessions with the AI Dungeon Master |
| **[Campaign Consistency](docs/campaign-consistency.md)** | Health monitoring, rollback, dynamic content, persistent consequences |
| **[Migration v2→v3](docs/migration-v2-to-v3.md)** | Migrate legacy `areas/` campaigns to `chapters/` |
| **[Walkthrough: la-hoja-de-vlad](docs/walkthroughs/la-hoja-de-vlad.md)** | Annotated tour of the reference campaign |
| **[Developer Guide](docs/developer-guide.md)** | Contributing to Grimorio |

### What's New

See the [CHANGELOG](CHANGELOG.md) for the full release history.

**Latest — v5.4.x (PDF Compiler Hardening)**
- **📐 Cover page fits one A4 sheet** — `position: absolute` + `@page :first { margin: 0 }` anchors the cover to the page, so the title, hero image, and "Generated by Grimorio" footer no longer split across two pages.
- **🖼️ Monster hero images inside stat blocks** — `peekHoistableMonsterImage` hoists `monster-*.png` files (e.g. El Rayo, Cabra de Dos Cabezas) into the just-emitted `.stat-block` div, matching WotC adventure convention.
- **📖 Inline stat-block flow** — Trait lines with inner `**bold**` spans (e.g. *Radiación Distorsionante (pasiva). El gólem tiene **resistencia** …*) render as a single `<p class="trait">` instead of being split into 3 `.property-line` divs.
- **🎯 Image MIME detection from magic bytes** — JPEG bytes embedded as PNG (the old "broken cover image" bug) are now detected and served as `data:image/jpeg`.

See **[Getting Started](docs/getting-started.md)** for a quick tour of the commands and tools.

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
El architect primero te preguntará el idioma (inglés o español), luego te hará 6 preguntas más (nombre, tipo, idea, nivel, tono, duración) y luego generará la campaña completa en batches:
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

**3. Validar y Actualizar Contenido** (después de cambios manuales o ediciones)

Valida tu campaña sin compilar el PDF:
```bash
grimorio validate ciudad-hundida                 # Todos los checks, salida humana
grimorio validate ciudad-hundida --scope=wotc    # Solo formato WotC
grimorio validate ciudad-hundida --scope=structure # Solo estructura
grimorio validate ciudad-hundida --json          # Salida machine-readable
```

Refrescá el canon después de ediciones manuales:
```bash
grimorio validate_canon --campaign ciudad-hundida
grimorio check_consistency --campaign ciudad-hundida
```

**4. Exportar a Otros Formatos** (opcional)

Además de PDF, exportá tu campaña a:
```bash
grimorio export_campaign --campaign ciudad-hundida --format=pdf       # Default
grimorio export_campaign --campaign ciudad-hundida --format=markdown  # .md concatenado
grimorio export_campaign --campaign ciudad-hundida --format=epub      # EPUB 3
```

**5. Chequear Salud de Campaña** (opcional)

Obtené un puntaje 0-100 en seis ejes de calidad de campaña:
```
Usá la tool MCP campaign_health_dashboard desde cualquier agente.
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

Ver el [CHANGELOG](CHANGELOG.md) para el historial completo de versiones.

**Último — v5.4.x (Hardening del Compilador PDF)**
- **📐 Portada cabe en una hoja A4** — `position: absolute` + `@page :first { margin: 0 }` ancla la portada a la página, así el título, imagen hero y footer "Generated by Grimorio" ya no se parten en dos páginas.
- **🖼️ Imágenes hero de monstruos dentro de los stat blocks** — `peekHoistableMonsterImage` sube los archivos `monster-*.png` (ej. El Rayo, Cabra de Dos Cabezas) dentro del `.stat-block` recién emitido, siguiendo la convención WotC.
- **📖 Flujo inline de stat blocks** — Las traits con `**bold**` internos (ej. *Radiación Distorsionante (pasiva). El gólem tiene **resistencia** …*) se renderizan como un único `<p class="trait">` en lugar de partirse en 3 `.property-line`.
- **🎯 Detección de MIME por magic bytes** — Bytes JPEG embebidos como PNG (el viejo bug de "imagen de portada rota") ahora se detectan y se sirven como `data:image/jpeg`.

Ver **[Empezando](docs/getting-started.md)** para un tour rápido de los comandos y herramientas.

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
    ├── chapters/              # Chapters with areas (WotC format) / Capítulos con áreas
    │   ├── drafts/            # Sequential generation drafts / Borradores de generación secuencial
    │   ├── chapter_00.md      # Prologue chapter (optional) / Capítulo de prólogo (opcional)
    │   └── chapter_01.md      # Chapter 1 / Capítulo 1
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
│  ├─ grimorio-maps       ├─ grimorio-chapters (WotC format)  │
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
│                                                              │
│  Quality & Health (v5.0):                                   │
│  ├─ campaign_health_dashboard                               │
│  ├─ export_campaign (pdf|markdown|epub)                     │
│  ├─ generate_treasure                                       │
│  └─ grimorio validate CLI                                   │
│                                                              │
│  Monster CR Validation (v5.2):                              │
│  ├─ validate_monster  (CR vs DMG cap. 9)                    │
│  ├─ suggest_monster_cr  (skeleton for target CR)            │
│  └─ audit_monster_cr  (full bestiary CR drift)               │
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

## License / Licencia

**Mozilla Public License 2.0** — See [LICENSE](LICENSE) for details.

Copyright (c) 2026 Pau Valls

---

## Contributing / Contribuir

See [Developer Guide](docs/developer-guide.md) for contribution guidelines. / Ver [Guía de Desarrollador](docs/developer-guide.md) para directrices de contribución.
