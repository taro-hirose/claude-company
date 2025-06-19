package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type OrchestratorConfig struct {
	Manager  ManagerConfig  `yaml:"manager"`
	Workers  WorkersConfig  `yaml:"workers"`
	Session  SessionConfig  `yaml:"session"`
	Defaults DefaultsConfig `yaml:"defaults"`
}

type ManagerConfig struct {
	Role        string `yaml:"role"`
	Prompt      string `yaml:"prompt"`
	MaxRetries  int    `yaml:"max_retries"`
	ReviewDepth int    `yaml:"review_depth"`
}

type WorkersConfig struct {
	MaxWorkers   int      `yaml:"max_workers"`
	Roles        []string `yaml:"roles"`
	TaskTimeout  int      `yaml:"task_timeout"`
	CoordinationMode string `yaml:"coordination_mode"`
}

type SessionConfig struct {
	Name           string `yaml:"name"`
	Layout         string `yaml:"layout"`
	WindowPrefix   string `yaml:"window_prefix"`
	PanePrefix     string `yaml:"pane_prefix"`
	AutoStartTmux  bool   `yaml:"auto_start_tmux"`
}

type DefaultsConfig struct {
	WorkingDir  string `yaml:"working_dir"`
	Shell       string `yaml:"shell"`
	ClaudeFlags []string `yaml:"claude_flags"`
	Language    string `yaml:"language"`
}

func NewOrchestratorConfig() *OrchestratorConfig {
	return &OrchestratorConfig{
		Manager: ManagerConfig{
			Role:        "project_manager",
			Prompt:      "あなたはプロジェクトマネージャーです。タスクを分析し、ワーカーに適切に割り当ててください。",
			MaxRetries:  3,
			ReviewDepth: 2,
		},
		Workers: WorkersConfig{
			MaxWorkers:   4,
			Roles:        []string{"developer", "tester", "reviewer"},
			TaskTimeout:  1800,
			CoordinationMode: "hierarchical",
		},
		Session: SessionConfig{
			Name:         "claude-squad",
			Layout:       "tiled",
			WindowPrefix: "work",
			PanePrefix:   "claude",
			AutoStartTmux: true,
		},
		Defaults: DefaultsConfig{
			WorkingDir:  ".",
			Shell:       "/bin/bash",
			ClaudeFlags: []string{"--dangerously-skip-permissions"},
			Language:    "ja",
		},
	}
}

func (c *OrchestratorConfig) LoadFromFile(filePath string) error {
	loader := NewLoader()
	return loader.LoadConfig(filePath, c)
}

func (c *OrchestratorConfig) GetConfigPath() (string, error) {
	configPaths := []string{
		".claude/orchestrator.yaml",
		".claude/orchestrator.yml",
		"orchestrator.yaml",
		"orchestrator.yml",
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return "", fmt.Errorf("設定ファイルの絶対パス取得に失敗: %w", err)
			}
			return absPath, nil
		}
	}

	return "", fmt.Errorf("設定ファイルが見つかりません: %v", configPaths)
}

func (c *OrchestratorConfig) Validate() error {
	if c.Manager.Role == "" {
		return fmt.Errorf("manager.role は必須です")
	}
	if c.Workers.MaxWorkers <= 0 {
		return fmt.Errorf("workers.max_workers は1以上である必要があります")
	}
	if c.Session.Name == "" {
		return fmt.Errorf("session.name は必須です")
	}
	if c.Defaults.WorkingDir == "" {
		return fmt.Errorf("defaults.working_dir は必須です")
	}
	return nil
}