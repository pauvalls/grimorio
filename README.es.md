# 📜 Grimorio

**Generador de campañas y one-shots de D&D 5e con IA para OpenCode.**

Convierte una idea en un libro de aventuras listo para imprimir: trasfondo, NPCs, bestiario,
encuentros, mapas, arte de portada y un PDF con estilo WotC. Funciona como servidor MCP, así que las
mismas herramientas están disponibles desde cualquier agente que pueda llamar a MCP.

**[English](README.md) · [Español](README.es.md)**

---

## Instalación

**Linux / macOS**

```bash
curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | sh
```

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/pauvalls/grimorio/main/install.ps1 | iex
```

El instalador deja el binario en `~/.grimorio/grimorio` y lo enlaza en `~/.local/bin/grimorio`.
En Linux y macOS además configura OpenCode: servidor MCP, skills, agents y el comando `/grimorio`.

Necesita `python3` para fusionar `opencode.json`. Si no está, imprime la entrada exacta para pegar
en vez de escribir una configuración rota.

## Verificación

```bash
grimorio --version        # grimorio version vX.Y.Z
opencode mcp list         # ✓ grimorio  connected
```

Dentro de OpenCode también deben aparecer los agentes `grimorio-architect` y `grimorio-dm`, las
skills `grimorio-*` y el comando `/grimorio`.

## Uso

**Generar una campaña** — se pide en lenguaje natural:

```
/grimorio Una ciudad hundida donde los nobles son vampiros acuáticos
```

El arquitecto pregunta el idioma, luego seis cosas (nombre, one-shot o campaña, rango de nivel,
tono, duración) y genera la campaña capítulo por capítulo. Todo queda en
`~/campaigns/<nombre-campaña>/` y termina en un PDF.

**Dirigir una sesión en vivo** — cambia al agente `grimorio-dm` dentro del directorio de la campaña.
Carga canon, áreas, NPCs y bestiario, lleva el estado de la sesión, oculta HP/CA de los enemigos y
mantiene la narración coherente entre sesiones. Narración por voz local opcional con Piper TTS.

**Validar y exportar** — la validación es un comando CLI; el formato de salida es una tool MCP:

```bash
grimorio validate sunken-city                  # consistencia + formato WotC + estructura
grimorio validate sunken-city --json           # machine-readable, para CI
grimorio campaign export sunken-city           # archivo tar.gz del directorio de campaña
```

El PDF (default), Markdown y EPUB salen de la tool MCP `export_campaign`, así que hay que pedírselos
al agente (por ejemplo, "exportá sunken-city como epub").

## Actualización

```bash
grimorio update all      # integración OpenCode + skills + agents + binario (recomendado)
grimorio update          # solo el binario
grimorio update commands # solo config de OpenCode y el comando /grimorio
grimorio update skills   # solo skills
grimorio update agents   # solo agents
```

Las actualizaciones son idempotentes: hacen backup de `opencode.json`, migran layouts anteriores y
conservan toda la configuración que no les pertenece. Si algo falla, se puede repetir el instalador
con `sh -s -- --update`.

## Documentación

**Para empezar**

| Documento | Qué cubre |
|-----------|-----------|
| [Empezando](docs/getting-started.md) | Instalación, requisitos, primera campaña, plantilla de brief, `validate`, exportaciones |
| [Tutorial de Sesión](docs/tutorials/session-tutorial.md) | Primera partida: preparación → juego → post-sesión |
| [Walkthrough: La Hoja de Vlad](docs/walkthroughs/la-hoja-de-vlad.md) | Recorrido comentado de una campaña terminada |

**Generar**

| Documento | Qué cubre |
|-----------|-----------|
| [Referencia de herramientas MCP](docs/features/mcp-tools.md) | Todas las herramientas, agrupadas por tarea |
| [Compilador PDF](docs/features/pdf-compiler.md) | Layout, CSS, secciones, control de salida |
| [Mejoras del compilador PDF](docs/pdf-compiler-enhancements.md) | Funciones de layout más recientes |
| [Formato de comandos save](docs/SAVE_COMMANDS_FORMAT.md) | Markdown que espera cada herramienta `save_*` |
| [Reglas de diseño de monstruos](docs/dnd-monster-design-rules.md) | Cálculo de CR y formato de stat block (DMG cap. 9) |
| [Creación de personajes](docs/tutorials/character-creation.md) | PJs, pre-generados, hojas de trabajo |

**Dirigir y jugar**

| Documento | Qué cubre |
|-----------|-----------|
| [Guía del agente DM](docs/dm-agent-guide.md) | Sesiones en vivo con `grimorio-dm` |
| [Guía de DM](docs/dm-guide.md) | Consejos generales para dirigir |
| [Consistencia de campaña](docs/campaign-consistency.md) | Puntaje de salud, rollback, consecuencias persistentes |
| [Generador de sesiones](docs/tutorials/session-generator.md) | Adaptar sesiones a personajes concretos |
| [TTS (experimental)](docs/tts-experimental.md) | Instalación de Piper, voces, script de narración |
| [Roadmap de TTS](docs/tts-roadmap.md) | Hacia dónde va la narración |

**Referencia y contribución**

| Documento | Qué cubre |
|-----------|-----------|
| [Arquitectura](docs/features/architecture.md) | Capas, almacenamiento, abstracciones clave |
| [Guía de desarrollo](docs/developer-guide.md) | Compilar, testear, lint, release |
| [Migración v2 → v3](docs/migration-v2-to-v3.md) | Pasar campañas con `areas/` a `chapters/` |
| [Changelog](CHANGELOG.md) | Historial de versiones |

## Requisitos

| Dependencia | Cuándo hace falta |
|-------------|-------------------|
| Chrome / Chromium | Compilar el PDF (motor preferido) |
| wkhtmltopdf | Solo como respaldo de PDF si no se puede instalar Chrome |
| `python3` | Merge de configuración durante instalación y actualización |
| Go 1.24+ | Solo para compilar desde el código fuente |
| `poppler-utils` | Solo para extraer texto de PDF en Windows/WSL |

## Compilar desde el código fuente

```bash
git clone https://github.com/pauvalls/grimorio.git
cd grimorio
make build     # ./grimorio
make install   # ejecuta install.sh y configura OpenCode
make test      # tests + cobertura
make lint      # golangci-lint
```

## Problemas frecuentes

**`grimorio: command not found`** — agregar `~/.local/bin` al `PATH`.

**`opencode mcp list` no muestra grimorio** — ejecutar `grimorio update commands` y reabrir OpenCode.

**La actualización falla o quedó desfasada** — cada merge deja
`~/.config/opencode/opencode.json.backup.*`. Comparar `grimorio --version` con
`~/.grimorio/grimorio --version`; `grimorio update` reemplaza todos los directorios de instalación.
Si sigue trabado, repetir el instalador con `--update`.

**No se genera el PDF** — instalar Chrome/Chromium, o `wkhtmltopdf` como respaldo.

## Contribuir

Leer la [Guía de desarrollo](docs/developer-guide.md). Commits convencionales, tests junto al código
que verifican, y `make test` antes de abrir un PR.

## Licencia

[LICENSE](LICENSE)
