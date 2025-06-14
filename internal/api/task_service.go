package api

import (
	"fmt"
	"strings"

	"claude-company/internal/database"
	"claude-company/internal/models"
	"claude-company/internal/session"
	"claude-company/internal/utils"
)

type TaskService struct {
	repo           *database.TaskRepository
	sessionManager *session.Manager
	paneFilter     *utils.PaneFilter // 統一ペインフィルター
}

func NewTaskService(sessionManager *session.Manager) *TaskService {
	return &TaskService{
		repo:           database.NewTaskRepository(),
		sessionManager: sessionManager,
		paneFilter:     utils.NewPaneFilter(),
	}
}

type TaskWithChildren struct {
	*models.Task
	Children []*TaskWithChildren `json:"children,omitempty"`
}

func (s *TaskService) GetTaskHierarchy(taskID string) (*TaskWithChildren, error) {
	task, err := s.repo.GetByID(taskID)
	if err != nil {
		return nil, err
	}

	hierarchy := &TaskWithChildren{Task: task}
	children, err := s.getTaskChildren(taskID)
	if err != nil {
		return nil, err
	}
	hierarchy.Children = children

	return hierarchy, nil
}

func (s *TaskService) getTaskChildren(parentID string) ([]*TaskWithChildren, error) {
	children, err := s.repo.GetChildren(parentID)
	if err != nil {
		return nil, err
	}

	var result []*TaskWithChildren
	for _, child := range children {
		childWithChildren := &TaskWithChildren{Task: child}
		
		grandChildren, err := s.getTaskChildren(child.ID)
		if err != nil {
			return nil, err
		}
		childWithChildren.Children = grandChildren
		
		result = append(result, childWithChildren)
	}

	return result, nil
}

func (s *TaskService) ShareTaskWithSiblings(taskID string) error {
	task, err := s.repo.GetByID(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	if task.ParentID == nil {
		return fmt.Errorf("task has no parent, cannot share with siblings")
	}

	siblings, err := s.repo.GetChildren(*task.ParentID)
	if err != nil {
		return fmt.Errorf("failed to get siblings: %w", err)
	}

	for _, sibling := range siblings {
		if sibling.ID != taskID && sibling.PaneID != task.PaneID {
			if err := s.repo.ShareTask(taskID, sibling.PaneID, "read"); err != nil {
				return fmt.Errorf("failed to share with sibling %s: %w", sibling.ID, err)
			}
		}
	}

	return nil
}

func (s *TaskService) ShareTaskWithFamily(taskID string) error {
	task, err := s.repo.GetByID(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	if task.ParentID != nil {
		parent, err := s.repo.GetByID(*task.ParentID)
		if err != nil {
			return fmt.Errorf("failed to get parent: %w", err)
		}

		if parent.PaneID != task.PaneID {
			if err := s.repo.ShareTask(taskID, parent.PaneID, "read"); err != nil {
				return fmt.Errorf("failed to share with parent: %w", err)
			}
		}

		if err := s.ShareTaskWithSiblings(taskID); err != nil {
			return fmt.Errorf("failed to share with siblings: %w", err)
		}
	}

	children, err := s.repo.GetChildren(taskID)
	if err != nil {
		return fmt.Errorf("failed to get children: %w", err)
	}

	for _, child := range children {
		if child.PaneID != task.PaneID {
			if err := s.repo.ShareTask(taskID, child.PaneID, "read"); err != nil {
				return fmt.Errorf("failed to share with child %s: %w", child.ID, err)
			}
		}
	}

	return nil
}

func (s *TaskService) PropagateStatusUpdate(taskID, newStatus string) error {
	task, err := s.repo.GetByID(taskID)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateStatus(taskID, newStatus); err != nil {
		return err
	}

	if newStatus == "completed" && task.ParentID != nil {
		siblings, err := s.repo.GetChildren(*task.ParentID)
		if err != nil {
			return err
		}

		allCompleted := true
		for _, sibling := range siblings {
			if sibling.Status != "completed" {
				allCompleted = false
				break
			}
		}

		if allCompleted {
			if err := s.PropagateStatusUpdate(*task.ParentID, "completed"); err != nil {
				return err
			}
		}
	}

	return nil
}

type TaskStats struct {
	TotalTasks         int            `json:"total_tasks"`
	TasksByStatus      map[string]int `json:"tasks_by_status"`
	TasksByPriority    map[int]int    `json:"tasks_by_priority"`
	CompletionRate     float64        `json:"completion_rate"`
	AverageTimeToComplete *float64    `json:"average_time_to_complete,omitempty"`
}

func (s *TaskService) GetTaskStatistics(paneID string) (*TaskStats, error) {
	tasks, err := s.repo.GetByPaneID(paneID)
	if err != nil {
		return nil, err
	}

	stats := &TaskStats{
		TotalTasks:      len(tasks),
		TasksByStatus:   make(map[string]int),
		TasksByPriority: make(map[int]int),
	}

	var completedCount int
	var totalCompletionTime float64
	var completedWithTime int

	for _, task := range tasks {
		stats.TasksByStatus[task.Status]++
		stats.TasksByPriority[task.Priority]++

		if task.Status == "completed" {
			completedCount++
			if task.CompletedAt != nil {
				duration := task.CompletedAt.Sub(task.CreatedAt).Hours()
				totalCompletionTime += duration
				completedWithTime++
			}
		}
	}

	if stats.TotalTasks > 0 {
		stats.CompletionRate = float64(completedCount) / float64(stats.TotalTasks) * 100
	}

	if completedWithTime > 0 {
		avgTime := totalCompletionTime / float64(completedWithTime)
		stats.AverageTimeToComplete = &avgTime
	}

	return stats, nil
}

// TaskFilter はタスクの種類を判定するためのフィルタ
type TaskFilter struct {
	ImplementationKeywords []string
	ManagementKeywords    []string
	ReviewKeywords        []string
}

func NewTaskFilter() *TaskFilter {
	return &TaskFilter{
		ImplementationKeywords: []string{
			"実装", "コード", "作成", "開発", "プログラム", "関数", "メソッド", 
			"クラス", "API", "データベース", "スクリプト", "機能追加", "バグ修正",
			"implement", "code", "develop", "create", "function", "method", 
			"class", "api", "database", "script", "feature", "bug fix",
		},
		ManagementKeywords: []string{
			"計画", "管理", "割り当て", "スケジュール", "レビュー", "承認", "監督",
			"plan", "manage", "assign", "schedule", "review", "approve", "supervise",
		},
		ReviewKeywords: []string{
			"確認", "検証", "テスト", "レビュー", "チェック", "監査",
			"verify", "test", "review", "check", "audit", "validate",
		},
	}
}

// ClassifyTask はタスクの種類を分類
func (tf *TaskFilter) ClassifyTask(description string) string {
	desc := strings.ToLower(description)
	
	implementationScore := 0
	managementScore := 0
	reviewScore := 0
	
	for _, keyword := range tf.ImplementationKeywords {
		if strings.Contains(desc, strings.ToLower(keyword)) {
			implementationScore++
		}
	}
	
	for _, keyword := range tf.ManagementKeywords {
		if strings.Contains(desc, strings.ToLower(keyword)) {
			managementScore++
		}
	}
	
	for _, keyword := range tf.ReviewKeywords {
		if strings.Contains(desc, strings.ToLower(keyword)) {
			reviewScore++
		}
	}
	
	if implementationScore > managementScore && implementationScore > reviewScore {
		return "implementation"
	} else if managementScore > reviewScore {
		return "management"
	} else if reviewScore > 0 {
		return "review"
	}
	
	return "unknown"
}

// FilterAndAssignTask はタスクをフィルタリングして適切なペインに割り当て（統一フィルター使用）
func (s *TaskService) FilterAndAssignTask(taskDescription, requestedPaneID string) (string, error) {
	// タスク割り当ての妥当性を検証
	isValid, message, err := s.paneFilter.ValidateTaskAssignment(taskDescription, requestedPaneID)
	if err != nil {
		return requestedPaneID, fmt.Errorf("validation failed: %v", err)
	}
	
	if !isValid {
		fmt.Printf("⚠️  %s\n", message)
		// 最適なペインを取得
		bestPane, err := s.paneFilter.GetBestPaneForTask(taskDescription)
		if err != nil {
			// フォールバック: 子ペインを作成
			if strings.Contains(err.Error(), "no worker panes available") {
				newPaneID, createErr := s.sessionManager.CreateNewPaneAndRegisterAsChild()
				if createErr != nil {
					return requestedPaneID, fmt.Errorf("failed to create new pane: %v", createErr)
				}
				fmt.Printf("🔄 Created new worker pane %s for task\n", newPaneID)
				return newPaneID, nil
			}
			return requestedPaneID, fmt.Errorf("failed to find suitable pane: %v", err)
		}
		fmt.Printf("🔄 Redirected task to pane %s\n", bestPane)
		return bestPane, nil
	}
	
	fmt.Printf("✅ %s\n", message)
	return requestedPaneID, nil
}

// EnforceRoleBasedAssignment は役割ベースのタスク割り当てを強制
func (s *TaskService) EnforceRoleBasedAssignment(taskDescription, requestedPaneID string) error {
	assignedPaneID, err := s.FilterAndAssignTask(taskDescription, requestedPaneID)
	if err != nil {
		return fmt.Errorf("failed to filter and assign task: %v", err)
	}
	
	if assignedPaneID != requestedPaneID {
		// タスクがリダイレクトされた場合、元のペインに通知
		notification := fmt.Sprintf("Task redirected to pane %s for proper execution", assignedPaneID)
		if err := s.sessionManager.SendToPane(requestedPaneID, notification); err != nil {
			fmt.Printf("Warning: failed to send redirect notification: %v\n", err)
		}
	}
	
	// 実際のタスクを適切なペインに送信
	return s.sessionManager.SendToFilteredPane(assignedPaneID, taskDescription)
}

// ValidateTaskAssignment はタスク割り当ての妥当性を検証（統一フィルター使用）
func (s *TaskService) ValidateTaskAssignment(taskDescription, paneID string) (bool, string, error) {
	return s.paneFilter.ValidateTaskAssignment(taskDescription, paneID)
}