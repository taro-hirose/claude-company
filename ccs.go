package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ccsMain() {
	if len(os.Args) < 2 {
		printCCSUsage()
		return
	}

	command := os.Args[1]
	switch command {
	case "list", "ls":
		listSessions()
	case "new", "create":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ccs new <session-name>")
			return
		}
		createSession(os.Args[2])
	case "attach", "a":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ccs attach <session-name>")
			return
		}
		attachSession(os.Args[2])
	case "kill", "k":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ccs kill <session-name>")
			return
		}
		killSession(os.Args[2])
	case "help", "-h", "--help":
		printCCSUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printCCSUsage()
	}
}

func printCCSUsage() {
	fmt.Println("ccs - tmux session management tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ccs list                    List all tmux sessions")
	fmt.Println("  ccs new <session-name>      Create new tmux session")
	fmt.Println("  ccs attach <session-name>   Attach to existing session")
	fmt.Println("  ccs kill <session-name>     Kill tmux session")
	fmt.Println("  ccs help                    Show this help message")
	fmt.Println()
	fmt.Println("Aliases:")
	fmt.Println("  ls  -> list")
	fmt.Println("  a   -> attach")
	fmt.Println("  k   -> kill")
}

func listSessions() {
	cmd := exec.Command("tmux", "list-sessions")
	output, err := cmd.Output()
	if err != nil {
		if strings.Contains(err.Error(), "no server running") {
			fmt.Println("No tmux sessions running")
			return
		}
		fmt.Printf("Error listing sessions: %v\n", err)
		return
	}
	
	if len(output) == 0 {
		fmt.Println("No tmux sessions running")
		return
	}
	
	fmt.Print(string(output))
}

func createSession(sessionName string) {
	cmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error creating session '%s': %v\n", sessionName, err)
		return
	}
	fmt.Printf("Created session: %s\n", sessionName)
}

func attachSession(sessionName string) {
	cmd := exec.Command("tmux", "attach-session", "-t", sessionName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error attaching to session '%s': %v\n", sessionName, err)
		return
	}
}

func killSession(sessionName string) {
	cmd := exec.Command("tmux", "kill-session", "-t", sessionName)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error killing session '%s': %v\n", sessionName, err)
		return
	}
	fmt.Printf("Killed session: %s\n", sessionName)
}