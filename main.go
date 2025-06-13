package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

type Task struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Mode        string `json:"mode"`
	PaneID      string `json:"pane_id"`
	Status      string `json:"status"`
}

type ClaudeCompany struct {
	sessionName string
	claudeCmd   string
}

func main() {
	var setup bool
	var aiMode, simpleMode bool
	var taskDesc, paneID string
	
	flag.BoolVar(&setup, "setup", false, "Setup Claude Company tmux session")
	flag.BoolVar(&aiMode, "ai", false, "Enable AI-assisted mode")
	flag.BoolVar(&simpleMode, "simple", false, "Enable simple mode")
	flag.StringVar(&taskDesc, "task", "", "Task description")
	flag.StringVar(&paneID, "pane", "", "Target pane ID")
	flag.Parse()

	// Default behavior: setup tmux session
	if len(os.Args) == 1 || setup {
		cc := &ClaudeCompany{
			sessionName: "claude-squad",
			claudeCmd:   "claude --dangerously-skip-permissions",
		}
		
		if err := cc.setupSession(); err != nil {
			log.Fatal(err)
		}
		return
	}

	if taskDesc == "" {
		log.Fatal("Task description is required (-task)")
	}

	if paneID == "" && !aiMode {
		log.Fatal("Pane ID is required (-pane) for non-AI mode")
	}

	if !aiMode && !simpleMode {
		aiMode = true
	}

	deploy := DeployCommand{
		aiMode:   aiMode,
		simpleMode: simpleMode,
		taskDesc: taskDesc,
		paneID:   paneID,
	}

	if err := deploy.execute(); err != nil {
		log.Fatal(err)
	}
}

type DeployCommand struct {
	aiMode     bool
	simpleMode bool
	taskDesc   string
	paneID     string
}

func (c *DeployCommand) execute() error {
	var mode string
	if c.aiMode {
		mode = "ai"
	} else {
		mode = "simple"
	}

	task := Task{
		ID:          generateTaskID(),
		Description: c.taskDesc,
		Mode:        mode,
		PaneID:      c.paneID,
		Status:      "assigned",
	}

	if mode == "ai" {
		return c.executeAIMode(task)
	}
	return c.executeSimpleMode(task)
}

func (c *DeployCommand) executeAIMode(task Task) error {
	// Get all panes from claude-squad session
	cmd := exec.Command("tmux", "list-panes", "-s", "-t", "claude-squad", "-F", "#{pane_id}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get panes: %v", err)
	}

	// Parse panes
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

	if len(panes) < 2 {
		return fmt.Errorf("need at least 2 panes for AI mode (manager + workers)")
	}

	// First pane is always the manager (parent pane)
	managerPane := panes[0]
	claudePane := panes[1] // Bottom pane with Claude
	
	// Update task to use the manager pane as parent
	task.PaneID = managerPane

	// Send AI manager prompt to first pane (top pane)
	aiManagerPrompt := fmt.Sprintf(`あなたは%s（親ペイン・マネージャー）です。

==== メインタスク ====
%s

==== あなたの役割 ====
1. このメインタスクを分析し、効率的にサブタスクに分解する
2. 必要に応じて子ペインを動的に作成する
3. 各子ペインに適切なサブタスクを自動で割り当てる
4. 進捗を管理し、最終的に統合・完了させる

==== 利用可能なClaude実行ペイン ====
%s

==== 子ペイン作成方法 ====
必要に応じてtmux split-windowコマンドで新しい子ペインを作成できます：
例：
- 横分割: tmux split-window -h -t claude-squad
- 縦分割: tmux split-window -v -t claude-squad
- 特定ペインを分割: tmux split-window -h -t %s

==== 新規ペイン作成後の手順 ====
新しいペインを作成したら、必ずClaude AIを起動してください：
1. ペイン作成後：tmux send-keys -t 新ペインID 'claude --dangerously-skip-permissions' Enter
2. Claude起動確認後にタスクを送信

==== タスク送信方法 ====
各子ペイン（%sまたは新規作成）にサブタスクを送信する場合は、以下のコマンド形式を使用してください：

tmux send-keys -t ペインID 'サブタスク: [具体的なサブタスク内容]' Enter

例：
tmux send-keys -t %s 'サブタスク: データベース設計を行い、結果を報告してください' Enter

==== 実行手順 ====
1. メインタスクを分析してサブタスクに分解
2. 必要に応じて子ペインを作成
3. 各子ペインに上記形式でサブタスクを送信
4. 進捗を管理し、最終的に統合・完了させる

それでは、メインタスクの分析と子ペインへの自動振り分けを開始してください。`, 
		managerPane, task.Description, claudePane, claudePane, claudePane)

	// Send manager prompt to Claude pane (bottom pane)
	if err := c.sendToPane(claudePane, aiManagerPrompt); err != nil {
		return err
	}

	fmt.Printf("AI助手モード開始: 親ペイン %s がタスクを分析し、子ペインを動的作成して配信します\n", managerPane)
	return nil
}

func (c *DeployCommand) executeSimpleMode(task Task) error {
	simpleCommand := fmt.Sprintf("echo \"タスク開始: %s (パネル: %s)\"", task.Description, task.PaneID)
	
	return c.sendToPane(task.PaneID, simpleCommand)
}

func (c *DeployCommand) sendToPane(paneID, command string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", paneID, command, "Enter")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tmux command failed: %v, output: %s", err, string(output))
	}
	
	fmt.Printf("Task assigned to pane %s\n", paneID)
	return nil
}

func generateTaskID() string {
	return fmt.Sprintf("task_%d", os.Getpid())
}

func (cc *ClaudeCompany) setupSession() error {
	// Check if tmux is installed
	if _, err := exec.LookPath("tmux"); err != nil {
		return fmt.Errorf("❌ Error: tmux is not installed")
	}

	// Check if session already exists
	cmd := exec.Command("tmux", "has-session", "-t", cc.sessionName)
	if cmd.Run() == nil {
		fmt.Printf("🔄 Session '%s' already exists.\n", cc.sessionName)
		
		// Show current pane status
		fmt.Println("📊 Current pane status:")
		statusCmd := exec.Command("tmux", "list-panes", "-s", "-t", cc.sessionName, "-F", "#{pane_index}: #{pane_id} #{pane_current_command}")
		if output, err := statusCmd.Output(); err == nil {
			fmt.Print(string(output))
		}
		
		// Attach to existing session
		return cc.attachSession()
	}

	fmt.Printf("🚀 Creating new Claude Code Company session '%s'...\n", cc.sessionName)

	// Create new session
	if err := cc.createSession(); err != nil {
		return err
	}

	// Setup pane layout
	fmt.Println("📐 Setting up pane layout...")
	if err := cc.setupPanes(); err != nil {
		return err
	}

	// Wait a bit for panes to be ready
	time.Sleep(time.Second)

	// Start Claude sessions in subordinate panes
	if err := cc.startClaudeSessions(); err != nil {
		return err
	}

	// Setup main pane
	if err := cc.setupMainPane(); err != nil {
		return err
	}

	fmt.Println("✅ Claude Code Company setup completed!")

	// Attach to session
	return cc.attachSession()
}

func (cc *ClaudeCompany) createSession() error {
	cmd := exec.Command("tmux", "new-session", "-d", "-s", cc.sessionName, "-n", "main")
	return cmd.Run()
}

func (cc *ClaudeCompany) setupPanes() error {
	commands := [][]string{
		{"tmux", "split-window", "-v", "-t", cc.sessionName + ":0.0"},
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to execute %v: %w", cmdArgs, err)
		}
	}

	return nil
}

func (cc *ClaudeCompany) startClaudeSessions() error {
	// Get subordinate panes (all except first)
	cmd := exec.Command("tmux", "list-panes", "-s", "-t", cc.sessionName, "-F", "#{pane_id}")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	// Parse output properly
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

	// Start Claude in the bottom pane (second pane)
	if len(lines) > 1 {
		bottomPaneID := lines[1]
		fmt.Printf("🤖 Starting Claude Code in bottom pane %s...\n", bottomPaneID)
		cmd := exec.Command("tmux", "send-keys", "-t", bottomPaneID, cc.claudeCmd, "Enter")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start Claude in pane %s: %w", bottomPaneID, err)
		}
	}

	return nil
}

func (cc *ClaudeCompany) setupMainPane() error {
	// Get main pane ID
	cmd := exec.Command("tmux", "list-panes", "-s", "-t", cc.sessionName, "-F", "#{pane_id}")
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
	
	// Select main pane
	selectCmd := exec.Command("tmux", "select-pane", "-t", mainPaneID)
	if err := selectCmd.Run(); err != nil {
		return err
	}

	// Send help command (if available) or show basic info
	helpCmd := exec.Command("tmux", "send-keys", "-t", mainPaneID, "echo '🚀 Claude Company Manager - Use deploy command to assign AI tasks'", "Enter")
	return helpCmd.Run()
}

func (cc *ClaudeCompany) attachSession() error {
	// Check if we're already in tmux
	if os.Getenv("TMUX") != "" {
		fmt.Printf("🔄 Switching to session '%s'...\n", cc.sessionName)
		cmd := exec.Command("tmux", "switch-client", "-t", cc.sessionName)
		return cmd.Run()
	} else {
		fmt.Printf("🔗 Attaching to session '%s'...\n", cc.sessionName)
		cmd := exec.Command("tmux", "attach-session", "-t", cc.sessionName)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}