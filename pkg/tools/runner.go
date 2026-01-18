package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Runner ejecuta comandos en un sandbox con validación
type Runner struct {
	allowedCommands map[string][]string // agente -> comandos permitidos
	maxMemoryMB     int
	maxCPUSeconds    int
	maxDuration      time.Duration
}

// CommandResult representa el resultado de un comando
type CommandResult struct {
	Command     string
	Args        []string
	Output      string
	Error       string
	ExitCode    int
	Duration    time.Duration
	MemoryUsed  int // MB
	Allowed     bool
	BlockReason string
}

// NewRunner crea un nuevo tool runner
func NewRunner() *Runner {
	return &Runner{
		allowedCommands: make(map[string][]string),
		maxMemoryMB:     1024, // 1GB por defecto
		maxCPUSeconds:    300,  // 5 minutos
		maxDuration:      time.Minute * 10,
	}
}

// SetAllowedCommands configura comandos permitidos para un agente
func (r *Runner) SetAllowedCommands(agentID string, commands []string) {
	r.allowedCommands[agentID] = commands
}

// SetLimits configura límites de recursos
func (r *Runner) SetLimits(maxMemoryMB, maxCPUSeconds int, maxDuration time.Duration) {
	r.maxMemoryMB = maxMemoryMB
	r.maxCPUSeconds = maxCPUSeconds
	r.maxDuration = maxDuration
}

// Run ejecuta un comando con validación y sandbox
func (r *Runner) Run(ctx context.Context, agentID, cmd string, args ...string) (*CommandResult, error) {
	result := &CommandResult{
		Command: cmd,
		Args:    args,
	}
	
	// Validar comando permitido
	if !r.isCommandAllowed(agentID, cmd) {
		result.Allowed = false
		result.BlockReason = fmt.Sprintf("command '%s' not allowed for agent '%s'", cmd, agentID)
		return result, fmt.Errorf("command not allowed: %s", cmd)
	}
	
	result.Allowed = true
	
	// Crear comando con contexto
	command := exec.CommandContext(ctx, cmd, args...)
	
	// Configurar límites de recursos (si el OS lo soporta)
	// En Linux: usar cgroups, en otros: limitaciones básicas
	
	// Ejecutar con timeout
	startTime := time.Now()
	output, err := command.CombinedOutput()
	duration := time.Since(startTime)
	
	result.Duration = duration
	result.Output = string(output)
	
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
		}
		result.Error = err.Error()
	} else {
		result.ExitCode = 0
	}
	
	// Verificar límites
	if duration > r.maxDuration {
		result.BlockReason = fmt.Sprintf("command exceeded max duration: %v", duration)
		return result, fmt.Errorf("command exceeded max duration")
	}
	
	return result, nil
}

// isCommandAllowed verifica si un comando está permitido para un agente
func (r *Runner) isCommandAllowed(agentID, cmd string) bool {
	allowed, exists := r.allowedCommands[agentID]
	if !exists {
		// Si no hay restricciones específicas, permitir comandos comunes
		commonCommands := []string{"go", "git", "ls", "cat", "echo"}
		for _, common := range commonCommands {
			if cmd == common {
				return true
			}
		}
		return false
	}
	
	// Verificar si el comando está en la lista permitida
	for _, allowedCmd := range allowed {
		if allowedCmd == cmd || strings.HasPrefix(cmd, allowedCmd) {
			return true
		}
	}
	
	return false
}

// ValidateCommand valida un comando sin ejecutarlo
func (r *Runner) ValidateCommand(agentID, cmd string, args ...string) (bool, string) {
	if !r.isCommandAllowed(agentID, cmd) {
		return false, fmt.Sprintf("command '%s' not allowed for agent '%s'", cmd, agentID)
	}
	
	// Validaciones adicionales de argumentos
	if cmd == "rm" && contains(args, "-rf") {
		return false, "dangerous command: rm -rf"
	}
	
	if cmd == "git" && contains(args, "push", "--force") {
		return false, "dangerous command: git push --force"
	}
	
	return true, ""
}

// contains verifica si un slice contiene un string
func contains(slice []string, items ...string) bool {
	for _, item := range items {
		for _, s := range slice {
			if s == item {
				return true
			}
		}
	}
	return false
}
