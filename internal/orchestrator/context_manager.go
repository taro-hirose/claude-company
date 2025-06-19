package orchestrator

import (
	"fmt"
	"sync"
	"time"
)

type ContextManager struct {
	tasks       map[string]*TaskSummary
	stepChain   []string
	stepContext map[int]*StepContext
	summarizer  *ContextSummarizer
	mu          sync.RWMutex
	maxHistory  int
	contextData map[string]*ContextData
}

type ContextData struct {
	Key       string                 `json:"key"`
	Value     string                 `json:"value"`
	Type      ContextDataType        `json:"type"`
	StepID    int                    `json:"step_id"`
	TaskID    string                 `json:"task_id"`
	CreatedAt time.Time              `json:"created_at"`
	ExpiresAt *time.Time             `json:"expires_at,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type ContextDataType string

const (
	ContextTypeResult   ContextDataType = "result"
	ContextTypeInput    ContextDataType = "input"
	ContextTypeOutput   ContextDataType = "output"
	ContextTypeMetadata ContextDataType = "metadata"
	ContextTypeError    ContextDataType = "error"
)

type StepTransition struct {
	FromStep    int                    `json:"from_step"`
	ToStep      int                    `json:"to_step"`
	TaskID      string                 `json:"task_id"`
	Summary     string                 `json:"summary"`
	Context     map[string]interface{} `json:"context"`
	Timestamp   time.Time              `json:"timestamp"`
	Success     bool                   `json:"success"`
	ErrorReason string                 `json:"error_reason,omitempty"`
}

func NewContextManager(maxHistory int) *ContextManager {
	if maxHistory <= 0 {
		maxHistory = 100
	}

	return &ContextManager{
		tasks:       make(map[string]*TaskSummary),
		stepChain:   make([]string, 0),
		stepContext: make(map[int]*StepContext),
		summarizer:  NewContextSummarizer(),
		maxHistory:  maxHistory,
		contextData: make(map[string]*ContextData),
	}
}

func (cm *ContextManager) AddTask(task *TaskSummary) error {
	if task == nil {
		return fmt.Errorf("タスクがnilです")
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	if err := task.Validate(); err != nil {
		return fmt.Errorf("タスクの検証に失敗: %w", err)
	}

	cm.tasks[task.ID] = task
	return nil
}

func (cm *ContextManager) GetTask(taskID string) (*TaskSummary, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	task, exists := cm.tasks[taskID]
	return task, exists
}

func (cm *ContextManager) UpdateTaskStatus(taskID string, status TaskStatus) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	task, exists := cm.tasks[taskID]
	if !exists {
		return fmt.Errorf("タスクID %s が見つかりません", taskID)
	}

	task.SetStatus(status)
	return nil
}

func (cm *ContextManager) SetStepContext(stepNumber int, dependencies, outputs []string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.stepContext[stepNumber] = &StepContext{
		StepNumber:    stepNumber,
		Dependencies:  dependencies,
		Outputs:       outputs,
		SharedContext: make(map[string]string),
	}
}

func (cm *ContextManager) AddContextData(key, value string, dataType ContextDataType, stepID int, taskID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	contextKey := fmt.Sprintf("%s:%d:%s", key, stepID, taskID)
	cm.contextData[contextKey] = &ContextData{
		Key:       key,
		Value:     value,
		Type:      dataType,
		StepID:    stepID,
		TaskID:    taskID,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	if stepContext, exists := cm.stepContext[stepID]; exists {
		stepContext.SharedContext[key] = value
	}

	if task, exists := cm.tasks[taskID]; exists {
		task.AddContextData(key, value)
	}

	cm.cleanupExpiredData()
}

func (cm *ContextManager) GetContextData(key string, stepID int, taskID string) (*ContextData, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	contextKey := fmt.Sprintf("%s:%d:%s", key, stepID, taskID)
	data, exists := cm.contextData[contextKey]
	return data, exists
}

func (cm *ContextManager) GetStepContext(stepNumber int) (*StepContext, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	context, exists := cm.stepContext[stepNumber]
	return context, exists
}

func (cm *ContextManager) ShareContextBetweenSteps(fromStep, toStep int, keys []string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	fromContext, exists := cm.stepContext[fromStep]
	if !exists {
		return fmt.Errorf("ステップ %d のコンテキストが見つかりません", fromStep)
	}

	toContext, exists := cm.stepContext[toStep]
	if !exists {
		cm.stepContext[toStep] = &StepContext{
			StepNumber:    toStep,
			SharedContext: make(map[string]string),
		}
		toContext = cm.stepContext[toStep]
	}

	for _, key := range keys {
		if value, exists := fromContext.SharedContext[key]; exists {
			toContext.SharedContext[key] = value
		}
	}

	return nil
}

func (cm *ContextManager) TransitionStep(fromStep, toStep int, taskID string) (*StepTransition, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	task, exists := cm.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("タスクID %s が見つかりません", taskID)
	}

	summary, err := cm.summarizer.SummarizeTask(task, &SummaryOptions{
		WordLimit: 100,
		Template:  "brief",
	})
	if err != nil {
		summary = fmt.Sprintf("タスク %s の要約生成に失敗", taskID)
	}

	context := make(map[string]interface{})
	if stepContext, exists := cm.stepContext[fromStep]; exists {
		for key, value := range stepContext.SharedContext {
			context[key] = value
		}
	}

	transition := &StepTransition{
		FromStep:  fromStep,
		ToStep:    toStep,
		TaskID:    taskID,
		Summary:   summary,
		Context:   context,
		Timestamp: time.Now(),
		Success:   true,
	}

	cm.stepChain = append(cm.stepChain, fmt.Sprintf("%d->%d:%s", fromStep, toStep, taskID))

	if len(cm.stepChain) > cm.maxHistory {
		cm.stepChain = cm.stepChain[1:]
	}

	return transition, nil
}

func (cm *ContextManager) GetStepChain() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	chain := make([]string, len(cm.stepChain))
	copy(chain, cm.stepChain)
	return chain
}

func (cm *ContextManager) GenerateContextSummary(stepID int, wordLimit int) (string, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var relevantTasks []*TaskSummary
	for _, task := range cm.tasks {
		if task.StepContext != nil && task.StepContext.StepNumber == stepID {
			relevantTasks = append(relevantTasks, task)
		}
	}

	if len(relevantTasks) == 0 {
		return "", fmt.Errorf("ステップ %d に関連するタスクがありません", stepID)
	}

	if len(relevantTasks) == 1 {
		return cm.summarizer.SummarizeTask(relevantTasks[0], &SummaryOptions{
			WordLimit: wordLimit,
			Template:  "detailed",
		})
	}

	return cm.summarizer.SummarizeMultipleTasks(relevantTasks, &SummaryOptions{
		WordLimit: wordLimit,
		Template:  "management",
	})
}

func (cm *ContextManager) GetTasksByStatus(status TaskStatus) []*TaskSummary {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var tasks []*TaskSummary
	for _, task := range cm.tasks {
		if task.Status == status {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

func (cm *ContextManager) GetTasksByStep(stepNumber int) []*TaskSummary {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var tasks []*TaskSummary
	for _, task := range cm.tasks {
		if task.StepContext != nil && task.StepContext.StepNumber == stepNumber {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

func (cm *ContextManager) GetOverallSummary(wordLimit int) (string, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var allTasks []*TaskSummary
	for _, task := range cm.tasks {
		allTasks = append(allTasks, task)
	}

	if len(allTasks) == 0 {
		return "現在管理されているタスクはありません。", nil
	}

	return cm.summarizer.SummarizeMultipleTasks(allTasks, &SummaryOptions{
		WordLimit:     wordLimit,
		Template:      "management",
		IncludeSteps:  true,
		IncludeStatus: true,
	})
}

func (cm *ContextManager) SetContextExpiration(key string, stepID int, taskID string, duration time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	contextKey := fmt.Sprintf("%s:%d:%s", key, stepID, taskID)
	if data, exists := cm.contextData[contextKey]; exists {
		expiresAt := time.Now().Add(duration)
		data.ExpiresAt = &expiresAt
	}
}

func (cm *ContextManager) cleanupExpiredData() {
	now := time.Now()
	for key, data := range cm.contextData {
		if data.ExpiresAt != nil && now.After(*data.ExpiresAt) {
			delete(cm.contextData, key)
		}
	}
}

func (cm *ContextManager) GetContextDataByStep(stepID int) []*ContextData {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var data []*ContextData
	for _, contextData := range cm.contextData {
		if contextData.StepID == stepID {
			data = append(data, contextData)
		}
	}
	return data
}

func (cm *ContextManager) GetContextDataByTask(taskID string) []*ContextData {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var data []*ContextData
	for _, contextData := range cm.contextData {
		if contextData.TaskID == taskID {
			data = append(data, contextData)
		}
	}
	return data
}

func (cm *ContextManager) GetStatistics() map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_tasks"] = len(cm.tasks)
	stats["total_steps"] = len(cm.stepContext)
	stats["total_context_data"] = len(cm.contextData)
	stats["step_chain_length"] = len(cm.stepChain)

	statusCounts := make(map[TaskStatus]int)
	for _, task := range cm.tasks {
		statusCounts[task.Status]++
	}
	stats["status_counts"] = statusCounts

	return stats
}