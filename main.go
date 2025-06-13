package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
)

type Task struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Mode        string `json:"mode"`
	PaneID      string `json:"pane_id"`
	Status      string `json:"status"`
}

type CCACommand struct {
	aiMode     bool
	simpleMode bool
	taskDesc   string
	paneID     string
}

func main() {
	var cca CCACommand
	
	flag.BoolVar(&cca.aiMode, "ai", false, "Enable AI-assisted mode")
	flag.BoolVar(&cca.simpleMode, "simple", false, "Enable simple mode")
	flag.StringVar(&cca.taskDesc, "task", "", "Task description")
	flag.StringVar(&cca.paneID, "pane", "", "Target pane ID")
	flag.Parse()

	if cca.taskDesc == "" {
		log.Fatal("Task description is required (-task)")
	}

	if cca.paneID == "" {
		log.Fatal("Pane ID is required (-pane)")
	}

	if !cca.aiMode && !cca.simpleMode {
		cca.aiMode = true // Default to AI mode
	}

	if err := cca.execute(); err != nil {
		log.Fatal(err)
	}
}

func (c *CCACommand) execute() error {
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

func (c *CCACommand) executeAIMode(task Task) error {
	aiPrompt := fmt.Sprintf("あなたは%sです。サブタスク: %s。完了時またはエラー時は報告: tmux send-keys -t %%3 \"%%4 %s完了または進捗報告\" Enter", 
		task.PaneID, task.Description, task.Description)
	
	return c.sendToPane(task.PaneID, aiPrompt)
}

func (c *CCACommand) executeSimpleMode(task Task) error {
	simpleCommand := fmt.Sprintf("echo \"タスク開始: %s (パネル: %s)\"", task.Description, task.PaneID)
	
	return c.sendToPane(task.PaneID, simpleCommand)
}

func (c *CCACommand) sendToPane(paneID, command string) error {
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