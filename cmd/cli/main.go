package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/nanochip/multi-agent/pkg/orchestrator"
	"github.com/nanochip/multi-agent/pkg/policies"
	"github.com/nanochip/multi-agent/pkg/types"
	"github.com/nanochip/multi-agent/pkg/workspace"
)

func main() {
	planCmd := flag.NewFlagSet("plan", flag.ExitOnError)
	planObj := planCmd.String("objective", "", "Objective to plan")

	statusCmd := flag.NewFlagSet("status", flag.ExitOnError)
	taskID := statusCmd.String("task", "", "Task ID to check")

	repoPath := flag.String("repo", ".", "Path to git repository")

	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  plan  - Create a plan for an objective")
		fmt.Println("  status - Check task status")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "plan":
		planCmd.Parse(os.Args[2:])
		if *planObj == "" {
			log.Fatal("--objective is required")
		}
		handlePlan(*repoPath, *planObj)

	case "status":
		statusCmd.Parse(os.Args[2:])
		if *taskID == "" {
			log.Fatal("--task is required")
		}
		handleStatus(*repoPath, *taskID)

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func handlePlan(repoPath, objective string) {
	ws, err := workspace.NewManager(repoPath)
	if err != nil {
		log.Fatalf("Failed to create workspace manager: %v", err)
	}

	policy := policies.NewEngine()
	orch := orchestrator.New(ws, policy)
	orch.Start()

	task := &types.Task{
		Type:        types.TaskPlan,
		Objective:   objective,
		Inputs:      make(map[string]interface{}),
		Constraints: make(map[string]interface{}),
		MaxRetries:  3,
	}

	if err := orch.SubmitTask(task); err != nil {
		log.Fatalf("Failed to submit task: %v", err)
	}

	fmt.Printf("Plan task submitted: %s\n", task.ID)
	fmt.Printf("Objective: %s\n", objective)

	// Esperar un poco para que se procese
	// En producción, usarías un mecanismo de espera más robusto
}

func handleStatus(repoPath, taskID string) {
	ws, err := workspace.NewManager(repoPath)
	if err != nil {
		log.Fatalf("Failed to create workspace manager: %v", err)
	}

	policy := policies.NewEngine()
	orch := orchestrator.New(ws, policy)

	task, result := orch.GetTaskState(taskID)
	if task == nil {
		log.Fatalf("Task %s not found", taskID)
	}

	fmt.Printf("Task ID: %s\n", task.ID)
	fmt.Printf("Type: %s\n", task.Type)
	fmt.Printf("State: %s\n", task.State)
	fmt.Printf("Objective: %s\n", task.Objective)

	if result != nil {
		fmt.Printf("\nResult:\n")
		fmt.Printf("  Success: %v\n", result.Success)
		fmt.Printf("  Duration: %v\n", result.Duration)
		if result.Error != "" {
			fmt.Printf("  Error: %s\n", result.Error)
		}

		if len(result.Outputs) > 0 {
			fmt.Printf("\nOutputs:\n")
			outputJSON, _ := json.MarshalIndent(result.Outputs, "  ", "  ")
			fmt.Println(string(outputJSON))
		}
	}
}
