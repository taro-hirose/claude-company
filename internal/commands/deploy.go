package commands

import (
	"context"
	"claude-company/internal/session"
	"fmt"
)

type DeployCommand struct {
	taskDesc string
	manager  *session.Manager
}

func NewDeployCommand(taskDesc string, manager *session.Manager) *DeployCommand {
	return &DeployCommand{
		taskDesc: taskDesc,
		manager:  manager,
	}
}

func (c *DeployCommand) Execute(ctx context.Context) error {
	panes, err := c.manager.GetPanes()
	if err != nil {
		return fmt.Errorf("failed to get panes: %w", err)
	}

	if len(panes) < 2 {
		return fmt.Errorf("need at least 2 panes for AI mode (manager + workers)")
	}

	// Check if orchestrator mode is enabled
	if c.manager.IsOrchestratorMode() {
		return c.executeOrchestratorMode(ctx, panes)
	}

	return c.executeTraditionalMode(panes)
}

func (c *DeployCommand) executeTraditionalMode(panes []string) error {
	workerPane := panes[1]

	// Set the main task in manager
	c.manager.SetMainTask(c.taskDesc)

	// Send task to manager pane using traditional manager prompt
	taskPrompt := c.manager.GetPromptForMode(workerPane)

	if err := c.manager.SendToPane(workerPane, taskPrompt); err != nil {
		return fmt.Errorf("failed to send task to manager pane: %w", err)
	}

	fmt.Printf("🎯 AI管理モード開始: 親ペイン %s がプロジェクトマネージャーとして機能します\n", workerPane)
	fmt.Printf("📋 役割分担: 親ペイン=マネジメント・レビュー専用, 子ペイン=実装作業専用\n")
	fmt.Printf("🔄 タスク: %s\n", c.taskDesc)
	fmt.Printf("⚡ ワーカーペイン %s で実装作業を行ってください\n", workerPane)

	return nil
}

func (c *DeployCommand) executeOrchestratorMode(ctx context.Context, panes []string) error {
	workerPane := panes[1]

	// Set the main task in manager
	c.manager.SetMainTask(c.taskDesc)

	// Initialize orchestrator if not already done
	if err := c.manager.InitializeOrchestrator(ctx); err != nil {
		return fmt.Errorf("failed to initialize orchestrator: %w", err)
	}

	// Send orchestrator prompt to manager pane
	orchestratorPrompt := c.manager.GetPromptForMode(workerPane)
	if err := c.manager.SendToPane(workerPane, orchestratorPrompt); err != nil {
		return fmt.Errorf("failed to send orchestrator prompt: %w", err)
	}

	fmt.Printf("🎯 オーケストレーターモード開始: 親ペイン %s がAIタスクオーケストレーターとして機能します\n", workerPane)
	fmt.Printf("📋 役割分担: 親ペイン=オーケストレーション専用, 子ペイン=ステップベース実装専用\n")
	fmt.Printf("🔄 タスク: %s\n", c.taskDesc)
	fmt.Printf("⚡ ワーカーペイン %s でステップベース実行を行ってください\n", workerPane)
	fmt.Printf("🧠 機能: 自動ステップ分解、並列実行最適化、品質監視、自動リトライ\n")
	fmt.Printf("📊 モード: %s\n", c.manager.GetModeStatus())

	return nil
}

// Legacy method maintained for backwards compatibility
func (c *DeployCommand) executeAIMode(panes []string) error {
	return c.executeTraditionalMode(panes)
}
