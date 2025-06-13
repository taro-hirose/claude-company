package main

import (
	"flag"
	"log"
	"os"
	"claude-company/internal/api"
	"claude-company/internal/commands"
	"claude-company/internal/database"
	"claude-company/internal/session"
)

func main() {
	var setup bool
	var taskDesc string
	var apiMode bool
	
	flag.BoolVar(&setup, "setup", false, "Setup Claude Company tmux session")
	flag.StringVar(&taskDesc, "task", "", "Task description")
	flag.BoolVar(&apiMode, "api", false, "Start API server mode")
	flag.Parse()

	// Initialize database
	dbConfig := database.NewConfig()
	if err := database.InitDB(dbConfig); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// API server mode
	if apiMode {
		server := api.NewServer(dbConfig)
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		if err := server.Start(port); err != nil {
			log.Fatal(err)
		}
		return
	}

	manager := session.NewManager("claude-squad", "claude --dangerously-skip-permissions")

	// Default behavior: setup tmux session
	if len(os.Args) == 1 || setup {
		if err := manager.Setup(); err != nil {
			log.Fatal(err)
		}
		return
	}

	if taskDesc != "" {
		deploy := commands.NewDeployCommand(taskDesc, manager)
		if err := deploy.Execute(); err != nil {
			log.Fatal(err)
		}
	}
}