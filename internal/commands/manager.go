package commands

import (
	"bytes"
	"claude-company/internal/api"
	"claude-company/internal/models"
	"claude-company/internal/session"
	"claude-company/internal/utils"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type AIManager struct {
	sessionManager *session.Manager
	taskTracker    *models.TaskTracker
	taskService    *api.TaskService
	parentPanes    map[string]bool   // Track parent panes to prevent task assignment (レガシー)
	paneFilter     *utils.PaneFilter // 統一ペインフィルター
}

func NewAIManager(sessionManager *session.Manager, mainTask models.Task, managerPane string) *AIManager {
	parentPanes := make(map[string]bool)
	parentPanes[managerPane] = true

	// Get initial panes and mark them as parents (deprecated, using session manager now)
	if panes, err := sessionManager.GetPanes(); err == nil {
		for _, pane := range panes {
			parentPanes[pane] = true
		}
	}

	manager := &AIManager{
		sessionManager: sessionManager,
		taskTracker:    models.NewTaskTracker(mainTask, managerPane),
		taskService:    api.NewTaskService(sessionManager),
		parentPanes:    parentPanes,
		paneFilter:     utils.NewPaneFilterWithLegacySupport(parentPanes),
	}

	return manager
}

func (m *AIManager) SendManagerPrompt(claudePane string) error {
	prompt := m.buildManagerPrompt()
	return m.sessionManager.SendToPane(claudePane, prompt)
}

func (m *AIManager) buildManagerPrompt() string {
	availablePanes, _ := m.sessionManager.GetPanes()
	var claudePane string
	if len(availablePanes) > 1 {
		claudePane = availablePanes[1]
	}

	return fmt.Sprintf(`あなたは%s（プロジェクトマネージャー）です。

🔐 **絶対的な役割制限** 🔐
以下の作業は一切禁止されています：
- コードの記述・編集
- ファイルの直接操作
- ビルド・テストの実行
- デプロイ作業
- 技術実装

✅ **許可されている役割** ✅
- タスクの分析・分解
- サブタスクの割り当て
- 進捗管理・監視
- 品質管理・レビュー指示
- 統合管理・完了判定

⚠️ 強化された制約 ⚠️
1. 実装関連のタスクが誤って親ペインに送られた場合、自動的に子ペインにリダイレクトされます
2. 親ペインではマネージメント・レビュー・品質管理のみ実行可能です
3. 子ペインでは実装・検証・テストのみ実行可能です
4. この役割分担は技術的に強制されており、違反は防止されます

==== メインタスク ====
%s

==== あなたの役割（マネージャー専用） ====
1. メインタスクを分析し、効率的なサブタスクに分解する
2. 必要に応じて子ペインを動的に作成する(並行作業できるものであれば複数立ち上げも可)  
3. 各子ペインに具体的なサブタスクを割り当てる
4. 子ペインの進捗を監視し、作業完了を確認する
5. 子ペインから提出された成果物をレビューする
6. 品質チェック・統合テストを指示する
7. 最終的な統合・完了判定を行う

==== 子ペイン作成方法 ====
必要に応じてtmux split-windowコマンドで新しい子ペインを作成できます：
例：
- 横分割: tmux split-window -h -t claude-squad
- 縦分割: tmux split-window -v -t claude-squad
- 特定ペインを分割: tmux split-window -h -t %s

==== 新規ペイン作成後の手順 ====
新しいペインを作成したら、必ずClaude AIを起動してください：
1. ペイン作成後：tmux send-keys -t 新ペインID 'claude --dangerously-skip-permissions' Enter
2. Claude起動確認後にサブタスクを送信
3. サブタスクを送信後、新規ペインでエンターを1秒後に送信してタスクを実行

==== サブタスクの作成方法 ====
必須条件：
1. サブタスクは必ず、子ペインを作り、子ペインで起動しているclaudeにやらせること
2. 親ペイン（%s）には絶対にサブタスクを送信しないこと（親ペインはマネージメント専用）

サブタスク送信テンプレート：
`+"`"+`
サブタスク: [タスク名]
目的: [このタスクで達成したいこと]
期待する成果物: [具体的な成果物の説明]
制約条件: [注意点や制約があれば]
完了条件: [完了と判断する基準]
`+"`"+`

- サブタスクの作成方法：tmux send-keys -t 子ペインID '[ここにタスクの内容]' Enter
- サブタスクを送信後、必ず送信先のペインでエンターを1秒後に送信してタスクを実行
==== タスク送信方法 ====
各子ペイン（%sまたは新規作成）にサブタスクを送信する場合は、以下のコマンド形式を使用してください：
tmux send-keys -t 子ペインID 'サブタスク: [具体的なサブタスク内容と期待する成果物]' Enter

例：
tmux send-keys -t %s 'サブタスク: internal/models/user.goファイルを作成し、User構造体を定義してください。完了後は「実装完了：ファイルパス」で報告してください' Enter

==== 進捗管理・レビュー方法 ====
1. 定期的に子ペインに進捗確認メッセージを送信
2. 子ペインから「実装完了」報告があったらレビュー指示を送信  
3. 問題があれば修正指示を送信
4. 全サブタスク完了後、統合テストを指示

==== 品質管理ガイドライン ====
- 各サブタスク完了後、必ず成果物のレビューを実施
- ビルドエラーがないか確認指示
- コード品質・設計一貫性の確認
- テスト実行の指示
- 必要に応じて修正・改善指示

==== 実行手順 ====
1. メインタスクを分析してサブタスクに分解
2. 必要に応じて子ペインを作成  
3. 各子ペインに具体的なサブタスクを送信
4. 定期的に進捗を確認し、レビュー・品質管理を実施
5. 全体の統合・完了判定を行う

🚨 **システム強制による役割分担** 🚨
- 実装タスクは自動的に子ペインに割り当てられます
- マネージメントタスクは親ペインでのみ実行されます
- この制限はコードレベルで強制されており、迂回不可能です
- 違反を試みるとエラーが発生し、適切なペインにリダイレクトされます

==== 作業状況報告フォーマット ====
子ペインからの報告は以下の形式で受け取ります：
- 「実装完了：[ファイルパス] - [簡単な説明]」
- 「進捗報告：[進捗状況] - [現在の作業内容]」
- 「エラー報告：[エラー内容] - [支援要請]」

それでは、メインタスクの分析と子ペインへの作業委託を開始してください。`,
		m.taskTracker.ManagerPane,
		m.taskTracker.MainTask.Description,
		m.taskTracker.ManagerPane,
		m.taskTracker.ManagerPane,
		claudePane, claudePane)
}

func (m *AIManager) AddSubTask(description, assignedPane string) (models.SubTask, error) {
	// 役割ベースのタスク割り当てを強制
	correctedPane, err := m.taskTracker.EnforceRoleBasedTaskAssignment(description, assignedPane)
	if err != nil {
		return models.SubTask{}, err
	}

	if correctedPane != assignedPane {
		fmt.Printf("⚠️ タスク '%s' のペインを %s から %s にリダイレクトしました\n", description, assignedPane, correctedPane)
	}

	return m.taskTracker.AddSubTask(description, correctedPane), nil
}

func (m *AIManager) UpdateTaskStatus(subTaskID string, status models.TaskStatus, result string) bool {
	return m.taskTracker.UpdateSubTaskStatus(subTaskID, status, result)
}

func (m *AIManager) SendProgressCheck(paneID string) error {
	// 親ペインからの進捗確認は許可
	if paneID == m.taskTracker.ManagerPane {
		return fmt.Errorf("マネージャーペイン %s に進捗確認を送信することはできません。子ペインのみ監視対象です", paneID)
	}

	checkMessage := fmt.Sprintf("進捗確認: 現在の作業状況を報告してください。完了した場合は「実装完了：[詳細]」、進行中の場合は「進捗報告：[状況]」で回答してください。")
	return m.sessionManager.SendToPane(paneID, checkMessage)
}

func (m *AIManager) SendReviewRequest(paneID, filePath string) error {
	// レビューは親ペインの役割
	if paneID != m.taskTracker.ManagerPane {
		return fmt.Errorf("レビュー要請は親ペイン %s からのみ送信可能です。現在のペイン: %s", m.taskTracker.ManagerPane, paneID)
	}

	reviewMessage := fmt.Sprintf("レビュー要請: %s が完成したとのことですが、以下を確認して報告してください：1. ビルドエラーがないか、2. コードの品質、3. 設計の一貫性。問題があれば具体的な修正指示をお願いします。", filePath)
	// レビューは子ペインに送信
	if len(m.taskTracker.AssignedPanes) > 0 {
		return m.sessionManager.SendToPane(m.taskTracker.AssignedPanes[0], reviewMessage)
	}
	return fmt.Errorf("レビュー対象の子ペインが見つかりません")
}

func (m *AIManager) SendIntegrationTest() error {
	panes, err := m.sessionManager.GetPanes()
	if err != nil || len(panes) < 2 {
		return fmt.Errorf("no available panes for integration test")
	}

	testMessage := "統合テスト実行: 全体のビルドテストを実行し、go build -o bin/ccs が成功することを確認してください。エラーがあれば詳細を報告してください。"
	return m.sessionManager.SendToPane(panes[1], testMessage)
}

func (m *AIManager) GetTaskSummary() string {
	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("=== タスク管理サマリー ===\n"))
	summary.WriteString(fmt.Sprintf("メインタスク: %s\n", m.taskTracker.MainTask.Description))
	summary.WriteString(fmt.Sprintf("サブタスク総数: %d\n", len(m.taskTracker.SubTasks)))

	pending := m.taskTracker.GetPendingTasks()
	needsReview := m.taskTracker.GetTasksNeedingReview()

	summary.WriteString(fmt.Sprintf("保留中: %d, レビュー待ち: %d\n", len(pending), len(needsReview)))
	summary.WriteString(fmt.Sprintf("全タスク完了: %t\n", m.taskTracker.AllTasksCompleted()))

	return summary.String()
}

// detectNewPane creates a new pane and returns its ID directly using tmux -P -F option
func (m *AIManager) detectNewPane(paneCreationCommand string) (string, error) {
	// Debug: Log the command being executed
	fmt.Printf("🔍 Executing pane creation command: %s\n", paneCreationCommand)

	// Execute the pane creation command and capture the new pane ID
	cmd := exec.Command("bash", "-c", paneCreationCommand)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create pane: %v\nstderr: %s", err, stderr.String())
	}

	// Get the new pane ID from stdout
	newPaneID := strings.TrimSpace(stdout.String())

	// Debug: Log the received pane ID
	fmt.Printf("✅ New pane created with ID: %s\n", newPaneID)

	// Validate the pane ID format
	if !strings.HasPrefix(newPaneID, "%") {
		return "", fmt.Errorf("invalid pane ID format: %s", newPaneID)
	}

	// Add the new pane to tracking if it's not a parent pane
	if !m.isParentPane(newPaneID) {
		// Track this as a child pane
		m.taskTracker.AssignedPanes = append(m.taskTracker.AssignedPanes, newPaneID)
		fmt.Printf("📝 Tracked new child pane: %s\n", newPaneID)
	} else {
		fmt.Printf("⚠️ Warning: Created pane %s is marked as parent pane\n", newPaneID)
	}

	return newPaneID, nil
}

// isParentPane checks if a pane ID is a parent pane (統一フィルター使用)
func (m *AIManager) isParentPane(paneID string) bool {
	return m.paneFilter.IsParentPane(paneID)
}

// SendTaskToChildPane sends a task to a specific child pane with enhanced filtering
func (m *AIManager) SendTaskToChildPane(paneID, taskDescription string) error {
	// Use the enhanced task service for filtering and assignment
	if m.taskService != nil {
		// Validate and get the appropriate pane for the task
		assignedPaneID, err := m.taskService.FilterAndAssignTask(taskDescription, paneID)
		if err != nil {
			return fmt.Errorf("task filtering failed: %v", err)
		}

		// If the task was redirected, log it
		if assignedPaneID != paneID {
			fmt.Printf("🔄 Task automatically redirected from %s to %s\n", paneID, assignedPaneID)
		}

		// Send the task to the assigned pane
		return m.sessionManager.SendToFilteredPane(assignedPaneID, taskDescription)
	}

	// Fallback to legacy validation
	if m.isParentPane(paneID) {
		// Find or create a suitable child pane
		childPane, err := m.findOrCreateChildPane()
		if err != nil {
			return fmt.Errorf("cannot send implementation task to parent pane %s and failed to create child pane: %v", paneID, err)
		}
		paneID = childPane
		fmt.Printf("⚠️ Redirected task from parent pane to child pane %s\n", paneID)
	}

	// Send the task
	return m.sessionManager.SendToPane(paneID, taskDescription)
}

// findOrCreateChildPane finds an existing child pane or creates a new one
func (m *AIManager) findOrCreateChildPane() (string, error) {
	// Get current panes
	panes, err := m.sessionManager.GetPanes()
	if err != nil {
		return "", fmt.Errorf("failed to get panes: %v", err)
	}

	// Debug: Log available panes
	fmt.Printf("🔍 Checking %d available panes for child panes\n", len(panes))

	// Look for existing child panes
	for _, pane := range panes {
		if !m.isParentPane(pane) {
			fmt.Printf("✅ Found existing child pane: %s\n", pane)
			return pane, nil
		}
		fmt.Printf("⏭️ Skipping parent pane: %s\n", pane)
	}

	// No child pane found, create a new one
	fmt.Printf("🔨 No child pane found, creating new one\n")
	splitCmd := "tmux split-window -h -t claude-squad -P -F \"#{pane_id}\""
	newPaneID, err := m.detectNewPane(splitCmd)
	if err != nil {
		return "", fmt.Errorf("failed to create new child pane: %v", err)
	}

	// Start Claude in the new pane
	time.Sleep(500 * time.Millisecond)
	claudeStartCmd := fmt.Sprintf("tmux send-keys -t %s 'claude --dangerously-skip-permissions' Enter", newPaneID)
	if err := m.sessionManager.ExecuteCommand(claudeStartCmd); err != nil {
		return "", fmt.Errorf("failed to start Claude in new pane %s: %v", newPaneID, err)
	}

	// Wait for Claude to be ready
	fmt.Printf("⏳ Waiting for Claude to start in pane %s\n", newPaneID)
	time.Sleep(2 * time.Second)

	return newPaneID, nil
}

// ValidateAndEnforceTaskAssignment は統合されたタスク割り当て検証・強制システム（統一フィルター使用）
func (m *AIManager) ValidateAndEnforceTaskAssignment(taskDescription, requestedPaneID string) error {
	// タスク割り当ての妥当性を検証
	isValid, message, err := m.paneFilter.ValidateTaskAssignment(taskDescription, requestedPaneID)
	if err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}

	if !isValid {
		fmt.Printf("⚠️  %s\n", message)
		// 最適なペインを取得してリダイレクト
		bestPane, err := m.paneFilter.GetBestPaneForTask(taskDescription)
		if err != nil {
			return fmt.Errorf("failed to find suitable pane: %v", err)
		}
		fmt.Printf("🔄 Redirecting task to pane %s\n", bestPane)
		requestedPaneID = bestPane
	} else {
		fmt.Printf("✅ %s\n", message)
	}

	return m.sessionManager.SendToPane(requestedPaneID, taskDescription)
}

// isWorkerPane checks if a pane ID is a worker pane (統一フィルター使用)
func (m *AIManager) isWorkerPane(paneID string) bool {
	return m.paneFilter.IsWorkerPane(paneID)
}

// GetPaneStatistics はペイン統計情報を取得（統一フィルター使用）
func (m *AIManager) GetPaneStatistics() (map[string]interface{}, error) {
	return m.paneFilter.GetPaneStatistics()
}
