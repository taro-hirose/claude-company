package session

import (
	"claude-company/internal/utils"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Manager struct {
	SessionName   string
	ClaudeCmd     string
	ParentPanes   map[string]bool  // 親ペイン追跡マップ（レガシー）
	InitialPanes  []string         // 初期ペイン状態
	paneFilter    *utils.PaneFilter // 統一ペインフィルター
}

func NewManager(sessionName, claudeCmd string) *Manager {
	parentPanes := make(map[string]bool)
	return &Manager{
		SessionName:  sessionName,
		ClaudeCmd:    claudeCmd,
		ParentPanes:  parentPanes,
		InitialPanes: []string{},
		paneFilter:   utils.NewPaneFilterWithLegacySupport(parentPanes),
	}
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
	// ペイン分割
	commands := [][]string{
		{"tmux", "split-window", "-v", "-t", m.SessionName + ":0.0"},
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to execute %v: %w", cmdArgs, err)
		}
	}

	// ペインタイトルを設定
	if err := m.setPaneTitles(); err != nil {
		return fmt.Errorf("failed to set pane titles: %w", err)
	}

	return nil
}

// setPaneTitles は各ペインに適切なタイトルを設定
func (m *Manager) setPaneTitles() error {
	// 現在のペイン一覧を取得
	cmd := exec.Command("tmux", "list-panes", "-s", "-t", m.SessionName, "-F", "#{pane_id}|#{pane_index}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get panes: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		parts := strings.Split(line, "|")
		if len(parts) != 2 {
			continue
		}
		
		paneID := parts[0]
		paneIndex := parts[1]
		
		var title string
		switch paneIndex {
		case "0":
			// 上部ペイン（最初のペイン）をコンソールペインとして設定
			title = "[CONSOLE] コンソールペイン"
		case "1":
			// 下部ペイン（分割されたペイン）をマネージャーペインとして設定
			title = "[MANAGER] マネージャーペイン"
		default:
			// その他のペインはワーカーペインとして設定
			title = "[WORKER] ワーカーペイン"
		}
		
		titleCmd := exec.Command("tmux", "select-pane", "-t", paneID, "-T", title)
		if err := titleCmd.Run(); err != nil {
			fmt.Printf("⚠️  Warning: failed to set title for pane %s: %v\n", paneID, err)
		} else {
			fmt.Printf("✅ Set title for pane %s (%s): %s\n", paneID, paneIndex, title)
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

	panes := []string{}
	current := ""
	for _, b := range output {
		if b == '\n' {
			if current != "" {
				panes = append(panes, current)
				current = ""
			}
		} else {
			current += string(b)
		}
	}
	if current != "" {
		panes = append(panes, current)
	}

	return panes, nil
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

	var newPaneID string
	for _, afterPane := range afterPanes {
		found := false
		for _, beforePane := range beforePanes {
			if afterPane == beforePane {
				found = true
				break
			}
		}
		if !found {
			newPaneID = afterPane
			break
		}
	}

	if newPaneID == "" {
		return "", fmt.Errorf("failed to identify new pane ID")
	}

	// 新しいペインにワーカータイトルを設定
	if err := m.setWorkerPaneTitle(newPaneID); err != nil {
		fmt.Printf("⚠️  Warning: failed to set title for new pane %s: %v\n", newPaneID, err)
	}

	return newPaneID, nil
}

// setWorkerPaneTitle は新しく作成されたペインにワーカータイトルを設定
func (m *Manager) setWorkerPaneTitle(paneID string) error {
	title := "[WORKER] ワーカーペイン"
	titleCmd := exec.Command("tmux", "select-pane", "-t", paneID, "-T", title)
	if err := titleCmd.Run(); err != nil {
		return fmt.Errorf("failed to set worker title: %v", err)
	}
	fmt.Printf("✅ Set worker title for new pane %s: %s\n", paneID, title)
	return nil
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

// IsParentPane は指定されたペインが親ペインかどうかを判定（統一フィルター使用）
func (m *Manager) IsParentPane(paneID string) bool {
	return m.paneFilter.IsParentPane(paneID)
}

// IsChildPane は指定されたペインが子ペインかどうかを判定（統一フィルター使用）
func (m *Manager) IsChildPane(paneID string) bool {
	return m.paneFilter.IsChildPane(paneID)
}

// IsWorkerPane はワーカーペインかどうかを判定（統一フィルター使用）
func (m *Manager) IsWorkerPane(paneID string) bool {
	return m.paneFilter.IsWorkerPane(paneID)
}

// IsManagerPane は管理ペインかどうかを判定（統一フィルター使用）
func (m *Manager) IsManagerPane(paneID string) bool {
	return m.paneFilter.IsManagerPane(paneID)
}

// IsConsolePane はコンソールペインかどうかを判定（統一フィルター使用）
func (m *Manager) IsConsolePane(paneID string) bool {
	return m.paneFilter.IsConsolePane(paneID)
}

// GetPaneType はペインタイプを取得（統一フィルター使用）
func (m *Manager) GetPaneType(paneID string) utils.PaneType {
	return m.paneFilter.GetPaneType(paneID)
}

// GetChildPanes は子ペイン一覧を取得（統一フィルター使用）
func (m *Manager) GetChildPanes() ([]string, error) {
	return m.paneFilter.GetWorkerPanes()
}

// GetManagerPanes は管理ペイン一覧を取得（統一フィルター使用）
func (m *Manager) GetManagerPanes() ([]string, error) {
	return m.paneFilter.GetManagerPanes()
}

// GetConsolePanes はコンソールペイン一覧を取得（統一フィルター使用）
func (m *Manager) GetConsolePanes() ([]string, error) {
	return m.paneFilter.GetConsolePanes()
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

// SendToFilteredPane はペインフィルタリング付きでタスクを送信（統一フィルター使用）
func (m *Manager) SendToFilteredPane(paneID, command string) error {
	// タスク割り当ての妥当性を検証
	isValid, message, err := m.paneFilter.ValidateTaskAssignment(command, paneID)
	if err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}
	
	if !isValid {
		fmt.Printf("⚠️  %s\n", message)
		// 最適なペインを取得
		bestPane, err := m.paneFilter.GetBestPaneForTask(command)
		if err != nil {
			return fmt.Errorf("failed to find suitable pane: %v", err)
		}
		fmt.Printf("🔄 Redirecting task to pane %s\n", bestPane)
		paneID = bestPane
	} else {
		fmt.Printf("✅ %s\n", message)
	}
	
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