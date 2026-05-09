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
- wkhtmltopdf (for PDF compilation)
- Grimorio binary to `~/.local/bin/grimorio`
- Plugin files to your AI assistant's plugin directory

## Requirements

| Dependency | Auto-installed | Purpose |
|------------|---------------|---------|
| Go 1.24+ | ✅ Yes | Build the MCP server binary |
| wkhtmltopdf | ✅ Yes | Compile HTML to PDF |
| Git | ❌ Must have | Clone the repository |

## Your First Campaign

Once installed, type in your AI assistant chat:

```
/grimorio A sunken city where the nobles are aquatic vampires
```

The AI will ask you 5 questions:
1. Campaign name (kebab-case, e.g., "sunken-city")
2. One-shot or full campaign?
3. Player level range?
4. Desired tone?
5. Duration?

Then it generates everything automatically:
- Lore, NPCs, bestiary, encounters, maps
- Acts with numbered areas (WotC format)
- AI-generated cover art and illustrations
- Procedural battle maps (SVG)
- Professional PDF (D&D-styled layout)

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
- wkhtmltopdf (para compilar PDFs)
- Binario de Grimorio en `~/.local/bin/grimorio`
- Archivos del plugin en el directorio de plugins de tu asistente IA

## Requisitos

| Dependencia | Auto-instalado | Propósito |
|------------|---------------|---------|
| Go 1.24+ | ✅ Sí | Compilar el binario del servidor MCP |
| wkhtmltopdf | ✅ Sí | Compilar HTML a PDF |
| Git | ❌ Debes tenerlo | Clonar el repositorio |

## Tu Primera Campaña

Una vez instalado, escribe en el chat de tu asistente IA:

```
/grimorio Una ciudad hundida donde los nobles son vampiros acuáticos
```

La IA te hará 5 preguntas:
1. Nombre de campaña (kebab-case, ej. "ciudad-hundida")
2. ¿One-shot o campaña completa?
3. ¿Rango de nivel de jugadores?
4. ¿Tono deseado?
5. ¿Duración?

Luego genera todo automáticamente:
- Trasfondo, NPCs, bestiario, encuentros, mapas
- Actos con áreas numeradas (formato WotC)
- Portada e ilustraciones generadas por IA
- Mapas de batalla procedimentales (SVG)
- PDF profesional (estilo D&D)

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
