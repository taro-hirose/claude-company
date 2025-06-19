package orchestrator

import (
	"context"
	"time"
)

// Orchestrator は AI タスクを統制・管理するメインインターフェース
type Orchestrator interface {
	// タスク管理
	CreateTask(ctx context.Context, req TaskRequest) (*TaskResponse, error)
	GetTask(ctx context.Context, taskID string) (*Task, error)
	ListTasks(ctx context.Context, filter TaskFilter) ([]*Task, error)
	UpdateTask(ctx context.Context, taskID string, updates TaskUpdate) error
	CancelTask(ctx context.Context, taskID string) error
	DeleteTask(ctx context.Context, taskID string) error

	// プラン管理
	CreatePlan(ctx context.Context, taskID string) (*TaskPlan, error)
	UpdatePlan(ctx context.Context, planID string, updates PlanUpdate) error
	ExecutePlan(ctx context.Context, planID string) error

	// ワーカー管理
	RegisterWorker(ctx context.Context, worker Worker) error
	UnregisterWorker(ctx context.Context, workerID string) error
	ListWorkers(ctx context.Context) ([]*Worker, error)
	AssignTask(ctx context.Context, taskID string, workerID string) error

	// イベント管理
	Subscribe(ctx context.Context, eventTypes []TaskEventType) (<-chan TaskEvent, error)
	PublishEvent(ctx context.Context, event TaskEvent) error

	// システム管理
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Status(ctx context.Context) (*SystemStatus, error)
}

// TaskPlanner はタスク計画を立案するインターフェース
type TaskPlanner interface {
	// 計画立案
	AnalyzeTask(ctx context.Context, task *Task) (*TaskAnalysis, error)
	CreatePlan(ctx context.Context, task *Task, analysis *TaskAnalysis) (*TaskPlan, error)
	OptimizePlan(ctx context.Context, plan *TaskPlan) (*TaskPlan, error)
	
	// 依存関係管理
	ResolveDependencies(ctx context.Context, tasks []*Task) (*DependencyGraph, error)
	ValidatePlan(ctx context.Context, plan *TaskPlan) error
}

// TaskExecutor はタスク実行を管理するインターフェース
type TaskExecutor interface {
	// 実行管理
	ExecuteTask(ctx context.Context, task *Task) (*TaskResult, error)
	ExecuteStep(ctx context.Context, step *TaskStep) (*StepOutput, error)
	
	// 進捗管理
	GetProgress(ctx context.Context, taskID string) (*TaskProgress, error)
	UpdateProgress(ctx context.Context, taskID string, progress *TaskProgress) error
	
	// リソース管理
	AllocateResources(ctx context.Context, task *Task) (*ResourceAllocation, error)
	ReleaseResources(ctx context.Context, allocationID string) error
}

// WorkerManager はワーカーを管理するインターフェース
type WorkerManager interface {
	// ワーカー管理
	CreateWorker(ctx context.Context, config WorkerConfig) (*Worker, error)
	GetWorker(ctx context.Context, workerID string) (*Worker, error)
	UpdateWorker(ctx context.Context, workerID string, updates WorkerUpdate) error
	RemoveWorker(ctx context.Context, workerID string) error
	
	// 割り当て管理
	FindAvailableWorker(ctx context.Context, requirements WorkerRequirements) (*Worker, error)
	AssignTask(ctx context.Context, workerID string, taskID string) error
	UnassignTask(ctx context.Context, workerID string) error
	
	// ヘルスチェック
	HealthCheck(ctx context.Context, workerID string) error
	MonitorWorkers(ctx context.Context) error
}

// EventBus はイベント配信を管理するインターフェース
type EventBus interface {
	// イベント配信
	Publish(ctx context.Context, event TaskEvent) error
	Subscribe(ctx context.Context, eventTypes []TaskEventType) (<-chan TaskEvent, error)
	Unsubscribe(ctx context.Context, subscription string) error
	
	// フィルタリング
	AddFilter(ctx context.Context, filter EventFilter) error
	RemoveFilter(ctx context.Context, filterID string) error
}

// Storage はデータ永続化を管理するインターフェース
type Storage interface {
	// タスク操作
	SaveTask(ctx context.Context, task *Task) error
	LoadTask(ctx context.Context, taskID string) (*Task, error)
	ListTasks(ctx context.Context, filter TaskFilter) ([]*Task, error)
	DeleteTask(ctx context.Context, taskID string) error
	
	// プラン操作
	SavePlan(ctx context.Context, plan *TaskPlan) error
	LoadPlan(ctx context.Context, planID string) (*TaskPlan, error)
	DeletePlan(ctx context.Context, planID string) error
	
	// ワーカー操作
	SaveWorker(ctx context.Context, worker *Worker) error
	LoadWorker(ctx context.Context, workerID string) (*Worker, error)
	ListWorkers(ctx context.Context) ([]*Worker, error)
	DeleteWorker(ctx context.Context, workerID string) error
	
	// イベント操作
	SaveEvent(ctx context.Context, event *TaskEvent) error
	ListEvents(ctx context.Context, filter EventFilter) ([]*TaskEvent, error)
	
	// クリーンアップ
	Cleanup(ctx context.Context) error
}

// 補助型定義

type TaskFilter struct {
	Status   []TaskStatus   `json:"status,omitempty"`
	Type     []TaskType     `json:"type,omitempty"`
	Priority []TaskPriority `json:"priority,omitempty"`
	CreatedAfter  *string   `json:"created_after,omitempty"`
	CreatedBefore *string   `json:"created_before,omitempty"`
	Limit    int            `json:"limit,omitempty"`
	Offset   int            `json:"offset,omitempty"`
}

type TaskUpdate struct {
	Title       *string      `json:"title,omitempty"`
	Description *string      `json:"description,omitempty"`
	Status      *TaskStatus  `json:"status,omitempty"`
	Priority    *TaskPriority `json:"priority,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type PlanUpdate struct {
	Strategy      *PlanStrategy `json:"strategy,omitempty"`
	Steps         []TaskStep    `json:"steps,omitempty"`
	EstimatedTime *int64        `json:"estimated_time,omitempty"`
	Dependencies  []string      `json:"dependencies,omitempty"`
}

type TaskAnalysis struct {
	Complexity   ComplexityLevel    `json:"complexity"`
	Requirements []string           `json:"requirements"`
	Dependencies []string           `json:"dependencies"`
	Risks        []Risk             `json:"risks"`
	Suggestions  []string           `json:"suggestions"`
}

type ComplexityLevel string

const (
	ComplexityLow    ComplexityLevel = "low"
	ComplexityMedium ComplexityLevel = "medium"
	ComplexityHigh   ComplexityLevel = "high"
)

type Risk struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Impact      string  `json:"impact"`
	Probability float64 `json:"probability"`
	Mitigation  string  `json:"mitigation"`
}

type DependencyGraph struct {
	Nodes []DependencyNode `json:"nodes"`
	Edges []DependencyEdge `json:"edges"`
}

type DependencyNode struct {
	TaskID string `json:"task_id"`
	Level  int    `json:"level"`
}

type DependencyEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

type TaskProgress struct {
	TaskID          string  `json:"task_id"`
	CompletedSteps  int     `json:"completed_steps"`
	TotalSteps      int     `json:"total_steps"`
	PercentComplete float64 `json:"percent_complete"`
	CurrentStep     *string `json:"current_step,omitempty"`
	EstimatedTimeRemaining *int64 `json:"estimated_time_remaining,omitempty"`
}

type ResourceAllocation struct {
	ID        string            `json:"id"`
	TaskID    string            `json:"task_id"`
	Resources map[string]any    `json:"resources"`
	AllocatedAt time.Time       `json:"allocated_at"`
}

type WorkerConfig struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	Capabilities []string          `json:"capabilities"`
	Config       map[string]any    `json:"config"`
}

type WorkerUpdate struct {
	Status       *WorkerStatus     `json:"status,omitempty"`
	Capabilities []string          `json:"capabilities,omitempty"`
	Config       map[string]any    `json:"config,omitempty"`
}

type WorkerRequirements struct {
	Capabilities []string          `json:"capabilities"`
	MinResources map[string]any    `json:"min_resources,omitempty"`
	Preferences  map[string]any    `json:"preferences,omitempty"`
}

type EventFilter struct {
	ID         string            `json:"id"`
	EventTypes []TaskEventType   `json:"event_types"`
	TaskIDs    []string          `json:"task_ids,omitempty"`
	Conditions map[string]any    `json:"conditions,omitempty"`
}

type SystemStatus struct {
	Version        string            `json:"version"`
	Uptime         int64             `json:"uptime"`
	ActiveTasks    int               `json:"active_tasks"`
	ActiveWorkers  int               `json:"active_workers"`
	SystemLoad     SystemLoad        `json:"system_load"`
	Health         HealthStatus      `json:"health"`
}

type SystemLoad struct {
	CPU     float64 `json:"cpu"`
	Memory  float64 `json:"memory"`
	Disk    float64 `json:"disk"`
	Network float64 `json:"network"`
}

type HealthStatus struct {
	Overall    string            `json:"overall"`
	Components map[string]string `json:"components"`
}