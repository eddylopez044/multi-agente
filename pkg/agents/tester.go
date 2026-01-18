package agents

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nanochip/multi-agent/pkg/policies"
	"github.com/nanochip/multi-agent/pkg/types"
	"github.com/nanochip/multi-agent/pkg/workspace"
)

// Tester ejecuta tests y genera reportes
type Tester struct {
	*BaseAgent
}

// NewTester crea un nuevo agente tester
func NewTester(ws *workspace.Manager, policy *policies.Engine) *Tester {
	contract := types.AgentContract{
		ID:             "tester",
		Name:           "Tester",
		AllowedPaths:   []string{"**/*_test.go"},
		ForbiddenPaths: []string{"**/*.go"}, // No puede modificar código de producción
		AllowedTools:   []string{"go", "go test"},
		RequiredTests:  false,
	}

	return &Tester{
		BaseAgent: NewBaseAgent(ws, policy, contract),
	}
}

// Execute ejecuta la tarea de testing
func (t *Tester) Execute(ctx context.Context, task *types.Task) *types.TaskResult {
	decision := types.Decision{
		Agent:      "tester",
		Reason:     "executing test suite",
		Action:     "test",
		Timestamp:  time.Now(),
		Confidence: 0.95,
	}

	startTime := time.Now()

	// Ejecutar tests
	testOutput, err := t.workspace.RunCommand("go", "test", "-v", "-cover", "./...")

	duration := time.Since(startTime)

	testResult := t.parseTestOutput(testOutput, err, duration)

	// Ejecutar coverage detallado
	coverageOutput, _ := t.workspace.RunCommand("go", "test", "-coverprofile=coverage.out", "./...")
	if coverageOutput != "" {
		coverage, _ := t.parseCoverage(coverageOutput)
		testResult.Coverage = coverage
	}

	evidence := []types.Evidence{
		{
			Type:        "report",
			Source:      "go test",
			Content:     []byte(testOutput),
			Timestamp:   time.Now(),
			Description: "Test execution output",
		},
	}

	outputs := map[string]interface{}{
		"test_result": testResult,
		"command":     "go test -v -cover ./...",
	}

	success := testResult.Failed == 0 && err == nil

	return &types.TaskResult{
		TaskID:    task.ID,
		State:     mapState(success),
		Success:   success,
		Outputs:   outputs,
		Evidence:  evidence,
		Decisions: []types.Decision{decision},
		Duration:  duration,
	}
}

// parseTestOutput parsea la salida de go test
func (t *Tester) parseTestOutput(output string, err error, duration time.Duration) *types.TestResult {
	result := &types.TestResult{
		Passed:   0,
		Failed:   0,
		Skipped:  0,
		Duration: duration,
		Failures: make([]types.TestFailure, 0),
	}

	if err != nil {
		// Extraer información de fallos de la salida
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "FAIL:") {
				result.Failed++
				// Intentar extraer nombre del test
				parts := strings.Fields(line)
				if len(parts) > 1 {
					testName := parts[1]
					result.Failures = append(result.Failures, types.TestFailure{
						Test:    testName,
						Message: line,
					})
				}
			} else if strings.Contains(line, "PASS:") || strings.Contains(line, "ok") {
				result.Passed++
			} else if strings.Contains(line, "SKIP:") {
				result.Skipped++
			}
		}
	} else {
		// Si no hay error, contar PASS
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "PASS:") || (strings.Contains(line, "ok") && strings.Contains(line, "coverage")) {
				result.Passed++
			}
		}
	}

	return result
}

// parseCoverage parsea el porcentaje de cobertura
func (t *Tester) parseCoverage(output string) (float64, error) {
	// Buscar patrón de cobertura: "coverage: XX.X%"
	re := regexp.MustCompile(`coverage:\s*(\d+\.?\d*)%`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		coverage, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			return 0, err
		}
		return coverage, nil
	}
	return 0, fmt.Errorf("coverage not found in output")
}
