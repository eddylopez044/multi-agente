package agents

import (
	"context"
	"fmt"
	"time"

	"github.com/nanochip/multi-agent/pkg/policies"
	"github.com/nanochip/multi-agent/pkg/types"
	"github.com/nanochip/multi-agent/pkg/workspace"
)

// Planner descompone objetivos en subtareas ejecutables
type Planner struct {
	*BaseAgent
}

// NewPlanner crea un nuevo agente planificador
func NewPlanner(ws *workspace.Manager, policy *policies.Engine) *Planner {
	contract := types.AgentContract{
		ID:            "planner",
		Name:          "Planner",
		AllowedTools:  []string{},
		RequiredTests: false,
	}

	return &Planner{
		BaseAgent: NewBaseAgent(ws, policy, contract),
	}
}

// Execute ejecuta la tarea de planificación
func (p *Planner) Execute(ctx context.Context, task *types.Task) *types.TaskResult {
	decision := types.Decision{
		Agent:      "planner",
		Reason:     "decomposing objective into subtasks",
		Action:     "plan",
		Timestamp:  time.Now(),
		Confidence: 0.9,
	}

	objective := task.Objective

	// Análisis básico del objetivo para crear subtareas
	subtasks := p.createSubtasks(objective, task.Inputs)

	outputs := map[string]interface{}{
		"subtasks":   subtasks,
		"plan_count": len(subtasks),
	}

	return &types.TaskResult{
		TaskID:    task.ID,
		State:     types.StateSuccess,
		Success:   true,
		Outputs:   outputs,
		Decisions: []types.Decision{decision},
	}
}

// createSubtasks crea subtareas basadas en el objetivo
func (p *Planner) createSubtasks(objective string, inputs map[string]interface{}) []*types.Task {
	subtasks := make([]*types.Task, 0)

	// Análisis simple: si el objetivo contiene "fix", "bug", "repair"
	// crear tarea de código seguida de tests
	if containsKeywords(objective, []string{"fix", "bug", "repair", "arreglar"}) {
		subtasks = append(subtasks, &types.Task{
			Type:       types.TaskCode,
			Objective:  fmt.Sprintf("implement fix for: %s", objective),
			Inputs:     inputs,
			MaxRetries: 3,
		})
	}

	// Si contiene "optimize", "performance", "slow"
	if containsKeywords(objective, []string{"optimize", "performance", "slow", "lento"}) {
		subtasks = append(subtasks, &types.Task{
			Type:       types.TaskCode,
			Objective:  fmt.Sprintf("analyze and optimize: %s", objective),
			Inputs:     inputs,
			MaxRetries: 3,
		})
		subtasks = append(subtasks, &types.Task{
			Type:       types.TaskOptimize,
			Objective:  fmt.Sprintf("apply optimizations for: %s", objective),
			Inputs:     inputs,
			MaxRetries: 2,
		})
	}

	// Si no hay subtareas específicas, crear flujo estándar
	if len(subtasks) == 0 {
		subtasks = append(subtasks, &types.Task{
			Type:       types.TaskCode,
			Objective:  objective,
			Inputs:     inputs,
			MaxRetries: 3,
		})
	}

	// Siempre terminar con tests y auditoría
	subtasks = append(subtasks, &types.Task{
		Type:       types.TaskTest,
		Objective:  "execute test suite",
		MaxRetries: 2,
	})

	subtasks = append(subtasks, &types.Task{
		Type:       types.TaskAudit,
		Objective:  "audit code changes",
		MaxRetries: 1,
	})

	return subtasks
}

// containsKeywords verifica si un string contiene alguna de las keywords
func containsKeywords(text string, keywords []string) bool {
	lowerText := toLower(text)
	for _, keyword := range keywords {
		if contains(lowerText, toLower(keyword)) {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		result[i] = c
	}
	return string(result)
}
