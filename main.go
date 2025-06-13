package main

import (
	"flag"
	"log"
	"os"
	"claude-company/internal/commands"
	"claude-company/internal/session"
)

func main() {
	var setup bool
	var taskDesc string
	
	flag.BoolVar(&setup, "setup", false, "Setup Claude Company tmux session")
	flag.StringVar(&taskDesc, "task", "", "Task description")
	flag.Parse()

	manager := session.NewManager("claude-squad", "claude --dangerously-skip-permissions")

	// Default behavior: setup tmux session
	if len(os.Args) == 1 || setup {
		if err := manager.Setup(); err != nil {
			log.Fatal(err)
		}
		return
	}

	if taskDesc == "" {
		log.Fatal("Task description is required (-task)")
	}

	deploy := commands.NewDeployCommand(taskDesc, manager)

	if err := deploy.Execute(); err != nil {
		log.Fatal(err)
	}
}