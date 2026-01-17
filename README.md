# Multi-Agent Development System

Sistema multi-agente autónomo para desarrollo de software con guardrails y auto-reparación.

## Arquitectura

```
orchestrator/     # FSM, cola de tareas, estado de ejecución
agents/           # planner, coder, tester, auditor, optimizer, repairer
workspace/        # git manager, branches, patches
policies/         # guardrails, gates obligatorios
tools/            # sandbox de comandos permitidos
```

## Agentes

- **Orchestrator**: Asigna tareas, decide siguiente paso, aplica políticas
- **Planner**: Descompone objetivos en tickets ejecutables
- **Coder**: Implementa cambios en el repo (solo /src, /cmd)
- **Tester**: Ejecuta tests, genera reportes (no modifica producción)
- **Auditor**: Revisa seguridad, estilo, dependencias, licencias, secretos
- **Repairer**: Auto-repara basado en fallos del pipeline
- **Optimizer**: Optimiza código con validación (benchmarks, profiling)

## Flujo de Ciclo Completo

1. **Objetivo** → Planificador crea subtareas
2. **Coder** implementa
3. **Tester** ejecuta pipeline (unit, integration, fuzz)
4. Si falla → **Repairer** analiza y propone fix
5. Si pasa → **Auditor** revisa (seguridad, estilos, deps)
6. Si auditor falla → vuelve a reparación
7. Si todo ok → **Optimizer** mejora (sin romper tests)
8. **Release** empaqueta y versiona

## Guardrails

- ✅ fmt/lint obligatorio
- ✅ tests deben pasar
- ✅ cobertura mínima
- ✅ escaneo de secretos
- ✅ análisis de dependencias (CVEs)
- ✅ revisión de cambios de riesgo

## Uso

```bash
# Instalar dependencias
go mod download

# Ejecutar sistema multi-agente
go run cmd/orchestrator/main.go --task "fix bug in /api/users"

# O usar CLI
go run cmd/cli/main.go plan --objective "optimize endpoint /api/search"
```
