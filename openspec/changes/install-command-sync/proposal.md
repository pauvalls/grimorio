# Proposal: Install Command Sync with grimorio-architect

## Intent
Sincronizar automáticamente el template del comando `/grimorio` en opencode.json con el contenido actualizado de `agents/grimorio-architect.md`, asegurando que las preguntas de Phase 1 (incluyendo "Campaign idea / brief description") estén siempre actualizadas.

## Scope
- **install.sh**: Agregar función `sync_command_from_agent` que lee el workflow de grimorio-architect.md
- **install.sh**: Actualizar template hardcoded para incluir pregunta 3 sobre brief description
- **configure_opencode_command**: Llamar a sync_command_from_agent antes de configurar el agente

## Approach
1. Extraer el workflow Phase 1 de agents/grimorio-architect.md
2. Generar el template del comando dinámicamente desde el archivo del agente
3. Mantener compatibilidad hacia atrás si el archivo del agente no existe

## Benefits
- ✅ Single source of truth: el workflow vive solo en grimorio-architect.md
- ✅ Actualizaciones automáticas: al cambiar el agente, el comando se actualiza en el próximo install
- ✅ Reduce duplicación: no hay que mantener dos copias del mismo workflow
- ✅ Previene drift: el comando y el agente siempre están sincronizados
