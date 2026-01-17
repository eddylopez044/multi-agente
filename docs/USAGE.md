# Guía de Uso

## Instalación

```bash
# Clonar el repositorio
git clone <repo-url>
cd multi-agent

# Instalar dependencias
go mod download

# Instalar herramientas de desarrollo (opcional)
make install-tools
```

## Uso Básico

### 1. Ejecutar Orchestrator

```bash
# Desde el directorio raíz del proyecto que quieres modificar
go run cmd/orchestrator/main.go \
  --task "fix bug in user authentication" \
  --repo .

# O construir y ejecutar
make build
./bin/orchestrator --task "optimize endpoint /api/search"
```

### 2. Usar CLI

```bash
# Crear un plan
go run cmd/cli/main.go plan --objective "fix memory leak in cache"

# Ver estado de una tarea
go run cmd/cli/main.go status --task task-1
```

### 3. Ejemplo Simple

```bash
go run examples/simple/main.go
```

## Configuración

### Políticas

Edita `policies.example.yaml` y renómbralo a `policies.yaml`:

```yaml
policies:
  - id: coder-policy
    enabled: true
    metadata:
      allowed_paths:
        - "src/**"
        - "cmd/**"
      
gates:
  - id: fmt-lint
    required: true
```

### Variables de Entorno

```bash
# Nivel de log
export LOG_LEVEL=debug

# Tiempo máximo de ejecución (segundos)
export MAX_EXECUTION_TIME=300

# Número máximo de retries
export MAX_RETRIES=3
```

## Casos de Uso

### Caso 1: Arreglar un Bug

```bash
./bin/orchestrator \
  --task "fix nil pointer in /api/users endpoint" \
  --repo /path/to/repo
```

El sistema:
1. Planificará la tarea
2. Identificará el archivo afectado
3. Aplicará el fix
4. Ejecutará tests
5. Reparará si algo falla
6. Auditará el código
7. Optimizará si es posible

### Caso 2: Optimizar Performance

```bash
./bin/orchestrator \
  --task "optimize slow endpoint /api/search" \
  --repo /path/to/repo
```

El sistema:
1. Ejecutará benchmarks
2. Identificará cuellos de botella
3. Aplicará optimizaciones
4. Validará con tests
5. Comparará benchmarks antes/después

### Caso 3: Arreglar Tests que Fallan

```bash
./bin/orchestrator \
  --task "fix failing tests in user package" \
  --repo /path/to/repo
```

El Repairer:
1. Analizará los fallos
2. Identificará el tipo de error
3. Aplicará fixes específicos
4. Re-ejecutará tests

## Integración con CI/CD

### GitHub Actions

```yaml
name: Multi-Agent Auto-Fix

on:
  workflow_dispatch:
    inputs:
      objective:
        description: 'Objective'
        required: true

jobs:
  autofix:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      
      - name: Run Multi-Agent
        run: |
          go run cmd/orchestrator/main.go \
            --task "${{ github.event.inputs.objective }}" \
            --repo .
      
      - name: Create PR
        uses: peter-evans/create-pull-request@v5
```

### GitLab CI

```yaml
autofix:
  script:
    - go run cmd/orchestrator/main.go --task "$OBJECTIVE" --repo .
  only:
    - schedules
```

## Monitoreo

### Ver Estado de Tareas

```bash
# Usar CLI
./bin/cli status --task task-123

# O consultar logs
tail -f artifacts/orchestrator.log
```

### Ver Memoria de Decisiones

Las decisiones se almacenan en el Orchestrator. Puedes acceder programáticamente:

```go
orch := orchestrator.New(ws, policy)
decisions := orch.GetMemory()
```

## Troubleshooting

### Tarea Bloqueada por Política

```
Error: task blocked by policy
```

**Solución**: Revisa las políticas en `policies.yaml` y ajusta las restricciones.

### Tests Siempre Fallan

**Solución**: Revisa los logs del Tester para ver qué tests fallan. El Repairer debería intentar arreglarlos automáticamente.

### Timeout

```
Error: execution timeout
```

**Solución**: Aumenta `MAX_EXECUTION_TIME` o simplifica la tarea.

## Mejores Prácticas

1. **Empieza Simple**: Usa objetivos específicos y pequeños
2. **Revisa Cambios**: Siempre revisa los cambios generados antes de mergear
3. **Configura Políticas**: Ajusta las políticas según tu proyecto
4. **Monitorea Memoria**: Revisa las decisiones para mejorar el sistema
5. **Itera**: El sistema mejora con más uso
