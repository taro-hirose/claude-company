# CLAUDE.md

必ず日本語で回答して
This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
# Build all binaries (storm + deploy)
make build

# Build individual components
make build-storm   # Builds tmux session manager
make build-deploy  # Builds AI task manager

# Cross-platform builds
make build-cross   # Builds for Linux, macOS (Intel/ARM), Windows

# Install to system
make install       # Installs to /usr/local/bin (requires sudo)

# Run tests
make test

# Clean build artifacts
make clean
```

## Architecture Overview

Claude Company consists of two main components:

### 1. Storm (Tmux Session Manager)

### 2. Deploy (AI Task Manager)

- **Entry point**: `/main.go`
- **Binary location**: `bin/deploy` or `claude-company`
- **Purpose**: Creates and manages tmux sessions with multiple Claude AI instances following a manager-worker pattern
- **Usage**:
  - `./claude-company` - Setup tmux session
  - `./claude-company --task "description"` - Assign task to AI team

### Key Directories

- `/internal/commands/` - Command implementations (deploy.go for AI task management)
- `/internal/session/` - Session management logic
  - `manager.go` - Core session manager with AI coordination
  - `session.go` - Tmux interface and implementation

### Important Architectural Patterns

1. **Manager-Worker Pattern**: Parent tmux panes act as project managers (coordination only), while child panes handle implementation
2. **Japanese Prompts**: The AI coordination prompts are in Japanese, indicating primary use in Japanese environments
3. **Tmux Integration**: Deep integration with tmux for terminal multiplexing and pane management
4. **Role Separation**: Manager AIs never write code directly; they only coordinate and review

## Development Notes

- Go version: 1.21 (as specified in go.mod)
- No external dependencies beyond Go standard library
- Binaries are built to `bin/` directory
- The project uses Go's internal package pattern for encapsulation
- Session name "claude-squad" is hardcoded in main.go
- Claude is invoked with `--dangerously-skip-permissions` flag
