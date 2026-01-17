package agents

import (
	"context"

	"github.com/nanochip/multi-agent/pkg/policies"
	"github.com/nanochip/multi-agent/pkg/types"
	"github.com/nanochip/multi-agent/pkg/workspace"
)

// Agent define la interfaz que todos los agentes deben implementar
type Agent interface {
	Execute(ctx context.Context, task *types.Task) *types.TaskResult
	GetContract() types.AgentContract
}

// BaseAgent proporciona funcionalidad común a todos los agentes
type BaseAgent struct {
	workspace workspace.Manager
	policy    policies.Engine
	contract  types.AgentContract
}

// NewBaseAgent crea un nuevo agente base
func NewBaseAgent(ws workspace.Manager, policy policies.Engine, contract types.AgentContract) *BaseAgent {
	return &BaseAgent{
		workspace: ws,
		policy:    policy,
		contract:  contract,
	}
}

// ValidatePath verifica si una ruta está permitida según el contrato
func (b *BaseAgent) ValidatePath(path string) bool {
	// Verificar rutas prohibidas primero
	for _, forbidden := range b.contract.ForbiddenPaths {
		if matched, _ := pathMatches(path, forbidden); matched {
			return false
		}
	}
	
	// Si hay rutas permitidas, verificar que esté en la lista
	if len(b.contract.AllowedPaths) > 0 {
		allowed := false
		for _, allowedPath := range b.contract.AllowedPaths {
			if matched, _ := pathMatches(path, allowedPath); matched {
				allowed = true
				break
			}
		}
		return allowed
	}
	
	return true
}

// pathMatches verifica si un path coincide con un patrón
func pathMatches(path, pattern string) (bool, error) {
	if pattern == "*" {
		return true, nil
	}
	if pattern == path {
		return true, nil
	}
	// Implementación más completa de glob matching
	return false, nil
}

// GetContract retorna el contrato del agente
func (b *BaseAgent) GetContract() types.AgentContract {
	return b.contract
}
