package commands

import (
	"fmt"
	"claude-company/internal/models"
	"claude-company/internal/session"
	"claude-company/internal/utils"
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
	task := models.Task{
		ID:          utils.GenerateTaskID(),
		Description: c.taskDesc,
		Mode:        "ai",
		Status:      "assigned",
	}

	return c.executeAIMode(task)
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
	
	task.PaneID = managerPane

	// 新しいAIマネージャーを作成
	aiManager := NewAIManager(c.manager, task, managerPane)
	
	// マネージャープロンプトを送信
	if err := aiManager.SendManagerPrompt(claudePane); err != nil {
		return err
	}

	fmt.Printf("🎯 AI管理モード開始: 親ペイン %s がプロジェクトマネージャーとして機能します\n", managerPane)
	fmt.Printf("📋 役割分担: 親ペイン=マネジメント・レビュー専用, 子ペイン=実装作業専用\n")
	fmt.Printf("🔄 子ペインが実装完了報告後、親ペインがレビュー・品質管理を実施します\n")
	
	return nil
}

