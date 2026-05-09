# Proposal: Clean Installer v2 - Complete MCP Installation

## Intent
Crear un instalador desde cero que sea **completo, limpio y autocontenido** para todo el sistema Grimorio MCP.

## Scope
- **install.sh**: Reescrito completo (no parches sobre el existente)
- **Objetivo**: Instalación limpia que borra todo lo anterior y regenera todo
- **Componentes**: Comandos, agentes, templates, binarios Go, configuración MCP

## Requirements
1. **Clean First**: Borrar TODA instalación anterior (plugins, binarios, config, shell)
2. **Download Latest**: Clonar repo fresco desde GitHub
3. **Build Go**: Compilar binarios con Go instalado
4. **Install All**: Copiar TODO (agents, commands, skills, templates, MCP config)
5. **Configure**: Actualizar opencode.json con commands Y agentes
6. **Idempotent**: Se puede ejecutar múltiples veces sin romper nada

## Benefits
- ✅ Single source: Un solo script hace todo
- ✅ Clean state: Sin residuos de instalaciones anteriores
- ✅ Complete: No faltan funciones ni configuraciones
- ✅ Maintainable: Código limpio, fácil de entender
- ✅ Testable: Fácil de verificar que funciona

## Risks
- Romper instalaciones existentes → **Mitigación**: Backup automático de campañas
- Perder configuraciones custom → **Mitigación**: Documentar que se resetea todo
