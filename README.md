# Claude Company Tools 🤖

**Cross-shell compatible command-line tools for team collaboration with Claude AI.**

Claude Company Tools provides two powerful commands (`ccs` and `cca`) that transform your terminal into an AI-powered team workspace using tmux sessions and Claude AI integration.

## ✨ Features

### 🚀 CCS (Claude Company Shell)
- **Multi-pane tmux session management** - Automatically creates and manages structured tmux sessions
- **Cross-shell compatibility** - Works seamlessly across bash, zsh, fish, and other shells
- **Session lifecycle management** - Create, attach, switch, rename, and manage sessions
- **Smart pane layout** - Optimized pane distribution for team collaboration

### 🎯 CCA (Claude Company Assistant) 
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
cp bin/ccs ~/bin/ccs
cp bin/cca ~/bin/cca
chmod +x ~/bin/ccs ~/bin/cca

# Add ~/bin to your PATH if not already added
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc  # or ~/.zshrc
```

### Cross-platform Build
```bash
make build-cross
```

## 🚀 Quick Start

### 1. Start a Claude Company Session
```bash
# Create and attach to a new session
ccs new my-project

# Or create the default 'claude-squad' session
ccs
```

### 2. Assign Tasks to Team Members
```bash
# AI-assisted mode (automatic task breakdown)
cca -ai -task "Implement user authentication system" -pane %1

# Simple mode (direct assignment)
cca -simple -task "Review code changes" -pane %2
```

### 3. Manage Sessions
```bash
# List all sessions
ccs list

# Attach to existing session
ccs attach my-project

# Switch between sessions
ccs switch another-project

# Clean up
ccs kill my-project
```

## 📚 Usage Examples

### Team Collaboration Workflow
```bash
# 1. Start a new project session
ccs new web-app-project

# 2. Assign backend development
cca -ai -task "Create REST API with authentication" -pane %1

# 3. Assign frontend development  
cca -ai -task "Build React components for user dashboard" -pane %2

# 4. Assign testing tasks
cca -simple -task "Write integration tests" -pane %3
```

### Session Management
```bash
# List active sessions
ccs ls

# Rename a session
ccs rename old-name new-name

# Kill specific session
ccs kill session-name
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
│   └── ccs/           # CCS command implementation
│       └── main.go
├── bin/               # Built binaries
├── main.go           # CCA command implementation
├── ccs.go            # CCS core logic
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
chmod +x ~/bin/cca ~/bin/ccs
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