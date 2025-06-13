# Claude Company 🤖

**AI-powered project management system with intelligent task delegation**

Claude Company transforms your development workflow by creating an AI-powered team where one Claude AI acts as a project manager, orchestrating multiple worker Claude AIs to collaboratively complete complex tasks.

## ✨ Key Features

### 🎯 **AI Project Manager System**
- **Smart Task Delegation**: Parent pane analyzes tasks and breaks them into manageable subtasks
- **Intelligent Worker Management**: Automatically creates and manages child panes for parallel work
- **Quality Control**: Built-in review system and integration testing
- **Real-time Progress Monitoring**: Track completion status across all workers

### ⚡ **STORM Session Manager**
- **Lightning-fast tmux session management**
- **Cross-shell compatibility** (bash, zsh, fish)
- **Clean command interface** for session lifecycle management

### 🔄 **Automated Workflow**
- **Role Separation**: Manager for oversight, workers for implementation
- **Automatic Pane Creation**: Dynamic scaling based on task complexity
- **Quality Assurance**: Mandatory code review and build testing
- **Seamless Integration**: Built-in tmux and Claude AI integration

## 🚀 Quick Start

### 1. Installation
```bash
# Clone and build
git clone https://github.com/yourusername/claude-company.git
cd claude-company
go build -o bin/ccs

# Or use the install script
./install.sh
```

### 2. Setup Claude Company Session
```bash
# Create tmux session with AI-powered workspace
./bin/ccs
# This creates a structured environment with manager and worker panes
```

### 3. Assign Tasks to AI Team
```bash
# AI Manager Mode (now the only mode)
./bin/ccs --task "Implement user authentication system with JWT tokens"
```

### 4. Watch the Magic Happen
1. **Manager pane** analyzes the task and creates a project plan
2. **Worker panes** are automatically created and assigned specific subtasks
3. **Implementation** happens in parallel across multiple Claude AIs
4. **Quality control** - Manager reviews all work and coordinates testing
5. **Integration** - Final build and validation

## 🏗️ Architecture Overview

### **Role-Based AI Team Structure**

```
┌─────────────────┬─────────────────┐
│ 🎯 Manager Pane │ 🔧 Worker Pane  │
│ (Claude AI #1)  │ (Claude AI #2)  │
│                 │                 │
│ • Task Analysis │ • Code Writing  │
│ • Planning      │ • File Creation │
│ • Review        │ • Implementation│
│ • Quality Check │ • Bug Fixes     │
├─────────────────┼─────────────────┤
│ 🔍 Review Pane  │ 🧪 Test Pane    │
│ (Claude AI #3)  │ (Claude AI #4)  │
│                 │                 │
│ • Code Review   │ • Unit Testing  │
│ • Standards     │ • Integration   │
│ • Optimization  │ • Validation    │
│ • Documentation │ • Build Checks  │
└─────────────────┴─────────────────┘
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
./bin/ccs --task "Create a REST API with authentication, user management, and a React frontend"

# Code refactoring and optimization  
./bin/ccs --task "Refactor the existing codebase for better maintainability and add comprehensive tests"

# Bug fixing and enhancement
./bin/ccs --task "Fix all build errors and add logging functionality throughout the application"
```

### **Project Management Tasks**

```bash
# Architecture design
./bin/ccs --task "Design a microservices architecture for the e-commerce platform and implement the user service"

# Documentation creation
./bin/ccs --task "Create comprehensive API documentation and add inline code comments"

# Performance optimization
./bin/ccs --task "Profile the application, identify bottlenecks, and implement performance improvements"
```

## 🛠️ Command Reference

### **Main Commands**
```bash
# Setup tmux session (default behavior)
./bin/ccs
./bin/ccs --setup

# Assign task to AI team
./bin/ccs --task "TASK_DESCRIPTION"
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
go build -o bin/ccs
```

3. **Setup PATH (optional)**
```bash
cp bin/ccs ~/bin/ccs
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

## 🎭 Real-World Example

Let's say you want to add a user authentication system:

```bash
./bin/ccs --task "Add JWT-based user authentication with registration, login, and protected routes"
```

**What happens automatically:**

1. **Manager Analysis** (Parent Pane):
   - "I need to break this into: user models, JWT service, auth middleware, registration endpoint, login endpoint, and tests"

2. **Worker Creation & Assignment**:
   - Creates 3 child panes
   - Assigns backend work to Worker #1
   - Assigns testing to Worker #2  
   - Assigns integration to Worker #3

3. **Parallel Implementation**:
   - Worker #1: Creates user models, JWT functions, endpoints
   - Worker #2: Writes unit tests and integration tests
   - Worker #3: Sets up middleware and route protection

4. **Quality Control**:
   - Manager reviews each component
   - Requests modifications if needed
   - Coordinates final integration testing

5. **Completion**:
   - All code is working and tested
   - Build passes successfully
   - Features are ready to use

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
./bin/ccs --setup
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
go build -o bin/ccs
./bin/ccs --task "Help improve this project"  # Meta! 😄
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