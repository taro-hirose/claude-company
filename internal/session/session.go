package session

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type SessionManager interface {
	CreateSession(name string) error
	AttachSession(name string) error
	KillSession(name string) error
	RenameSession(oldName, newName string) error
	SwitchSession(name string) error
	ListSessions() ([]string, error)
	SessionExists(name string) bool
	SendKeysToPane(paneID, command string) error
	GetPanes(sessionName string) ([]string, error)
}

type TmuxSessionManager struct{}

func NewTmuxSessionManager() *TmuxSessionManager {
	return &TmuxSessionManager{}
}

func (t *TmuxSessionManager) CreateSession(name string) error {
	if t.SessionExists(name) {
		return fmt.Errorf("session '%s' already exists", name)
	}
	
	cmd := exec.Command("tmux", "new-session", "-d", "-s", name)
	return cmd.Run()
}

func (t *TmuxSessionManager) AttachSession(name string) error {
	if !t.SessionExists(name) {
		return fmt.Errorf("session '%s' does not exist", name)
	}
	
	var cmd *exec.Cmd
	if os.Getenv("TMUX") != "" {
		cmd = exec.Command("tmux", "switch-client", "-t", name)
	} else {
		cmd = exec.Command("tmux", "attach-session", "-t", name)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	
	return cmd.Run()
}

func (t *TmuxSessionManager) KillSession(name string) error {
	if !t.SessionExists(name) {
		return fmt.Errorf("session '%s' does not exist", name)
	}
	
	cmd := exec.Command("tmux", "kill-session", "-t", name)
	return cmd.Run()
}

func (t *TmuxSessionManager) RenameSession(oldName, newName string) error {
	if !t.SessionExists(oldName) {
		return fmt.Errorf("session '%s' does not exist", oldName)
	}
	
	if t.SessionExists(newName) {
		return fmt.Errorf("session '%s' already exists", newName)
	}
	
	cmd := exec.Command("tmux", "rename-session", "-t", oldName, newName)
	return cmd.Run()
}

func (t *TmuxSessionManager) SwitchSession(name string) error {
	if !t.SessionExists(name) {
		return fmt.Errorf("session '%s' does not exist", name)
	}
	
	cmd := exec.Command("tmux", "switch-client", "-t", name)
	return cmd.Run()
}

func (t *TmuxSessionManager) ListSessions() ([]string, error) {
	cmd := exec.Command("tmux", "list-sessions")
	output, err := cmd.Output()
	if err != nil {
		if strings.Contains(err.Error(), "no server running") {
			return []string{}, nil
		}
		return nil, err
	}
	
	if len(output) == 0 {
		return []string{}, nil
	}
	
	sessions := strings.Split(strings.TrimSpace(string(output)), "\n")
	return sessions, nil
}

func (t *TmuxSessionManager) SessionExists(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	err := cmd.Run()
	return err == nil
}

func (t *TmuxSessionManager) SendKeysToPane(paneID, command string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", paneID, command, "Enter")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tmux command failed: %v, output: %s", err, string(output))
	}
	return nil
}

func (t *TmuxSessionManager) GetPanes(sessionName string) ([]string, error) {
	cmd := exec.Command("tmux", "list-panes", "-s", "-t", sessionName, "-F", "#{pane_id}")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := []string{}
	current := ""
	for _, b := range output {
		if b == '\n' {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
		} else {
			current += string(b)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}

	return lines, nil
}