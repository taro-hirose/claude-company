package commands

import (
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

func (c *DeployCommand) Execute() error {
	panes, err := c.manager.GetPanes()
	if err != nil {
		return fmt.Errorf("failed to get panes: %w", err)
	}

	if len(panes) < 2 {
		return fmt.Errorf("need at least 2 panes for AI mode (manager + workers)")
	}

	return c.executeAIMode(panes)
}

func (c *DeployCommand) executeAIMode(panes []string) error {
	workerPane := panes[1]

	// Set the main task in manager
	c.manager.SetMainTask(c.taskDesc)

	// Send task to manager pane using BuildManagerPrompt
	taskPrompt := c.manager.BuildManagerPrompt(workerPane)

	if err := c.manager.SendToPane(workerPane, taskPrompt); err != nil {
		return fmt.Errorf("failed to send task to manager pane: %w", err)
	}

	fmt.Printf("🎯 AI管理モード開始: 親ペイン %s がプロジェクトマネージャーとして機能します\n", workerPane)
	fmt.Printf("📋 役割分担: 親ペイン=マネジメント・レビュー専用, 子ペイン=実装作業専用\n")
	fmt.Printf("🔄 タスク: %s\n", c.taskDesc)
	fmt.Printf("⚡ ワーカーペイン %s で実装作業を行ってください\n", workerPane)

	return nil
}
