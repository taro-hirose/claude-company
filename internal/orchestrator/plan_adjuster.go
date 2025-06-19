package orchestrator

import (
	"fmt"
	"sort"
	"time"
)

// Plan represents a complete execution plan
type Plan struct {
	ID          string
	Name        string
	Description string
	Steps       []*Step
	Dependencies map[string][]string // stepID -> dependent stepIDs
	Priority    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Status      PlanStatus
	Metadata    map[string]interface{}
}

// Step represents a single execution step
type Step struct {
	ID               string
	Name             string
	Description      string
	Type             StepType
	Status           StepStatus
	Priority         int
	EstimatedTime    time.Duration
	ActualTime       time.Duration
	Dependencies     []string
	Resources        []string
	Deliverables     []string
	CompletionCriteria []string
	RetryCount       int
	MaxRetries       int
	AssignedPane     string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Result           *StepResult
	Metadata         map[string]interface{}
}

// StepType represents the type of step
type StepType int

const (
	StepTypeResearch StepType = iota
	StepTypeImplementation
	StepTypeTesting
	StepTypeDocumentation
	StepTypeReview
	StepTypeDeployment
	StepTypeCustom
)

func (st StepType) String() string {
	switch st {
	case StepTypeResearch:
		return "research"
	case StepTypeImplementation:
		return "implementation"
	case StepTypeTesting:
		return "testing"
	case StepTypeDocumentation:
		return "documentation"
	case StepTypeReview:
		return "review"
	case StepTypeDeployment:
		return "deployment"
	case StepTypeCustom:
		return "custom"
	default:
		return "unknown"
	}
}

// PlanStatus represents the status of a plan
type PlanStatus int

const (
	PlanStatusDraft PlanStatus = iota
	PlanStatusActive
	PlanStatusPaused
	PlanStatusCompleted
	PlanStatusFailed
	PlanStatusCancelled
)

func (ps PlanStatus) String() string {
	switch ps {
	case PlanStatusDraft:
		return "draft"
	case PlanStatusActive:
		return "active"
	case PlanStatusPaused:
		return "paused"
	case PlanStatusCompleted:
		return "completed"
	case PlanStatusFailed:
		return "failed"
	case PlanStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// AdjustmentStrategy represents different strategies for plan adjustment
type AdjustmentStrategy int

const (
	StrategyConservative AdjustmentStrategy = iota
	StrategyBalanced
	StrategyAggressive
	StrategyAdaptive
)

// PlanAdjuster handles dynamic plan adjustments based on step evaluations
type PlanAdjuster struct {
	strategy        AdjustmentStrategy
	adjustmentRules []AdjustmentRule
	historySize     int
	adjustmentHistory []*AdjustmentRecord
}

// AdjustmentRule defines rules for plan adjustments
type AdjustmentRule struct {
	Name        string
	Condition   AdjustmentCondition
	Action      AdjustmentAction
	Priority    int
	Weight      float64
	Description string
}

// AdjustmentCondition checks if adjustment should be applied
type AdjustmentCondition func(step *Step, result *StepResult, plan *Plan) bool

// AdjustmentAction performs the adjustment
type AdjustmentAction func(step *Step, result *StepResult, plan *Plan) (*Plan, error)

// AdjustmentRecord tracks adjustment history
type AdjustmentRecord struct {
	Timestamp   time.Time
	StepID      string
	RuleName    string
	Action      string
	Reason      string
	OldPlan     string
	NewPlan     string
	Success     bool
	Impact      float64
}

// NewPlanAdjuster creates a new plan adjuster
func NewPlanAdjuster(strategy AdjustmentStrategy) *PlanAdjuster {
	adjuster := &PlanAdjuster{
		strategy:          strategy,
		adjustmentRules:   make([]AdjustmentRule, 0),
		historySize:       100,
		adjustmentHistory: make([]*AdjustmentRecord, 0),
	}
	
	adjuster.initializeDefaultRules()
	
	return adjuster
}

// initializeDefaultRules sets up default adjustment rules
func (pa *PlanAdjuster) initializeDefaultRules() {
	pa.adjustmentRules = []AdjustmentRule{
		{
			Name:     "failed_step_retry",
			Priority: 1,
			Weight:   1.0,
			Condition: func(step *Step, result *StepResult, plan *Plan) bool {
				return result.Status == StepStatusFailed && step.RetryCount < step.MaxRetries
			},
			Action: pa.retryFailedStep,
			Description: "Retry failed step if retries available",
		},
		{
			Name:     "low_quality_rework",
			Priority: 2,
			Weight:   0.8,
			Condition: func(step *Step, result *StepResult, plan *Plan) bool {
				return result.Quality == QualityPoor || result.Quality == QualityUnacceptable
			},
			Action: pa.reworkLowQualityStep,
			Description: "Rework step with poor quality",
		},
		{
			Name:     "blocked_step_reorder",
			Priority: 3,
			Weight:   0.7,
			Condition: func(step *Step, result *StepResult, plan *Plan) bool {
				return result.Status == StepStatusBlocked
			},
			Action: pa.reorderBlockedStep,
			Description: "Reorder blocked step to later",
		},
		{
			Name:     "efficient_step_parallelize",
			Priority: 4,
			Weight:   0.6,
			Condition: func(step *Step, result *StepResult, plan *Plan) bool {
				return result.Status == StepStatusCompleted && result.EfficiencyScore > 0.8
			},
			Action: pa.parallelizeEfficentSteps,
			Description: "Parallelize efficiently completed steps",
		},
		{
			Name:     "slow_step_optimize",
			Priority: 5,
			Weight:   0.5,
			Condition: func(step *Step, result *StepResult, plan *Plan) bool {
				return result.ExecutionTime > step.EstimatedTime*2
			},
			Action: pa.optimizeSlowStep,
			Description: "Optimize steps taking too long",
		},
		{
			Name:     "add_missing_dependencies",
			Priority: 6,
			Weight:   0.4,
			Condition: func(step *Step, result *StepResult, plan *Plan) bool {
				return len(result.NextActions) > 0
			},
			Action: pa.addMissingDependencies,
			Description: "Add missing dependencies based on feedback",
		},
	}
	
	// Sort rules by priority
	sort.Slice(pa.adjustmentRules, func(i, j int) bool {
		return pa.adjustmentRules[i].Priority < pa.adjustmentRules[j].Priority
	})
}

// AdjustPlan adjusts the plan based on step evaluation results
func (pa *PlanAdjuster) AdjustPlan(plan *Plan, step *Step, result *StepResult) (*Plan, error) {
	adjustedPlan := pa.clonePlan(plan)
	adjustmentsMade := false
	
	for _, rule := range pa.adjustmentRules {
		if rule.Condition(step, result, adjustedPlan) {
			newPlan, err := rule.Action(step, result, adjustedPlan)
			if err != nil {
				pa.recordAdjustment(step.ID, rule.Name, "failed", err.Error(), false, 0.0)
				continue
			}
			
			if newPlan != nil {
				impact := pa.calculateImpact(adjustedPlan, newPlan)
				pa.recordAdjustment(step.ID, rule.Name, "applied", 
					fmt.Sprintf("Rule applied successfully"), true, impact)
				adjustedPlan = newPlan
				adjustmentsMade = true
				
				// Apply only one rule per adjustment cycle for conservative approach
				if pa.strategy == StrategyConservative {
					break
				}
			}
		}
	}
	
	if adjustmentsMade {
		adjustedPlan.UpdatedAt = time.Now()
		return adjustedPlan, nil
	}
	
	return plan, nil
}

// retryFailedStep retries a failed step
func (pa *PlanAdjuster) retryFailedStep(step *Step, result *StepResult, plan *Plan) (*Plan, error) {
	stepToUpdate := pa.findStepInPlan(plan, step.ID)
	if stepToUpdate == nil {
		return nil, fmt.Errorf("step %s not found in plan", step.ID)
	}
	
	stepToUpdate.RetryCount++
	stepToUpdate.Status = StepStatusPending
	stepToUpdate.Result = nil
	stepToUpdate.UpdatedAt = time.Now()
	
	return plan, nil
}

// reworkLowQualityStep reworks a step with poor quality
func (pa *PlanAdjuster) reworkLowQualityStep(step *Step, result *StepResult, plan *Plan) (*Plan, error) {
	stepToUpdate := pa.findStepInPlan(plan, step.ID)
	if stepToUpdate == nil {
		return nil, fmt.Errorf("step %s not found in plan", step.ID)
	}
	
	// Create a rework step
	reworkStep := &Step{
		ID:               step.ID + "_rework",
		Name:             "Rework: " + step.Name,
		Description:      "Reworking step due to quality issues: " + result.Feedback,
		Type:             step.Type,
		Status:           StepStatusPending,
		Priority:         step.Priority + 1,
		EstimatedTime:    step.EstimatedTime,
		Dependencies:     []string{step.ID},
		Resources:        step.Resources,
		Deliverables:     step.Deliverables,
		CompletionCriteria: step.CompletionCriteria,
		MaxRetries:       step.MaxRetries,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Metadata:         map[string]interface{}{"original_step": step.ID, "rework_reason": "quality"},
	}
	
	plan.Steps = append(plan.Steps, reworkStep)
	
	return plan, nil
}

// reorderBlockedStep moves a blocked step to later in the execution order
func (pa *PlanAdjuster) reorderBlockedStep(step *Step, result *StepResult, plan *Plan) (*Plan, error) {
	stepToUpdate := pa.findStepInPlan(plan, step.ID)
	if stepToUpdate == nil {
		return nil, fmt.Errorf("step %s not found in plan", step.ID)
	}
	
	// Lower priority to move it later
	stepToUpdate.Priority += 10
	stepToUpdate.Status = StepStatusPending
	stepToUpdate.UpdatedAt = time.Now()
	
	// Sort steps by priority
	sort.Slice(plan.Steps, func(i, j int) bool {
		return plan.Steps[i].Priority < plan.Steps[j].Priority
	})
	
	return plan, nil
}

// parallelizeEfficentSteps identifies steps that can be parallelized
func (pa *PlanAdjuster) parallelizeEfficentSteps(step *Step, result *StepResult, plan *Plan) (*Plan, error) {
	// Find similar steps that can be parallelized
	for _, planStep := range plan.Steps {
		if planStep.Status == StepStatusPending && 
		   planStep.Type == step.Type && 
		   !pa.hasDependency(plan, planStep.ID, step.ID) {
			// Mark for parallel execution
			if planStep.Metadata == nil {
				planStep.Metadata = make(map[string]interface{})
			}
			planStep.Metadata["parallel_group"] = step.ID + "_group"
		}
	}
	
	return plan, nil
}

// optimizeSlowStep creates optimization suggestions for slow steps
func (pa *PlanAdjuster) optimizeSlowStep(step *Step, result *StepResult, plan *Plan) (*Plan, error) {
	stepToUpdate := pa.findStepInPlan(plan, step.ID)
	if stepToUpdate == nil {
		return nil, fmt.Errorf("step %s not found in plan", step.ID)
	}
	
	// Update estimated time based on actual performance
	stepToUpdate.EstimatedTime = result.ExecutionTime
	
	// Add optimization step if needed
	if result.ExecutionTime > step.EstimatedTime*3 {
		optimizationStep := &Step{
			ID:          step.ID + "_optimize",
			Name:        "Optimize: " + step.Name,
			Description: "Optimization analysis for slow step",
			Type:        StepTypeReview,
			Status:      StepStatusPending,
			Priority:    step.Priority - 1,
			EstimatedTime: 15 * time.Minute,
			Dependencies: []string{},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Metadata:    map[string]interface{}{"optimization_target": step.ID},
		}
		
		plan.Steps = append(plan.Steps, optimizationStep)
	}
	
	return plan, nil
}

// addMissingDependencies adds steps based on feedback
func (pa *PlanAdjuster) addMissingDependencies(step *Step, result *StepResult, plan *Plan) (*Plan, error) {
	for i, action := range result.NextActions {
		newStep := &Step{
			ID:          fmt.Sprintf("%s_followup_%d", step.ID, i+1),
			Name:        action,
			Description: fmt.Sprintf("Follow-up action from %s", step.Name),
			Type:        StepTypeCustom,
			Status:      StepStatusPending,
			Priority:    step.Priority + i + 1,
			EstimatedTime: 30 * time.Minute,
			Dependencies: []string{step.ID},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Metadata:    map[string]interface{}{"generated_from": step.ID},
		}
		
		plan.Steps = append(plan.Steps, newStep)
	}
	
	return plan, nil
}

// Helper methods

// clonePlan creates a deep copy of a plan
func (pa *PlanAdjuster) clonePlan(plan *Plan) *Plan {
	newPlan := &Plan{
		ID:          plan.ID,
		Name:        plan.Name,
		Description: plan.Description,
		Priority:    plan.Priority,
		CreatedAt:   plan.CreatedAt,
		UpdatedAt:   plan.UpdatedAt,
		Status:      plan.Status,
		Steps:       make([]*Step, len(plan.Steps)),
		Dependencies: make(map[string][]string),
		Metadata:    make(map[string]interface{}),
	}
	
	// Clone steps
	for i, step := range plan.Steps {
		newPlan.Steps[i] = pa.cloneStep(step)
	}
	
	// Clone dependencies
	for k, v := range plan.Dependencies {
		newPlan.Dependencies[k] = make([]string, len(v))
		copy(newPlan.Dependencies[k], v)
	}
	
	// Clone metadata
	for k, v := range plan.Metadata {
		newPlan.Metadata[k] = v
	}
	
	return newPlan
}

// cloneStep creates a deep copy of a step
func (pa *PlanAdjuster) cloneStep(step *Step) *Step {
	newStep := &Step{
		ID:               step.ID,
		Name:             step.Name,
		Description:      step.Description,
		Type:             step.Type,
		Status:           step.Status,
		Priority:         step.Priority,
		EstimatedTime:    step.EstimatedTime,
		ActualTime:       step.ActualTime,
		RetryCount:       step.RetryCount,
		MaxRetries:       step.MaxRetries,
		AssignedPane:     step.AssignedPane,
		CreatedAt:        step.CreatedAt,
		UpdatedAt:        step.UpdatedAt,
		Result:           step.Result,
		Dependencies:     make([]string, len(step.Dependencies)),
		Resources:        make([]string, len(step.Resources)),
		Deliverables:     make([]string, len(step.Deliverables)),
		CompletionCriteria: make([]string, len(step.CompletionCriteria)),
		Metadata:         make(map[string]interface{}),
	}
	
	copy(newStep.Dependencies, step.Dependencies)
	copy(newStep.Resources, step.Resources)
	copy(newStep.Deliverables, step.Deliverables)
	copy(newStep.CompletionCriteria, step.CompletionCriteria)
	
	for k, v := range step.Metadata {
		newStep.Metadata[k] = v
	}
	
	return newStep
}

// findStepInPlan finds a step by ID in the plan
func (pa *PlanAdjuster) findStepInPlan(plan *Plan, stepID string) *Step {
	for _, step := range plan.Steps {
		if step.ID == stepID {
			return step
		}
	}
	return nil
}

// hasDependency checks if stepA depends on stepB
func (pa *PlanAdjuster) hasDependency(plan *Plan, stepA, stepB string) bool {
	deps, exists := plan.Dependencies[stepA]
	if !exists {
		return false
	}
	
	for _, dep := range deps {
		if dep == stepB {
			return true
		}
	}
	
	return false
}

// calculateImpact calculates the impact of plan changes
func (pa *PlanAdjuster) calculateImpact(oldPlan, newPlan *Plan) float64 {
	impact := 0.0
	
	// Count step changes
	if len(newPlan.Steps) != len(oldPlan.Steps) {
		impact += float64(abs(len(newPlan.Steps) - len(oldPlan.Steps))) * 0.2
	}
	
	// Count priority changes
	for i, newStep := range newPlan.Steps {
		if i < len(oldPlan.Steps) {
			if newStep.Priority != oldPlan.Steps[i].Priority {
				impact += 0.1
			}
		}
	}
	
	return impact
}

// recordAdjustment records adjustment history
func (pa *PlanAdjuster) recordAdjustment(stepID, ruleName, action, reason string, success bool, impact float64) {
	record := &AdjustmentRecord{
		Timestamp: time.Now(),
		StepID:    stepID,
		RuleName:  ruleName,
		Action:    action,
		Reason:    reason,
		Success:   success,
		Impact:    impact,
	}
	
	pa.adjustmentHistory = append(pa.adjustmentHistory, record)
	
	// Keep history size limited
	if len(pa.adjustmentHistory) > pa.historySize {
		pa.adjustmentHistory = pa.adjustmentHistory[1:]
	}
}

// GetAdjustmentHistory returns recent adjustment history
func (pa *PlanAdjuster) GetAdjustmentHistory(limit int) []*AdjustmentRecord {
	if limit <= 0 || limit > len(pa.adjustmentHistory) {
		limit = len(pa.adjustmentHistory)
	}
	
	return pa.adjustmentHistory[len(pa.adjustmentHistory)-limit:]
}

// AddAdjustmentRule adds a custom adjustment rule
func (pa *PlanAdjuster) AddAdjustmentRule(rule AdjustmentRule) {
	pa.adjustmentRules = append(pa.adjustmentRules, rule)
	
	// Re-sort by priority
	sort.Slice(pa.adjustmentRules, func(i, j int) bool {
		return pa.adjustmentRules[i].Priority < pa.adjustmentRules[j].Priority
	})
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}