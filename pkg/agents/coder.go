package agents

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nanochip/multi-agent/pkg/policies"
	"github.com/nanochip/multi-agent/pkg/types"
	"github.com/nanochip/multi-agent/pkg/workspace"
)

// Coder implementa cambios en el código
type Coder struct {
	*BaseAgent
}

// NewCoder crea un nuevo agente codificador
func NewCoder(ws workspace.Manager, policy policies.Engine) *Coder {
	contract := types.AgentContract{
		ID:           "coder",
		Name:         "Coder",
		AllowedPaths: []string{"src/**", "cmd/**", "internal/**", "pkg/**"},
		ForbiddenPaths: []string{"**/*_test.go", "vendor/**"},
		AllowedTools:   []string{"go", "git"},
		RequiredTests:  true,
	}
	
	return &Coder{
		BaseAgent: NewBaseAgent(ws, policy, contract),
	}
}

// Execute ejecuta la tarea de codificación
func (c *Coder) Execute(ctx context.Context, task *types.Task) *types.TaskResult {
	decision := types.Decision{
		Agent:      "coder",
		Reason:     fmt.Sprintf("implementing: %s", task.Objective),
		Action:     "code",
		Timestamp:  time.Now(),
		Confidence: 0.8,
	}
	
	// Crear rama para los cambios
	branchName := fmt.Sprintf("agent-code-%d", time.Now().Unix())
	if err := c.workspace.CheckoutBranch(branchName); err != nil {
		return &types.TaskResult{
			TaskID:   task.ID,
			State:    types.StateFailed,
			Success:  false,
			Error:    fmt.Sprintf("failed to create branch: %v", err),
			Decisions: []types.Decision{decision},
		}
	}
	
	// Analizar el objetivo para determinar qué archivos modificar
	filesToModify := c.analyzeObjective(task.Objective, task.Inputs)
	
	changes := make([]string, 0)
	evidence := make([]types.Evidence, 0)
	
	for _, file := range filesToModify {
		// Validar que el archivo está permitido
		fullPath := filepath.Join(c.workspace.GetRepoPath(), file)
		if !c.ValidatePath(file) {
			continue
		}
		
		// Intentar aplicar cambios (por ahora solo simular)
		change := c.applyChange(file, task.Objective)
		if change != "" {
			changes = append(changes, file)
			
			evidence = append(evidence, types.Evidence{
				Type:        "diff",
				Source:      file,
				Description: fmt.Sprintf("Change applied to %s", file),
				Timestamp:   time.Now(),
			})
		}
	}
	
	// Ejecutar fmt
	if output, err := c.workspace.RunCommand("go", "fmt", "./..."); err != nil {
		evidence = append(evidence, types.Evidence{
			Type:        "log",
			Source:      "go fmt",
			Content:     []byte(output),
			Timestamp:   time.Now(),
		})
	}
	
	outputs := map[string]interface{}{
		"files_changed": changes,
		"branch":        branchName,
	}
	
	success := len(changes) > 0
	
	return &types.TaskResult{
		TaskID:    task.ID,
		State:     mapState(success),
		Success:   success,
		Outputs:   outputs,
		Evidence:  evidence,
		Decisions: []types.Decision{decision},
	}
}

// analyzeObjective determina qué archivos deben modificarse
func (c *Coder) analyzeObjective(objective string, inputs map[string]interface{}) []string {
	files := make([]string, 0)
	
	// Si hay archivos específicos en los inputs
	if fileInput, ok := inputs["files"].([]interface{}); ok {
		for _, f := range fileInput {
			if fileStr, ok := f.(string); ok {
				files = append(files, fileStr)
			}
		}
	}
	
	// Buscar archivos relevantes en el workspace
	repoPath := c.workspace.GetRepoPath()
	
	// Buscar en directorios permitidos
	for _, allowedPath := range c.contract.AllowedPaths {
		searchPath := filepath.Join(repoPath, allowedPath)
		if matches := c.findRelevantFiles(searchPath, objective); len(matches) > 0 {
			files = append(files, matches...)
		}
	}
	
	// Si no se encontraron archivos, buscar archivos Go en src o cmd
	if len(files) == 0 {
		candidateDirs := []string{"src", "cmd", "internal", "pkg"}
		for _, dir := range candidateDirs {
			dirPath := filepath.Join(repoPath, dir)
			if _, err := os.Stat(dirPath); err == nil {
				if matches := c.findGoFiles(dirPath); len(matches) > 0 {
					files = append(files, matches[0]) // Tomar el primero como ejemplo
					break
				}
			}
		}
	}
	
	return files
}

// findRelevantFiles busca archivos relevantes para el objetivo
func (c *Coder) findRelevantFiles(searchPath string, objective string) []string {
	// Por ahora, búsqueda simple
	files := make([]string, 0)
	
	filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			// Simplificado: devolver algunos archivos
			if len(files) < 3 {
				files = append(files, path)
			}
		}
		return nil
	})
	
	return files
}

// findGoFiles encuentra archivos Go en un directorio
func (c *Coder) findGoFiles(dir string) []string {
	files := make([]string, 0)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			relPath, _ := filepath.Rel(dir, path)
			files = append(files, relPath)
		}
		return nil
	})
	return files
}

// applyChange aplica un cambio a un archivo (simulado por ahora)
func (c *Coder) applyChange(file, objective string) string {
	// Por ahora, solo retornar el nombre del archivo si existe
	fullPath := filepath.Join(c.workspace.GetRepoPath(), file)
	if _, err := os.Stat(fullPath); err == nil {
		return file
	}
	return ""
}

// mapState convierte un bool a TaskState
func mapState(success bool) types.TaskState {
	if success {
		return types.StateSuccess
	}
	return types.StateFailed
}
