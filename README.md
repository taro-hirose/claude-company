# Claude Company 🤖

**AI-powered tmux session management with intelligent task delegation**

Claude Company provides a streamlined way to manage tmux sessions and delegate tasks to Claude AI within structured panes for collaborative development work.

## ✨ Key Features

### ⚡ **STORM Session Manager**
- **Lightning-fast tmux session management**
- **Cross-shell compatibility** (bash, zsh, fish)
- **Clean command interface** for session lifecycle management

### 🔄 **AI Task Management**
- **Role Separation**: Manager for oversight, workers for implementation
- **Structured Pane Creation**: Creates organized workspace with manager and worker panes
- **Task Assignment**: Simple interface for assigning tasks to AI workers

## 🚀 Quick Start

### 1. Installation
```bash
# Clone and build
git clone https://github.com/yourusername/claude-company.git
cd claude-company
go build -o claude-company .

# Or use the build system
make build
```

### 2. Setup Claude Company Session
```bash
# Create tmux session with AI-powered workspace
./claude-company
# This creates a structured environment with manager and worker panes
```

### 3. Assign Tasks to AI Team
```bash
# Simple task assignment
./claude-company --task "Implement user authentication system"
```

## 🏗️ Architecture Overview

### **Role-Based AI Team Structure**

```
┌─────────────────┬─────────────────┐
│ 🎯 Manager Pane │ 🔧 Worker Pane  │
│ (Claude AI #1)  │ (Claude AI #2+) │
│                 │                 │
│ • Task Analysis │ • Code Writing  │
│ • Planning      │ • File Creation │
│ • Review        │ • Implementation│
│ • Quality Check │ • Bug Fixes     │
│ • Coordination  │ • Task Execution│
│ • Final Check   │ • Progress Report│
└─────────────────┴─────────────────┘

**Note**: Manager creates additional worker panes dynamically as needed
```

### **Workflow Process**

1. **📋 Task Input**: You provide a high-level task description
2. **🧠 Analysis**: Manager AI breaks down the task into subtasks  
3. **🏭 Scaling**: Manager creates additional worker panes as needed
4. **⚡ Parallel Execution**: Multiple Claude AIs work simultaneously
5. **🔍 Quality Control**: Manager reviews all implementation work
6. **🧪 Testing**: Automated build verification and testing
7. **✅ Integration**: Final validation and completion

## 📚 Usage Examples

### **Software Development Tasks**

```bash
# Full-stack application development
./claude-company --task "Create a REST API with authentication, user management, and a React frontend"

# Code refactoring and optimization  
./claude-company --task "Refactor the existing codebase for better maintainability and add comprehensive tests"

# Bug fixing and enhancement
./claude-company --task "Fix all build errors and add logging functionality throughout the application"
```

### **Project Management Tasks**

```bash
# Architecture design
./claude-company --task "Design a microservices architecture for the e-commerce platform and implement the user service"

# Documentation creation
./claude-company --task "Create comprehensive API documentation and add inline code comments"

# Performance optimization
./claude-company --task "Profile the application, identify bottlenecks, and implement performance improvements"
```

## 🛠️ Command Reference

### **Main Commands**
```bash
# Setup tmux session (default behavior)
./claude-company
./claude-company --setup

# Assign task to AI team
./claude-company --task "TASK_DESCRIPTION"
```

### **STORM Session Management**
```bash
# List all sessions
./bin/storm list     # or 'ls'

# Create new session  
./bin/storm new <session-name>

# Attach to session
./bin/storm attach <session-name>  # or 'a'

# Kill session
./bin/storm kill <session-name>    # or 'k'

# Switch between sessions
./bin/storm switch <session-name>  # or 's'

# Rename session
./bin/storm rename <old-name> <new-name>  # or 'r'
```

## 🎯 How It Works

### **Manager AI Responsibilities**
- ❌ **Never writes code directly**  
- ✅ **Analyzes and breaks down tasks**
- ✅ **Creates and manages worker panes**
- ✅ **Assigns specific subtasks to workers**
- ✅ **Reviews completed work for quality**
- ✅ **Coordinates testing and integration**
- ✅ **Provides final approval and completion**

### **Worker AI Responsibilities**  
- ✅ **Implements assigned subtasks**
- ✅ **Writes actual code and creates files**
- ✅ **Reports completion with deliverables**
- ✅ **Responds to feedback and revision requests**
- ✅ **Fixes issues identified during review**

### **Communication Protocol**
```bash
# Worker → Manager reporting format
"実装完了：internal/auth/jwt.go - JWT token generation and validation implemented"

# Manager → Worker task assignment format  
"サブタスク: Create user authentication middleware in internal/auth/middleware.go. Include JWT validation and error handling. Report completion when done."

# Manager → Worker review requests
"レビュー要請: Please review internal/auth/jwt.go for code quality, security best practices, and integration compatibility."
```

## 🔧 Installation & Setup

### **System Requirements**
- **Go 1.21+** for building from source
- **tmux** - Required for pane management
- **Claude AI access** - Via Claude CLI tool
- **Unix-like OS** - Linux, macOS, or WSL

### **Step-by-Step Installation**

1. **Install Dependencies**
```bash
# macOS
brew install tmux go

# Ubuntu/Debian  
sudo apt install tmux golang-go

# Install Claude CLI (follow official docs)
```

2. **Build Claude Company**
```bash
git clone https://github.com/yourusername/claude-company.git
cd claude-company
go mod tidy
make build
```

3. **Setup PATH (optional)**
```bash
# Option 1: Copy main binary
cp claude-company ~/bin/claude-company

# Option 2: Install via make (recommended)
make install

# Update PATH if needed
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```


## 🎭 Real-World Example

```bash
./claude-company --task "Add user authentication system"
```

The manager pane will analyze the task and coordinate with worker panes to implement the authentication system in a structured way.

## 🚨 Troubleshooting

### **Common Issues**

**❌ "tmux: command not found"**
```bash
# Install tmux first
brew install tmux      # macOS
sudo apt install tmux  # Ubuntu
```

**❌ "claude: command not found"**
```bash
# Install Claude CLI following official documentation
# Ensure it's available in your PATH
```

**❌ Panes not responding**
```bash
# Check if Claude is running in each pane
tmux list-panes -s -t claude-squad -F '#{pane_id}: #{pane_current_command}'

# Restart session if needed
./claude-company --setup
```

**❌ Tasks not being distributed**
```bash
# Ensure at least 2 panes exist for manager/worker separation
# Manager needs worker panes to delegate tasks to
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md).

### **Development Setup**
```bash
git clone https://github.com/yourusername/claude-company.git
cd claude-company
go mod tidy
make build
./claude-company --task "Help improve this project"  # Meta! 😄
```

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **Claude AI** for making intelligent collaboration possible
- **tmux** for robust terminal multiplexing
- **Go community** for excellent tooling and libraries

---

**Transform your development workflow with AI-powered team collaboration** 🚀

*Made with ❤️ for developers who want to work with AI, not just use it*