package evaluation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nanochip/multi-agent/pkg/types"
)

// Engine parsea resultados y clasifica fallas
type Engine struct {
	patterns map[string]*FailurePattern
}

// FailurePattern representa un patrón de fallo conocido
type FailurePattern struct {
	ID          string
	Name        string
	Category    string // "compilation", "test", "runtime", "lint", "security"
	Severity    types.Severity
	Regex       *regexp.Regexp
	Description string
	Remediation string
}

// FailureClassification representa la clasificación de un fallo
type FailureClassification struct {
	Pattern     *FailurePattern
	Matches     []string
	Confidence  float64
	Suggestions []string
}

// NewEngine crea un nuevo evaluation engine
func NewEngine() *Engine {
	engine := &Engine{
		patterns: make(map[string]*FailurePattern),
	}
	
	// Cargar patrones conocidos
	engine.loadDefaultPatterns()
	
	return engine
}

// loadDefaultPatterns carga patrones de fallo comunes
func (e *Engine) loadDefaultPatterns() {
	patterns := []*FailurePattern{
		{
			ID:          "nil-pointer",
			Name:        "Nil Pointer Dereference",
			Category:    "runtime",
			Severity:    types.SeverityCritical,
			Regex:       regexp.MustCompile(`panic: runtime error: invalid memory address or nil pointer dereference`),
			Description: "Attempt to dereference a nil pointer",
			Remediation: "Add nil check before dereferencing",
		},
		{
			ID:          "undefined-variable",
			Name:        "Undefined Variable",
			Category:    "compilation",
			Severity:    types.SeverityHigh,
			Regex:       regexp.MustCompile(`undefined: (\w+)`),
			Description: "Variable or function is not defined",
			Remediation: "Define the variable or import the package",
		},
		{
			ID:          "type-mismatch",
			Name:        "Type Mismatch",
			Category:    "compilation",
			Severity:    types.SeverityHigh,
			Regex:       regexp.MustCompile(`cannot use .* \(type .*\) as type .*`),
			Description: "Type mismatch in assignment or function call",
			Remediation: "Fix type conversion or use correct type",
		},
		{
			ID:          "import-error",
			Name:        "Import Error",
			Category:    "compilation",
			Severity:    types.SeverityMedium,
			Regex:       regexp.MustCompile(`cannot find package|package .* is not in GOROOT`),
			Description: "Package import error",
			Remediation: "Run go mod tidy or install missing package",
		},
		{
			ID:          "test-failure",
			Name:        "Test Failure",
			Category:    "test",
			Severity:    types.SeverityHigh,
			Regex:       regexp.MustCompile(`FAIL:\s+(\S+)`),
			Description: "Unit test failed",
			Remediation: "Fix the test or the code being tested",
		},
		{
			ID:          "lint-error",
			Name:        "Lint Error",
			Category:    "lint",
			Severity:    types.SeverityMedium,
			Regex:       regexp.MustCompile(`(golangci-lint|go vet|staticcheck).*error`),
			Description: "Code style or lint error",
			Remediation: "Fix linting issues",
		},
		{
			ID:          "race-condition",
			Name:        "Race Condition",
			Category:    "runtime",
			Severity:    types.SeverityCritical,
			Regex:       regexp.MustCompile(`WARNING: DATA RACE|race detected`),
			Description: "Data race detected",
			Remediation: "Add synchronization (mutex, channel, etc.)",
		},
		{
			ID:          "out-of-bounds",
			Name:        "Index Out of Bounds",
			Category:    "runtime",
			Severity:    types.SeverityCritical,
			Regex:       regexp.MustCompile(`index out of range|runtime error: index out of range`),
			Description: "Array or slice index out of bounds",
			Remediation: "Add bounds checking before indexing",
		},
		{
			ID:          "timeout",
			Name:        "Timeout",
			Category:    "test",
			Severity:    types.SeverityMedium,
			Regex:       regexp.MustCompile(`timeout|context deadline exceeded`),
			Description: "Operation timed out",
			Remediation: "Increase timeout or optimize slow operation",
		},
		{
			ID:          "deadlock",
			Name:        "Deadlock",
			Category:    "runtime",
			Severity:    types.SeverityCritical,
			Regex:       regexp.MustCompile(`fatal error: all goroutines are asleep - deadlock`),
			Description: "Deadlock detected",
			Remediation: "Review synchronization logic",
		},
	}
	
	for _, pattern := range patterns {
		e.patterns[pattern.ID] = pattern
	}
}

// ParseResult parsea un resultado de tarea y clasifica fallas
func (e *Engine) ParseResult(result *types.TaskResult) []*FailureClassification {
	classifications := make([]*FailureClassification, 0)
	
	if result.Success {
		return classifications
	}
	
	// Buscar en el error
	if result.Error != "" {
		classifications = append(classifications, e.classifyFailure(result.Error)...)
	}
	
	// Buscar en evidence
	for _, evidence := range result.Evidence {
		if evidence.Type == "log" || evidence.Type == "report" {
			content := string(evidence.Content)
			classifications = append(classifications, e.classifyFailure(content)...)
		}
	}
	
	// Buscar en outputs
	if testResult, ok := result.Outputs["test_result"].(*types.TestResult); ok {
		for _, failure := range testResult.Failures {
			classifications = append(classifications, e.classifyFailure(failure.Message)...)
			classifications = append(classifications, e.classifyFailure(failure.Output)...)
		}
	}
	
	return classifications
}

// classifyFailure clasifica un fallo específico
func (e *Engine) classifyFailure(text string) []*FailureClassification {
	classifications := make([]*FailureClassification, 0)
	
	text = strings.ToLower(text)
	
	for _, pattern := range e.patterns {
		matches := pattern.Regex.FindAllString(text, -1)
		if len(matches) > 0 {
			confidence := e.calculateConfidence(pattern, matches, text)
			classification := &FailureClassification{
				Pattern:    pattern,
				Matches:    matches,
				Confidence: confidence,
				Suggestions: []string{pattern.Remediation},
			}
			classifications = append(classifications, classification)
		}
	}
	
	return classifications
}

// calculateConfidence calcula la confianza de una clasificación
func (e *Engine) calculateConfidence(pattern *FailurePattern, matches []string, text string) float64 {
	// Confianza base basada en severidad
	confidence := 0.5
	
	if pattern.Severity == types.SeverityCritical {
		confidence = 0.9
	} else if pattern.Severity == types.SeverityHigh {
		confidence = 0.8
	} else if pattern.Severity == types.SeverityMedium {
		confidence = 0.7
	}
	
	// Aumentar confianza si hay múltiples matches
	if len(matches) > 1 {
		confidence += 0.1
		if confidence > 1.0 {
			confidence = 1.0
		}
	}
	
	return confidence
}

// GetSuggestions retorna sugerencias de reparación basadas en clasificaciones
func (e *Engine) GetSuggestions(classifications []*FailureClassification) []string {
	suggestions := make([]string, 0)
	seen := make(map[string]bool)
	
	for _, classification := range classifications {
		if classification.Confidence > 0.7 {
			for _, suggestion := range classification.Suggestions {
				if !seen[suggestion] {
					suggestions = append(suggestions, suggestion)
					seen[suggestion] = true
				}
			}
		}
	}
	
	return suggestions
}

// AddPattern añade un nuevo patrón de fallo
func (e *Engine) AddPattern(pattern *FailurePattern) {
	e.patterns[pattern.ID] = pattern
}

// GetPattern retorna un patrón por ID
func (e *Engine) GetPattern(id string) *FailurePattern {
	return e.patterns[id]
}
