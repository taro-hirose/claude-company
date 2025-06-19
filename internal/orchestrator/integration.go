package orchestrator

import (
	"fmt"
	"time"
)

// SessionIntegration provides integration with session management
type SessionIntegration struct {
	planner     *AdaptivePlanner
	sessionName string
	paneMapping map[string]string // stepID -> paneID
}

// NewSessionIntegration creates a new session integration
func NewSessionIntegration(planner *AdaptivePlanner, sessionName string) *SessionIntegration {
	return &SessionIntegration{
		planner:     planner,
		sessionName: sessionName,
		paneMapping: make(map[string]string),
	}
}

// AssignStepToPane assigns a step to a specific tmux pane
func (si *SessionIntegration) AssignStepToPane(stepID, paneID string) error {
	step := si.planner.findStep(stepID)
	if step == nil {
		return fmt.Errorf("step %s not found", stepID)
	}
	
	step.AssignedPane = paneID
	si.paneMapping[stepID] = paneID
	
	return nil
}

// GetStepProgress returns progress information for session display
func (si *SessionIntegration) GetStepProgress() map[string]interface{} {
	status := si.planner.GetPlanStatus()
	
	return map[string]interface{}{
		"total_steps":      status.TotalSteps,
		"completed_steps":  status.CompletedSteps,
		"failed_steps":     status.FailedSteps,
		"blocked_steps":    status.BlockedSteps,
		"progress":         status.Progress,
		"adjustments":      status.Adjustments,
		"quality_score":    status.QualityScore,
		"efficiency_score": status.EfficiencyScore,
	}
}

// GenerateProgressReport creates a session-friendly progress report
func (si *SessionIntegration) GenerateProgressReport() string {
	status := si.planner.GetPlanStatus()
	
	report := fmt.Sprintf("📊 実行進捗: %.1f%% (%d/%d完了)\n", 
		status.Progress*100, status.CompletedSteps, status.TotalSteps)
	
	if status.FailedSteps > 0 {
		report += fmt.Sprintf("❌ 失敗: %d個\n", status.FailedSteps)
	}
	
	if status.BlockedSteps > 0 {
		report += fmt.Sprintf("🚫 ブロック: %d個\n", status.BlockedSteps)
	}
	
	if status.Adjustments > 0 {
		report += fmt.Sprintf("🔄 計画調整: %d回\n", status.Adjustments)
	}
	
	report += fmt.Sprintf("🎯 品質スコア: %.2f\n", status.QualityScore)
	report += fmt.Sprintf("⚡ 効率スコア: %.2f\n", status.EfficiencyScore)
	
	return report
}

// GetNextStepForPane returns the next step that should be executed in a specific pane
func (si *SessionIntegration) GetNextStepForPane(paneID string) (*Step, error) {
	nextSteps, err := si.planner.GetNextSteps(10)
	if err != nil {
		return nil, err
	}
	
	// Find step already assigned to this pane
	for _, step := range nextSteps {
		if step.AssignedPane == paneID {
			return step, nil
		}
	}
	
	// Assign first available step to this pane
	for _, step := range nextSteps {
		if step.AssignedPane == "" {
			step.AssignedPane = paneID
			si.paneMapping[step.ID] = paneID
			return step, nil
		}
	}
	
	return nil, fmt.Errorf("no available steps for pane %s", paneID)
}

// HandleStepCompletion processes step completion from session
func (si *SessionIntegration) HandleStepCompletion(stepID, output string, 
	startTime, endTime time.Time) (*StepResult, error) {
	
	return si.planner.ExecuteStep(stepID, output, startTime, endTime)
}

// GetPaneAssignments returns current pane assignments
func (si *SessionIntegration) GetPaneAssignments() map[string]string {
	assignments := make(map[string]string)
	for stepID, paneID := range si.paneMapping {
		assignments[stepID] = paneID
	}
	return assignments
}

// PromptIntegration provides integration with prompt generation
type PromptIntegration struct {
	planner *AdaptivePlanner
}

// NewPromptIntegration creates a new prompt integration
func NewPromptIntegration(planner *AdaptivePlanner) *PromptIntegration {
	return &PromptIntegration{
		planner: planner,
	}
}

// GetStepContextForPrompt returns context information for prompt generation
func (pi *PromptIntegration) GetStepContextForPrompt(stepID string) map[string]interface{} {
	step := pi.planner.findStep(stepID)
	if step == nil {
		return nil
	}
	
	context := map[string]interface{}{
		"step_id":             step.ID,
		"step_name":           step.Name,
		"step_type":           step.Type.String(),
		"step_description":    step.Description,
		"deliverables":        step.Deliverables,
		"completion_criteria": step.CompletionCriteria,
		"dependencies":        step.Dependencies,
		"resources":           step.Resources,
		"estimated_time":      step.EstimatedTime.String(),
		"priority":            step.Priority,
		"retry_count":         step.RetryCount,
		"max_retries":         step.MaxRetries,
	}
	
	// Add execution history if available
	logs := pi.planner.GetExecutionLog(10)
	stepLogs := make([]map[string]interface{}, 0)
	for _, log := range logs {
		if log.StepID == stepID {
			stepLogs = append(stepLogs, map[string]interface{}{
				"timestamp": log.Timestamp,
				"action":    log.Action.String(),
				"metrics":   log.Metrics,
			})
		}
	}
	context["execution_history"] = stepLogs
	
	// Add learning insights
	insights := pi.planner.GetLearningInsights()
	if successPatterns, ok := insights["success_patterns"].(map[string]float64); ok {
		patternKey := fmt.Sprintf("%s_%d_deps", step.Type.String(), len(step.Dependencies))
		if rate, exists := successPatterns[patternKey]; exists {
			context["success_rate"] = rate
		}
	}
	
	return context
}

// GetPlanOverviewForPrompt returns plan overview for manager prompts
func (pi *PromptIntegration) GetPlanOverviewForPrompt() map[string]interface{} {
	status := pi.planner.GetPlanStatus()
	
	overview := map[string]interface{}{
		"plan_id":         status.PlanID,
		"total_steps":     status.TotalSteps,
		"completed_steps": status.CompletedSteps,
		"failed_steps":    status.FailedSteps,
		"blocked_steps":   status.BlockedSteps,
		"progress":        status.Progress,
		"adjustments":     status.Adjustments,
		"quality_score":   status.QualityScore,
	}
	
	// Add next steps information
	nextSteps, err := pi.planner.GetNextSteps(5)
	if err == nil {
		stepSummaries := make([]map[string]interface{}, len(nextSteps))
		for i, step := range nextSteps {
			stepSummaries[i] = map[string]interface{}{
				"id":          step.ID,
				"name":        step.Name,
				"type":        step.Type.String(),
				"priority":    step.Priority,
				"assigned_pane": step.AssignedPane,
			}
		}
		overview["next_steps"] = stepSummaries
	}
	
	// Add recent adjustments
	adjustments := pi.planner.planAdjuster.GetAdjustmentHistory(3)
	adjustmentSummaries := make([]map[string]interface{}, len(adjustments))
	for i, adj := range adjustments {
		adjustmentSummaries[i] = map[string]interface{}{
			"timestamp": adj.Timestamp,
			"step_id":   adj.StepID,
			"rule_name": adj.RuleName,
			"action":    adj.Action,
			"success":   adj.Success,
		}
	}
	overview["recent_adjustments"] = adjustmentSummaries
	
	return overview
}