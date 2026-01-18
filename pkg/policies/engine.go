package policies

import (
	"github.com/nanochip/multi-agent/pkg/types"
)

// Engine gestiona políticas y guardrails
type Engine struct {
	policies []types.Policy
	gates    []Gate
}

// Gate representa un gate obligatorio
type Gate struct {
	ID          string
	Name        string
	Description string
	Validator   func(*types.TaskResult) bool
	Required    bool
}

// NewEngine crea un nuevo motor de políticas
func NewEngine() *Engine {
	engine := &Engine{
		policies: make([]types.Policy, 0),
		gates:    make([]Gate, 0),
	}

	// Configurar gates por defecto
	engine.setupDefaultGates()

	return engine
}

// setupDefaultGates configura los gates obligatorios
func (e *Engine) setupDefaultGates() {
	e.gates = []Gate{
		{
			ID:          "fmt-lint",
			Name:        "Format and Lint",
			Description: "Code must pass fmt and lint checks",
			Required:    true,
			Validator: func(result *types.TaskResult) bool {
				if findings, ok := result.Outputs["lint_errors"].([]types.AuditFinding); ok {
					for _, finding := range findings {
						if finding.Severity == types.SeverityCritical || finding.Severity == types.SeverityHigh {
							return false
						}
					}
				}
				return true
			},
		},
		{
			ID:          "tests-pass",
			Name:        "Tests Must Pass",
			Description: "All tests must pass",
			Required:    true,
			Validator: func(result *types.TaskResult) bool {
				if testResult, ok := result.Outputs["test_result"].(*types.TestResult); ok {
					return testResult.Failed == 0
				}
				// Si no hay test result, permitir (puede que no haya tests)
				return true
			},
		},
		{
			ID:          "coverage",
			Name:        "Minimum Coverage",
			Description: "Code coverage must meet minimum threshold",
			Required:    true,
			Validator: func(result *types.TaskResult) bool {
				const minCoverage = 70.0
				if testResult, ok := result.Outputs["test_result"].(*types.TestResult); ok {
					return testResult.Coverage >= minCoverage
				}
				return true
			},
		},
		{
			ID:          "secrets",
			Name:        "No Secrets",
			Description: "No secrets should be exposed",
			Required:    true,
			Validator: func(result *types.TaskResult) bool {
				if findings, ok := result.Outputs["secret_findings"].([]types.AuditFinding); ok {
					for _, finding := range findings {
						if finding.Category == "secret" &&
							(finding.Severity == types.SeverityCritical || finding.Severity == types.SeverityHigh) {
							return false
						}
					}
				}
				return true
			},
		},
		{
			ID:          "dependencies",
			Name:        "Dependency Security",
			Description: "No critical CVEs in dependencies",
			Required:    true,
			Validator: func(result *types.TaskResult) bool {
				if findings, ok := result.Outputs["dependency_findings"].([]types.AuditFinding); ok {
					for _, finding := range findings {
						if finding.Category == "dependency" && finding.Severity == types.SeverityCritical {
							return false
						}
					}
				}
				return true
			},
		},
		{
			ID:          "risk-review",
			Name:        "Risk Review",
			Description: "High-risk changes require review",
			Required:    false, // Warning, no bloquea
			Validator: func(result *types.TaskResult) bool {
				if riskLevel, ok := result.Outputs["risk_level"].(string); ok {
					if riskLevel == "high" {
						// Log warning pero no bloquea
						return true
					}
				}
				return true
			},
		},
	}
}

// AllowTask verifica si una tarea está permitida según las políticas
func (e *Engine) AllowTask(task *types.Task) bool {
	for _, policy := range e.policies {
		if !policy.Enabled {
			continue
		}

		// Verificar restricciones de rutas si están en los inputs
		if files, ok := task.Inputs["files"].([]interface{}); ok {
			if forbiddenPaths, ok := policy.Metadata["forbidden_paths"].([]interface{}); ok {
				// Validar que la tarea no toque rutas prohibidas
				for _, file := range files {
					if fileStr, ok := file.(string); ok {
						for _, forbiddenPath := range forbiddenPaths {
							if forbiddenPathStr, ok := forbiddenPath.(string); ok {
								if matched, _ := pathMatches(fileStr, forbiddenPathStr); matched {
									return false
								}
							}
						}
					}
				}
			}
		}
	}
	return true
}

// ValidateResult valida un resultado contra todos los gates obligatorios
func (e *Engine) ValidateResult(result *types.TaskResult) bool {
	for _, gate := range e.gates {
		if gate.Required && !gate.Validator(result) {
			return false
		}
	}
	return true
}

// AddPolicy añade una nueva política
func (e *Engine) AddPolicy(policy types.Policy) {
	e.policies = append(e.policies, policy)
}

// GetGates retorna los gates configurados
func (e *Engine) GetGates() []Gate {
	return e.gates
}

// ValidatePath verifica si una ruta está permitida para un agente
func (e *Engine) ValidatePath(agentID string, path string, policy types.Policy) bool {
	// Verificar rutas permitidas
	if allowedPaths, ok := policy.Metadata["allowed_paths"].([]interface{}); ok {
		allowed := false
		for _, allowedPath := range allowedPaths {
			if allowedPathStr, ok := allowedPath.(string); ok {
				if matched, _ := pathMatches(path, allowedPathStr); matched {
					allowed = true
					break
				}
			}
		}
		if !allowed {
			return false
		}
	}

	// Verificar rutas prohibidas
	if forbiddenPaths, ok := policy.Metadata["forbidden_paths"].([]interface{}); ok {
		for _, forbiddenPath := range forbiddenPaths {
			if forbiddenPathStr, ok := forbiddenPath.(string); ok {
				if matched, _ := pathMatches(path, forbiddenPathStr); matched {
					return false
				}
			}
		}
	}

	return true
}

// pathMatches verifica si un path coincide con un patrón (simple glob)
func pathMatches(path, pattern string) (bool, error) {
	// Implementación simple de glob matching
	// Por ahora, soporte básico de "*"
	if pattern == "*" {
		return true, nil
	}
	if pattern == path {
		return true, nil
	}
	// Implementación más completa iría aquí
	return false, nil
}

// GetPolicyForAgent retorna la política para un agente específico
func (e *Engine) GetPolicyForAgent(agentID string) *types.Policy {
	// Buscar política que coincida con el agente
	for _, policy := range e.policies {
		if agentName, ok := policy.Metadata["agent_id"].(string); ok && agentName == agentID {
			return &policy
		}
	}
	return nil
}
