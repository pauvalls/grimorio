<div align="center">

# 📜 Grimorio

**D&D One-shot & Campaign Generator**

[![Version v3.0.0](https://img.shields.io/badge/version-v3.0.0-purple?style=for-the-badge)](#)

[English](#english) · [Español](#español)

</div>

---

<a name="english"></a>

## 🇬🇧 English

AI-powered D&D 5e campaign and one-shot generator. Turn a spark of an idea into a fully-formatted, print-ready PDF adventure book — complete with lore, NPCs, bestiary, encounters, and styled layouts inspired by official Wizards of the Coast manuals.

### Quick Start

```bash
curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | bash
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
- **100% Local** — No cloud dependencies required

### Documentation

| Topic | Description |
|-------|-------------|
| **[Getting Started](docs/getting-started.md)** | Installation, first campaign, brief template |
| **[Session Tutorial](docs/tutorials/session-tutorial.md)** | Run your first game (prep → play → post-session) |
| **[Character Creation](docs/tutorials/character-creation.md)** | Generate PCs, pre-gens, character worksheets |
| **[Session Generator](docs/tutorials/session-generator.md)** | Adapt sessions to specific characters |
| **[PDF Compiler](docs/features/pdf-compiler.md)** | Customize PDF output, CSS styles, sections |
| **[MCP Tools](docs/features/mcp-tools.md)** | Full tool reference (29 tools) |
| **[Architecture](docs/features/architecture.md)** | How Grimorio works internally |
| **[DM Guide](docs/dm-guide.md)** | General advice for running games |
| **[Developer Guide](docs/developer-guide.md)** | Contributing to Grimorio |

### What's New in v3.0.0

**Major Features:**
- **Milestone XP Tracking** — Per-chapter XP tables with party level progression
- **WotC Unified Area Format** — Sequential numbering (1-15 per chapter)
- **Enhanced Magic Items** — Full stat blocks with rarity and attunement
- **Combat Tactics Engine** — Intelligence-based tactics (instinctive to strategic)
- **Player-Facing Maps** — Automatic secret feature redaction
- **Session Zero Generator** — Campaign-specific guides with safety tools
- **Structured Quests** — 3+ distinct approaches (combat, social, stealth)
- **Consequence Tables** — Act transition tracking with faction reputation
- **PDF Compiler Enhancements** — Characters/Quests in PDF, WotC-style CSS, Shock Points
- **WotC Format Validation** — `check_consistency scope=full` validates developments (3-5 IF-THEN branches), multiple solutions (stealth/social/combat), character hooks, boxed text (100-600 words), NPC word count, and integration cross-references (act ↔ bestiary ↔ NPCs)
- **HTML Rendering Fix** — Fixed invalid `<p><div>` nesting for proper CSS rendering in PDF

See [CHANGELOG.md](CHANGELOG.md) for full details.

---

<a name="español"></a>

## 🇪🇸 Español

Generador de campañas y one-shots de D&D 5e potenciado por IA. Convierte una chispa de idea en un libro de aventuras en PDF completamente formateado, listo para imprimir — con trasfondo, NPCs, bestiario, encuentros, y diseños con estilo inspirado en los manuales oficiales de Wizards of the Coast.

### Inicio Rápido

```bash
curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | bash
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
- **100% Local** — Sin dependencias de nube requeridas

### Documentación

| Tema | Descripción |
|------|-------------|
| **[Empezando](docs/getting-started.md)** | Instalación, primera campaña, plantilla de brief |
| **[Tutorial de Sesión](docs/tutorials/session-tutorial.md)** | Ejecuta tu primer juego (prep → jugar → post-sesión) |
| **[Creación de Personajes](docs/tutorials/character-creation.md)** | Genera PJs, pre-generados, hojas de trabajo |
| **[Generador de Sesiones](docs/tutorials/session-generator.md)** | Adapta sesiones a personajes específicos |
| **[Compilador PDF](docs/features/pdf-compiler.md)** | Personaliza salida PDF, estilos CSS, secciones |
| **[Herramientas MCP](docs/features/mcp-tools.md)** | Referencia completa de herramientas (29 herramientas) |
| **[Arquitectura](docs/features/architecture.md)** | Cómo funciona Grimorio internamente |
| **[Guía de DM](docs/dm-guide.md)** | Consejos generales para dirigir juegos |
| **[Guía de Desarrollador](docs/developer-guide.md)** | Contribuir a Grimorio |

### Nuevo en v3.0.0

**Características Principales:**
- **Seguimiento de XP por Hitos** — Tablas de XP por capítulo con progresión de nivel
- **Formato Unificado de Áreas WotC** — Numeración secuencial (1-15 por capítulo)
- **Objetos Mágicos Mejorados** — Estadísticas completas con rareza y sintonización
- **Motor de Tácticas de Combate** — Tácticas basadas en inteligencia (instintivo a estratégico)
- **Mapas para Jugadores** — Redacción automática de características secretas
- **Generador de Sesión Cero** — Guías específicas de campaña con herramientas de seguridad
- **Misiones Estructuradas** — 3+ enfoques distintos (combate, social, sigilo)
- **Tablas de Consecuencias** — Seguimiento de transición de actos con reputación de facciones
- **Mejoras al Compilador PDF** — Personajes/Misiones en PDF, CSS estilo WotC, Puntos de Shock
- **Validación de Formato WotC** — `check_consistency scope=full` valida developments (3-5 ramas IF-THEN), múltiples soluciones (sigilo/social/combate), ganchos de personaje, boxed text (100-600 palabras), cantidad de palabras en NPCs, e integración cross-reference (act ↔ bestiary ↔ NPCs)
- **Fix de Renderizado HTML** — Corregido nesting inválido `<p><div>` para renderizado CSS correcto en PDF

Ver [CHANGELOG.md](CHANGELOG.md) para detalles completos.

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
│  └─ generate_flowchart                                      │
└─────────────────────────────────────────────────────────────┘
```

---

## License / Licencia

MIT License — See [LICENSE](LICENSE) for details.

---

## Contributing / Contribuir

See [Developer Guide](docs/developer-guide.md) for contribution guidelines.

Ver [Guía de Desarrollador](docs/developer-guide.md) para directrices de contribución.
