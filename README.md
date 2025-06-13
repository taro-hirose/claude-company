# Claude Company Tools 🤖

**Cross-shell compatible command-line tools for team collaboration with Claude AI.**

Claude Company Tools provides two powerful commands (`storm` and `deploy`) that transform your terminal into an AI-powered team workspace using tmux sessions and Claude AI integration.

## ✨ Features

### ⚡ Storm (Session Manager)
- **Multi-pane tmux session management** - Lightning-fast tmux session management
- **Cross-shell compatibility** - Works seamlessly across bash, zsh, fish, and other shells
- **Session lifecycle management** - Create, attach, switch, rename, and manage sessions
- **Clean command interface** - Simple and intuitive command structure

### 🚀 Deploy (AI Task Manager) 
- **AI-powered task assignment** - Intelligently distribute tasks across team members
- **Dual operation modes**:
  - **AI-assisted mode** - Automatic task breakdown and intelligent distribution
  - **Simple mode** - Direct task assignment to specific team members
- **Real-time progress tracking** - Monitor task completion and team status
- **Tmux integration** - Seamless integration with tmux pane management

## 🔧 Installation

### Quick Install (Recommended)
```bash
# Clone the repository
git clone https://github.com/yourusername/claude-company.git
cd claude-company

# Run the installation script
./install.sh
```

### Manual Installation
```bash
# Build the binaries
make build

# Copy to your PATH
cp bin/storm ~/bin/storm
cp bin/deploy ~/bin/deploy
chmod +x ~/bin/storm ~/bin/deploy

# Add ~/bin to your PATH if not already added
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc  # or ~/.zshrc
```

### Cross-platform Build
```bash
make build-cross
```

## 🚀 Quick Start

### 1. Start a Session
```bash
# Create and attach to a new session
storm new my-project

# List all sessions
storm list

# Attach to existing session
storm attach my-project
```

### 2. Setup Claude Company Environment
```bash
# Setup tmux session with Claude AI integration
./bin/deploy

# This creates a structured workspace with:
# - Top pane: Management interface
# - Bottom pane: Claude AI assistant
```

### 3. Assign Tasks to Team Members
```bash
# AI-assisted mode (intelligent task processing)
./bin/deploy -ai -task "Implement user authentication system" -pane %1

# Simple mode (direct task assignment)  
./bin/deploy -simple -task "Review code changes" -pane %2
```

### 4. Manage Sessions
```bash
# List all sessions
storm list

# Attach to existing session
storm attach my-project

# Kill session
storm kill my-project
```

## 📚 Usage Examples

### Task Assignment Modes

#### 🤖 AI-Assisted Mode
For complex tasks that benefit from intelligent processing:
```bash
./bin/deploy -ai -task "Design and implement user authentication system" -pane %1
```
**Features:**
- Intelligent task analysis and breakdown
- Automatic sub-task identification
- Context-aware implementation strategies
- Enhanced error handling and reporting

#### 🎯 Simple Mode  
For straightforward, direct task assignments:
```bash
./bin/deploy -simple -task "Run unit tests and fix failures" -pane %2
```
**Features:**
- Direct task execution
- Minimal processing overhead
- Quick task assignment
- Basic status reporting

### Team Collaboration Workflow
```bash
# 1. Start a new project session
storm new web-app-project

# 2. Setup Claude Company environment
./bin/deploy

# 3. Assign backend development (AI mode)
./bin/deploy -ai -task "Create REST API with authentication" -pane %1

# 4. Assign frontend development (AI mode)
./bin/deploy -ai -task "Build React components for user dashboard" -pane %2

# 5. Assign testing tasks (Simple mode)
./bin/deploy -simple -task "Write integration tests" -pane %3

# 6. Monitor progress - check pane output
tmux capture-pane -t %1 -p | tail -10
```

### Session Management
```bash
# List active sessions
storm list

# Kill specific session
storm kill session-name
```

## 🛠️ Command Reference

### Storm Command (Session Manager)
```bash
# List all sessions
storm list
storm ls

# Create new session
storm new <session-name>

# Attach to session
storm attach <session-name>
storm a <session-name>

# Kill session
storm kill <session-name>
storm k <session-name>

# Show help
storm help
```

### Deploy Command (AI Task Manager)
```bash
# Setup tmux session (default behavior)
./bin/deploy

# AI-assisted task assignment
./bin/deploy -ai -task "TASK_DESCRIPTION" -pane PANE_ID

# Simple task assignment  
./bin/deploy -simple -task "TASK_DESCRIPTION" -pane PANE_ID

# Setup session explicitly
./bin/deploy -setup
```

### Parameters
- `-ai`: Enable AI-assisted mode (default if no mode specified)
- `-simple`: Enable simple mode
- `-task`: Task description (required for assignments)
- `-pane`: Target pane ID (required for assignments)
- `-setup`: Explicitly setup tmux session

### Pane ID Format
Use tmux pane IDs like `%1`, `%2`, `%3`, etc. You can find these with:
```bash
tmux list-panes -s -t claude-squad -F '#{pane_id}'
```

## 🛠️ System Requirements

- **Go 1.21+** for building from source
- **tmux** - Required for session management
- **Unix-like OS** - Linux, macOS, or WSL on Windows
- **Shell support** - bash, zsh, fish, or compatible shells

## 🏗️ Development

### Build from Source
```bash
# Clone the repository
git clone https://github.com/yourusername/claude-company.git
cd claude-company

# Install dependencies
go mod tidy

# Build all binaries
make build

# Run tests
make test

# Clean build artifacts
make clean
```

### Project Structure
```
claude-company/
├── cmd/
│   └── ccs/           # Storm command implementation (tmux session manager)
│       └── main.go
├── bin/               # Built binaries (storm, deploy)
├── main.go           # Deploy command implementation (AI task manager)
├── ccs.go            # Storm core logic
├── Makefile          # Build configuration
├── install.sh        # Installation script
├── go.mod            # Go module definition
└── README.md         # This file
```

## 🤝 Contributing

We welcome contributions! Here's how to get started:

1. **Fork the repository**
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Make your changes** with tests
4. **Commit your changes** (`git commit -m 'Add amazing feature'`)
5. **Push to the branch** (`git push origin feature/amazing-feature`)
6. **Open a Pull Request**

### Development Guidelines
- Follow Go conventions and best practices
- Add tests for new functionality
- Update documentation for API changes
- Ensure cross-shell compatibility

## 📋 Roadmap

- [ ] Web UI for session management
- [ ] Plugin system for custom task types
- [ ] Integration with more AI providers
- [ ] Docker containerization
- [ ] Configuration file support
- [ ] Team analytics and reporting

## 🐛 Troubleshooting

### Common Issues

**Q: `tmux: command not found`**
```bash
# Install tmux on your system
# macOS: brew install tmux
# Ubuntu: sudo apt install tmux  
# CentOS: sudo yum install tmux
```

**Q: Commands not found after installation**
```bash
# Ensure ~/bin is in your PATH
echo $PATH | grep "$HOME/bin"

# If not found, add to your shell config
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

**Q: Permission denied errors**
```bash
# Make binaries executable
chmod +x ~/bin/storm ~/bin/deploy
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by modern DevOps collaboration workflows
- Built for seamless integration with Claude AI
- Community-driven development approach

---

**Made with ❤️ for developer productivity and AI-human collaboration**

For more examples and advanced usage, check out our [Wiki](https://github.com/yourusername/claude-company/wiki).