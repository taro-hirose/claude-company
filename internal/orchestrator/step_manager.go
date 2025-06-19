package orchestrator

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

type StepManager struct {
	mu               sync.RWMutex
	steps            map[string]*TaskStep
	stepsByTask      map[string][]*TaskStep
	stepExecutions   map[string]*StepExecution
	eventBus         EventBus
	storage          Storage
	config           StepManagerConfig
	executorPool     *ExecutorPool
}

type StepManagerConfig struct {
	MaxConcurrentSteps int           `json:"max_concurrent_steps"`
	StepTimeout        time.Duration `json:"step_timeout"`
	RetryPolicy        RetryPolicy   `json:"retry_policy"`
	ExecutorPoolSize   int           `json:"executor_pool_size"`
}

type StepExecution struct {
	Step       *TaskStep         `json:"step"`
	Context    context.Context   `json:"-"`
	Cancel     context.CancelFunc `json:"-"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    *time.Time        `json:"end_time,omitempty"`
	Progress   float64           `json:"progress"`
	Output     *StepOutput       `json:"output,omitempty"`
	Error      error             `json:"-"`
	RetryCount int               `json:"retry_count"`
}

type ExecutorPool struct {
	workers   chan struct{}
	executing sync.Map
	wg        sync.WaitGroup
}

func NewStepManager(eventBus EventBus, storage Storage, config StepManagerConfig) *StepManager {
	if config.MaxConcurrentSteps <= 0 {
		config.MaxConcurrentSteps = 10
	}
	if config.StepTimeout <= 0 {
		config.StepTimeout = 30 * time.Minute
	}
	if config.ExecutorPoolSize <= 0 {
		config.ExecutorPoolSize = 5
	}

	return &StepManager{
		steps:          make(map[string]*TaskStep),
		stepsByTask:    make(map[string][]*TaskStep),
		stepExecutions: make(map[string]*StepExecution),
		eventBus:       eventBus,
		storage:        storage,
		config:         config,
		executorPool:   newExecutorPool(config.ExecutorPoolSize),
	}
}

func newExecutorPool(size int) *ExecutorPool {
	return &ExecutorPool{
		workers: make(chan struct{}, size),
	}
}

func (sm *StepManager) CreateStep(ctx context.Context, step *TaskStep) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if step.ID == "" {
		step.ID = generateStepID()
	}

	if step.Status == "" {
		step.Status = TaskStatusPending
	}

	sm.steps[step.ID] = step

	if step.ParentTaskID != "" {
		sm.stepsByTask[step.ParentTaskID] = append(sm.stepsByTask[step.ParentTaskID], step)
		sort.Slice(sm.stepsByTask[step.ParentTaskID], func(i, j int) bool {
			return sm.stepsByTask[step.ParentTaskID][i].Order < sm.stepsByTask[step.ParentTaskID][j].Order
		})
	}

	if sm.storage != nil {
		if err := sm.saveStepToStorage(ctx, step); err != nil {
			return fmt.Errorf("failed to save step to storage: %w", err)
		}
	}

	if sm.eventBus != nil {
		event := TaskEvent{
			ID:        generateEventID(),
			TaskID:    step.ParentTaskID,
			Type:      TaskEventCreated,
			Timestamp: time.Now(),
			Data: map[string]any{
				"step_id":   step.ID,
				"step_name": step.Name,
				"order":     step.Order,
			},
		}
		sm.eventBus.Publish(ctx, event)
	}

	return nil
}

func (sm *StepManager) GetStep(ctx context.Context, stepID string) (*TaskStep, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	step, exists := sm.steps[stepID]
	if !exists {
		return nil, fmt.Errorf("step not found: %s", stepID)
	}

	return step, nil
}

func (sm *StepManager) UpdateStep(ctx context.Context, stepID string, updates StepUpdate) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	step, exists := sm.steps[stepID]
	if !exists {
		return fmt.Errorf("step not found: %s", stepID)
	}

	if updates.Status != nil {
		step.Status = *updates.Status
		if *updates.Status == TaskStatusInProgress && step.StartedAt == nil {
			now := time.Now()
			step.StartedAt = &now
		}
		if (*updates.Status == TaskStatusCompleted || *updates.Status == TaskStatusFailed) && step.CompletedAt == nil {
			now := time.Now()
			step.CompletedAt = &now
		}
	}

	if updates.Output != nil {
		step.Output = updates.Output
	}

	if updates.Error != nil {
		step.Error = updates.Error
	}

	if sm.storage != nil {
		if err := sm.saveStepToStorage(ctx, step); err != nil {
			return fmt.Errorf("failed to update step in storage: %w", err)
		}
	}

	if sm.eventBus != nil {
		event := TaskEvent{
			ID:        generateEventID(),
			TaskID:    step.ParentTaskID,
			Type:      TaskEventProgress,
			Timestamp: time.Now(),
			Data: map[string]any{
				"step_id": step.ID,
				"status":  step.Status,
				"updates": updates,
			},
		}
		sm.eventBus.Publish(ctx, event)
	}

	return nil
}

func (sm *StepManager) ExecuteStep(ctx context.Context, stepID string, executor StepExecutorFunc) error {
	step, err := sm.GetStep(ctx, stepID)
	if err != nil {
		return err
	}

	if step.Status != TaskStatusPending {
		return fmt.Errorf("step %s is not in pending status: %s", stepID, step.Status)
	}

	select {
	case sm.executorPool.workers <- struct{}{}:
		sm.executorPool.wg.Add(1)
		go sm.executeStepAsync(ctx, step, executor)
		return nil
	default:
		return fmt.Errorf("executor pool is full")
	}
}

func (sm *StepManager) executeStepAsync(ctx context.Context, step *TaskStep, executor StepExecutorFunc) {
	defer func() {
		<-sm.executorPool.workers
		sm.executorPool.wg.Done()
	}()

	stepCtx, cancel := context.WithTimeout(ctx, sm.config.StepTimeout)
	defer cancel()

	execution := &StepExecution{
		Step:      step,
		Context:   stepCtx,
		Cancel:    cancel,
		StartTime: time.Now(),
		Progress:  0.0,
	}

	sm.mu.Lock()
	sm.stepExecutions[step.ID] = execution
	sm.mu.Unlock()

	defer func() {
		sm.mu.Lock()
		delete(sm.stepExecutions, step.ID)
		sm.mu.Unlock()
	}()

	sm.UpdateStep(stepCtx, step.ID, StepUpdate{
		Status: &[]TaskStatus{TaskStatusInProgress}[0],
	})

	output, err := sm.executeWithRetry(stepCtx, step, executor, execution)

	if err != nil {
		sm.UpdateStep(stepCtx, step.ID, StepUpdate{
			Status: &[]TaskStatus{TaskStatusFailed}[0],
			Error: &StepError{
				Code:    "execution_failed",
				Message: err.Error(),
			},
		})
		return
	}

	sm.UpdateStep(stepCtx, step.ID, StepUpdate{
		Status: &[]TaskStatus{TaskStatusCompleted}[0],
		Output: output,
	})
}

func (sm *StepManager) executeWithRetry(ctx context.Context, step *TaskStep, executor StepExecutorFunc, execution *StepExecution) (*StepOutput, error) {
	var lastErr error

	for attempt := 0; attempt <= sm.config.RetryPolicy.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := sm.calculateBackoff(attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}

			execution.RetryCount = attempt
			if sm.eventBus != nil {
				event := TaskEvent{
					ID:        generateEventID(),
					TaskID:    step.ParentTaskID,
					Type:      TaskEventRetried,
					Timestamp: time.Now(),
					Data: map[string]any{
						"step_id": step.ID,
						"attempt": attempt,
					},
				}
				sm.eventBus.Publish(ctx, event)
			}
		}

		output, err := executor(ctx, step)
		if err == nil {
			return output, nil
		}

		lastErr = err

		if !sm.isRetryableError(err) {
			break
		}
	}

	return nil, lastErr
}

func (sm *StepManager) calculateBackoff(attempt int) time.Duration {
	backoff := float64(sm.config.RetryPolicy.InitialBackoff) * 
		pow(sm.config.RetryPolicy.BackoffFactor, float64(attempt-1))
	
	if backoff > float64(sm.config.RetryPolicy.MaxBackoff) {
		backoff = float64(sm.config.RetryPolicy.MaxBackoff)
	}

	return time.Duration(backoff)
}

func (sm *StepManager) isRetryableError(err error) bool {
	return true
}

func (sm *StepManager) GetStepsByTask(ctx context.Context, taskID string) ([]*TaskStep, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	steps, exists := sm.stepsByTask[taskID]
	if !exists {
		return []*TaskStep{}, nil
	}

	result := make([]*TaskStep, len(steps))
	copy(result, steps)
	return result, nil
}

func (sm *StepManager) GetStepProgress(ctx context.Context, stepID string) (*StepProgress, error) {
	sm.mu.RLock()
	execution, exists := sm.stepExecutions[stepID]
	sm.mu.RUnlock()

	if !exists {
		step, err := sm.GetStep(ctx, stepID)
		if err != nil {
			return nil, err
		}
		return &StepProgress{
			StepID:   stepID,
			Status:   step.Status,
			Progress: sm.getStaticProgress(step.Status),
		}, nil
	}

	return &StepProgress{
		StepID:           stepID,
		Status:           execution.Step.Status,
		Progress:         execution.Progress,
		StartTime:        execution.StartTime,
		ElapsedTime:      time.Since(execution.StartTime),
		EstimatedTimeRemaining: sm.estimateRemainingTime(execution),
	}, nil
}

func (sm *StepManager) getStaticProgress(status TaskStatus) float64 {
	switch status {
	case TaskStatusPending:
		return 0.0
	case TaskStatusInProgress:
		return 0.5
	case TaskStatusCompleted:
		return 1.0
	case TaskStatusFailed, TaskStatusCancelled:
		return 0.0
	default:
		return 0.0
	}
}

func (sm *StepManager) estimateRemainingTime(execution *StepExecution) *time.Duration {
	if execution.Progress <= 0 {
		return nil
	}

	elapsed := time.Since(execution.StartTime)
	totalEstimated := time.Duration(float64(elapsed) / execution.Progress)
	remaining := totalEstimated - elapsed

	if remaining < 0 {
		remaining = 0
	}

	return &remaining
}

func (sm *StepManager) CancelStep(ctx context.Context, stepID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	execution, exists := sm.stepExecutions[stepID]
	if exists {
		execution.Cancel()
	}

	return sm.UpdateStep(ctx, stepID, StepUpdate{
		Status: &[]TaskStatus{TaskStatusCancelled}[0],
	})
}

func (sm *StepManager) WaitForCompletion(ctx context.Context, stepIDs []string) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			allCompleted := true
			for _, stepID := range stepIDs {
				step, err := sm.GetStep(ctx, stepID)
				if err != nil {
					return err
				}
				if step.Status == TaskStatusPending || step.Status == TaskStatusInProgress {
					allCompleted = false
					break
				}
			}
			if allCompleted {
				return nil
			}
		}
	}
}

func (sm *StepManager) Shutdown(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, execution := range sm.stepExecutions {
		execution.Cancel()
	}

	done := make(chan struct{})
	go func() {
		sm.executorPool.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (sm *StepManager) saveStepToStorage(ctx context.Context, step *TaskStep) error {
	return nil
}

type StepExecutorFunc func(ctx context.Context, step *TaskStep) (*StepOutput, error)

type StepUpdate struct {
	Status *TaskStatus  `json:"status,omitempty"`
	Output *StepOutput  `json:"output,omitempty"`
	Error  *StepError   `json:"error,omitempty"`
}

type StepProgress struct {
	StepID                 string         `json:"step_id"`
	Status                 TaskStatus     `json:"status"`
	Progress               float64        `json:"progress"`
	StartTime              time.Time      `json:"start_time"`
	ElapsedTime            time.Duration  `json:"elapsed_time"`
	EstimatedTimeRemaining *time.Duration `json:"estimated_time_remaining,omitempty"`
}

func generateStepID() string {
	return fmt.Sprintf("step_%d", time.Now().UnixNano())
}

func generateEventID() string {
	return fmt.Sprintf("event_%d", time.Now().UnixNano())
}


func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}