package orchestrator

import (
	"context"
	"time"
)

type TaskType string

const (
	TaskTypeFeature      TaskType = "feature"
	TaskTypeBugFix       TaskType = "bugfix"
	TaskTypeRefactoring  TaskType = "refactoring"
	TaskTypeDocumentation TaskType = "documentation"
	TaskTypeResearch     TaskType = "research"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

type TaskPriority string

const (
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityLow    TaskPriority = "low"
)

type Task struct {
	ID          string       `json:"id"`
	Type        TaskType     `json:"type"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Status      TaskStatus   `json:"status"`
	Priority    TaskPriority `json:"priority"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
	Plan        *TaskPlan    `json:"plan,omitempty"`
	Context     TaskContext  `json:"context"`
}

type TaskContext struct {
	ProjectPath string            `json:"project_path"`
	Branch      string            `json:"branch"`
	Environment map[string]string `json:"environment"`
	Metadata    map[string]any    `json:"metadata"`
}

type TaskPlan struct {
	ID              string          `json:"id"`
	TaskID          string          `json:"task_id"`
	Strategy        PlanStrategy    `json:"strategy"`
	Steps           []TaskStep      `json:"steps"`
	EstimatedTime   time.Duration   `json:"estimated_time"`
	ActualTime      *time.Duration  `json:"actual_time,omitempty"`
	SubTasks        []SubTask       `json:"subtasks"`
	Dependencies    []string        `json:"dependencies"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type PlanStrategy string

const (
	PlanStrategySequential PlanStrategy = "sequential"
	PlanStrategyParallel   PlanStrategy = "parallel"
	PlanStrategyHybrid     PlanStrategy = "hybrid"
)

type TaskStep struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Order        int          `json:"order"`
	Status       TaskStatus   `json:"status"`
	ParentTaskID string       `json:"parent_task_id"`
	Dependencies []string     `json:"dependencies"`
	StartedAt    *time.Time   `json:"started_at,omitempty"`
	CompletedAt  *time.Time   `json:"completed_at,omitempty"`
	Output       *StepOutput  `json:"output,omitempty"`
	Error        *StepError   `json:"error,omitempty"`
}

type SubTask struct {
	ID              string       `json:"id"`
	ParentTaskID    string       `json:"parent_task_id"`
	Title           string       `json:"title"`
	Description     string       `json:"description"`
	Status          TaskStatus   `json:"status"`
	AssignedWorker  string       `json:"assigned_worker"`
	StartedAt       *time.Time   `json:"started_at,omitempty"`
	CompletedAt     *time.Time   `json:"completed_at,omitempty"`
	Result          *TaskResult  `json:"result,omitempty"`
	Dependencies    []string     `json:"dependencies"`
}

type StepOutput struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Data    any    `json:"data,omitempty"`
}

type StepError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type TaskResult struct {
	Success bool              `json:"success"`
	Summary string            `json:"summary"`
	Details map[string]any    `json:"details"`
	Artifacts []TaskArtifact  `json:"artifacts"`
}

type TaskArtifact struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Metadata map[string]any `json:"metadata"`
}

type OrchestratorConfig struct {
	MaxConcurrentTasks int           `json:"max_concurrent_tasks"`
	TaskTimeout        time.Duration `json:"task_timeout"`
	RetryPolicy        RetryPolicy   `json:"retry_policy"`
	LogLevel           string        `json:"log_level"`
}

type RetryPolicy struct {
	MaxRetries     int           `json:"max_retries"`
	InitialBackoff time.Duration `json:"initial_backoff"`
	MaxBackoff     time.Duration `json:"max_backoff"`
	BackoffFactor  float64       `json:"backoff_factor"`
}

type WorkerStatus string

const (
	WorkerStatusIdle     WorkerStatus = "idle"
	WorkerStatusBusy     WorkerStatus = "busy"
	WorkerStatusOffline  WorkerStatus = "offline"
)

type Worker struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	Status       WorkerStatus `json:"status"`
	Capabilities []string     `json:"capabilities"`
	CurrentTask  *string      `json:"current_task,omitempty"`
	LastSeen     time.Time    `json:"last_seen"`
}

type TaskEvent struct {
	ID        string          `json:"id"`
	TaskID    string          `json:"task_id"`
	Type      TaskEventType   `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      map[string]any  `json:"data"`
}

type TaskEventType string

const (
	TaskEventCreated    TaskEventType = "task_created"
	TaskEventStarted    TaskEventType = "task_started"
	TaskEventProgress   TaskEventType = "task_progress"
	TaskEventCompleted  TaskEventType = "task_completed"
	TaskEventFailed     TaskEventType = "task_failed"
	TaskEventCancelled  TaskEventType = "task_cancelled"
	TaskEventRetried    TaskEventType = "task_retried"
)

type TaskRequest struct {
	Type        TaskType              `json:"type"`
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Priority    TaskPriority          `json:"priority"`
	Context     context.Context       `json:"-"`
	Metadata    map[string]any        `json:"metadata"`
}

type TaskResponse struct {
	TaskID      string        `json:"task_id"`
	Status      TaskStatus    `json:"status"`
	Message     string        `json:"message"`
	Plan        *TaskPlan     `json:"plan,omitempty"`
	Error       error         `json:"-"`
}