package session

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Manager struct {
	SessionName   string
	ClaudeCmd     string
	ParentPanes   map[string]bool  // 親ペイン追跡マップ
	InitialPanes  []string         // 初期ペイン状態
	mainTask      string           // メインタスク
}

func NewManager(sessionName, claudeCmd string) *Manager {
	return &Manager{
		SessionName:  sessionName,
		ClaudeCmd:    claudeCmd,
		ParentPanes:  make(map[string]bool),
		InitialPanes: []string{},
		mainTask:     "",
	}
}

func (m *Manager) SetMainTask(task string) {
	m.mainTask = task
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
完了報告手順: 
1. tmux send-keys -t %s で報告すること
2. 必ず送信先のペインでエンターを1秒後に送信してタスクを実行
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

==== 作業状況報告フォーマット ====
子ペインからの報告は以下の形式で受け取ります：
- 「実装完了：[ファイルパス] - [簡単な説明]」
- 「進捗報告：[進捗状況] - [現在の作業内容]」
- 「エラー報告：[エラー内容] - [支援要請]」

それでは、メインタスクの分析と子ペインへの作業委託を開始してください。`,
		claudePane,
		m.mainTask,
		claudePane,
		claudePane,
		claudePane,
		claudePane, claudePane)
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

	cmd := exec.Command("tmux", "split-window", "-v", "-t", m.SessionName + ":0.0")
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