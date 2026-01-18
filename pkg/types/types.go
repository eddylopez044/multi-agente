package types

import (
	"encoding/json"
	"time"
)

// TaskType representa el tipo de tarea
type TaskType string

const (
	TaskPlan     TaskType = "plan"
	TaskCode     TaskType = "code"
	TaskTest     TaskType = "test"
	TaskAudit    TaskType = "audit"
	TaskRepair   TaskType = "repair"
	TaskOptimize TaskType = "optimize"
	TaskRelease  TaskType = "release"
)

// TaskState representa el estado de una tarea
type TaskState string

const (
	StatePending   TaskState = "pending"
	StateRunning   TaskState = "running"
	StateSuccess   TaskState = "success"
	StateFailed    TaskState = "failed"
	StateRetrying  TaskState = "retrying"
	StateCancelled TaskState = "cancelled"
)

// Severity representa la severidad de un hallazgo
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// Task representa una tarea en el sistema
type Task struct {
	ID          string                 `json:"id"`
	Type        TaskType               `json:"type"`
	State       TaskState              `json:"state"`
	Objective   string                 `json:"objective"`
	Inputs      map[string]interface{} `json:"inputs"`
	Constraints map[string]interface{} `json:"constraints"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	ParentID    string                 `json:"parent_id,omitempty"`
}

// TaskResult representa el resultado de una tarea
type TaskResult struct {
	TaskID    string                 `json:"task_id"`
	State     TaskState              `json:"state"`
	Success   bool                   `json:"success"`
	Outputs   map[string]interface{} `json:"outputs"`
	Evidence  []Evidence             `json:"evidence"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Decisions []Decision             `json:"decisions,omitempty"`
}

// Evidence representa evidencia de ejecución (logs, reportes, diffs)
type Evidence struct {
	Type        string          `json:"type"` // "log", "report", "diff", "metric"
	Source      string          `json:"source"`
	Content     json.RawMessage `json:"content"`
	Timestamp   time.Time       `json:"timestamp"`
	Description string          `json:"description,omitempty"`
}

// Decision representa una decisión tomada por un agente
type Decision struct {
	Agent      string                 `json:"agent"`
	Reason     string                 `json:"reason"`
	Action     string                 `json:"action"`
	Timestamp  time.Time              `json:"timestamp"`
	Confidence float64                `json:"confidence,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// AuditFinding representa un hallazgo de auditoría
type AuditFinding struct {
	ID          string                 `json:"id"`
	Severity    Severity               `json:"severity"`
	Category    string                 `json:"category"` // "security", "style", "dependency", "license", "secret"
	Rule        string                 `json:"rule"`
	File        string                 `json:"file,omitempty"`
	Line        int                    `json:"line,omitempty"`
	Message     string                 `json:"message"`
	Remediation string                 `json:"remediation,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TestResult representa el resultado de tests
type TestResult struct {
	Passed   int                    `json:"passed"`
	Failed   int                    `json:"failed"`
	Skipped  int                    `json:"skipped"`
	Duration time.Duration          `json:"duration"`
	Coverage float64                `json:"coverage"`
	Failures []TestFailure          `json:"failures,omitempty"`
	Command  string                 `json:"command"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TestFailure representa un test que falló
type TestFailure struct {
	Test    string `json:"test"`
	Package string `json:"package"`
	Message string `json:"message"`
	Output  string `json:"output,omitempty"`
}

// Policy representa una política de guardrail
type Policy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"` // "gate", "constraint", "rule"
	Enabled     bool                   `json:"enabled"`
	Rules       []PolicyRule           `json:"rules"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PolicyRule representa una regla de política
type PolicyRule struct {
	Condition string                 `json:"condition"`
	Action    string                 `json:"action"` // "allow", "deny", "warn"
	Message   string                 `json:"message,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AgentContract define el contrato de un agente
type AgentContract struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	AllowedPaths   []string               `json:"allowed_paths,omitempty"`
	ForbiddenPaths []string               `json:"forbidden_paths,omitempty"`
	AllowedTools   []string               `json:"allowed_tools,omitempty"`
	RequiredTests  bool                   `json:"required_tests"`
	Constraints    map[string]interface{} `json:"constraints,omitempty"`
}
