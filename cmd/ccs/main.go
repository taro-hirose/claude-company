package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]
	switch command {
	case "list", "ls":
		listSessions()
	case "new", "create":
		if len(os.Args) < 3 {
			fmt.Println("Usage: storm new <session-name>")
			return
		}
		createSession(os.Args[2])
	case "attach", "a":
		if len(os.Args) < 3 {
			fmt.Println("Usage: storm attach <session-name>")
			return
		}
		attachSession(os.Args[2])
	case "kill", "k":
		if len(os.Args) < 3 {
			fmt.Println("Usage: storm kill <session-name>")
			return
		}
		killSession(os.Args[2])
	case "rename", "r":
		if len(os.Args) < 4 {
			fmt.Println("Usage: storm rename <old-session-name> <new-session-name>")
			return
		}
		renameSession(os.Args[2], os.Args[3])
	case "switch", "s":
		if len(os.Args) < 3 {
			fmt.Println("Usage: storm switch <session-name>")
			return
		}
		switchSession(os.Args[2])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("🌪️  STORM - Claude Company Session Manager")
	fmt.Println("    Lightning-fast tmux session management")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  storm list                            List all active sessions")
	fmt.Println("  storm new <session-name>              Create new session")
	fmt.Println("  storm attach <session-name>           Attach to session")
	fmt.Println("  storm kill <session-name>             Terminate session")
	fmt.Println("  storm rename <old-name> <new-name>    Rename session")
	fmt.Println("  storm switch <session-name>           Switch to session")
	fmt.Println("  storm help                            Show this help")
	fmt.Println()
	fmt.Println("⚡ Aliases:")
	fmt.Println("  ls  → list     a  → attach     k  → kill")
	fmt.Println("  r   → rename   s  → switch")
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
	
	sessions := strings.Split(strings.TrimSpace(string(output)), "\n")
	fmt.Printf("Active tmux sessions (%d):\n", len(sessions))
	for i, session := range sessions {
		fmt.Printf("  %d. %s\n", i+1, session)
	}
}

func createSession(sessionName string) {
	if sessionExists(sessionName) {
		fmt.Printf("Session '%s' already exists\n", sessionName)
		return
	}
	
	cmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error creating session '%s': %v\n", sessionName, err)
		return
	}
	fmt.Printf("Created session: %s\n", sessionName)
}

func attachSession(sessionName string) {
	if !sessionExists(sessionName) {
		fmt.Printf("Session '%s' does not exist\n", sessionName)
		return
	}
	
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
	if !sessionExists(sessionName) {
		fmt.Printf("Session '%s' does not exist\n", sessionName)
		return
	}
	
	cmd := exec.Command("tmux", "kill-session", "-t", sessionName)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error killing session '%s': %v\n", sessionName, err)
		return
	}
	fmt.Printf("Killed session: %s\n", sessionName)
}

func renameSession(oldName, newName string) {
	if !sessionExists(oldName) {
		fmt.Printf("Session '%s' does not exist\n", oldName)
		return
	}
	
	if sessionExists(newName) {
		fmt.Printf("Session '%s' already exists\n", newName)
		return
	}
	
	cmd := exec.Command("tmux", "rename-session", "-t", oldName, newName)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error renaming session '%s' to '%s': %v\n", oldName, newName, err)
		return
	}
	fmt.Printf("Renamed session '%s' to '%s'\n", oldName, newName)
}

func switchSession(sessionName string) {
	if !sessionExists(sessionName) {
		fmt.Printf("Session '%s' does not exist\n", sessionName)
		return
	}
	
	cmd := exec.Command("tmux", "switch-client", "-t", sessionName)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error switching to session '%s': %v\n", sessionName, err)
		return
	}
	fmt.Printf("Switched to session: %s\n", sessionName)
}

func sessionExists(sessionName string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", sessionName)
	err := cmd.Run()
	return err == nil
}