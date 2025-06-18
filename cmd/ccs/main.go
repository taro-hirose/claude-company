package main

import (
	"fmt"
	"os"
	"strings"
	"claude-company/internal/session"
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
	case "ccs":
		if len(os.Args) < 3 {
			printUsage()
			return
		}
		handleCCSCommand(os.Args[2:])
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
	fmt.Println("  storm ccs <command>                   Basic tmux session management")
	fmt.Println("  storm help                            Show this help")
	fmt.Println()
	fmt.Println("⚡ Aliases:")
	fmt.Println("  ls  → list     a  → attach     k  → kill")
	fmt.Println("  r   → rename   s  → switch")
	fmt.Println()
	fmt.Println("💡 For basic tmux functionality, use: storm ccs help")
}

func listSessions() {
	listSessionsImpl(true)
}

func listSessionsImpl(verbose bool) {
	sessionManager := session.NewTmuxSessionManager()
	sessions, err := sessionManager.ListSessions()
	if err != nil {
		if strings.Contains(err.Error(), "no server running") {
			fmt.Println("No tmux sessions running")
			return
		}
		fmt.Printf("Error listing sessions: %v\n", err)
		return
	}
	
	if len(sessions) == 0 {
		fmt.Println("No tmux sessions running")
		return
	}
	
	if verbose {
		fmt.Printf("Active tmux sessions (%d):\n", len(sessions))
		for i, sessionName := range sessions {
			fmt.Printf("  %d. %s\n", i+1, sessionName)
		}
	} else {
		for _, sessionName := range sessions {
			fmt.Println(sessionName)
		}
	}
}

func createSession(sessionName string) {
	createSessionImpl(sessionName, true)
}

func createSessionImpl(sessionName string, verbose bool) {
	sessionManager := session.NewTmuxSessionManager()
	if sessionManager.SessionExists(sessionName) {
		if verbose {
			fmt.Printf("Session '%s' already exists\n", sessionName)
		}
		return
	}
	
	err := sessionManager.CreateSession(sessionName)
	if err != nil {
		fmt.Printf("Error creating session '%s': %v\n", sessionName, err)
		return
	}
	fmt.Printf("Created session: %s\n", sessionName)
}

func attachSession(sessionName string) {
	attachSessionImpl(sessionName, true)
}

func attachSessionImpl(sessionName string, verbose bool) {
	sessionManager := session.NewTmuxSessionManager()
	if !sessionManager.SessionExists(sessionName) {
		if verbose {
			fmt.Printf("Session '%s' does not exist\n", sessionName)
		}
		return
	}
	
	err := sessionManager.AttachSession(sessionName)
	if err != nil {
		fmt.Printf("Error attaching to session '%s': %v\n", sessionName, err)
		return
	}
}

func killSession(sessionName string) {
	killSessionImpl(sessionName, true)
}

func killSessionImpl(sessionName string, verbose bool) {
	sessionManager := session.NewTmuxSessionManager()
	if !sessionManager.SessionExists(sessionName) {
		if verbose {
			fmt.Printf("Session '%s' does not exist\n", sessionName)
		}
		return
	}
	
	err := sessionManager.KillSession(sessionName)
	if err != nil {
		fmt.Printf("Error killing session '%s': %v\n", sessionName, err)
		return
	}
	fmt.Printf("Killed session: %s\n", sessionName)
}

func renameSession(oldName, newName string) {
	sessionManager := session.NewTmuxSessionManager()
	if !sessionManager.SessionExists(oldName) {
		fmt.Printf("Session '%s' does not exist\n", oldName)
		return
	}
	
	if sessionManager.SessionExists(newName) {
		fmt.Printf("Session '%s' already exists\n", newName)
		return
	}
	
	err := sessionManager.RenameSession(oldName, newName)
	if err != nil {
		fmt.Printf("Error renaming session '%s' to '%s': %v\n", oldName, newName, err)
		return
	}
	fmt.Printf("Renamed session '%s' to '%s'\n", oldName, newName)
}

func switchSession(sessionName string) {
	sessionManager := session.NewTmuxSessionManager()
	if !sessionManager.SessionExists(sessionName) {
		fmt.Printf("Session '%s' does not exist\n", sessionName)
		return
	}
	
	err := sessionManager.SwitchSession(sessionName)
	if err != nil {
		fmt.Printf("Error switching to session '%s': %v\n", sessionName, err)
		return
	}
	fmt.Printf("Switched to session: %s\n", sessionName)
}

func sessionExists(sessionName string) bool {
	sessionManager := session.NewTmuxSessionManager()
	return sessionManager.SessionExists(sessionName)
}

func handleCCSCommand(args []string) {
	if len(args) < 1 {
		printCCSUsage()
		return
	}

	command := args[0]
	switch command {
	case "list", "ls":
		listSessionsBasic()
	case "new", "create":
		if len(args) < 2 {
			fmt.Println("Usage: ccs new <session-name>")
			return
		}
		createSessionBasic(args[1])
	case "attach", "a":
		if len(args) < 2 {
			fmt.Println("Usage: ccs attach <session-name>")
			return
		}
		attachSessionBasic(args[1])
	case "kill", "k":
		if len(args) < 2 {
			fmt.Println("Usage: ccs kill <session-name>")
			return
		}
		killSessionBasic(args[1])
	case "help", "-h", "--help":
		printCCSUsage()
	default:
		fmt.Printf("Unknown ccs command: %s\n", command)
		printCCSUsage()
	}
}

func printCCSUsage() {
	fmt.Println("ccs - tmux session management tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  storm ccs list                    List all tmux sessions")
	fmt.Println("  storm ccs new <session-name>      Create new tmux session")
	fmt.Println("  storm ccs attach <session-name>   Attach to existing session")
	fmt.Println("  storm ccs kill <session-name>     Kill tmux session")
	fmt.Println("  storm ccs help                    Show this help message")
	fmt.Println()
	fmt.Println("Aliases:")
	fmt.Println("  ls  -> list")
	fmt.Println("  a   -> attach")
	fmt.Println("  k   -> kill")
}

func listSessionsBasic() {
	listSessionsImpl(false)
}

func createSessionBasic(sessionName string) {
	createSessionImpl(sessionName, false)
}

func attachSessionBasic(sessionName string) {
	attachSessionImpl(sessionName, false)
}

func killSessionBasic(sessionName string) {
	killSessionImpl(sessionName, false)
}