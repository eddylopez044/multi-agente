package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nanochip/multi-agent/pkg/agents"
	"github.com/nanochip/multi-agent/pkg/policies"
	"github.com/nanochip/multi-agent/pkg/types"
	"github.com/nanochip/multi-agent/pkg/workspace"
)

// Orchestrator coordina todos los agentes y gestiona el flujo de trabajo
type Orchestrator struct {
	workspace  workspace.Manager
	policy     policies.Engine
	taskQueue  chan *types.Task
	taskState  map[string]*types.Task
	results    map[string]*types.TaskResult
	memory     []types.Decision
	mu         sync.RWMutex
	agents     map[string]agents.Agent
	ctx        context.Context
	cancel     context.CancelFunc
}

// New crea un nuevo Orchestrator
func New(ws workspace.Manager, policyEngine policies.Engine) *Orchestrator {
	ctx, cancel := context.WithCancel(context.Background())
	
	o := &Orchestrator{
		workspace: ws,
		policy:    policyEngine,
		taskQueue: make(chan *types.Task, 100),
		taskState: make(map[string]*types.Task),
		results:   make(map[string]*types.TaskResult),
		memory:    make([]types.Decision, 0),
		agents:    make(map[string]agents.Agent),
		ctx:       ctx,
		cancel:    cancel,
	}
	
	// Registrar agentes
	o.registerAgents()
	
	return o
}

// registerAgents registra todos los agentes disponibles
func (o *Orchestrator) registerAgents() {
	o.agents["planner"] = agents.NewPlanner(o.workspace, o.policy)
	o.agents["coder"] = agents.NewCoder(o.workspace, o.policy)
	o.agents["tester"] = agents.NewTester(o.workspace, o.policy)
	o.agents["auditor"] = agents.NewAuditor(o.workspace, o.policy)
	o.agents["repairer"] = agents.NewRepairer(o.workspace, o.policy)
	o.agents["optimizer"] = agents.NewOptimizer(o.workspace, o.policy)
}

// Start inicia el orchestrator
func (o *Orchestrator) Start() error {
	go o.processQueue()
	return nil
}

// Stop detiene el orchestrator
func (o *Orchestrator) Stop() {
	o.cancel()
	close(o.taskQueue)
}

// SubmitTask envía una nueva tarea al sistema
func (o *Orchestrator) SubmitTask(task *types.Task) error {
	task.CreatedAt = time.Now()
	task.State = types.StatePending
	
	o.mu.Lock()
	task.ID = fmt.Sprintf("task-%d", len(o.taskState)+1)
	o.taskState[task.ID] = task
	o.mu.Unlock()
	
	select {
	case o.taskQueue <- task:
		return nil
	case <-o.ctx.Done():
		return o.ctx.Err()
	default:
		return fmt.Errorf("task queue full")
	}
}

// processQueue procesa la cola de tareas
func (o *Orchestrator) processQueue() {
	for {
		select {
		case task := <-o.taskQueue:
			go o.executeTask(task)
		case <-o.ctx.Done():
			return
		}
	}
}

// executeTask ejecuta una tarea usando el agente apropiado
func (o *Orchestrator) executeTask(task *types.Task) {
	startTime := time.Now()
	
	// Actualizar estado
	o.updateTaskState(task.ID, types.StateRunning, nil)
	now := time.Now()
	task.StartedAt = &now
	
	// Verificar políticas antes de ejecutar
	if !o.policy.AllowTask(task) {
		result := &types.TaskResult{
			TaskID:   task.ID,
			State:    types.StateFailed,
			Success:  false,
			Error:    "task blocked by policy",
			Duration: time.Since(startTime),
		}
		o.recordResult(result)
		return
	}
	
	// Seleccionar agente
	agent := o.selectAgent(task.Type)
	if agent == nil {
		result := &types.TaskResult{
			TaskID:   task.ID,
			State:    types.StateFailed,
			Success:  false,
			Error:    fmt.Sprintf("no agent available for task type: %s", task.Type),
			Duration: time.Since(startTime),
		}
		o.recordResult(result)
		return
	}
	
	// Ejecutar agente
	result := agent.Execute(o.ctx, task)
	result.Duration = time.Since(startTime)
	
	// Validar resultado contra políticas (gates)
	if result.Success && !o.policy.ValidateResult(result) {
		result.Success = false
		result.State = types.StateFailed
		result.Error = "result failed policy validation"
	}
	
	// Registrar decisión
	if len(result.Decisions) > 0 {
		o.mu.Lock()
		o.memory = append(o.memory, result.Decisions...)
		o.mu.Unlock()
	}
	
	// Si falla y hay retries, reintentar
	if !result.Success && task.RetryCount < task.MaxRetries {
		task.RetryCount++
		task.State = types.StateRetrying
		time.Sleep(time.Second * time.Duration(task.RetryCount))
		go o.executeTask(task)
		return
	}
	
	// Actualizar estado final
	now = time.Now()
	task.CompletedAt = &now
	o.updateTaskState(task.ID, result.State, task.CompletedAt)
	o.recordResult(result)
	
	// Si hay subtareas, ejecutarlas
	if nextTasks := o.getNextTasks(task, result); len(nextTasks) > 0 {
		for _, nextTask := range nextTasks {
			o.SubmitTask(nextTask)
		}
	}
}

// selectAgent selecciona el agente apropiado para el tipo de tarea
func (o *Orchestrator) selectAgent(taskType types.TaskType) agents.Agent {
	switch taskType {
	case types.TaskPlan:
		return o.agents["planner"]
	case types.TaskCode:
		return o.agents["coder"]
	case types.TaskTest:
		return o.agents["tester"]
	case types.TaskAudit:
		return o.agents["auditor"]
	case types.TaskRepair:
		return o.agents["repairer"]
	case types.TaskOptimize:
		return o.agents["optimizer"]
	default:
		return nil
	}
}

// getNextTasks determina las siguientes tareas basadas en el resultado
func (o *Orchestrator) getNextTasks(task *types.Task, result *types.TaskResult) []*types.Task {
	nextTasks := make([]*types.Task, 0)
	
	switch task.Type {
	case types.TaskPlan:
		// Después de planificar, ejecutar las subtareas
		if subtasks, ok := result.Outputs["subtasks"].([]*types.Task); ok {
			return subtasks
		}
		
	case types.TaskCode:
		// Después de codificar, ejecutar tests
		if result.Success {
			nextTasks = append(nextTasks, &types.Task{
				Type:      types.TaskTest,
				Objective: "test code changes",
				Inputs:    map[string]interface{}{"task_id": task.ID},
				ParentID:  task.ID,
			})
		}
		
	case types.TaskTest:
		// Después de testear, decidir: repair, audit o continuar
		if !result.Success {
			// Si falla, reparar
			nextTasks = append(nextTasks, &types.Task{
				Type:      types.TaskRepair,
				Objective: "repair failing tests",
				Inputs:    map[string]interface{}{"test_result": result},
				ParentID:  task.ID,
			})
		} else {
			// Si pasa, auditar
			nextTasks = append(nextTasks, &types.Task{
				Type:      types.TaskAudit,
				Objective: "audit code changes",
				Inputs:    map[string]interface{}{"task_id": task.ID},
				ParentID:  task.ID,
			})
		}
		
	case types.TaskRepair:
		// Después de reparar, volver a testear
		if result.Success {
			nextTasks = append(nextTasks, &types.Task{
				Type:      types.TaskTest,
				Objective: "verify repair",
				Inputs:    map[string]interface{}{"task_id": task.ID},
				ParentID:  task.ID,
			})
		}
		
	case types.TaskAudit:
		// Después de auditar, decidir: repair o optimize
		if auditFailures, ok := result.Outputs["critical_findings"].([]types.AuditFinding); ok && len(auditFailures) > 0 {
			// Si hay hallazgos críticos, reparar
			nextTasks = append(nextTasks, &types.Task{
				Type:      types.TaskRepair,
				Objective: "repair audit findings",
				Inputs:    map[string]interface{}{"audit_result": result},
				ParentID:  task.ID,
			})
		} else if result.Success {
			// Si todo ok, optimizar
			nextTasks = append(nextTasks, &types.Task{
				Type:      types.TaskOptimize,
				Objective: "optimize code",
				Inputs:    map[string]interface{}{"task_id": task.ID},
				ParentID:  task.ID,
			})
		}
		
	case types.TaskOptimize:
		// Después de optimizar, verificar que los tests aún pasan
		if result.Success {
			nextTasks = append(nextTasks, &types.Task{
				Type:      types.TaskTest,
				Objective: "verify optimization didn't break tests",
				Inputs:    map[string]interface{}{"task_id": task.ID},
				ParentID:  task.ID,
			})
		}
	}
	
	return nextTasks
}

// updateTaskState actualiza el estado de una tarea
func (o *Orchestrator) updateTaskState(taskID string, state types.TaskState, completedAt *time.Time) {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	if task, exists := o.taskState[taskID]; exists {
		task.State = state
		if completedAt != nil {
			task.CompletedAt = completedAt
		}
	}
}

// recordResult registra un resultado de tarea
func (o *Orchestrator) recordResult(result *types.TaskResult) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.results[result.TaskID] = result
}

// GetTaskState retorna el estado de una tarea
func (o *Orchestrator) GetTaskState(taskID string) (*types.Task, *types.TaskResult) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	task := o.taskState[taskID]
	result := o.results[taskID]
	return task, result
}

// GetMemory retorna la memoria de decisiones
func (o *Orchestrator) GetMemory() []types.Decision {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return append([]types.Decision{}, o.memory...)
}
