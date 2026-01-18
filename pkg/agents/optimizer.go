package agents

import (
	"context"
	"strings"
	"time"

	"github.com/nanochip/multi-agent/pkg/policies"
	"github.com/nanochip/multi-agent/pkg/types"
	"github.com/nanochip/multi-agent/pkg/workspace"
)

// Optimizer optimiza código con validación
type Optimizer struct {
	*BaseAgent
}

// NewOptimizer crea un nuevo agente optimizador
func NewOptimizer(ws *workspace.Manager, policy *policies.Engine) *Optimizer {
	contract := types.AgentContract{
		ID:            "optimizer",
		Name:          "Optimizer",
		AllowedPaths:  []string{"src/**", "cmd/**", "internal/**", "pkg/**"},
		AllowedTools:  []string{"go", "go test", "go tool pprof"},
		RequiredTests: true,
	}

	return &Optimizer{
		BaseAgent: NewBaseAgent(ws, policy, contract),
	}
}

// Execute ejecuta la tarea de optimización
func (o *Optimizer) Execute(ctx context.Context, task *types.Task) *types.TaskResult {
	decision := types.Decision{
		Agent:      "optimizer",
		Reason:     "analyzing and optimizing code",
		Action:     "optimize",
		Timestamp:  time.Now(),
		Confidence: 0.7,
	}

	// Ejecutar benchmark si está disponible
	benchmarkResult := o.runBenchmarks()

	// Identificar optimizaciones potenciales
	optimizations := o.identifyOptimizations(task.Objective)

	// Aplicar optimizaciones (solo si son seguras)
	appliedOpts := make([]string, 0)
	for _, opt := range optimizations {
		if o.isSafeOptimization(opt) {
			if o.applyOptimization(opt) {
				appliedOpts = append(appliedOpts, opt)
			}
		}
	}

	// Validar que los tests aún pasan después de optimizar
	testOutput, _ := o.workspace.RunCommand("go", "test", "./...")
	testsStillPass := !strings.Contains(testOutput, "FAIL")

	// Comparar benchmark antes/después
	benchmarkAfter := o.runBenchmarks()
	improvement := o.compareBenchmarks(benchmarkResult, benchmarkAfter)

	outputs := map[string]interface{}{
		"optimizations":    appliedOpts,
		"benchmark_before": benchmarkResult,
		"benchmark_after":  benchmarkAfter,
		"improvement":      improvement,
		"tests_still_pass": testsStillPass,
	}

	// Solo considerar éxito si los tests aún pasan
	success := testsStillPass && len(appliedOpts) > 0

	return &types.TaskResult{
		TaskID:    task.ID,
		State:     mapState(success),
		Success:   success,
		Outputs:   outputs,
		Decisions: []types.Decision{decision},
	}
}

// runBenchmarks ejecuta benchmarks
func (o *Optimizer) runBenchmarks() map[string]interface{} {
	// Ejecutar go test -bench
	benchOutput, _ := o.workspace.RunCommand("go", "test", "-bench=.", "-benchmem", "./...")

	result := map[string]interface{}{
		"output": benchOutput,
	}

	return result
}

// identifyOptimizations identifica optimizaciones potenciales
func (o *Optimizer) identifyOptimizations(objective string) []string {
	optimizations := make([]string, 0)

	// Analizar objetivo para tipos de optimización
	if strings.Contains(objective, "slow") || strings.Contains(objective, "performance") {
		optimizations = append(optimizations, "optimize_loops")
		optimizations = append(optimizations, "reduce_allocations")
		optimizations = append(optimizations, "cache_results")
	}

	// Siempre sugerir algunas optimizaciones comunes
	optimizations = append(optimizations, "remove_unused_imports")
	optimizations = append(optimizations, "simplify_expressions")

	return optimizations
}

// isSafeOptimization verifica si una optimización es segura
func (o *Optimizer) isSafeOptimization(opt string) bool {
	// Optimizaciones que no cambian comportamiento
	safeOpts := []string{
		"remove_unused_imports",
		"simplify_expressions",
		"reduce_allocations",
	}

	for _, safe := range safeOpts {
		if opt == safe {
			return true
		}
	}

	// Optimizaciones que requieren tests
	testRequiredOpts := []string{
		"optimize_loops",
		"cache_results",
	}

	for _, testOpt := range testRequiredOpts {
		if opt == testOpt {
			// Verificar que hay tests disponibles
			testOutput, _ := o.workspace.RunCommand("go", "test", "-list", ".", "./...")
			return strings.Contains(testOutput, "Test")
		}
	}

	return false
}

// applyOptimization aplica una optimización específica
func (o *Optimizer) applyOptimization(opt string) bool {
	switch opt {
	case "remove_unused_imports":
		// goimports lo hace automáticamente
		o.workspace.RunCommand("goimports", "-w", ".")
		return true
	case "simplify_expressions":
		// gofmt simplifica algunas expresiones
		o.workspace.RunCommand("go", "fmt", "./...")
		return true
	default:
		// Optimizaciones más complejas requerirían análisis de AST
		return false
	}
}

// compareBenchmarks compara resultados de benchmarks
func (o *Optimizer) compareBenchmarks(before, after map[string]interface{}) string {
	// Por ahora simplificado
	return "benchmarks_compared"
}
