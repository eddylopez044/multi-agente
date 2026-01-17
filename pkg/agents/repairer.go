package agents

import (
	"context"
	"fmt"
	"time"

	"github.com/nanochip/multi-agent/pkg/policies"
	"github.com/nanochip/multi-agent/pkg/types"
	"github.com/nanochip/multi-agent/pkg/workspace"
)

// Repairer auto-repara fallos del pipeline
type Repairer struct {
	*BaseAgent
}

// NewRepairer crea un nuevo agente reparador
func NewRepairer(ws workspace.Manager, policy policies.Engine) *Repairer {
	contract := types.AgentContract{
		ID:           "repairer",
		Name:         "Repairer",
		AllowedPaths: []string{"src/**", "cmd/**", "internal/**", "pkg/**"},
		AllowedTools: []string{"go", "go fmt", "go fix"},
		RequiredTests: true,
	}
	
	return &Repairer{
		BaseAgent: NewBaseAgent(ws, policy, contract),
	}
}

// Execute ejecuta la tarea de reparación
func (r *Repairer) Execute(ctx context.Context, task *types.Task) *types.TaskResult {
	decision := types.Decision{
		Agent:      "repairer",
		Reason:     "analyzing and repairing failures",
		Action:     "repair",
		Timestamp:  time.Now(),
		Confidence: 0.75,
	}
	
	// Analizar el tipo de fallo
	var repairStrategy string
	var fixes []string
	
	// Si viene de test failure
	if testResult, ok := task.Inputs["test_result"].(*types.TaskResult); ok {
		if tr, ok := testResult.Outputs["test_result"].(*types.TestResult); ok {
			strategy, fixesList := r.analyzeTestFailures(tr)
			repairStrategy = strategy
			fixes = fixesList
		}
	}
	
	// Si viene de audit failure
	if auditResult, ok := task.Inputs["audit_result"].(*types.TaskResult); ok {
		if findings, ok := auditResult.Outputs["critical_findings"].([]types.AuditFinding); ok {
			strategy, fixesList := r.analyzeAuditFailures(findings)
			repairStrategy = strategy
			fixes = append(fixes, fixesList...)
		}
	}
	
	// Aplicar fixes
	appliedFixes := make([]string, 0)
	for _, fix := range fixes {
		if r.applyFix(fix) {
			appliedFixes = append(appliedFixes, fix)
		}
	}
	
	// Ejecutar go fmt automáticamente
	r.workspace.RunCommand("go", "fmt", "./...")
	
	// Ejecutar go fix para correcciones automáticas
	r.workspace.RunCommand("go", "fix", "./...")
	
	outputs := map[string]interface{}{
		"strategy":      repairStrategy,
		"applied_fixes": appliedFixes,
	}
	
	success := len(appliedFixes) > 0 || repairStrategy != ""
	
	return &types.TaskResult{
		TaskID:    task.ID,
		State:     mapState(success),
		Success:   success,
		Outputs:   outputs,
		Decisions: []types.Decision{decision},
	}
}

// analyzeTestFailures analiza fallos de tests y propone fixes
func (r *Repairer) analyzeTestFailures(testResult *types.TestResult) (string, []string) {
	strategies := make([]string, 0)
	fixes := make([]string, 0)
	
	if testResult.Failed > 0 {
		for _, failure := range testResult.Failures {
			// Análisis básico de tipos de fallos comunes
			if contains(failure.Message, "nil pointer") {
				strategies = append(strategies, "nil_pointer_check")
				fixes = append(fixes, "add nil pointer checks")
			} else if contains(failure.Message, "undefined") {
				strategies = append(strategies, "missing_definition")
				fixes = append(fixes, "add missing definitions")
			} else if contains(failure.Message, "cannot use") {
				strategies = append(strategies, "type_mismatch")
				fixes = append(fixes, "fix type mismatches")
			} else {
				strategies = append(strategies, "generic_fix")
				fixes = append(fixes, fmt.Sprintf("fix test in %s", failure.Package))
			}
		}
	}
	
	// Si la cobertura es baja, sugerir agregar tests
	if testResult.Coverage < 70.0 {
		fixes = append(fixes, "increase test coverage")
	}
	
	return fmt.Sprintf("repair_%d_failures", len(strategies)), fixes
}

// analyzeAuditFailures analiza hallazgos de auditoría y propone fixes
func (r *Repairer) analyzeAuditFailures(findings []types.AuditFinding) (string, []string) {
	fixes := make([]string, 0)
	
	for _, finding := range findings {
		switch finding.Category {
		case "lint":
			fixes = append(fixes, fmt.Sprintf("fix lint in %s: %s", finding.File, finding.Message))
		case "security":
			if finding.Remediation != "" {
				fixes = append(fixes, finding.Remediation)
			} else {
				fixes = append(fixes, fmt.Sprintf("fix security issue: %s", finding.Message))
			}
		case "secret":
			fixes = append(fixes, "remove exposed secrets, use environment variables")
		}
	}
	
	return "repair_audit_findings", fixes
}

// applyFix aplica un fix específico (simulado)
func (r *Repairer) applyFix(fix string) bool {
	// En producción, aquí se aplicarían fixes reales
	// Por ahora solo simular
	return true
}

// contains verifica si un string contiene un substring
func contains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// mapState convierte un bool a TaskState
func mapState(success bool) types.TaskState {
	if success {
		return types.StateSuccess
	}
	return types.StateFailed
}
