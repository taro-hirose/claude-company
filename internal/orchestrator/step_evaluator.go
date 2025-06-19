package orchestrator

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// StepStatus represents the status of a step execution
type StepStatus int

const (
	StepStatusPending StepStatus = iota
	StepStatusInProgress
	StepStatusCompleted
	StepStatusFailed
	StepStatusBlocked
	StepStatusSkipped
)

func (s StepStatus) String() string {
	switch s {
	case StepStatusPending:
		return "pending"
	case StepStatusInProgress:
		return "in_progress"
	case StepStatusCompleted:
		return "completed"
	case StepStatusFailed:
		return "failed"
	case StepStatusBlocked:
		return "blocked"
	case StepStatusSkipped:
		return "skipped"
	default:
		return "unknown"
	}
}

// StepQuality represents the quality assessment of a completed step
type StepQuality int

const (
	QualityExcellent StepQuality = iota
	QualityGood
	QualityAcceptable
	QualityPoor
	QualityUnacceptable
)

func (q StepQuality) String() string {
	switch q {
	case QualityExcellent:
		return "excellent"
	case QualityGood:
		return "good"
	case QualityAcceptable:
		return "acceptable"
	case QualityPoor:
		return "poor"
	case QualityUnacceptable:
		return "unacceptable"
	default:
		return "unknown"
	}
}

// StepResult contains the result and evaluation of a step execution
type StepResult struct {
	StepID           string
	Status           StepStatus
	Quality          StepQuality
	Output           string
	ExecutionTime    time.Duration
	StartTime        time.Time
	EndTime          time.Time
	ErrorMessage     string
	Warnings         []string
	Deliverables     []string
	CompletionRate   float64 // 0.0 to 1.0
	EfficiencyScore  float64 // 0.0 to 1.0
	QualityMetrics   map[string]float64
	Feedback         string
	NextActions      []string
}

// StepEvaluator evaluates step execution results and provides feedback
type StepEvaluator struct {
	qualityRules    []QualityRule
	performanceRules []PerformanceRule
	feedbackPatterns map[string]*regexp.Regexp
}

// QualityRule defines criteria for quality assessment
type QualityRule struct {
	Name        string
	Pattern     *regexp.Regexp
	Weight      float64
	ScoreFunc   func(match string) float64
	Description string
}

// PerformanceRule defines criteria for performance assessment
type PerformanceRule struct {
	Name           string
	MaxDuration    time.Duration
	MinCompletionRate float64
	Weight         float64
	Description    string
}

// NewStepEvaluator creates a new step evaluator with default rules
func NewStepEvaluator() *StepEvaluator {
	evaluator := &StepEvaluator{
		qualityRules:     make([]QualityRule, 0),
		performanceRules: make([]PerformanceRule, 0),
		feedbackPatterns: make(map[string]*regexp.Regexp),
	}
	
	evaluator.initializeDefaultRules()
	evaluator.initializeFeedbackPatterns()
	
	return evaluator
}

// initializeDefaultRules sets up default quality and performance rules
func (se *StepEvaluator) initializeDefaultRules() {
	// Quality rules
	se.qualityRules = []QualityRule{
		{
			Name:        "completion_indicator",
			Pattern:     regexp.MustCompile(`(?i)(完了|完了しました|実装完了|テスト完了|作成完了)`),
			Weight:      0.3,
			ScoreFunc:   func(match string) float64 { return 1.0 },
			Description: "Completion indicators in output",
		},
		{
			Name:        "error_indicator", 
			Pattern:     regexp.MustCompile(`(?i)(エラー|失敗|問題|error|failed|exception)`),
			Weight:      0.4,
			ScoreFunc:   func(match string) float64 { return 0.0 },
			Description: "Error indicators in output",
		},
		{
			Name:        "success_indicator",
			Pattern:     regexp.MustCompile(`(?i)(成功|正常|✓|✅|success|successful)`),
			Weight:      0.3,
			ScoreFunc:   func(match string) float64 { return 1.0 },
			Description: "Success indicators in output",
		},
		{
			Name:        "progress_indicator",
			Pattern:     regexp.MustCompile(`(?i)(進行中|作業中|実装中|進捗)`),
			Weight:      0.2,
			ScoreFunc:   func(match string) float64 { return 0.7 },
			Description: "Progress indicators in output",
		},
		{
			Name:        "warning_indicator",
			Pattern:     regexp.MustCompile(`(?i)(警告|注意|warning|caution)`),
			Weight:      0.1,
			ScoreFunc:   func(match string) float64 { return 0.8 },
			Description: "Warning indicators in output",
		},
	}
	
	// Performance rules
	se.performanceRules = []PerformanceRule{
		{
			Name:              "quick_completion",
			MaxDuration:       5 * time.Minute,
			MinCompletionRate: 0.9,
			Weight:            0.3,
			Description:       "Quick completion with high quality",
		},
		{
			Name:              "standard_completion",
			MaxDuration:       15 * time.Minute,
			MinCompletionRate: 0.8,
			Weight:            0.4,
			Description:       "Standard completion time",
		},
		{
			Name:              "extended_completion",
			MaxDuration:       30 * time.Minute,
			MinCompletionRate: 0.7,
			Weight:            0.3,
			Description:       "Extended completion time",
		},
	}
}

// initializeFeedbackPatterns sets up patterns for feedback extraction
func (se *StepEvaluator) initializeFeedbackPatterns() {
	se.feedbackPatterns = map[string]*regexp.Regexp{
		"deliverables": regexp.MustCompile(`(?i)(?:成果物|deliverable|output)[:：]\s*(.+)`),
		"next_steps":   regexp.MustCompile(`(?i)(?:次のステップ|next\s+step|todo)[:：]\s*(.+)`),
		"issues":       regexp.MustCompile(`(?i)(?:問題|issue|problem)[:：]\s*(.+)`),
		"suggestions":  regexp.MustCompile(`(?i)(?:提案|suggestion|recommend)[:：]\s*(.+)`),
	}
}

// EvaluateStep evaluates a step execution result
func (se *StepEvaluator) EvaluateStep(stepID, output string, startTime, endTime time.Time) *StepResult {
	result := &StepResult{
		StepID:         stepID,
		Output:         output,
		StartTime:      startTime,
		EndTime:        endTime,
		ExecutionTime:  endTime.Sub(startTime),
		QualityMetrics: make(map[string]float64),
		Warnings:       make([]string, 0),
		Deliverables:   make([]string, 0),
		NextActions:    make([]string, 0),
	}
	
	// Evaluate status and quality
	se.evaluateStatus(result)
	se.evaluateQuality(result)
	se.evaluatePerformance(result)
	se.extractFeedback(result)
	
	return result
}

// evaluateStatus determines the step status based on output
func (se *StepEvaluator) evaluateStatus(result *StepResult) {
	output := strings.ToLower(result.Output)
	
	if strings.Contains(output, "完了") || strings.Contains(output, "成功") || 
	   strings.Contains(output, "completed") || strings.Contains(output, "success") {
		result.Status = StepStatusCompleted
	} else if strings.Contains(output, "エラー") || strings.Contains(output, "失敗") ||
	          strings.Contains(output, "error") || strings.Contains(output, "failed") {
		result.Status = StepStatusFailed
		result.ErrorMessage = se.extractErrorMessage(result.Output)
	} else if strings.Contains(output, "進行中") || strings.Contains(output, "作業中") ||
	          strings.Contains(output, "in progress") {
		result.Status = StepStatusInProgress
	} else if strings.Contains(output, "ブロック") || strings.Contains(output, "blocked") {
		result.Status = StepStatusBlocked
	} else {
		result.Status = StepStatusPending
	}
}

// evaluateQuality assesses the quality of step execution
func (se *StepEvaluator) evaluateQuality(result *StepResult) {
	totalScore := 0.0
	totalWeight := 0.0
	
	for _, rule := range se.qualityRules {
		matches := rule.Pattern.FindAllString(result.Output, -1)
		if len(matches) > 0 {
			score := rule.ScoreFunc(matches[0])
			result.QualityMetrics[rule.Name] = score
			totalScore += score * rule.Weight
			totalWeight += rule.Weight
		}
	}
	
	if totalWeight > 0 {
		averageScore := totalScore / totalWeight
		result.Quality = se.scoreToQuality(averageScore)
	} else {
		result.Quality = QualityAcceptable
	}
}

// evaluatePerformance assesses the performance of step execution
func (se *StepEvaluator) evaluatePerformance(result *StepResult) {
	bestScore := 0.0
	
	for _, rule := range se.performanceRules {
		if result.ExecutionTime <= rule.MaxDuration {
			score := rule.Weight
			if result.CompletionRate >= rule.MinCompletionRate {
				score *= 1.2 // Bonus for meeting completion rate
			}
			if score > bestScore {
				bestScore = score
			}
		}
	}
	
	result.EfficiencyScore = bestScore
	
	// Calculate completion rate based on deliverables and output quality
	result.CompletionRate = se.calculateCompletionRate(result)
}

// extractFeedback extracts structured feedback from output
func (se *StepEvaluator) extractFeedback(result *StepResult) {
	for patternName, pattern := range se.feedbackPatterns {
		matches := pattern.FindStringSubmatch(result.Output)
		if len(matches) > 1 {
			content := strings.TrimSpace(matches[1])
			switch patternName {
			case "deliverables":
				result.Deliverables = append(result.Deliverables, content)
			case "next_steps":
				result.NextActions = append(result.NextActions, content)
			case "issues":
				result.Warnings = append(result.Warnings, content)
			case "suggestions":
				result.Feedback = content
			}
		}
	}
}

// scoreToQuality converts numerical score to quality enum
func (se *StepEvaluator) scoreToQuality(score float64) StepQuality {
	if score >= 0.9 {
		return QualityExcellent
	} else if score >= 0.8 {
		return QualityGood
	} else if score >= 0.6 {
		return QualityAcceptable
	} else if score >= 0.4 {
		return QualityPoor
	} else {
		return QualityUnacceptable
	}
}

// calculateCompletionRate calculates completion rate based on various factors
func (se *StepEvaluator) calculateCompletionRate(result *StepResult) float64 {
	rate := 0.5 // Base rate
	
	// Adjust based on status
	switch result.Status {
	case StepStatusCompleted:
		rate = 1.0
	case StepStatusInProgress:
		rate = 0.6
	case StepStatusFailed:
		rate = 0.2
	case StepStatusBlocked:
		rate = 0.1
	}
	
	// Adjust based on deliverables
	if len(result.Deliverables) > 0 {
		rate += 0.1
	}
	
	// Adjust based on quality metrics
	if qualityScore, exists := result.QualityMetrics["completion_indicator"]; exists && qualityScore > 0 {
		rate += 0.2
	}
	
	if rate > 1.0 {
		rate = 1.0
	}
	
	return rate
}

// extractErrorMessage extracts error message from output
func (se *StepEvaluator) extractErrorMessage(output string) string {
	errorPattern := regexp.MustCompile(`(?i)(?:エラー|error)[:：]\s*(.+)`)
	matches := errorPattern.FindStringSubmatch(output)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return "Unknown error"
}

// AddQualityRule adds a custom quality rule
func (se *StepEvaluator) AddQualityRule(rule QualityRule) {
	se.qualityRules = append(se.qualityRules, rule)
}

// AddPerformanceRule adds a custom performance rule
func (se *StepEvaluator) AddPerformanceRule(rule PerformanceRule) {
	se.performanceRules = append(se.performanceRules, rule)
}

// GetEvaluationSummary returns a summary of the evaluation
func (se *StepEvaluator) GetEvaluationSummary(result *StepResult) string {
	summary := fmt.Sprintf("Step %s evaluation:\n", result.StepID)
	summary += fmt.Sprintf("Status: %s\n", result.Status.String())
	summary += fmt.Sprintf("Quality: %s\n", result.Quality.String())
	summary += fmt.Sprintf("Completion Rate: %.1f%%\n", result.CompletionRate*100)
	summary += fmt.Sprintf("Efficiency Score: %.2f\n", result.EfficiencyScore)
	summary += fmt.Sprintf("Execution Time: %v\n", result.ExecutionTime)
	
	if len(result.Warnings) > 0 {
		summary += fmt.Sprintf("Warnings: %d\n", len(result.Warnings))
	}
	
	if len(result.NextActions) > 0 {
		summary += fmt.Sprintf("Next Actions: %d\n", len(result.NextActions))
	}
	
	return summary
}