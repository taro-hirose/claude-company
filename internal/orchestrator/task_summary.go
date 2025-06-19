package orchestrator

import (
	"fmt"
	"strings"
	"time"
)

type TaskSummary struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Summary     string            `json:"summary"`
	WordCount   int               `json:"word_count"`
	Status      TaskStatus        `json:"status"`
	Priority    TaskPriority      `json:"priority"`
	AssignedTo  string            `json:"assigned_to"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
	StepContext *StepContext      `json:"step_context,omitempty"`
}

// 既存のtypes.goで定義された型を使用
const (
	StatusPending    = TaskStatusPending
	StatusInProgress = TaskStatusInProgress
	StatusCompleted  = TaskStatusCompleted
	StatusFailed     = TaskStatusFailed
	StatusCancelled  = TaskStatusCancelled
)

const (
	PriorityLow      = TaskPriorityLow
	PriorityMedium   = TaskPriorityMedium
	PriorityHigh     = TaskPriorityHigh
	PriorityCritical = TaskPriorityHigh  // types.goにはcriticalがないのでhighにマップ
)

type StepContext struct {
	StepNumber    int               `json:"step_number"`
	Dependencies  []string          `json:"dependencies"`
	Outputs       []string          `json:"outputs"`
	SharedContext map[string]string `json:"shared_context"`
}

func NewTaskSummary(id, title, description string) *TaskSummary {
	return &TaskSummary{
		ID:          id,
		Title:       title,
		Description: description,
		Status:      StatusPending,
		Priority:    PriorityMedium,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Tags:        make([]string, 0),
		Metadata:    make(map[string]string),
	}
}

func (ts *TaskSummary) UpdateSummary(summary string) error {
	if summary == "" {
		return fmt.Errorf("要約が空です")
	}

	wordCount := countWords(summary)
	if wordCount > 200 {
		return fmt.Errorf("要約が長すぎます (現在: %d語, 上限: 200語)", wordCount)
	}
	if wordCount < 10 {
		return fmt.Errorf("要約が短すぎます (現在: %d語, 下限: 10語)", wordCount)
	}

	ts.Summary = summary
	ts.WordCount = wordCount
	ts.UpdatedAt = time.Now()
	return nil
}

func (ts *TaskSummary) SetStatus(status TaskStatus) {
	ts.Status = status
	ts.UpdatedAt = time.Now()

	if status == StatusCompleted {
		now := time.Now()
		ts.CompletedAt = &now
	}
}

func (ts *TaskSummary) SetPriority(priority TaskPriority) {
	ts.Priority = priority
	ts.UpdatedAt = time.Now()
}

func (ts *TaskSummary) AssignTo(assignee string) {
	ts.AssignedTo = assignee
	ts.UpdatedAt = time.Now()
}

func (ts *TaskSummary) AddTag(tag string) {
	if tag == "" {
		return
	}

	for _, existingTag := range ts.Tags {
		if existingTag == tag {
			return
		}
	}

	ts.Tags = append(ts.Tags, tag)
	ts.UpdatedAt = time.Now()
}

func (ts *TaskSummary) RemoveTag(tag string) {
	for i, existingTag := range ts.Tags {
		if existingTag == tag {
			ts.Tags = append(ts.Tags[:i], ts.Tags[i+1:]...)
			ts.UpdatedAt = time.Now()
			break
		}
	}
}

func (ts *TaskSummary) SetMetadata(key, value string) {
	if ts.Metadata == nil {
		ts.Metadata = make(map[string]string)
	}
	ts.Metadata[key] = value
	ts.UpdatedAt = time.Now()
}

func (ts *TaskSummary) GetMetadata(key string) (string, bool) {
	if ts.Metadata == nil {
		return "", false
	}
	value, exists := ts.Metadata[key]
	return value, exists
}

func (ts *TaskSummary) IsCompleted() bool {
	return ts.Status == StatusCompleted
}

func (ts *TaskSummary) IsFailed() bool {
	return ts.Status == StatusFailed
}

func (ts *TaskSummary) IsActive() bool {
	return ts.Status == StatusInProgress
}

func (ts *TaskSummary) Duration() time.Duration {
	if ts.CompletedAt != nil {
		return ts.CompletedAt.Sub(ts.CreatedAt)
	}
	return time.Since(ts.CreatedAt)
}

func (ts *TaskSummary) SetStepContext(stepNumber int, dependencies, outputs []string) {
	ts.StepContext = &StepContext{
		StepNumber:    stepNumber,
		Dependencies:  dependencies,
		Outputs:       outputs,
		SharedContext: make(map[string]string),
	}
	ts.UpdatedAt = time.Now()
}

func (ts *TaskSummary) AddContextData(key, value string) {
	if ts.StepContext == nil {
		ts.StepContext = &StepContext{
			SharedContext: make(map[string]string),
		}
	}
	if ts.StepContext.SharedContext == nil {
		ts.StepContext.SharedContext = make(map[string]string)
	}
	ts.StepContext.SharedContext[key] = value
	ts.UpdatedAt = time.Now()
}

func (ts *TaskSummary) GetContextData(key string) (string, bool) {
	if ts.StepContext == nil || ts.StepContext.SharedContext == nil {
		return "", false
	}
	value, exists := ts.StepContext.SharedContext[key]
	return value, exists
}

func (ts *TaskSummary) Validate() error {
	if ts.ID == "" {
		return fmt.Errorf("タスクIDが必須です")
	}
	if ts.Title == "" {
		return fmt.Errorf("タスクタイトルが必須です")
	}
	if ts.Description == "" {
		return fmt.Errorf("タスク説明が必須です")
	}
	if ts.Summary != "" && (ts.WordCount < 10 || ts.WordCount > 200) {
		return fmt.Errorf("要約の語数が範囲外です (現在: %d語, 範囲: 10-200語)", ts.WordCount)
	}
	return nil
}

func countWords(text string) int {
	if text == "" {
		return 0
	}

	// 日本語と英語を分けて処理
	japaneseChars := 0
	englishWords := 0
	
	// 日本語文字をカウント
	for _, r := range text {
		if isJapanese(r) {
			japaneseChars++
		}
	}
	
	// 英語単語をカウント（日本語文字を除去してから）
	cleanText := ""
	for _, r := range text {
		if !isJapanese(r) {
			cleanText += string(r)
		} else {
			cleanText += " " // 日本語文字をスペースに置換
		}
	}
	
	englishWordsSlice := strings.Fields(cleanText)
	for _, word := range englishWordsSlice {
		if strings.TrimSpace(word) != "" {
			englishWords++
		}
	}
	
	// 日本語は約2.5文字で1語とカウント、英語は1単語で1語
	japaneseWordEquivalent := japaneseChars / 3
	if japaneseChars%3 > 0 {
		japaneseWordEquivalent++
	}
	
	return japaneseWordEquivalent + englishWords
}

func isJapanese(r rune) bool {
	return (r >= 0x3040 && r <= 0x309F) || // ひらがな
		(r >= 0x30A0 && r <= 0x30FF) || // カタカナ
		(r >= 0x4E00 && r <= 0x9FAF) // 漢字
}