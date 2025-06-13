package models

import (
	"fmt"
	"strings"
	"time"
)

type Task struct {
	ID          string     `json:"id" db:"id"`
	ParentID    *string    `json:"parent_id,omitempty" db:"parent_id"`
	Description string     `json:"description" db:"description"`
	Mode        string     `json:"mode" db:"mode"`
	PaneID      string     `json:"pane_id" db:"pane_id"`
	Status      string     `json:"status" db:"status"`
	Priority    int        `json:"priority" db:"priority"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	Result      string     `json:"result,omitempty" db:"result"`
	Metadata    string     `json:"metadata,omitempty" db:"metadata"`
}

func NewTask(description, mode, paneID string) *Task {
	now := time.Now()
	return &Task{
		ID:          GenerateULID(),
		Description: description,
		Mode:        mode,
		PaneID:      paneID,
		Status:      "pending",
		Priority:    1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func NewSubTask(parentID, description, mode, paneID string) *Task {
	// 子ペイン専用タスクの検証
	if !isChildPaneTask(description) {
		panic(fmt.Sprintf("SubTask '%s' contains management keywords and should be assigned to parent pane, not child pane %s", description, paneID))
	}
	
	now := time.Now()
	return &Task{
		ID:          GenerateULID(),
		ParentID:    &parentID,
		Description: description,
		Mode:        mode,
		PaneID:      paneID,
		Status:      "pending",
		Priority:    1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// 子ペイン用タスクかどうかを判定するヘルパー関数
func isChildPaneTask(description string) bool {
	childKeywords := []string{"実装", "検証", "テスト", "コーディング", "ビルド", "デプロイ", "implement", "code", "test", "build", "deploy", "verify", "create", "develop", "write"}
	managerKeywords := []string{"マネージメント", "レビュー", "品質管理", "進捗管理", "スケジュール", "計画", "management", "review", "quality", "schedule", "plan", "monitor", "supervise"}
	
	descLower := strings.ToLower(description)
	
	// 管理系キーワードが含まれていれば子ペイン用ではない
	for _, keyword := range managerKeywords {
		if strings.Contains(descLower, strings.ToLower(keyword)) {
			return false
		}
	}
	
	// 実装系キーワードが含まれていれば子ペイン用
	for _, keyword := range childKeywords {
		if strings.Contains(descLower, strings.ToLower(keyword)) {
			return true
		}
	}
	
	// デフォルトでは子ペイン用（実装作業とみなす）
	return true
}

func (t *Task) IsSubTask() bool {
	return t.ParentID != nil
}

func (t *Task) MarkCompleted(result string) {
	now := time.Now()
	t.Status = "completed"
	t.Result = result
	t.CompletedAt = &now
	t.UpdatedAt = now
}

func (t *Task) UpdateStatus(status string) {
	t.Status = status
	t.UpdatedAt = time.Now()
}