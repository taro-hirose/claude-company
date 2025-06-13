package commands

import (
	"fmt"
	"time"
	"claude-company/internal/database"
	"claude-company/internal/models"
	"claude-company/internal/session"
	"claude-company/internal/utils"
	"claude-company/internal/api"
)

type DeployCommand struct {
	taskDesc    string
	manager     *session.Manager
	taskRepo    *database.TaskRepository
	executor    *AsyncTaskExecutor
	taskService *api.TaskService
}

func NewDeployCommand(taskDesc string, manager *session.Manager) *DeployCommand {
	executor := NewAsyncTaskExecutor(5)
	executor.Start()
	
	return &DeployCommand{
		taskDesc:    taskDesc,
		manager:     manager,
		taskRepo:    database.NewTaskRepository(),
		executor:    executor,
		taskService: api.NewTaskService(manager),
	}
}

func (c *DeployCommand) Execute() error {
	panes, err := c.manager.GetPanes()
	if err != nil {
		return fmt.Errorf("failed to get panes: %w", err)
	}
	
	if len(panes) < 1 {
		return fmt.Errorf("need at least 1 pane for task execution")
	}
	
	// 役割強制: タスクの適切なペインを決定
	assignedPaneID, err := c.taskService.FilterAndAssignTask(c.taskDesc, panes[0])
	if err != nil {
		return fmt.Errorf("failed to assign task to appropriate pane: %w", err)
	}
	
	if assignedPaneID != panes[0] {
		fmt.Printf("🔄 Task automatically assigned to appropriate pane: %s\n", assignedPaneID)
	}
	
	task := models.Task{
		ID:          utils.GenerateTaskID(),
		Description: c.taskDesc,
		Mode:        "ai",
		PaneID:      assignedPaneID,  // 強制割り当てされたペインを使用
		Status:      "assigned",
		Priority:    1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    "{}",
	}

	// Save task to database
	if err := c.taskRepo.Create(&task); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	// Submit task for async execution
	if err := c.executor.SubmitTask(task.ID); err != nil {
		return fmt.Errorf("failed to submit task for execution: %w", err)
	}

	return c.executeAIModeWithRoleEnforcement(task)
}

func (c *DeployCommand) executeAIMode(task models.Task) error {
	panes, err := c.manager.GetPanes()
	if err != nil {
		return err
	}

	if len(panes) < 2 {
		return fmt.Errorf("need at least 2 panes for AI mode (manager + workers)")
	}

	managerPane := panes[0]
	claudePane := panes[1]

	// TaskTrackerを作成してペイン分離を実装
	taskTracker := models.NewTaskTracker(task, managerPane)
	
	// 子ペイン作成前のスナップショットを取得
	beforeTasks := c.getCurrentPaneTasks(claudePane)
	taskTracker.CapturePreSubtaskSnapshot(claudePane, beforeTasks)
	
	// 新しいAIマネージャーを作成
	aiManager := NewAIManager(c.manager, task, managerPane)
	
	// マネージャープロンプトを送信
	if err := aiManager.SendManagerPrompt(claudePane); err != nil {
		return err
	}
	
	// 子ペイン作成後のスナップショットを取得
	afterTasks := c.getCurrentPaneTasks(claudePane)
	taskTracker.CapturePostSubtaskSnapshot(claudePane, afterTasks)
	
	// 新しく追加されたタスクを検出
	newTasks := taskTracker.GetPaneDiff(claudePane)
	if len(newTasks) > 0 {
		fmt.Printf("🔍 子ペインで検出された新規タスク: %v\n", newTasks)
		// これらのタスクがAIによって認識できるようになる
	}

	fmt.Printf("🎯 AI管理モード開始: 親ペイン %s がプロジェクトマネージャーとして機能します\n", managerPane)
	fmt.Printf("📋 役割分担: 親ペイン=マネジメント・レビュー専用, 子ペイン=実装作業専用\n")
	fmt.Printf("🔄 子ペインが実装完了報告後、親ペインがレビュー・品質管理を実施します\n")
	
	return nil
}

// 現在のペインのタスクを取得するヘルパー関数
func (c *DeployCommand) getCurrentPaneTasks(paneID string) []string {
	// 実際の実装では、tmuxまたはペイン管理システムからタスクリストを取得
	// ここでは簡略化してサンプルを返す
	tasks, _ := c.taskRepo.GetByPane(paneID)
	var taskDescs []string
	for _, task := range tasks {
		taskDescs = append(taskDescs, task.Description)
	}
	return taskDescs
}

// executeAIModeWithRoleEnforcement は役割強制機能付きのAIモード実行
func (c *DeployCommand) executeAIModeWithRoleEnforcement(task models.Task) error {
	panes, err := c.manager.GetPanes()
	if err != nil {
		return err
	}

	if len(panes) < 2 {
		return fmt.Errorf("need at least 2 panes for AI mode (manager + workers)")
	}

	managerPane := panes[0]
	claudePane := panes[1]

	// TaskTrackerを作成してペイン分離を実装
	taskTracker := models.NewTaskTracker(task, managerPane)
	
	// 役割強制機能付きAIマネージャーを作成
	aiManager := NewAIManager(c.manager, task, managerPane)
	
	// タスクの適切性を検証
	isValid, message, err := c.taskService.ValidateTaskAssignment(task.Description, task.PaneID)
	if err != nil {
		return fmt.Errorf("task validation failed: %v", err)
	}
	
	if !isValid {
		fmt.Printf("⚠️  Task assignment validation failed: %s\n", message)
		fmt.Println("🔄 Enforcing role-based task assignment...")
		
		// 役割ベースの強制割り当てを実行
		if err := c.taskService.EnforceRoleBasedAssignment(task.Description, task.PaneID); err != nil {
			return fmt.Errorf("role enforcement failed: %v", err)
		}
	}
	
	// 子ペイン作成前のスナップショットを取得
	beforeTasks := c.getCurrentPaneTasks(claudePane)
	taskTracker.CapturePreSubtaskSnapshot(claudePane, beforeTasks)
	
	// 拡張マネージャープロンプトを送信
	if err := aiManager.SendManagerPrompt(claudePane); err != nil {
		return err
	}
	
	// 子ペイン作成後のスナップショットを取得
	afterTasks := c.getCurrentPaneTasks(claudePane)
	taskTracker.CapturePostSubtaskSnapshot(claudePane, afterTasks)
	
	// 新しく追加されたタスクを検出
	newTasks := taskTracker.GetPaneDiff(claudePane)
	if len(newTasks) > 0 {
		fmt.Printf("🔍 子ペインで検出された新規タスク: %v\n", newTasks)
	}

	// ペイン統計を表示
	if stats, err := aiManager.GetPaneStatistics(); err == nil {
		fmt.Printf("📊 Pane Statistics: %v\n", stats)
	}

	fmt.Printf("🎯 役割強制AI管理モード開始: 親ペイン %s がプロジェクトマネージャーとして機能します\n", managerPane)
	fmt.Printf("🛡️  システム強制: 実装タスクは子ペインに、管理タスクは親ペインに自動割り当て\n")
	fmt.Printf("📋 役割分担: 親ペイン=マネジメント・レビュー専用, 子ペイン=実装作業専用\n")
	fmt.Printf("🔄 子ペインが実装完了報告後、親ペインがレビュー・品質管理を実施します\n")
	fmt.Printf("⚠️  親ペインへの実装タスク送信は自動的にブロック・リダイレクトされます\n")
	
	return nil
}

