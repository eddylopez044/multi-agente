package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nanochip/multi-agent/pkg/orchestrator"
	"github.com/nanochip/multi-agent/pkg/policies"
	"github.com/nanochip/multi-agent/pkg/types"
	"github.com/nanochip/multi-agent/pkg/workspace"
)

func main() {
	taskObj := flag.String("task", "", "Task objective to execute")
	repoPath := flag.String("repo", ".", "Path to git repository")
	flag.Parse()

	if *taskObj == "" {
		log.Fatal("--task is required")
	}

	// Crear workspace manager
	ws, err := workspace.NewManager(*repoPath)
	if err != nil {
		log.Fatalf("Failed to create workspace manager: %v", err)
	}

	// Crear policy engine
	policy := policies.NewEngine()

	// Configurar políticas por defecto para agentes
	setupDefaultPolicies(policy)

	// Crear orchestrator
	orch := orchestrator.New(ws, policy)

	// Iniciar orchestrator
	if err := orch.Start(); err != nil {
		log.Fatalf("Failed to start orchestrator: %v", err)
	}

	// Manejar señales para shutdown graceful
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Enviar tarea inicial
	task := &types.Task{
		Type:        types.TaskPlan,
		Objective:   *taskObj,
		Inputs:      make(map[string]interface{}),
		Constraints: make(map[string]interface{}),
		MaxRetries:  3,
	}

	if err := orch.SubmitTask(task); err != nil {
		log.Fatalf("Failed to submit task: %v", err)
	}

	fmt.Printf("Task submitted: %s\n", task.ID)
	fmt.Printf("Objective: %s\n", *taskObj)

	// Esperar señal de shutdown
	<-sigChan
	fmt.Println("\nShutting down...")
	orch.Stop()
	ws.Cleanup()
}

func setupDefaultPolicies(policy *policies.Engine) {
	// Política para Coder
	coderPolicy := types.Policy{
		ID:          "coder-policy",
		Name:        "Coder Agent Policy",
		Description: "Restricts coder to src, cmd, internal, pkg directories",
		Type:        "constraint",
		Enabled:     true,
		Metadata: map[string]interface{}{
			"agent_id":     "coder",
			"allowed_paths": []interface{}{"src/**", "cmd/**", "internal/**", "pkg/**"},
			"forbidden_paths": []interface{}{"**/*_test.go", "vendor/**"},
		},
	}
	policy.AddPolicy(coderPolicy)

	// Política para Tester
	testerPolicy := types.Policy{
		ID:          "tester-policy",
		Name:        "Tester Agent Policy",
		Description: "Tester can only modify test files",
		Type:        "constraint",
		Enabled:     true,
		Metadata: map[string]interface{}{
			"agent_id":     "tester",
			"allowed_paths": []interface{}{"**/*_test.go"},
			"forbidden_paths": []interface{}{"**/*.go"},
		},
	}
	policy.AddPolicy(testerPolicy)
}
