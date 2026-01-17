# Arquitectura del Sistema Multi-Agente

## Visión General

El sistema multi-agente está diseñado para proporcionar autonomía controlada en el desarrollo de software, con guardrails y auto-reparación.

## Componentes Principales

### 1. Orchestrator (`pkg/orchestrator/`)

El cerebro del sistema que:
- Gestiona la cola de tareas
- Asigna tareas a agentes apropiados
- Aplica políticas y guardrails
- Mantiene estado de ejecución
- Decide el siguiente paso basado en resultados

**Flujo de Estado:**
```
PENDING → RUNNING → SUCCESS/FAILED
              ↓
         RETRYING (si falla y hay retries)
```

### 2. Agents (`pkg/agents/`)

Cada agente tiene un contrato específico que define:
- Rutas permitidas/prohibidas
- Herramientas permitidas
- Si requiere tests
- Restricciones adicionales

#### Planner
- **Rol**: Descompone objetivos en subtareas ejecutables
- **No modifica código**
- **Salida**: Lista de subtareas ordenadas

#### Coder
- **Rol**: Implementa cambios en el código
- **Permitido**: `src/**`, `cmd/**`, `internal/**`, `pkg/**`
- **Prohibido**: `**/*_test.go`, `vendor/**`
- **Requiere**: Tests para cambios de lógica

#### Tester
- **Rol**: Ejecuta tests y genera reportes
- **Permitido**: Solo archivos `**/*_test.go`
- **No modifica**: Código de producción
- **Herramientas**: `go test`, coverage tools

#### Auditor
- **Rol**: Revisa seguridad, estilo, dependencias
- **Solo lectura**: No modifica código
- **Checks**: Lint, SAST, secrets, CVEs
- **Salida**: Hallazgos con severidad

#### Repairer
- **Rol**: Auto-repara fallos del pipeline
- **Permitido**: Mismo que Coder
- **Estrategia**: Analiza fallos → Propone fixes → Aplica

#### Optimizer
- **Rol**: Optimiza código con validación
- **Permitido**: Mismo que Coder
- **Valida**: Tests deben seguir pasando
- **Métricas**: Benchmarks antes/después

### 3. Policies Engine (`pkg/policies/`)

Gestiona guardrails y gates obligatorios:

#### Gates Obligatorios
1. **fmt/lint**: Código debe pasar format y lint
2. **tests-pass**: Todos los tests deben pasar
3. **coverage**: Cobertura mínima (70%)
4. **secrets**: No secretos expuestos
5. **dependencies**: No CVEs críticos
6. **risk-review**: Warning para cambios de alto riesgo

#### Políticas por Agente
- Restricciones de rutas
- Herramientas permitidas
- Límites de recursos
- Validaciones específicas

### 4. Workspace Manager (`pkg/workspace/`)

Gestiona el repositorio git:
- Crear/cambiar ramas
- Aplicar patches
- Ejecutar comandos en sandbox
- Obtener diffs
- Crear commits

## Flujo de Ciclo Completo

```
1. Objetivo recibido
   ↓
2. Planner crea subtareas:
   - TaskCode
   - TaskTest
   - TaskAudit
   ↓
3. Coder implementa cambios
   ↓
4. Tester ejecuta tests
   ↓
5a. Si falla → Repairer
     ↓
     Vuelve a 4
   ↓
5b. Si pasa → Auditor
   ↓
6a. Si auditor encuentra críticos → Repairer
     ↓
     Vuelve a 4
   ↓
6b. Si todo ok → Optimizer
   ↓
7. Optimizer mejora (validado con tests)
   ↓
8. Release (empaquetar, versionar)
```

## Auto-Reparación

El Repairer funciona en ciclos cortos:

1. **Detecta fallo**: Test rojo, crash, métrica mala
2. **Genera hipótesis**: "nil pointer por path X"
3. **Aplica fix mínimo**
4. **Agrega test de regresión**
5. **Repite hasta éxito o max retries**

**Estrategias comunes:**
- Nil pointer checks
- Missing definitions
- Type mismatches
- Lint errors
- Security issues

## Auto-Auditoría

El Auditor ejecuta herramientas reales:

- **SAST**: `go vet`, `golangci-lint`
- **Secrets**: `gitleaks`, `trufflehog`
- **Dependencies**: `nancy`, `snyk`
- **Policy checks**: Reglas personalizadas

## Auto-Optimización

El Optimizer solo optimiza si:
- Los tests siguen pasando
- El benchmark no empeora >3%
- Es una optimización segura (p. ej., remove unused imports)

**Tipos de optimización:**
- Safe: remove unused, simplify expressions
- Tested: optimize loops, cache results (requiere tests)

## Tipos de Mensajes

Todos los agentes comunican usando `TaskResult`:

```go
type TaskResult struct {
    TaskID    string
    State     TaskState
    Success   bool
    Outputs   map[string]interface{}
    Evidence  []Evidence  // logs, reports, diffs
    Decisions []Decision  // decisiones tomadas
    Error     string
    Duration  time.Duration
}
```

## Memoria de Decisiones

El Orchestrator mantiene un log estructurado de todas las decisiones:
- Qué agente
- Por qué razón
- Qué acción
- Nivel de confianza

Útil para debugging y mejorar políticas.
