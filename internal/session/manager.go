package session

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"claude-company/internal/orchestrator"
)

type Manager struct {
	SessionName      string
	ClaudeCmd        string
	ParentPanes      map[string]bool               // 親ペイン追跡マップ
	InitialPanes     []string                      // 初期ペイン状態
	mainTask         string                        // メインタスク
	orchestratorMode bool                          // オーケストレーターモードフラグ
	orchestrator     orchestrator.Orchestrator     // オーケストレーターインスタンス
	currentTask      *orchestrator.Task            // 現在実行中のタスク
	stepManager      *orchestrator.StepManager     // ステップマネージャー
	taskPlanManager  *orchestrator.TaskPlanManager // タスクプランマネージャー
}

func NewManager(sessionName, claudeCmd string) *Manager {
	return &Manager{
		SessionName:      sessionName,
		ClaudeCmd:        claudeCmd,
		ParentPanes:      make(map[string]bool),
		InitialPanes:     []string{},
		mainTask:         "",
		orchestratorMode: false,
	}
}

func (m *Manager) SetMainTask(task string) {
	m.mainTask = task
}

// SetOrchestratorMode enables or disables orchestrator mode
func (m *Manager) SetOrchestratorMode(enabled bool) {
	m.orchestratorMode = enabled
}

// IsOrchestratorMode returns whether orchestrator mode is enabled
func (m *Manager) IsOrchestratorMode() bool {
	return m.orchestratorMode
}

// InitializeOrchestrator initializes the orchestrator system
func (m *Manager) InitializeOrchestrator(ctx context.Context) error {
	if m.orchestrator != nil {
		return nil // Already initialized
	}

	// Create event bus (mock implementation for now)
	eventBus := &mockEventBus{}

	// Create storage (mock implementation for now)
	storage := &mockStorage{}

	// Initialize step manager
	stepConfig := orchestrator.StepManagerConfig{
		MaxConcurrentSteps: 5,
		StepTimeout:        30 * time.Minute,
		ExecutorPoolSize:   3,
		RetryPolicy: orchestrator.RetryPolicy{
			MaxRetries:     3,
			InitialBackoff: 1 * time.Second,
			MaxBackoff:     30 * time.Second,
			BackoffFactor:  2.0,
		},
	}
	m.stepManager = orchestrator.NewStepManager(eventBus, storage, stepConfig)

	// Initialize task plan manager
	m.taskPlanManager = orchestrator.NewTaskPlanManager(eventBus, storage, m.stepManager)

	fmt.Println("✅ Orchestrator system initialized")
	return nil
}

func (m *Manager) parseOutputLines(output []byte) []string {
	lines := []string{}
	current := ""
	for _, b := range output {
		if b == '\n' {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
		} else {
			current += string(b)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func (m *Manager) BuildManagerPrompt(claudePane string) string {
	_, _ = m.GetPanes()

	return fmt.Sprintf(`
ultrathink

プロジェクトマネージャー(%s)として機能してください。

## 制限事項
禁止: コード編集、ファイル操作、ビルド、テスト、デプロイ、技術実装
許可: コード解析、タスク分析・分解、割り当て、進捗管理、品質管理、統合判定

## メインタスク
%s

## 管理フロー
1. コードの理解
2. タスク分析→サブタスク分解
3. 子ペイン作成(並行可能なら複数)
4. サブタスク割り当て
5. 子ペインに依頼したサブタスクの進捗監視・成果物レビュー
6. 統合テスト指示・完了判定

## ペイン操作
**重要**: 新ペインIDのみに送信、親ペイン(%s)は管理専用なので'claude --dangerously-skip-permissions'の送信は不可
**作成**: tmux split-window -v -t claude-squad
**起動**: tmux send-keys -t 新ペインID 'claude --dangerously-skip-permissions' Enter
**送信**: tmux send-keys -t 新ペインID Enter

サブタスクを作成するときの起動、送信は必須

## サブタスク送信
**重要**: 子ペインのみに送信、親ペイン(%s)は管理専用なのでサブタスクの送信は不可

テンプレート:
`+"`"+`
サブタスク: [タスク名]
目的: [達成目標]
成果物: [具体的な成果物]
完了条件: [完了基準]
報告方法: tmux send-keys -t %s '[報告内容]' Enter; sleep 1; tmux send-keys -t %s '' Enter
送信方法: tmux send-keys -t %s Enter

報告の時の送信は必須
`+"`"+`

## 進捗管理
- 定期進捗確認
- 完了報告時のレビュー指示
- 問題発生時の修正指示
- 全体統合テスト指示

## 報告フォーマット
- 実装完了: [ファイルパス] - [説明]
- 進捗報告: [状況] - [作業内容]
- エラー報告: [内容] - [支援要請]

メインタスクの分析とサブタスク委託を開始してください。`,
		claudePane,
		m.mainTask,
		claudePane,
		claudePane,
		claudePane,
		claudePane,
		claudePane)
}

// BuildOrchestratorPrompt builds the orchestrator-specific prompt
func (m *Manager) BuildOrchestratorPrompt(claudePane string) string {
	_, _ = m.GetPanes()

	return fmt.Sprintf(`
ultrathink

AIタスクオーケストレーター(%s)として機能してください。

## 制限事項
禁止: コード編集、ファイル操作、ビルド、テスト、デプロイ、技術実装
許可: タスク分析、計画立案、ステップベース実行管理、進捗監視、品質管理

## メインタスク
%s

## オーケストレーション機能
1. タスク分析と計画立案
2. ステップベースのタスク分解
3. 並列実行可能な作業の特定
4. 依存関係の解決
5. 進捗監視とレポート
6. 品質保証とレビュー

## 実行戦略
- **Sequential**: 依存関係がある場合の逐次実行
- **Parallel**: 独立した作業の並列実行  
- **Hybrid**: 依存関係を考慮した最適化実行

## ペイン操作（従来通り）
**作成**: tmux split-window -v -t claude-squad
**起動**: tmux send-keys -t 新ペインID 'claude --dangerously-skip-permissions' Enter
**送信**: tmux send-keys -t 新ペインID Enter
※送信は起動の1秒後に実行することを必須とする

## ステップベースタスク管理
**重要**: 子ペイン(%s以外)のみに送信、親ペイン(%s)は管理専用

新しいステップベーステンプレート:
`+"`"+`
サブタスク: [タスク名]
目的: [達成目標]
成果物: [具体的な成果物]
完了条件: [完了基準]
依存関係: [前提となるタスク]
実行戦略: [Sequential/Parallel/Hybrid]
報告方法: tmux send-keys -t %s '[報告内容]' Enter; sleep 1; tmux send-keys -t %s '' Enter
送信方法: tmux send-keys -t %s Enter
※送信は報告の1秒後に実行することを必須とする。
`+"`"+`

従来テンプレート（後方互換性維持）:
`+"`"+`
サブタスク: [タスク名]
目的: [達成目標]
成果物: [具体的な成果物]
完了条件: [完了基準]
報告方法: tmux send-keys -t %s '[報告内容]' Enter; sleep 1; tmux send-keys -t %s '' Enter
送信方法: tmux send-keys -t %s Enter
※送信は必須
`+"`"+`

## 進捗管理の強化
- リアルタイム進捗トラッキング
- ステップ完了の自動検出
- 並列タスクの同期管理
- エラー発生時の自動リトライ
- 全体統合の品質チェック

## 報告フォーマット（拡張）
- 実装完了: [ファイルパス] - [説明]
- ステップ完了: [ステップ名] - [成果物]
- 進捗報告: [全体進捗%%] - [現在のステップ]
- 並列完了: [タスク群] - [同期状況]
- エラー報告: [内容] - [リトライ状況]

## オーケストレーター特有の指示
1. 最初にタスクを分析し、最適な実行計画を立案
2. 依存関係グラフを作成して並列化を最大化
3. ステップごとの完了を確認して次のステップに進行
4. 全体の進捗を定期的にレポート
5. 最終的な統合テストで品質を保証

メインタスクの分析とステップベース実行計画の立案を開始してください。`,
		claudePane,
		m.mainTask,
		claudePane,
		claudePane,
		claudePane,
		claudePane,
		claudePane,
		claudePane,
		claudePane,
		claudePane)
}

func (m *Manager) Setup() error {
	if _, err := exec.LookPath("tmux"); err != nil {
		return fmt.Errorf("❌ Error: tmux is not installed")
	}

	// 初期状態のペインを記録
	if err := m.recordInitialPanes(); err != nil {
		return fmt.Errorf("failed to record initial panes: %v", err)
	}

	cmd := exec.Command("tmux", "has-session", "-t", m.SessionName)
	if cmd.Run() == nil {
		fmt.Printf("🔄 Session '%s' already exists.\n", m.SessionName)

		fmt.Println("📊 Current pane status:")
		statusCmd := exec.Command("tmux", "list-panes", "-s", "-t", m.SessionName, "-F", "#{pane_index}: #{pane_id} #{pane_current_command}")
		if output, err := statusCmd.Output(); err == nil {
			fmt.Print(string(output))
		}

		return m.attach()
	}

	fmt.Printf("🚀 Creating new Claude Code Company session '%s'...\n", m.SessionName)

	if err := m.createSession(); err != nil {
		return err
	}

	fmt.Println("📐 Setting up pane layout...")
	if err := m.setupPanes(); err != nil {
		return err
	}

	time.Sleep(time.Second)

	if err := m.startClaudeSessions(); err != nil {
		return err
	}

	if err := m.setupMainPane(); err != nil {
		return err
	}

	fmt.Println("✅ Claude Code Company setup completed!")

	return m.attach()
}

func (m *Manager) createSession() error {
	cmd := exec.Command("tmux", "new-session", "-d", "-s", m.SessionName, "-n", "main")
	return cmd.Run()
}

func (m *Manager) setupPanes() error {
	commands := [][]string{
		{"tmux", "split-window", "-v", "-t", m.SessionName + ":0.0"},
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to execute %v: %w", cmdArgs, err)
		}
	}

	return nil
}

func (m *Manager) startClaudeSessions() error {
	cmd := exec.Command("tmux", "list-panes", "-s", "-t", m.SessionName, "-F", "#{pane_id}")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := m.parseOutputLines(output)

	if len(lines) > 1 {
		bottomPaneID := lines[1]
		fmt.Printf("🤖 Starting Claude Code in bottom pane %s...\n", bottomPaneID)
		cmd := exec.Command("tmux", "send-keys", "-t", bottomPaneID, m.ClaudeCmd, "Enter")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start Claude in pane %s: %w", bottomPaneID, err)
		}
	}

	return nil
}

func (m *Manager) setupMainPane() error {
	cmd := exec.Command("tmux", "list-panes", "-s", "-t", m.SessionName, "-F", "#{pane_id}")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := m.parseOutputLines(output)

	if len(lines) == 0 {
		return fmt.Errorf("no panes found")
	}

	mainPaneID := lines[0]

	fmt.Println("📝 Setting up main pane with management commands...")

	selectCmd := exec.Command("tmux", "select-pane", "-t", mainPaneID)
	if err := selectCmd.Run(); err != nil {
		return err
	}

	helpCmd := exec.Command("tmux", "send-keys", "-t", mainPaneID, "echo '🚀 Claude Company Manager - Use deploy command to assign AI tasks'", "Enter")
	return helpCmd.Run()
}

func (m *Manager) attach() error {
	if os.Getenv("TMUX") != "" {
		fmt.Printf("🔄 Switching to session '%s'...\n", m.SessionName)
		cmd := exec.Command("tmux", "switch-client", "-t", m.SessionName)
		return cmd.Run()
	} else {
		fmt.Printf("🔗 Attaching to session '%s'...\n", m.SessionName)
		cmd := exec.Command("tmux", "attach-session", "-t", m.SessionName)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}

func (m *Manager) SendToPane(paneID, command string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", paneID, command, "Enter")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tmux command failed: %v, output: %s", err, string(output))
	}

	fmt.Printf("Task assigned to pane %s\n", paneID)
	return nil
}

func (m *Manager) SendToNewPaneOnly(command string) error {
	newPaneID, err := m.CreateNewPaneAndGetID()
	if err != nil {
		return fmt.Errorf("failed to create new pane: %v", err)
	}

	if err := m.StartClaudeInNewPane(newPaneID); err != nil {
		return fmt.Errorf("failed to start Claude in new pane: %v", err)
	}

	cmd := exec.Command("tmux", "send-keys", "-t", newPaneID, command, "Enter")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tmux command failed: %v, output: %s", err, string(output))
	}

	fmt.Printf("📤 Task assigned to new pane %s only\n", newPaneID)
	return nil
}

func (m *Manager) GetPanes() ([]string, error) {
	cmd := exec.Command("tmux", "list-panes", "-s", "-t", m.SessionName, "-F", "#{pane_id}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get panes: %v", err)
	}

	return m.parseOutputLines(output), nil
}

func (m *Manager) GetAllPanes() ([]string, error) {
	cmd := exec.Command("tmux", "list-panes", "-a", "-F", "#{pane_id}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get all panes: %v", err)
	}

	return strings.Fields(strings.TrimSpace(string(output))), nil
}

func (m *Manager) CreateNewPaneAndGetID() (string, error) {
	beforePanes, err := m.GetAllPanes()
	if err != nil {
		return "", fmt.Errorf("failed to get panes before creation: %v", err)
	}

	cmd := exec.Command("tmux", "split-window", "-v", "-t", m.SessionName+":0.0")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create new pane: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	afterPanes, err := m.GetAllPanes()
	if err != nil {
		return "", fmt.Errorf("failed to get panes after creation: %v", err)
	}

	for _, afterPane := range afterPanes {
		found := false
		for _, beforePane := range beforePanes {
			if afterPane == beforePane {
				found = true
				break
			}
		}
		if !found {
			return afterPane, nil
		}
	}

	return "", fmt.Errorf("failed to identify new pane ID")
}

func (m *Manager) StartClaudeInNewPane(paneID string) error {
	fmt.Printf("🤖 Starting Claude Code in new pane %s...\n", paneID)
	cmd := exec.Command("tmux", "send-keys", "-t", paneID, m.ClaudeCmd, "Enter")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start Claude in pane %s: %w", paneID, err)
	}

	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		if m.isClaudeReady(paneID) {
			fmt.Printf("✅ Claude is ready in pane %s\n", paneID)
			return nil
		}
		fmt.Printf("⏳ Waiting for Claude to start in pane %s... (%d/10)\n", paneID, i+1)
	}

	return fmt.Errorf("Claude failed to start within timeout in pane %s", paneID)
}

func (m *Manager) isClaudeReady(paneID string) bool {
	cmd := exec.Command("tmux", "capture-pane", "-t", paneID, "-p")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	content := string(output)
	return strings.Contains(content, "claude") || strings.Contains(content, "ready") || strings.Contains(content, "$")
}

// recordInitialPanes は初期状態のペインを記録し、親ペインとして設定
func (m *Manager) recordInitialPanes() error {
	panes, err := m.GetAllPanes()
	if err != nil {
		// セッションが存在しない場合は問題なし
		return nil
	}

	m.InitialPanes = make([]string, len(panes))
	copy(m.InitialPanes, panes)

	// 初期ペインを親ペインとして記録
	for _, pane := range panes {
		m.ParentPanes[pane] = true
	}

	fmt.Printf("🔍 Recorded %d initial parent panes\n", len(panes))
	return nil
}

// IsParentPane は指定されたペインが親ペインかどうかを判定
func (m *Manager) IsParentPane(paneID string) bool {
	return m.ParentPanes[paneID]
}

// IsChildPane は指定されたペインが子ペインかどうかを判定（差分検出）
func (m *Manager) IsChildPane(paneID string) bool {
	return !m.ParentPanes[paneID]
}

// GetChildPanes は子ペイン一覧を取得
func (m *Manager) GetChildPanes() ([]string, error) {
	allPanes, err := m.GetPanes()
	if err != nil {
		return nil, err
	}

	var childPanes []string
	for _, pane := range allPanes {
		if m.IsChildPane(pane) {
			childPanes = append(childPanes, pane)
		}
	}

	return childPanes, nil
}

// SendToChildPaneOnly は子ペインにのみタスクを送信
func (m *Manager) SendToChildPaneOnly(command string) error {
	childPanes, err := m.GetChildPanes()
	if err != nil {
		return fmt.Errorf("failed to get child panes: %v", err)
	}

	if len(childPanes) == 0 {
		// 子ペインが存在しない場合は新しく作成
		return m.SendToNewPaneOnly(command)
	}

	// 最初の子ペインに送信
	targetPane := childPanes[0]
	return m.SendToPane(targetPane, command)
}

// SendToFilteredPane はペインフィルタリング付きでタスクを送信
func (m *Manager) SendToFilteredPane(paneID, command string) error {
	if m.IsParentPane(paneID) {
		fmt.Printf("⚠️  Blocked task assignment to parent pane %s\n", paneID)
		fmt.Println("🔄 Redirecting to child pane...")
		return m.SendToChildPaneOnly(command)
	}

	fmt.Printf("✅ Task assigned to child pane %s\n", paneID)
	return m.SendToPane(paneID, command)
}

// CreateNewPaneAndRegisterAsChild は新しいペインを作成し子ペインとして登録
func (m *Manager) CreateNewPaneAndRegisterAsChild() (string, error) {
	newPaneID, err := m.CreateNewPaneAndGetID()
	if err != nil {
		return "", err
	}

	// 新しいペインは自動的に子ペインとして扱われる（parentPanesに含まれない）
	fmt.Printf("📝 Registered new child pane: %s\n", newPaneID)
	return newPaneID, nil
}

// ExecuteCommand executes a shell command directly
func (m *Manager) ExecuteCommand(command string) error {
	cmd := exec.Command("bash", "-c", command)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("command failed: %v, output: %s", err, string(output))
	}
	return nil
}

// CreateTask creates a new orchestrated task
func (m *Manager) CreateTask(ctx context.Context, req orchestrator.TaskRequest) (*orchestrator.TaskResponse, error) {
	if !m.orchestratorMode {
		return nil, fmt.Errorf("orchestrator mode is not enabled")
	}

	if m.orchestrator == nil {
		if err := m.InitializeOrchestrator(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize orchestrator: %w", err)
		}
	}

	// Create task using orchestrator
	resp, err := m.orchestrator.CreateTask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Store current task reference
	if task, err := m.orchestrator.GetTask(ctx, resp.TaskID); err == nil {
		m.currentTask = task
	}

	return resp, nil
}

// GetCurrentTask returns the currently active task
func (m *Manager) GetCurrentTask() *orchestrator.Task {
	return m.currentTask
}

// CreatePlanForCurrentTask creates a plan for the current task
func (m *Manager) CreatePlanForCurrentTask(ctx context.Context) (*orchestrator.TaskPlan, error) {
	if m.currentTask == nil {
		return nil, fmt.Errorf("no current task available")
	}

	if m.taskPlanManager == nil {
		return nil, fmt.Errorf("task plan manager not initialized")
	}

	plan := &orchestrator.TaskPlan{
		TaskID:    m.currentTask.ID,
		Strategy:  orchestrator.PlanStrategyHybrid,
		Steps:     []orchestrator.TaskStep{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := m.taskPlanManager.CreatePlan(ctx, plan)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

// ExecutePlan executes a task plan with step-based management
func (m *Manager) ExecutePlan(ctx context.Context, planID string) error {
	if m.taskPlanManager == nil {
		return fmt.Errorf("task plan manager not initialized")
	}

	return m.taskPlanManager.ExecutePlan(ctx, planID)
}

// SendTaskToPane sends an orchestrated task to a specific pane
func (m *Manager) SendTaskToPane(ctx context.Context, paneID string, task *orchestrator.Task) error {
	if m.IsParentPane(paneID) {
		fmt.Printf("⚠️  Blocked orchestrated task assignment to parent pane %s\n", paneID)
		fmt.Println("🔄 Redirecting to child pane...")
		return m.SendTaskToChildPane(ctx, task)
	}

	// Build task command based on mode
	var command string
	if m.orchestratorMode {
		command = m.buildOrchestratedTaskCommand(task)
	} else {
		command = m.buildTraditionalTaskCommand(task)
	}

	fmt.Printf("✅ Orchestrated task assigned to child pane %s\n", paneID)
	return m.SendToPane(paneID, command)
}

// SendTaskToChildPane sends a task to any available child pane
func (m *Manager) SendTaskToChildPane(ctx context.Context, task *orchestrator.Task) error {
	childPanes, err := m.GetChildPanes()
	if err != nil {
		return fmt.Errorf("failed to get child panes: %v", err)
	}

	if len(childPanes) == 0 {
		// Create new pane if no child panes exist
		newPaneID, err := m.CreateNewPaneAndRegisterAsChild()
		if err != nil {
			return fmt.Errorf("failed to create new pane: %v", err)
		}

		if err := m.StartClaudeInNewPane(newPaneID); err != nil {
			return fmt.Errorf("failed to start Claude in new pane: %v", err)
		}

		return m.SendTaskToPane(ctx, newPaneID, task)
	}

	// Use the first available child pane
	return m.SendTaskToPane(ctx, childPanes[0], task)
}

// buildOrchestratedTaskCommand builds a command string for orchestrated tasks
func (m *Manager) buildOrchestratedTaskCommand(task *orchestrator.Task) string {
	return fmt.Sprintf(`サブタスク: %s
目的: %s
成果物: タスク完了時の具体的成果物
完了条件: %s
実行戦略: Hybrid
報告方法: tmux send-keys -t %%1 "実装完了: %s - %s" Enter; sleep 1; tmux send-keys -t %%1 "" Enter`,
		task.Title,
		task.Description,
		"実装とテストが完了していること",
		task.Title,
		"実装完了")
}

// buildTraditionalTaskCommand builds a command string for traditional tasks
func (m *Manager) buildTraditionalTaskCommand(task *orchestrator.Task) string {
	return fmt.Sprintf(`サブタスク: %s
目的: %s
成果物: タスク完了時の具体的成果物
完了条件: %s
報告方法: tmux send-keys -t %%1 "実装完了: %s - %s" Enter; sleep 1; tmux send-keys -t %%1 "" Enter`,
		task.Title,
		task.Description,
		"実装とテストが完了していること",
		task.Title,
		"実装完了")
}

// GetPromptForMode returns the appropriate prompt based on the current mode
func (m *Manager) GetPromptForMode(claudePane string) string {
	if m.orchestratorMode {
		return m.BuildOrchestratorPrompt(claudePane)
	}
	return m.BuildManagerPrompt(claudePane)
}

// ToggleOrchestratorMode toggles between orchestrator and traditional manager mode
func (m *Manager) ToggleOrchestratorMode(ctx context.Context) error {
	m.orchestratorMode = !m.orchestratorMode

	if m.orchestratorMode {
		fmt.Println("🔄 Switching to Orchestrator Mode...")
		if err := m.InitializeOrchestrator(ctx); err != nil {
			m.orchestratorMode = false // Revert on error
			return fmt.Errorf("failed to initialize orchestrator: %w", err)
		}
		fmt.Println("✅ Orchestrator Mode enabled")
	} else {
		fmt.Println("🔄 Switching to Traditional Manager Mode...")
		fmt.Println("✅ Traditional Manager Mode enabled")
	}

	return nil
}

// GetModeStatus returns the current mode status
func (m *Manager) GetModeStatus() string {
	if m.orchestratorMode {
		return "Orchestrator Mode (Step-based execution)"
	}
	return "Traditional Manager Mode (Basic task delegation)"
}

// Mock implementations for orchestrator interfaces
type mockEventBus struct{}

func (m *mockEventBus) Publish(ctx context.Context, event orchestrator.TaskEvent) error {
	fmt.Printf("📡 Event: %s for task %s\n", event.Type, event.TaskID)
	return nil
}

func (m *mockEventBus) Subscribe(ctx context.Context, eventTypes []orchestrator.TaskEventType) (<-chan orchestrator.TaskEvent, error) {
	ch := make(chan orchestrator.TaskEvent, 10)
	return ch, nil
}

func (m *mockEventBus) Unsubscribe(ctx context.Context, subscription string) error {
	return nil
}

func (m *mockEventBus) AddFilter(ctx context.Context, filter orchestrator.EventFilter) error {
	return nil
}

func (m *mockEventBus) RemoveFilter(ctx context.Context, filterID string) error {
	return nil
}

type mockStorage struct{}

func (m *mockStorage) SaveTask(ctx context.Context, task *orchestrator.Task) error {
	return nil
}

func (m *mockStorage) LoadTask(ctx context.Context, taskID string) (*orchestrator.Task, error) {
	return nil, fmt.Errorf("task not found")
}

func (m *mockStorage) ListTasks(ctx context.Context, filter orchestrator.TaskFilter) ([]*orchestrator.Task, error) {
	return []*orchestrator.Task{}, nil
}

func (m *mockStorage) DeleteTask(ctx context.Context, taskID string) error {
	return nil
}

func (m *mockStorage) SavePlan(ctx context.Context, plan *orchestrator.TaskPlan) error {
	return nil
}

func (m *mockStorage) LoadPlan(ctx context.Context, planID string) (*orchestrator.TaskPlan, error) {
	return nil, fmt.Errorf("plan not found")
}

func (m *mockStorage) DeletePlan(ctx context.Context, planID string) error {
	return nil
}

func (m *mockStorage) SaveWorker(ctx context.Context, worker *orchestrator.Worker) error {
	return nil
}

func (m *mockStorage) LoadWorker(ctx context.Context, workerID string) (*orchestrator.Worker, error) {
	return nil, fmt.Errorf("worker not found")
}

func (m *mockStorage) ListWorkers(ctx context.Context) ([]*orchestrator.Worker, error) {
	return []*orchestrator.Worker{}, nil
}

func (m *mockStorage) DeleteWorker(ctx context.Context, workerID string) error {
	return nil
}

func (m *mockStorage) SaveEvent(ctx context.Context, event *orchestrator.TaskEvent) error {
	return nil
}

func (m *mockStorage) ListEvents(ctx context.Context, filter orchestrator.EventFilter) ([]*orchestrator.TaskEvent, error) {
	return []*orchestrator.TaskEvent{}, nil
}

func (m *mockStorage) Cleanup(ctx context.Context) error {
	return nil
}
