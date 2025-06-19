package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"claude-company/internal/commands"
	"claude-company/internal/session"
)

func main() {
	var setup bool
	var taskDesc string
	var orchestrate bool
	var help bool
	
	flag.BoolVar(&setup, "setup", false, "Setup Claude Company tmux session")
	flag.StringVar(&taskDesc, "task", "", "Task description")
	flag.BoolVar(&orchestrate, "orchestrate", false, "Enable orchestrator mode for step-based task management")
	flag.BoolVar(&help, "help", false, "Show help information")
	flag.Parse()

	// Show help if requested
	if help {
		showHelp()
		return
	}

	manager := session.NewManager("claude-squad", "claude --dangerously-skip-permissions")

	// Set orchestrator mode if requested
	if orchestrate {
		manager.SetOrchestratorMode(true)
		fmt.Println("🔧 Orchestrator mode enabled")
	}

	// Default behavior: setup tmux session
	if len(os.Args) == 1 || setup {
		if err := manager.Setup(); err != nil {
			log.Fatal(err)
		}
		return
	}

	if taskDesc != "" {
		ctx := context.Background()
		deploy := commands.NewDeployCommand(taskDesc, manager)
		if err := deploy.Execute(ctx); err != nil {
			log.Fatal(err)
		}
	}
}

func showHelp() {
	fmt.Println("Claude Company - AI Task Management System")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  claude-company [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  --setup              Setup Claude Company tmux session (default behavior)")
	fmt.Println("  --task <description> Assign a task to AI team")
	fmt.Println("  --orchestrate        Enable orchestrator mode for step-based task management")
	fmt.Println("  --help               Show this help information")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  claude-company")
	fmt.Println("    Set up tmux session with traditional manager mode")
	fmt.Println()
	fmt.Println("  claude-company --orchestrate")
	fmt.Println("    Set up tmux session with orchestrator mode enabled")
	fmt.Println()
	fmt.Println("  claude-company --task \"Implement user authentication\"")
	fmt.Println("    Assign task using traditional delegation mode")
	fmt.Println()
	fmt.Println("  claude-company --orchestrate --task \"Implement user authentication\"")
	fmt.Println("    Assign task using orchestrator mode with step-based execution")
	fmt.Println()
	fmt.Println("MODES:")
	fmt.Println("  Traditional Manager Mode:")
	fmt.Println("    - Basic task delegation to child panes")
	fmt.Println("    - Simple progress monitoring")
	fmt.Println("    - Manual coordination")
	fmt.Println()
	fmt.Println("  Orchestrator Mode:")
	fmt.Println("    - Step-based task decomposition")
	fmt.Println("    - Automated dependency management")
	fmt.Println("    - Parallel execution optimization")
	fmt.Println("    - Quality monitoring and automatic retries")
	fmt.Println("    - Learning-based improvement")
}