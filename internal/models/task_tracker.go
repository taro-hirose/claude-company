package models

import (
	"fmt"
	"strings"
	"time"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusNeedsReview TaskStatus = "needs_review"
	TaskStatusRevisionRequired TaskStatus = "revision_required"
)

type SubTask struct {
	ID          string    `json:"id"`
	ParentTaskID string   `json:"parent_task_id"`
	Description string    `json:"description"`
	AssignedPane string   `json:"assigned_pane"`
	Status      TaskStatus `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Result      string    `json:"result,omitempty"`
	ReviewNotes string    `json:"review_notes,omitempty"`
}

type TaskTracker struct {
	MainTask Task         `json:"main_task"`
	SubTasks []SubTask    `json:"sub_tasks"`
	ManagerPane string    `json:"manager_pane"`
	AssignedPanes []string `json:"assigned_panes"`
	PaneSnapshot map[string][]string `json:"pane_snapshot"`
}

func NewTaskTracker(mainTask Task, managerPane string) *TaskTracker {
	return &TaskTracker{
		MainTask:      mainTask,
		SubTasks:      make([]SubTask, 0),
		ManagerPane:   managerPane,
		AssignedPanes: make([]string, 0),
		PaneSnapshot:  make(map[string][]string),
	}
}

func (t *TaskTracker) AddSubTask(description, assignedPane string) SubTask {
	// 子ペインの場合のみサブタスクを作成
	if assignedPane == t.ManagerPane {
		panic(fmt.Sprintf("Cannot assign subtask to manager pane %s. Subtasks must be assigned to child panes only.", t.ManagerPane))
	}
	
	subTask := SubTask{
		ID:           GenerateULID(),
		ParentTaskID: t.MainTask.ID,
		Description:  description,
		AssignedPane: assignedPane,
		Status:       TaskStatusPending,
		CreatedAt:    time.Now(),
	}
	
	// 新しい子ペインを記録
	if !contains(t.AssignedPanes, assignedPane) {
		t.AssignedPanes = append(t.AssignedPanes, assignedPane)
	}
	
	t.SubTasks = append(t.SubTasks, subTask)
	return subTask
}

func (t *TaskTracker) UpdateSubTaskStatus(subTaskID string, status TaskStatus, result string) bool {
	for i, task := range t.SubTasks {
		if task.ID == subTaskID {
			t.SubTasks[i].Status = status
			if result != "" {
				t.SubTasks[i].Result = result
			}
			if status == TaskStatusCompleted || status == TaskStatusNeedsReview {
				now := time.Now()
				t.SubTasks[i].CompletedAt = &now
			}
			return true
		}
	}
	return false
}

func (t *TaskTracker) GetPendingTasks() []SubTask {
	var pending []SubTask
	for _, task := range t.SubTasks {
		if task.Status == TaskStatusPending || task.Status == TaskStatusRevisionRequired {
			pending = append(pending, task)
		}
	}
	return pending
}

func (t *TaskTracker) GetTasksNeedingReview() []SubTask {
	var needsReview []SubTask
	for _, task := range t.SubTasks {
		if task.Status == TaskStatusNeedsReview {
			needsReview = append(needsReview, task)
		}
	}
	return needsReview
}

func (t *TaskTracker) AllTasksCompleted() bool {
	for _, task := range t.SubTasks {
		if task.Status != TaskStatusCompleted {
			return false
		}
	}
	return len(t.SubTasks) > 0
}

func (t *TaskTracker) GetProgressSummary() map[TaskStatus]int {
	summary := make(map[TaskStatus]int)
	for _, task := range t.SubTasks {
		summary[task.Status]++
	}
	return summary
}

func (t *TaskTracker) GetCompletionPercentage() float64 {
	if len(t.SubTasks) == 0 {
		return 0.0
	}
	
	completed := 0
	for _, task := range t.SubTasks {
		if task.Status == TaskStatusCompleted {
			completed++
		}
	}
	
	return float64(completed) / float64(len(t.SubTasks)) * 100.0
}

func (t *TaskTracker) GetInProgressTasks() []SubTask {
	var inProgress []SubTask
	for _, task := range t.SubTasks {
		if task.Status == TaskStatusInProgress {
			inProgress = append(inProgress, task)
		}
	}
	return inProgress
}

func (t *TaskTracker) GetCompletedTasks() []SubTask {
	var completed []SubTask
	for _, task := range t.SubTasks {
		if task.Status == TaskStatusCompleted {
			completed = append(completed, task)
		}
	}
	return completed
}

func (t *TaskTracker) GetSubTaskByID(id string) *SubTask {
	for i, task := range t.SubTasks {
		if task.ID == id {
			return &t.SubTasks[i]
		}
	}
	return nil
}

func (t *TaskTracker) AddReviewNotes(subTaskID, notes string) bool {
	for i, task := range t.SubTasks {
		if task.ID == subTaskID {
			t.SubTasks[i].ReviewNotes = notes
			return true
		}
	}
	return false
}

func (t *TaskTracker) GetTaskDuration(subTaskID string) *time.Duration {
	task := t.GetSubTaskByID(subTaskID)
	if task == nil || task.CompletedAt == nil {
		return nil
	}
	
	duration := task.CompletedAt.Sub(task.CreatedAt)
	return &duration
}

// 子ペイン作成前のスナップショットを保存
func (t *TaskTracker) CapturePreSubtaskSnapshot(paneID string, tasks []string) {
	t.PaneSnapshot[paneID+"_before"] = make([]string, len(tasks))
	copy(t.PaneSnapshot[paneID+"_before"], tasks)
}

// 子ペイン作成後のスナップショットを保存
func (t *TaskTracker) CapturePostSubtaskSnapshot(paneID string, tasks []string) {
	t.PaneSnapshot[paneID+"_after"] = make([]string, len(tasks))
	copy(t.PaneSnapshot[paneID+"_after"], tasks)
}

// ペイン内の差分を取得
func (t *TaskTracker) GetPaneDiff(paneID string) []string {
	before := t.PaneSnapshot[paneID+"_before"]
	after := t.PaneSnapshot[paneID+"_after"]
	
	var diff []string
	for _, task := range after {
		if !contains(before, task) {
			diff = append(diff, task)
		}
	}
	return diff
}

// 親ペインのタスクをフィルタリング
func (t *TaskTracker) IsManagerTask(taskDesc string) bool {
	managerKeywords := []string{"マネージメント", "レビュー", "品質管理", "進捗管理", "スケジュール", "計画", "management", "review", "quality", "schedule", "plan"}
	
	for _, keyword := range managerKeywords {
		if strings.Contains(strings.ToLower(taskDesc), strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// 子ペインのタスクをフィルタリング
func (t *TaskTracker) IsChildPaneTask(taskDesc string) bool {
	childKeywords := []string{"実装", "検証", "テスト", "コーディング", "ビルド", "デプロイ", "implement", "code", "test", "build", "deploy", "verify"}
	
	for _, keyword := range childKeywords {
		if strings.Contains(strings.ToLower(taskDesc), strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// ペインの役割を強制する
func (t *TaskTracker) EnforceRoleBasedTaskAssignment(taskDesc, requestedPane string) (string, error) {
	isManagerTask := t.IsManagerTask(taskDesc)
	isChildTask := t.IsChildPaneTask(taskDesc)
	
	// 親ペインに子タスクが、子ペインに親タスクが流れ込むのを防ぐ
	if requestedPane == t.ManagerPane && isChildTask {
		// 実装タスクを子ペインにリダイレクト
		if len(t.AssignedPanes) > 0 {
			return t.AssignedPanes[0], nil
		}
		return "", fmt.Errorf("実装タスク '%s' は子ペインにとィサインされる必要があります", taskDesc)
	}
	
	if requestedPane != t.ManagerPane && isManagerTask {
		// マネージメントタスクを親ペインにリダイレクト
		return t.ManagerPane, nil
	}
	
	return requestedPane, nil
}

// ヘルパー関数
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (t *TaskTracker) GetTotalDuration() *time.Duration {
	if !t.AllTasksCompleted() {
		return nil
	}
	
	if len(t.SubTasks) == 0 {
		return nil
	}
	
	earliestStart := t.SubTasks[0].CreatedAt
	latestEnd := t.SubTasks[0].CreatedAt
	
	for _, task := range t.SubTasks {
		if task.CreatedAt.Before(earliestStart) {
			earliestStart = task.CreatedAt
		}
		if task.CompletedAt != nil && task.CompletedAt.After(latestEnd) {
			latestEnd = *task.CompletedAt
		}
	}
	
	duration := latestEnd.Sub(earliestStart)
	return &duration
}

// UpdateSubTaskPane updates the assigned pane for a subtask
func (t *TaskTracker) UpdateSubTaskPane(subTaskID string, newPaneID string) bool {
	for i, task := range t.SubTasks {
		if task.ID == subTaskID {
			t.SubTasks[i].AssignedPane = newPaneID
			
			// Add the new pane to assigned panes if not already present
			if !contains(t.AssignedPanes, newPaneID) {
				t.AssignedPanes = append(t.AssignedPanes, newPaneID)
			}
			
			return true
		}
	}
	return false
}