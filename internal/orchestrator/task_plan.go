package orchestrator

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

type TaskPlanManager struct {
	mu       sync.RWMutex
	plans    map[string]*TaskPlan
	plansByTask map[string]*TaskPlan
	eventBus EventBus
	storage  Storage
	stepManager *StepManager
}

type PlanExecution struct {
	Plan      *TaskPlan         `json:"plan"`
	Context   context.Context   `json:"-"`
	Cancel    context.CancelFunc `json:"-"`
	StartTime time.Time         `json:"start_time"`
	EndTime   *time.Time        `json:"end_time,omitempty"`
	Status    TaskStatus        `json:"status"`
	Progress  *PlanProgress     `json:"progress"`
}

type PlanProgress struct {
	PlanID               string        `json:"plan_id"`
	TotalSteps           int           `json:"total_steps"`
	CompletedSteps       int           `json:"completed_steps"`
	FailedSteps          int           `json:"failed_steps"`
	InProgressSteps      int           `json:"in_progress_steps"`
	PercentComplete      float64       `json:"percent_complete"`
	EstimatedTimeRemaining *time.Duration `json:"estimated_time_remaining,omitempty"`
	CurrentStep          *string       `json:"current_step,omitempty"`
}

func NewTaskPlanManager(eventBus EventBus, storage Storage, stepManager *StepManager) *TaskPlanManager {
	return &TaskPlanManager{
		plans:       make(map[string]*TaskPlan),
		plansByTask: make(map[string]*TaskPlan),
		eventBus:    eventBus,
		storage:     storage,
		stepManager: stepManager,
	}
}

func (tpm *TaskPlanManager) CreatePlan(ctx context.Context, plan *TaskPlan) error {
	tpm.mu.Lock()
	defer tpm.mu.Unlock()

	if plan.ID == "" {
		plan.ID = generatePlanID()
	}

	plan.CreatedAt = time.Now()
	plan.UpdatedAt = time.Now()

	if err := tpm.validatePlan(plan); err != nil {
		return fmt.Errorf("invalid plan: %w", err)
	}

	tpm.plans[plan.ID] = plan
	tpm.plansByTask[plan.TaskID] = plan

	if tpm.storage != nil {
		if err := tpm.storage.SavePlan(ctx, plan); err != nil {
			return fmt.Errorf("failed to save plan: %w", err)
		}
	}

	if tpm.eventBus != nil {
		event := TaskEvent{
			ID:        generateEventID(),
			TaskID:    plan.TaskID,
			Type:      TaskEventCreated,
			Timestamp: time.Now(),
			Data: map[string]any{
				"plan_id":  plan.ID,
				"strategy": plan.Strategy,
				"steps":    len(plan.Steps),
			},
		}
		tpm.eventBus.Publish(ctx, event)
	}

	return nil
}

func (tpm *TaskPlanManager) GetPlan(ctx context.Context, planID string) (*TaskPlan, error) {
	tpm.mu.RLock()
	defer tpm.mu.RUnlock()

	plan, exists := tpm.plans[planID]
	if !exists {
		if tpm.storage != nil {
			loadedPlan, err := tpm.storage.LoadPlan(ctx, planID)
			if err != nil {
				return nil, fmt.Errorf("plan not found: %s", planID)
			}
			tpm.plans[planID] = loadedPlan
			return loadedPlan, nil
		}
		return nil, fmt.Errorf("plan not found: %s", planID)
	}

	return plan, nil
}

func (tpm *TaskPlanManager) GetPlanByTask(ctx context.Context, taskID string) (*TaskPlan, error) {
	tpm.mu.RLock()
	defer tpm.mu.RUnlock()

	plan, exists := tpm.plansByTask[taskID]
	if !exists {
		return nil, fmt.Errorf("no plan found for task: %s", taskID)
	}

	return plan, nil
}

func (tpm *TaskPlanManager) UpdatePlan(ctx context.Context, planID string, updates PlanUpdate) error {
	tpm.mu.Lock()
	defer tpm.mu.Unlock()

	plan, exists := tpm.plans[planID]
	if !exists {
		return fmt.Errorf("plan not found: %s", planID)
	}

	plan.UpdatedAt = time.Now()

	if updates.Strategy != nil {
		plan.Strategy = *updates.Strategy
	}

	if updates.Steps != nil {
		plan.Steps = updates.Steps
		sort.Slice(plan.Steps, func(i, j int) bool {
			return plan.Steps[i].Order < plan.Steps[j].Order
		})
	}

	if updates.EstimatedTime != nil {
		plan.EstimatedTime = time.Duration(*updates.EstimatedTime)
	}

	if updates.Dependencies != nil {
		plan.Dependencies = updates.Dependencies
	}

	if err := tpm.validatePlan(plan); err != nil {
		return fmt.Errorf("invalid plan update: %w", err)
	}

	if tpm.storage != nil {
		if err := tpm.storage.SavePlan(ctx, plan); err != nil {
			return fmt.Errorf("failed to update plan: %w", err)
		}
	}

	if tpm.eventBus != nil {
		event := TaskEvent{
			ID:        generateEventID(),
			TaskID:    plan.TaskID,
			Type:      TaskEventProgress,
			Timestamp: time.Now(),
			Data: map[string]any{
				"plan_id": plan.ID,
				"updates": updates,
			},
		}
		tpm.eventBus.Publish(ctx, event)
	}

	return nil
}

func (tpm *TaskPlanManager) ExecutePlan(ctx context.Context, planID string) error {
	plan, err := tpm.GetPlan(ctx, planID)
	if err != nil {
		return err
	}

	if len(plan.Steps) == 0 {
		return fmt.Errorf("plan has no steps to execute")
	}

	planCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	execution := &PlanExecution{
		Plan:      plan,
		Context:   planCtx,
		Cancel:    cancel,
		StartTime: time.Now(),
		Status:    TaskStatusInProgress,
	}

	if tpm.eventBus != nil {
		event := TaskEvent{
			ID:        generateEventID(),
			TaskID:    plan.TaskID,
			Type:      TaskEventStarted,
			Timestamp: time.Now(),
			Data: map[string]any{
				"plan_id": plan.ID,
			},
		}
		tpm.eventBus.Publish(ctx, event)
	}

	var executeErr error
	switch plan.Strategy {
	case PlanStrategySequential:
		executeErr = tpm.executeSequential(planCtx, plan)
	case PlanStrategyParallel:
		executeErr = tpm.executeParallel(planCtx, plan)
	case PlanStrategyHybrid:
		executeErr = tpm.executeHybrid(planCtx, plan)
	default:
		executeErr = fmt.Errorf("unknown plan strategy: %s", plan.Strategy)
	}

	now := time.Now()
	execution.EndTime = &now
	execution.Status = TaskStatusCompleted
	if executeErr != nil {
		execution.Status = TaskStatusFailed
	}

	plan.ActualTime = &[]time.Duration{time.Since(execution.StartTime)}[0]

	if tpm.storage != nil {
		tpm.storage.SavePlan(ctx, plan)
	}

	if tpm.eventBus != nil {
		eventType := TaskEventCompleted
		if executeErr != nil {
			eventType = TaskEventFailed
		}

		event := TaskEvent{
			ID:        generateEventID(),
			TaskID:    plan.TaskID,
			Type:      eventType,
			Timestamp: time.Now(),
			Data: map[string]any{
				"plan_id":    plan.ID,
				"duration":   time.Since(execution.StartTime),
				"error":      executeErr,
			},
		}
		tpm.eventBus.Publish(ctx, event)
	}

	return executeErr
}

func (tpm *TaskPlanManager) executeSequential(ctx context.Context, plan *TaskPlan) error {
	for _, step := range plan.Steps {
		if err := tpm.stepManager.CreateStep(ctx, &step); err != nil {
			return fmt.Errorf("failed to create step %s: %w", step.ID, err)
		}

		executor := tpm.createStepExecutor(step)
		if err := tpm.stepManager.ExecuteStep(ctx, step.ID, executor); err != nil {
			return fmt.Errorf("failed to execute step %s: %w", step.ID, err)
		}

		if err := tpm.stepManager.WaitForCompletion(ctx, []string{step.ID}); err != nil {
			return fmt.Errorf("step %s execution failed: %w", step.ID, err)
		}

		updatedStep, err := tpm.stepManager.GetStep(ctx, step.ID)
		if err != nil {
			return fmt.Errorf("failed to get step status: %w", err)
		}

		if updatedStep.Status == TaskStatusFailed {
			return fmt.Errorf("step %s failed", step.ID)
		}
	}

	return nil
}

func (tpm *TaskPlanManager) executeParallel(ctx context.Context, plan *TaskPlan) error {
	stepIDs := make([]string, len(plan.Steps))

	for i, step := range plan.Steps {
		if err := tpm.stepManager.CreateStep(ctx, &step); err != nil {
			return fmt.Errorf("failed to create step %s: %w", step.ID, err)
		}
		stepIDs[i] = step.ID
	}

	for _, step := range plan.Steps {
		executor := tpm.createStepExecutor(step)
		if err := tpm.stepManager.ExecuteStep(ctx, step.ID, executor); err != nil {
			return fmt.Errorf("failed to execute step %s: %w", step.ID, err)
		}
	}

	if err := tpm.stepManager.WaitForCompletion(ctx, stepIDs); err != nil {
		return fmt.Errorf("parallel execution failed: %w", err)
	}

	for _, stepID := range stepIDs {
		step, err := tpm.stepManager.GetStep(ctx, stepID)
		if err != nil {
			return fmt.Errorf("failed to get step status: %w", err)
		}

		if step.Status == TaskStatusFailed {
			return fmt.Errorf("step %s failed", stepID)
		}
	}

	return nil
}

func (tpm *TaskPlanManager) executeHybrid(ctx context.Context, plan *TaskPlan) error {
	dependencyGraph := tpm.buildDependencyGraph(plan.Steps)
	
	executed := make(map[string]bool)
	executing := make(map[string]bool)

	for len(executed) < len(plan.Steps) {
		readySteps := tpm.findReadySteps(plan.Steps, dependencyGraph, executed, executing)
		if len(readySteps) == 0 {
			return fmt.Errorf("no steps ready for execution - possible circular dependency")
		}

		stepIDs := make([]string, len(readySteps))
		for i, step := range readySteps {
			if err := tpm.stepManager.CreateStep(ctx, step); err != nil {
				return fmt.Errorf("failed to create step %s: %w", step.ID, err)
			}
			stepIDs[i] = step.ID
			executing[step.ID] = true
		}

		for _, step := range readySteps {
			executor := tpm.createStepExecutor(*step)
			if err := tpm.stepManager.ExecuteStep(ctx, step.ID, executor); err != nil {
				return fmt.Errorf("failed to execute step %s: %w", step.ID, err)
			}
		}

		if err := tpm.stepManager.WaitForCompletion(ctx, stepIDs); err != nil {
			return fmt.Errorf("hybrid execution batch failed: %w", err)
		}

		for _, stepID := range stepIDs {
			step, err := tpm.stepManager.GetStep(ctx, stepID)
			if err != nil {
				return fmt.Errorf("failed to get step status: %w", err)
			}

			if step.Status == TaskStatusFailed {
				return fmt.Errorf("step %s failed", stepID)
			}

			executed[stepID] = true
			delete(executing, stepID)
		}
	}

	return nil
}

func (tpm *TaskPlanManager) createStepExecutor(step TaskStep) StepExecutorFunc {
	return func(ctx context.Context, s *TaskStep) (*StepOutput, error) {
		time.Sleep(100 * time.Millisecond)
		
		return &StepOutput{
			Type:    "execution_result",
			Content: fmt.Sprintf("Step %s executed successfully", s.Name),
			Data: map[string]any{
				"step_id": s.ID,
				"name":    s.Name,
				"status":  "completed",
			},
		}, nil
	}
}

func (tpm *TaskPlanManager) buildDependencyGraph(steps []TaskStep) map[string][]string {
	graph := make(map[string][]string)
	for _, step := range steps {
		if step.Dependencies == nil {
			graph[step.ID] = []string{}
		} else {
			graph[step.ID] = step.Dependencies
		}
	}
	return graph
}

func (tpm *TaskPlanManager) findReadySteps(steps []TaskStep, graph map[string][]string, executed, executing map[string]bool) []*TaskStep {
	var ready []*TaskStep

	for i := range steps {
		step := &steps[i]
		if executed[step.ID] || executing[step.ID] {
			continue
		}

		allDepsReady := true
		if step.Dependencies != nil {
			for _, dep := range step.Dependencies {
				if !executed[dep] {
					allDepsReady = false
					break
				}
			}
		}

		if allDepsReady {
			ready = append(ready, step)
		}
	}

	return ready
}

func (tpm *TaskPlanManager) GetPlanProgress(ctx context.Context, planID string) (*PlanProgress, error) {
	plan, err := tpm.GetPlan(ctx, planID)
	if err != nil {
		return nil, err
	}

	progress := &PlanProgress{
		PlanID:     planID,
		TotalSteps: len(plan.Steps),
	}

	for _, step := range plan.Steps {
		switch step.Status {
		case TaskStatusCompleted:
			progress.CompletedSteps++
		case TaskStatusFailed:
			progress.FailedSteps++
		case TaskStatusInProgress:
			progress.InProgressSteps++
			if progress.CurrentStep == nil {
				progress.CurrentStep = &step.Name
			}
		}
	}

	if progress.TotalSteps > 0 {
		progress.PercentComplete = float64(progress.CompletedSteps) / float64(progress.TotalSteps) * 100
	}

	if progress.InProgressSteps > 0 && plan.EstimatedTime > 0 {
		elapsed := time.Since(plan.CreatedAt)
		if progress.PercentComplete > 0 {
			totalEstimated := time.Duration(float64(elapsed) / (progress.PercentComplete / 100))
			remaining := totalEstimated - elapsed
			if remaining > 0 {
				progress.EstimatedTimeRemaining = &remaining
			}
		}
	}

	return progress, nil
}

func (tpm *TaskPlanManager) validatePlan(plan *TaskPlan) error {
	if plan.TaskID == "" {
		return fmt.Errorf("plan must have a task ID")
	}

	if len(plan.Steps) == 0 {
		return fmt.Errorf("plan must have at least one step")
	}

	stepIDs := make(map[string]bool)
	for _, step := range plan.Steps {
		if step.ID == "" {
			return fmt.Errorf("all steps must have an ID")
		}
		if stepIDs[step.ID] {
			return fmt.Errorf("duplicate step ID: %s", step.ID)
		}
		stepIDs[step.ID] = true
	}

	for _, step := range plan.Steps {
		if step.Dependencies != nil {
			for _, dep := range step.Dependencies {
				if !stepIDs[dep] {
					return fmt.Errorf("step %s depends on non-existent step %s", step.ID, dep)
				}
			}
		}
	}

	if tpm.hasCyclicDependencies(plan.Steps) {
		return fmt.Errorf("plan has cyclic dependencies")
	}

	return nil
}

func (tpm *TaskPlanManager) hasCyclicDependencies(steps []TaskStep) bool {
	graph := make(map[string][]string)
	for _, step := range steps {
		graph[step.ID] = step.Dependencies
	}

	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for _, step := range steps {
		if !visited[step.ID] {
			if tpm.hasCyclicDependenciesUtil(step.ID, graph, visited, recStack) {
				return true
			}
		}
	}

	return false
}

func (tpm *TaskPlanManager) hasCyclicDependenciesUtil(stepID string, graph map[string][]string, visited, recStack map[string]bool) bool {
	visited[stepID] = true
	recStack[stepID] = true

	deps, exists := graph[stepID]
	if exists && deps != nil {
		for _, dep := range deps {
			if !visited[dep] {
				if tpm.hasCyclicDependenciesUtil(dep, graph, visited, recStack) {
					return true
				}
			} else if recStack[dep] {
				return true
			}
		}
	}

	recStack[stepID] = false
	return false
}

func (tpm *TaskPlanManager) DeletePlan(ctx context.Context, planID string) error {
	tpm.mu.Lock()
	defer tpm.mu.Unlock()

	plan, exists := tpm.plans[planID]
	if !exists {
		return fmt.Errorf("plan not found: %s", planID)
	}

	delete(tpm.plans, planID)
	delete(tpm.plansByTask, plan.TaskID)

	if tpm.storage != nil {
		if err := tpm.storage.DeletePlan(ctx, planID); err != nil {
			return fmt.Errorf("failed to delete plan from storage: %w", err)
		}
	}

	return nil
}

func generatePlanID() string {
	return fmt.Sprintf("plan_%d", time.Now().UnixNano())
}

