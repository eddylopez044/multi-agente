package main

import (
	"fmt"
	"log"

	"github.com/nanochip/multi-agent/pkg/orchestrator"
	"github.com/nanochip/multi-agent/pkg/policies"
	"github.com/nanochip/multi-agent/pkg/types"
	"github.com/nanochip/multi-agent/pkg/workspace"
)

func main() {
	// Ejemplo simple de uso del sistema multi-agente

	repoPath := "." // Cambiar a la ruta de tu repo
	objective := "fix bug in user authentication"

	// 1. Crear workspace manager
	ws, err := workspace.NewManager(repoPath)
	if err != nil {
		log.Fatalf("Failed to create workspace: %v", err)
	}
	defer ws.Cleanup()

	// 2. Crear policy engine
	policy := policies.NewEngine()

	// 3. Crear orchestrator
	orch := orchestrator.New(ws, policy)

	// 4. Iniciar orchestrator
	if err := orch.Start(); err != nil {
		log.Fatalf("Failed to start orchestrator: %v", err)
	}
	defer orch.Stop()

	// 5. Crear tarea inicial
	task := &types.Task{
		Type:        types.TaskPlan,
		Objective:   objective,
		Inputs:      make(map[string]interface{}),
		Constraints: make(map[string]interface{}),
		MaxRetries:  3,
	}

	// 6. Enviar tarea
	if err := orch.SubmitTask(task); err != nil {
		log.Fatalf("Failed to submit task: %v", err)
	}

	fmt.Printf("Task submitted: %s\n", task.ID)
	fmt.Printf("Objective: %s\n", objective)
	fmt.Printf("The multi-agent system will now:")
	fmt.Printf("\n  1. Plan the task")
	fmt.Printf("\n  2. Code the changes")
	fmt.Printf("\n  3. Test the changes")
	fmt.Printf("\n  4. Audit the code")
	fmt.Printf("\n  5. Repair if needed")
	fmt.Printf("\n  6. Optimize if possible\n")

	// En producción, aquí esperarías a que se complete o usarías un mecanismo de notificación
}
