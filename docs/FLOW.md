# Diagrama de Flujo del Sistema Multi-Agente

## Flujo Principal

```
┌─────────────────────────────────────────────────────────────┐
│                    USUARIO/CI/CD                            │
│            Envía objetivo: "fix bug X"                      │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                  ORCHESTRATOR                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Cola de Tareas                                      │  │
│  │  [Task-1] → [Task-2] → [Task-3]                      │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Policy Engine                                       │  │
│  │  ✅ Allow Task?                                      │  │
│  │  ✅ Validate Result?                                 │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Estado & Memoria                                    │  │
│  │  - Task State                                        │  │
│  │  - Decision Log                                      │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────┬────────────────────────────────────────┘
                     │
                     │ Selecciona Agente
                     ▼
        ┌────────────────────────┐
        │   PLANNER AGENT        │
        │   (TaskPlan)           │
        │                        │
        │  Input:  Objetivo      │
        │  Output: Subtareas []  │
        └───────────┬────────────┘
                    │
                    ▼ Subtareas
        ┌────────────────────────┐
        │    CODER AGENT         │
        │    (TaskCode)          │
        │                        │
        │  Input:  Subtarea      │
        │  Output: Files changed │
        │          Branch        │
        └───────────┬────────────┘
                    │
                    ▼
        ┌────────────────────────┐
        │   TESTER AGENT         │
        │   (TaskTest)           │
        │                        │
        │  Input:  Changed files │
        │  Output: TestResult    │
        │          Coverage      │
        └───────────┬────────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
        ▼                       ▼
    Tests ❌              Tests ✅
        │                       │
        │                       │
        ▼                       ▼
┌───────────────┐       ┌───────────────┐
│ REPAIRER      │       │ AUDITOR       │
│ AGENT         │       │ AGENT         │
│               │       │               │
│ Analyze       │       │ Lint Check    │
│ Failures      │       │ Security      │
│ Apply Fix     │       │ Dependencies  │
│               │       │ Secrets       │
└───────┬───────┘       └───────┬───────┘
        │                       │
        │                       │
        ▼                       ▼
    Re-test            ┌────────┴────────┐
        │              │                 │
        │              ▼                 ▼
        │         Critical          No Critical
        │         Findings          Findings
        │              │                 │
        │              │                 │
        └──────┬───────┘                 │
               │                         │
               ▼                         ▼
        ┌──────────────────────────────┐
        │     REPAIRER                 │
        │     (Fix audit findings)     │
        └─────────────┬────────────────┘
                      │
                      ▼
        ┌──────────────────────────────┐
        │    OPTIMIZER AGENT           │
        │    (TaskOptimize)            │
        │                              │
        │  Run Benchmarks              │
        │  Apply Safe Optimizations    │
        │  Validate Tests Still Pass   │
        └─────────────┬────────────────┘
                      │
                      ▼
        ┌──────────────────────────────┐
        │    RELEASE                   │
        │    (TaskRelease)             │
        │                              │
        │  Commit Changes              │
        │  Create PR                   │
        │  Tag Version                 │
        └──────────────────────────────┘
```

## Ciclo de Auto-Reparación

```
┌──────────────────────────────────────┐
│   TEST FAILURE DETECTED              │
│   State: FAILED                      │
└──────────────┬───────────────────────┘
               │
               ▼
┌──────────────────────────────────────┐
│   REPAIRER ANALYZES                  │
│   ┌────────────────────────────────┐ │
│   │ • Parse error message          │ │
│   │ • Identify error type          │ │
│   │ • Generate hypothesis          │ │
│   └────────────────────────────────┘ │
└──────────────┬───────────────────────┘
               │
               ▼
┌──────────────────────────────────────┐
│   APPLY FIX                          │
│   ┌────────────────────────────────┐ │
│   │ Strategy: nil_pointer_check    │ │
│   │ Fix: Add nil check             │ │
│   │ Files: pkg/user/service.go     │ │
│   └────────────────────────────────┘ │
└──────────────┬───────────────────────┘
               │
               ▼
┌──────────────────────────────────────┐
│   RE-RUN TESTS                       │
│   go test ./...                      │
└──────────────┬───────────────────────┘
               │
        ┌──────┴──────┐
        │             │
        ▼             ▼
    PASS ✅      FAIL ❌
        │             │
        │             │
        │             ▼
        │      ┌──────────────┐
        │      │ Retry Count  │
        │      │ < MaxRetries?│
        │      └──────┬───────┘
        │             │
        │        ┌────┴────┐
        │        │         │
        │        YES       NO
        │        │         │
        │        │         ▼
        │        │    ┌─────────────┐
        │        │    │ Give Up     │
        │        │    │ State: FAIL │
        │        │    └─────────────┘
        │        │
        │        ▼
        │   Back to ANALYZE
        │
        ▼
┌──────────────────────────────────────┐
│   SUCCESS                            │
│   State: SUCCESS                     │
│   Next: AUDIT                        │
└──────────────────────────────────────┘
```

## Gates de Validación

```
┌──────────────────────────────────────┐
│   TASK RESULT                        │
└──────────────┬───────────────────────┘
               │
               ▼
┌──────────────────────────────────────┐
│   POLICY ENGINE                      │
│   Validate Result                    │
└──────────────┬───────────────────────┘
               │
               ▼
   ┌───────────────────────┐
   │   GATE 1: fmt/lint    │
   │   ✅ Pass             │
   └───────────┬───────────┘
               │
               ▼
   ┌───────────────────────┐
   │   GATE 2: tests-pass  │
   │   ✅ Pass             │
   └───────────┬───────────┘
               │
               ▼
   ┌───────────────────────┐
   │   GATE 3: coverage    │
   │   ✅ 75% >= 70%       │
   └───────────┬───────────┘
               │
               ▼
   ┌───────────────────────┐
   │   GATE 4: secrets     │
   │   ✅ No secrets found │
   └───────────┬───────────┘
               │
               ▼
   ┌───────────────────────┐
   │   GATE 5: deps        │
   │   ✅ No CVEs          │
   └───────────┬───────────┘
               │
               ▼
        ┌──────────────┐
        │  ALL GATES   │
        │  PASSED ✅   │
        └──────┬───────┘
               │
               ▼
        Continue Flow
```

## Estados de Tarea

```
PENDING
  │
  ▼
RUNNING ────────┐
  │             │
  │             │ (Success)
  │             ▼
  │         SUCCESS
  │             │
  │             ▼
  │         (Next Task)
  │
  │ (Failure)
  ▼
FAILED
  │
  │ (Retries < Max)
  ▼
RETRYING ───┐
  │         │
  │         │ (After delay)
  └─────────┘
  │
  │ (Retries >= Max)
  ▼
FAILED (Final)
```

## Comunicación Entre Agentes

```
Task → TaskResult → Next Task

Cada TaskResult contiene:
  - Success: bool
  - Outputs: map[string]interface{}
  - Evidence: []Evidence
    - Logs
    - Reports
    - Diffs
  - Decisions: []Decision
    - Agent
    - Reason
    - Action
    - Confidence

El Orchestrator decide el siguiente paso
basado en el TaskResult y el tipo de tarea.
```
