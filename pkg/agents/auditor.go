package agents

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nanochip/multi-agent/pkg/policies"
	"github.com/nanochip/multi-agent/pkg/types"
	"github.com/nanochip/multi-agent/pkg/workspace"
)

// Auditor revisa seguridad, estilo, dependencias, licencias, secretos
type Auditor struct {
	*BaseAgent
}

// NewAuditor crea un nuevo agente auditor
func NewAuditor(ws *workspace.Manager, policy *policies.Engine) *Auditor {
	contract := types.AgentContract{
		ID:            "auditor",
		Name:          "Auditor",
		AllowedPaths:  []string{}, // No modifica código, solo lee
		AllowedTools:  []string{"go", "go vet", "golangci-lint"},
		RequiredTests: false,
	}

	return &Auditor{
		BaseAgent: NewBaseAgent(ws, policy, contract),
	}
}

// Execute ejecuta la tarea de auditoría
func (a *Auditor) Execute(ctx context.Context, task *types.Task) *types.TaskResult {
	decision := types.Decision{
		Agent:      "auditor",
		Reason:     "auditing code changes",
		Action:     "audit",
		Timestamp:  time.Now(),
		Confidence: 0.9,
	}

	findings := make([]types.AuditFinding, 0)

	// 1. Lint check
	lintFindings := a.checkLint()
	findings = append(findings, lintFindings...)

	// 2. Security check (búsqueda de patrones peligrosos)
	securityFindings := a.checkSecurity()
	findings = append(findings, securityFindings...)

	// 3. Secrets scan (simulado)
	secretFindings := a.checkSecrets()
	findings = append(findings, secretFindings...)

	// 4. Dependency check (simulado)
	dependencyFindings := a.checkDependencies()
	findings = append(findings, dependencyFindings...)

	// Clasificar hallazgos
	criticalFindings := make([]types.AuditFinding, 0)
	lintErrors := make([]types.AuditFinding, 0)
	secretFindingsList := make([]types.AuditFinding, 0)
	dependencyFindingsList := make([]types.AuditFinding, 0)

	for _, finding := range findings {
		switch finding.Category {
		case "security", "secret":
			if finding.Severity == types.SeverityCritical || finding.Severity == types.SeverityHigh {
				criticalFindings = append(criticalFindings, finding)
			}
			if finding.Category == "secret" {
				secretFindingsList = append(secretFindingsList, finding)
			}
		case "style", "lint":
			lintErrors = append(lintErrors, finding)
		case "dependency":
			dependencyFindingsList = append(dependencyFindingsList, finding)
		}
	}

	outputs := map[string]interface{}{
		"findings":            findings,
		"critical_findings":   criticalFindings,
		"lint_errors":         lintErrors,
		"secret_findings":     secretFindingsList,
		"dependency_findings": dependencyFindingsList,
	}

	success := len(criticalFindings) == 0

	return &types.TaskResult{
		TaskID:    task.ID,
		State:     mapState(success),
		Success:   success,
		Outputs:   outputs,
		Decisions: []types.Decision{decision},
	}
}

// checkLint ejecuta verificaciones de lint
func (a *Auditor) checkLint() []types.AuditFinding {
	findings := make([]types.AuditFinding, 0)

	// Ejecutar go vet
	vetOutput, err := a.workspace.RunCommand("go", "vet", "./...")
	if err != nil && vetOutput != "" {
		// Parsear salida de go vet
		lines := strings.Split(vetOutput, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" && !strings.Contains(line, "no packages") {
				findings = append(findings, types.AuditFinding{
					ID:       fmt.Sprintf("lint-%d", len(findings)),
					Severity: types.SeverityHigh,
					Category: "lint",
					Rule:     "go vet",
					Message:  line,
				})
			}
		}
	}

	// Intentar golangci-lint si está disponible
	lintOutput, _ := a.workspace.RunCommand("golangci-lint", "run")
	if lintOutput != "" {
		lines := strings.Split(lintOutput, "\n")
		for _, line := range lines {
			if strings.Contains(line, ":") && (strings.Contains(line, "warning") || strings.Contains(line, "error")) {
				findings = append(findings, types.AuditFinding{
					ID:       fmt.Sprintf("lint-%d", len(findings)),
					Severity: types.SeverityMedium,
					Category: "lint",
					Rule:     "golangci-lint",
					Message:  line,
				})
			}
		}
	}

	return findings
}

// checkSecurity busca patrones peligrosos
func (a *Auditor) checkSecurity() []types.AuditFinding {
	findings := make([]types.AuditFinding, 0)

	// Buscar uso de eval, exec peligrosos, etc.
	// Por ahora simulado, pero en producción usaría SAST tools

	dangerousPatterns := []struct {
		pattern  string
		message  string
		severity types.Severity
	}{
		{"exec.Command", "Direct command execution detected", types.SeverityHigh},
		{"eval(", "Use of eval detected", types.SeverityCritical},
		{"os.Getenv", "Environment variable access", types.SeverityLow},
	}

	// En producción, escanear archivos reales
	for _, pattern := range dangerousPatterns {
		// Simulado por ahora
		if strings.Contains(pattern.pattern, "eval") {
			findings = append(findings, types.AuditFinding{
				ID:          fmt.Sprintf("security-%d", len(findings)),
				Severity:    pattern.severity,
				Category:    "security",
				Rule:        "dangerous_pattern",
				Message:     pattern.message,
				Remediation: "Review usage of dangerous functions",
			})
		}
	}

	return findings
}

// checkSecrets busca secretos expuestos
func (a *Auditor) checkSecrets() []types.AuditFinding {
	findings := make([]types.AuditFinding, 0)

	// En producción, usar herramientas como gitleaks, trufflehog
	// Por ahora simulado
	// Patrones comunes de secretos que se buscarían:
	// - api[_-]?key
	// - password\\s*=
	// - secret[_-]?key
	// - bearer[_-]?token

	return findings
}

// checkDependencies verifica dependencias vulnerables
func (a *Auditor) checkDependencies() []types.AuditFinding {
	findings := make([]types.AuditFinding, 0)

	// En producción, usar nancy, snyk, o go list -json -m all | nancy sleuth
	// Por ahora retornar lista vacía
	// Se ejecutaría: go list -m all para obtener dependencias

	return findings
}
