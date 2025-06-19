package orchestrator

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// AdaptivePlanner coordinates step evaluation and plan adjustment
type AdaptivePlanner struct {
	stepEvaluator *StepEvaluator
	planAdjuster  *PlanAdjuster
	currentPlan   *Plan
	executionLog  []*ExecutionEntry
	feedbackLoop  *FeedbackLoop
	mutex         sync.RWMutex
	config        *PlannerConfig
}

// PlannerConfig contains configuration for the adaptive planner
type PlannerConfig struct {
	MaxConcurrentSteps   int
	EvaluationInterval   time.Duration
	AdjustmentThreshold  float64
	LearningRate         float64
	MaxPlanRevisions     int
	EnableFeedbackLoop   bool
	EnableLearning       bool
	ConservativeMode     bool
}

// ExecutionEntry logs step execution details
type ExecutionEntry struct {
	Timestamp     time.Time
	StepID        string
	Action        ExecutionAction
	PreviousState StepStatus
	NewState      StepStatus
	Result        *StepResult
	Adjustments   []string
	Metrics       map[string]float64
}

// ExecutionAction represents what action was taken
type ExecutionAction int

const (
	ActionStarted ExecutionAction = iota
	ActionCompleted
	ActionFailed
	ActionRetried
	ActionSkipped
	ActionAdjusted
)

func (ea ExecutionAction) String() string {
	switch ea {
	case ActionStarted:
		return "started"
	case ActionCompleted:
		return "completed"
	case ActionFailed:
		return "failed"
	case ActionRetried:
		return "retried"
	case ActionSkipped:
		return "skipped"
	case ActionAdjusted:
		return "adjusted"
	default:
		return "unknown"
	}
}

// FeedbackLoop handles continuous learning and adaptation
type FeedbackLoop struct {
	patternHistory    []ExecutionPattern
	successPatterns   map[string]float64
	failurePatterns   map[string]float64
	adjustmentImpact  map[string]float64
	learningEnabled   bool
	adaptationStrength float64
}

// ExecutionPattern represents a pattern in execution history
type ExecutionPattern struct {
	StepType      StepType
	Dependencies  []string
	Duration      time.Duration
	Quality       StepQuality
	SuccessRate   float64
	CommonIssues  []string
	BestPractices []string
}

// NewAdaptivePlanner creates a new adaptive planner
func NewAdaptivePlanner(config *PlannerConfig) *AdaptivePlanner {
	if config == nil {
		config = DefaultPlannerConfig()
	}
	
	strategy := StrategyBalanced
	if config.ConservativeMode {
		strategy = StrategyConservative
	}
	
	return &AdaptivePlanner{
		stepEvaluator: NewStepEvaluator(),
		planAdjuster:  NewPlanAdjuster(strategy),
		executionLog:  make([]*ExecutionEntry, 0),
		feedbackLoop:  NewFeedbackLoop(config.EnableLearning),
		config:        config,
	}
}

// DefaultPlannerConfig returns default configuration
func DefaultPlannerConfig() *PlannerConfig {
	return &PlannerConfig{
		MaxConcurrentSteps:  3,
		EvaluationInterval:  30 * time.Second,
		AdjustmentThreshold: 0.5,
		LearningRate:        0.1,
		MaxPlanRevisions:    10,
		EnableFeedbackLoop:  true,
		EnableLearning:      true,
		ConservativeMode:    false,
	}
}

// NewFeedbackLoop creates a new feedback loop
func NewFeedbackLoop(learningEnabled bool) *FeedbackLoop {
	return &FeedbackLoop{
		patternHistory:     make([]ExecutionPattern, 0),
		successPatterns:    make(map[string]float64),
		failurePatterns:    make(map[string]float64),
		adjustmentImpact:   make(map[string]float64),
		learningEnabled:    learningEnabled,
		adaptationStrength: 0.1,
	}
}

// SetPlan sets the current plan for execution
func (ap *AdaptivePlanner) SetPlan(plan *Plan) {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()
	
	ap.currentPlan = plan
	ap.logExecution("", ActionStarted, StepStatusPending, StepStatusPending, nil, []string{"plan_set"})
}

// ExecuteStep executes a single step and evaluates the result
func (ap *AdaptivePlanner) ExecuteStep(stepID string, output string, startTime, endTime time.Time) (*StepResult, error) {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()
	
	if ap.currentPlan == nil {
		return nil, fmt.Errorf("no plan set")
	}
	
	step := ap.findStep(stepID)
	if step == nil {
		return nil, fmt.Errorf("step %s not found in current plan", stepID)
	}
	
	// Evaluate the step
	result := ap.stepEvaluator.EvaluateStep(stepID, output, startTime, endTime)
	step.Result = result
	step.ActualTime = result.ExecutionTime
	step.UpdatedAt = time.Now()
	
	// Log execution
	previousStatus := step.Status
	step.Status = result.Status
	ap.logExecution(stepID, ActionCompleted, previousStatus, result.Status, result, []string{})
	
	// Update feedback loop
	if ap.config.EnableFeedbackLoop {
		ap.feedbackLoop.UpdatePattern(step, result)
	}
	
	// Check if plan adjustment is needed
	if ap.shouldAdjustPlan(step, result) {
		adjustedPlan, err := ap.planAdjuster.AdjustPlan(ap.currentPlan, step, result)
		if err != nil {
			ap.logExecution(stepID, ActionFailed, step.Status, step.Status, result, 
				[]string{"adjustment_failed: " + err.Error()})
		} else if adjustedPlan != ap.currentPlan {
			ap.currentPlan = adjustedPlan
			ap.logExecution(stepID, ActionAdjusted, step.Status, step.Status, result, 
				[]string{"plan_adjusted"})
		}
	}
	
	return result, nil
}

// GetNextSteps returns the next steps ready for execution
func (ap *AdaptivePlanner) GetNextSteps(maxSteps int) ([]*Step, error) {
	ap.mutex.RLock()
	defer ap.mutex.RUnlock()
	
	if ap.currentPlan == nil {
		return nil, fmt.Errorf("no plan set")
	}
	
	if maxSteps <= 0 {
		maxSteps = ap.config.MaxConcurrentSteps
	}
	
	availableSteps := ap.getAvailableSteps()
	
	// Apply learning-based prioritization
	if ap.config.EnableLearning {
		ap.optimizeStepOrder(availableSteps)
	}
	
	// Return up to maxSteps
	if len(availableSteps) > maxSteps {
		return availableSteps[:maxSteps], nil
	}
	
	return availableSteps, nil
}

// GetPlanStatus returns the current plan status and progress
func (ap *AdaptivePlanner) GetPlanStatus() *ExecutionStatus {
	ap.mutex.RLock()
	defer ap.mutex.RUnlock()
	
	if ap.currentPlan == nil {
		return &ExecutionStatus{
			PlanID:           "",
			Status:           PlanStatusDraft,
			TotalSteps:       0,
			CompletedSteps:   0,
			FailedSteps:      0,
			Progress:         0.0,
			EstimatedCompletion: time.Time{},
		}
	}
	
	return ap.calculatePlanStatus()
}

// ExecutionStatus represents the current status of plan execution
type ExecutionStatus struct {
	PlanID              string
	Status              PlanStatus
	TotalSteps          int
	CompletedSteps      int
	FailedSteps         int
	BlockedSteps        int
	Progress            float64
	EstimatedCompletion time.Time
	AverageStepTime     time.Duration
	QualityScore        float64
	EfficiencyScore     float64
	Adjustments         int
	LastAdjustment      time.Time
}

// shouldAdjustPlan determines if the plan should be adjusted
func (ap *AdaptivePlanner) shouldAdjustPlan(step *Step, result *StepResult) bool {
	// Always adjust for failed steps
	if result.Status == StepStatusFailed {
		return true
	}
	
	// Adjust for poor quality
	if result.Quality == QualityPoor || result.Quality == QualityUnacceptable {
		return true
	}
	
	// Adjust for blocked steps
	if result.Status == StepStatusBlocked {
		return true
	}
	
	// Adjust based on efficiency threshold
	if result.EfficiencyScore < ap.config.AdjustmentThreshold {
		return true
	}
	
	// Adjust if there are next actions suggested
	if len(result.NextActions) > 0 {
		return true
	}
	
	return false
}

// getAvailableSteps returns steps that are ready for execution
func (ap *AdaptivePlanner) getAvailableSteps() []*Step {
	available := make([]*Step, 0)
	
	for _, step := range ap.currentPlan.Steps {
		if step.Status == StepStatusPending && ap.areDependenciesMet(step) {
			available = append(available, step)
		}
	}
	
	// Sort by priority
	sort.Slice(available, func(i, j int) bool {
		return available[i].Priority < available[j].Priority
	})
	
	return available
}

// areDependenciesMet checks if all dependencies for a step are completed
func (ap *AdaptivePlanner) areDependenciesMet(step *Step) bool {
	for _, depID := range step.Dependencies {
		depStep := ap.findStep(depID)
		if depStep == nil || depStep.Status != StepStatusCompleted {
			return false
		}
	}
	return true
}

// optimizeStepOrder optimizes step execution order based on learning
func (ap *AdaptivePlanner) optimizeStepOrder(steps []*Step) {
	if !ap.feedbackLoop.learningEnabled {
		return
	}
	
	// Apply learning-based scoring
	for _, step := range steps {
		patternKey := ap.getPatternKey(step)
		
		successRate, hasSuccess := ap.feedbackLoop.successPatterns[patternKey]
		failureRate, hasFailure := ap.feedbackLoop.failurePatterns[patternKey]
		
		if hasSuccess && hasFailure {
			// Adjust priority based on historical success rate
			adjustment := int((successRate - failureRate) * 10)
			step.Priority -= adjustment // Lower priority number = higher priority
		}
	}
	
	// Re-sort with adjusted priorities
	sort.Slice(steps, func(i, j int) bool {
		return steps[i].Priority < steps[j].Priority
	})
}

// calculatePlanStatus calculates current plan execution status
func (ap *AdaptivePlanner) calculatePlanStatus() *ExecutionStatus {
	status := &ExecutionStatus{
		PlanID:     ap.currentPlan.ID,
		Status:     ap.currentPlan.Status,
		TotalSteps: len(ap.currentPlan.Steps),
	}
	
	var totalTime time.Duration
	var totalQuality float64
	var totalEfficiency float64
	completedCount := 0
	
	for _, step := range ap.currentPlan.Steps {
		switch step.Status {
		case StepStatusCompleted:
			status.CompletedSteps++
			completedCount++
			if step.Result != nil {
				totalQuality += float64(step.Result.Quality)
				totalEfficiency += step.Result.EfficiencyScore
			}
			totalTime += step.ActualTime
		case StepStatusFailed:
			status.FailedSteps++
		case StepStatusBlocked:
			status.BlockedSteps++
		}
	}
	
	if status.TotalSteps > 0 {
		status.Progress = float64(status.CompletedSteps) / float64(status.TotalSteps)
	}
	
	if completedCount > 0 {
		status.AverageStepTime = totalTime / time.Duration(completedCount)
		status.QualityScore = totalQuality / float64(completedCount)
		status.EfficiencyScore = totalEfficiency / float64(completedCount)
		
		// Estimate completion time
		remaining := status.TotalSteps - status.CompletedSteps
		if remaining > 0 {
			estimatedRemaining := time.Duration(remaining) * status.AverageStepTime
			status.EstimatedCompletion = time.Now().Add(estimatedRemaining)
		}
	}
	
	// Count adjustments
	status.Adjustments = len(ap.planAdjuster.GetAdjustmentHistory(0))
	if status.Adjustments > 0 {
		history := ap.planAdjuster.GetAdjustmentHistory(1)
		status.LastAdjustment = history[0].Timestamp
	}
	
	return status
}

// Helper methods

// findStep finds a step by ID in the current plan
func (ap *AdaptivePlanner) findStep(stepID string) *Step {
	if ap.currentPlan == nil {
		return nil
	}
	
	for _, step := range ap.currentPlan.Steps {
		if step.ID == stepID {
			return step
		}
	}
	return nil
}

// logExecution logs an execution event
func (ap *AdaptivePlanner) logExecution(stepID string, action ExecutionAction, 
	previousState, newState StepStatus, result *StepResult, adjustments []string) {
	
	entry := &ExecutionEntry{
		Timestamp:     time.Now(),
		StepID:        stepID,
		Action:        action,
		PreviousState: previousState,
		NewState:      newState,
		Result:        result,
		Adjustments:   adjustments,
		Metrics:       make(map[string]float64),
	}
	
	if result != nil {
		entry.Metrics["completion_rate"] = result.CompletionRate
		entry.Metrics["efficiency_score"] = result.EfficiencyScore
		entry.Metrics["quality_score"] = float64(result.Quality)
	}
	
	ap.executionLog = append(ap.executionLog, entry)
	
	// Keep log size manageable
	if len(ap.executionLog) > 1000 {
		ap.executionLog = ap.executionLog[100:]
	}
}

// getPatternKey generates a pattern key for learning
func (ap *AdaptivePlanner) getPatternKey(step *Step) string {
	return fmt.Sprintf("%s_%d_deps", step.Type.String(), len(step.Dependencies))
}

// UpdatePattern updates execution patterns for learning
func (fl *FeedbackLoop) UpdatePattern(step *Step, result *StepResult) {
	if !fl.learningEnabled {
		return
	}
	
	patternKey := fmt.Sprintf("%s_%d_deps", step.Type.String(), len(step.Dependencies))
	
	// Update success/failure patterns
	if result.Status == StepStatusCompleted && result.Quality >= QualityAcceptable {
		fl.successPatterns[patternKey] += fl.adaptationStrength
		if fl.successPatterns[patternKey] > 1.0 {
			fl.successPatterns[patternKey] = 1.0
		}
	} else if result.Status == StepStatusFailed || result.Quality <= QualityPoor {
		fl.failurePatterns[patternKey] += fl.adaptationStrength
		if fl.failurePatterns[patternKey] > 1.0 {
			fl.failurePatterns[patternKey] = 1.0
		}
	}
	
	// Decay older patterns
	for key := range fl.successPatterns {
		fl.successPatterns[key] *= 0.99
	}
	for key := range fl.failurePatterns {
		fl.failurePatterns[key] *= 0.99
	}
}

// GetExecutionLog returns recent execution log entries
func (ap *AdaptivePlanner) GetExecutionLog(limit int) []*ExecutionEntry {
	ap.mutex.RLock()
	defer ap.mutex.RUnlock()
	
	if limit <= 0 || limit > len(ap.executionLog) {
		limit = len(ap.executionLog)
	}
	
	return ap.executionLog[len(ap.executionLog)-limit:]
}

// GetLearningInsights returns insights from the feedback loop
func (ap *AdaptivePlanner) GetLearningInsights() map[string]interface{} {
	ap.mutex.RLock()
	defer ap.mutex.RUnlock()
	
	insights := make(map[string]interface{})
	
	if ap.feedbackLoop.learningEnabled {
		insights["success_patterns"] = ap.feedbackLoop.successPatterns
		insights["failure_patterns"] = ap.feedbackLoop.failurePatterns
		insights["adaptation_strength"] = ap.feedbackLoop.adaptationStrength
		insights["pattern_count"] = len(ap.feedbackLoop.patternHistory)
	}
	
	insights["total_adjustments"] = len(ap.planAdjuster.GetAdjustmentHistory(0))
	insights["execution_log_size"] = len(ap.executionLog)
	
	return insights
}