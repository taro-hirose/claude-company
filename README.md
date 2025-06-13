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
# AI Manager Mode (full functionality with database)
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

# Assign task to AI team (requires database)
./bin/ccs --task "TASK_DESCRIPTION"

# API server mode
./bin/ccs --api
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
- **Docker & Docker Compose** - For database services
- **Unix-like OS** - Linux, macOS, or WSL

### **Step-by-Step Installation**

1. **Install Dependencies**
```bash
# macOS
brew install tmux go docker

# Ubuntu/Debian  
sudo apt install tmux golang-go docker.io docker-compose

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

## 🗄️ Database Setup & Configuration

### **Prerequisites**

Before starting, ensure you have the following installed:

```bash
# Check Docker version (minimum required: 20.10+)
docker --version

# Check Docker Compose version (minimum required: 2.0+)
docker-compose --version

# Verify Docker daemon is running
docker info
```

**Required versions:**
- **Docker**: 20.10.0 or higher
- **Docker Compose**: 2.0.0 or higher
- **Available ports**: 5432 (PostgreSQL), 8080 (pgAdmin)

### **Database Configuration**

This project uses PostgreSQL as the database backend:

**🟢 PostgreSQL**
- Advanced JSON support and complex queries
- Better performance for analytical workloads
- Includes pgAdmin web interface
- Default choice for Claude Company features

### **1. Initial Setup Process**

#### **Step 1: Environment Configuration**

Create environment file for database settings:

```bash
# Create .env file in project root
cat > .env << 'EOF'
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=claude_user
DB_PASSWORD=claude_password
DB_NAME=claude_company
DB_SSLMODE=disable

# API Server Configuration
PORT=8081
GIN_MODE=release

# Optional: Choose database type (postgres/mysql)
DB_TYPE=postgres
EOF
```

#### **Step 2: Start Database Services**

```bash
# Start all database services in background
docker-compose up -d

# Start PostgreSQL services
docker-compose up -d postgres pgadmin
```

#### **Step 3: Verify Service Status**

```bash
# Check all services status
docker-compose ps

# Expected output:
# NAME                     IMAGE               STATUS
# claude-company-db        postgres:15-alpine  Up (healthy)
# claude-company-pgadmin   dpage/pgadmin4      Up
```

**Services Started:**
- **PostgreSQL Database**: `localhost:5432`
- **pgAdmin Web Interface**: `http://localhost:8080`

### **2. Connection Verification & Testing**

#### **Step 4: Verify Database Connection**

**PostgreSQL Connection Test:**
```bash
# Wait for services to be fully ready (may take 30-60 seconds)
docker-compose logs postgres | grep "ready to accept connections"

# Test connection using docker exec
docker exec claude-company-db psql -U claude_user -d claude_company -c "SELECT version();"

# Test connection from host machine (requires psql client)
PGPASSWORD=claude_password psql -h localhost -p 5432 -U claude_user -d claude_company -c "\\dt"
```


#### **Step 5: Access Web Interfaces**

**pgAdmin (PostgreSQL Management):**
1. Open `http://localhost:8080` in your browser
2. Login credentials:
   - **Email**: `admin@claude-company.local`
   - **Password**: `admin123`
3. Add server connection:
   - **Host**: `postgres` (Docker service name)
   - **Port**: `5432`
   - **Username**: `claude_user`
   - **Password**: `claude_password`


### **3. Environment Variables Reference**

Configure database connection settings:

```bash
# Required environment variables for PostgreSQL
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=claude_user
export DB_PASSWORD=claude_password
export DB_NAME=claude_company
export DB_SSLMODE=disable


# API server settings
export PORT=8081
export GIN_MODE=release
```

**Environment file (.env) template:**
```bash
# PostgreSQL Configuration (default)
DB_HOST=localhost
DB_PORT=5432
DB_USER=claude_user
DB_PASSWORD=claude_password
DB_NAME=claude_company
DB_SSLMODE=disable


# API Configuration
PORT=8081
GIN_MODE=release

# Development settings
DB_TYPE=postgres
```

### **4. Database Schema & Initialization**

The database schema is automatically initialized when starting the containers. The initialization includes:

- **Task tables** with hierarchical structure and ULID primary keys
- **Progress tracking** tables for real-time monitoring
- **Indexes** for performance optimization
- **Functions** for task hierarchy management and calculations
- **Sample data** for testing and demonstration

**Automatic Initialization Process:**
```bash
# Schema files are automatically executed on first startup
# Location: ./db/init/01_schema.sql and ./db/init/02_sample_data.sql

# Check initialization logs
docker-compose logs postgres | grep -i "database system is ready"
```

**Manual schema inspection:**
```bash
# PostgreSQL - View table structure
docker exec claude-company-db psql -U claude_user -d claude_company -c "\\dt"

# PostgreSQL - Check specific table details
docker exec claude-company-db psql -U claude_user -d claude_company -c "\\d tasks"

```

### **5. Development Environment Configuration**

#### **Recommended Development Settings**

For development work, use these optimized settings:

```bash
# Development .env configuration
cat > .env << 'EOF'
# Database (Development optimized)
DB_HOST=localhost
DB_PORT=5432
DB_USER=claude_user
DB_PASSWORD=claude_password
DB_NAME=claude_company_dev
DB_SSLMODE=disable
DB_MAX_CONNECTIONS=20
DB_IDLE_TIMEOUT=300

# API Server (Development mode)
PORT=8081
GIN_MODE=debug
LOG_LEVEL=debug
ENABLE_CORS=true

# Development features
AUTO_MIGRATE=true
SEED_DATABASE=true
EOF
```

#### **Performance Tuning for Development**

```bash
# PostgreSQL development optimizations
docker exec claude-company-db psql -U claude_user -d claude_company -c "
  -- Increase work memory for complex queries
  ALTER SYSTEM SET work_mem = '64MB';
  
  -- Enable query logging for debugging
  ALTER SYSTEM SET log_statement = 'all';
  ALTER SYSTEM SET log_min_duration_statement = 1000;
  
  -- Reload configuration
  SELECT pg_reload_conf();
"
```

### **6. Troubleshooting Common Issues**

#### **🔴 Database Connection Issues**

**Problem**: "Connection refused" errors
```bash
# Solution 1: Check if containers are running
docker-compose ps

# Solution 2: Verify port availability
netstat -tulpn | grep :5432

# Solution 3: Restart services with clean state
docker-compose down -v && docker-compose up -d
```

**Problem**: "Password authentication failed"
```bash
# Solution: Reset database container
docker-compose down
docker volume rm claude-company_postgres_data
docker-compose up -d postgres
```

#### **🔴 Docker Issues**

**Problem**: "Port already in use"
```bash
# Find process using the port
lsof -i :5432

# Kill the process or change port in docker-compose.yml
# Example: Change "5432:5432" to "5433:5432"
```

**Problem**: "Volume mount failed"
```bash
# Ensure Docker has permission to access project directory
# On macOS: Docker Desktop > Settings > Resources > File Sharing
# On Linux: Check SELinux/AppArmor permissions
```

#### **🔴 Performance Issues**

**Problem**: Slow database queries
```bash
# Check database stats
docker exec claude-company-db psql -U claude_user -d claude_company -c "
  SELECT schemaname, tablename, attname, n_distinct, correlation 
  FROM pg_stats WHERE tablename = 'tasks';
"

# Analyze query performance
docker exec claude-company-db psql -U claude_user -d claude_company -c "
  EXPLAIN ANALYZE SELECT * FROM tasks WHERE status = 'pending';
"
```

#### **🔴 Web Interface Issues**

**Problem**: Cannot access pgAdmin
```bash
# Check container logs
docker-compose logs pgadmin

# Verify correct URL and ports
echo "pgAdmin: http://localhost:8080"

# Clear browser cache and try again
```

### **7. Quick Health Check Commands**

```bash
# Complete system health check
#!/bin/bash
echo "=== Claude Company Database Health Check ==="

# Check Docker
echo "Docker version: $(docker --version)"

# Check containers
echo -e "\nContainer status:"
docker-compose ps

# Check database connectivity
echo -e "\nDatabase connectivity:"
docker exec claude-company-db psql -U claude_user -d claude_company -c "SELECT 'PostgreSQL Connected' as status;"

# Check web interfaces
echo -e "\nWeb interface availability:"
curl -s -o /dev/null -w "pgAdmin: %{http_code}\n" http://localhost:8080

echo -e "\nSetup complete! ✅"
```

## 🌐 API Usage Guide

### **Starting the API Server**

```bash
# Start with database enabled
./bin/ccs --api --port 8081

# Or with environment variables
PORT=8081 ./bin/ccs --api
```

**API Base URL:** `http://localhost:8081/api/v1`

### **Core Task Management Endpoints**

#### **Create Task**
```bash
# Create main task
curl -X POST http://localhost:8081/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Implement user authentication system",
    "mode": "manager",
    "pane_id": "pane_1",
    "priority": 3
  }'

# Create subtask
curl -X POST http://localhost:8081/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "parent_id": "01HXAMPLE123456789",
    "description": "Create JWT middleware",
    "mode": "worker",
    "pane_id": "pane_2",
    "priority": 2
  }'
```

#### **Get Tasks**
```bash
# Get tasks by pane
curl "http://localhost:8081/api/v1/tasks?pane_id=pane_1"

# Get tasks by status
curl "http://localhost:8081/api/v1/tasks?status=in_progress"

# Get child tasks
curl "http://localhost:8081/api/v1/tasks?parent_id=01HXAMPLE123456789"

# Get specific task
curl "http://localhost:8081/api/v1/tasks/01HXAMPLE123456789"
```

#### **Update Task**
```bash
# Update task details
curl -X PUT http://localhost:8081/api/v1/tasks/01HXAMPLE123456789 \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Updated task description",
    "status": "in_progress",
    "result": "Middleware implemented successfully"
  }'

# Update just status
curl -X PATCH http://localhost:8081/api/v1/tasks/01HXAMPLE123456789/status/completed

# Update status with propagation to related tasks
curl -X PATCH http://localhost:8081/api/v1/tasks/01HXAMPLE123456789/status-propagate/completed
```

#### **Task Hierarchy**
```bash
# Get complete task hierarchy
curl "http://localhost:8081/api/v1/tasks/01HXAMPLE123456789/hierarchy"
```

### **Progress Monitoring Endpoints**

#### **Get Progress Summary**
```bash
# Get progress for specific pane
curl "http://localhost:8081/api/v1/progress?pane_id=pane_1"

# Response example:
{
  "total_tasks": 10,
  "completed_tasks": 7,
  "pending_tasks": 2,
  "in_progress_tasks": 1,
  "progress_percent": 70.0
}
```

#### **Get Task Statistics**
```bash
# Get detailed statistics
curl "http://localhost:8081/api/v1/statistics?pane_id=pane_1"
```

## 🤝 Task Sharing Features

### **Share Individual Tasks**

```bash
# Share task with specific pane
curl -X POST http://localhost:8081/api/v1/tasks/01HXAMPLE123456789/share \
  -H "Content-Type: application/json" \
  -d '{
    "pane_id": "pane_3",
    "permission": "write"
  }'

# Share with all sibling tasks (same parent)
curl -X POST http://localhost:8081/api/v1/tasks/01HXAMPLE123456789/share-siblings

# Share with entire task family (parent + children)
curl -X POST http://localhost:8081/api/v1/tasks/01HXAMPLE123456789/share-family
```

### **Manage Task Shares**

```bash
# Get all shares for a task
curl "http://localhost:8081/api/v1/tasks/01HXAMPLE123456789/shares"

# Get all tasks shared with a pane
curl "http://localhost:8081/api/v1/shared-tasks?pane_id=pane_2"

# Remove share
curl -X DELETE http://localhost:8081/api/v1/tasks/01HXAMPLE123456789/share/pane_2
```

### **Permission Levels**
- **`read`**: View task details only (default)
- **`write`**: Can update task status and add progress
- **`admin`**: Full control including sharing and deletion

## ⚡ Asynchronous Execution Guide

### **Async Task Processing**

Claude Company supports asynchronous task execution for background processing and parallel work distribution:

#### **1. Enable Async Mode**
```bash
# Start with async processing enabled
./bin/ccs --async --workers 4

# Or combine with API mode
./bin/ccs --api --async --workers 4
```

#### **2. Create Async Tasks**
```bash
# Create task with async flag
curl -X POST http://localhost:8081/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Process large dataset analysis",
    "mode": "async_worker",
    "pane_id": "async_1",
    "priority": 1,
    "metadata": "{\"async\": true, \"timeout\": 300}"
  }'
```

#### **3. Monitor Async Progress**
```bash
# Check async task status
curl "http://localhost:8081/api/v1/tasks/01HXAMPLE123456789"

# Monitor all async tasks
curl "http://localhost:8081/api/v1/tasks?status=in_progress"
```

### **Async Execution Patterns**

#### **Parallel Task Distribution**
```bash
# Create parent task
PARENT_ID=$(curl -X POST http://localhost:8081/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Parallel data processing job",
    "mode": "manager",
    "pane_id": "manager_pane"
  }' | jq -r '.id')

# Create multiple async subtasks
for i in {1..4}; do
  curl -X POST http://localhost:8081/api/v1/tasks \
    -H "Content-Type: application/json" \
    -d "{
      \"parent_id\": \"$PARENT_ID\",
      \"description\": \"Process data chunk $i\",
      \"mode\": \"async_worker\",
      \"pane_id\": \"worker_$i\"
    }"
done
```

#### **Task Coordination with Auto-sharing**
```bash
# Create coordinated task that auto-shares with family
curl -X POST http://localhost:8081/api/v1/tasks/01HXAMPLE123456789/share-family

# All related tasks now have visibility into each other's progress
```

### **Health Check & Service Status**
```bash
# Check API health
curl "http://localhost:8081/health"

# Response:
{
  "status": "ok",
  "message": "Claude Company API is running"
}
```

## 🔄 Database Management

### **Stop/Start Services**
```bash
# Stop all services
docker-compose down

# Stop and remove volumes (clears all data)
docker-compose down -v

# Restart services
docker-compose restart

# View resource usage
docker-compose top
```

### **Backup & Restore**
```bash
# Backup database
docker exec claude-company-db pg_dump -U claude_user claude_company > backup.sql

# Restore database
docker exec -i claude-company-db psql -U claude_user claude_company < backup.sql
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