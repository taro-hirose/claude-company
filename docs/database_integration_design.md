# Enhanced Database Integration Patterns for Claude Company Manager

## Executive Summary

This document outlines improved database integration patterns for the `buildManagerPrompt` method, leveraging the existing shell scripts and database infrastructure to provide the manager AI with real-time task context, hierarchical awareness, and intelligent decision-making capabilities.

## Current State Analysis

### Available Database Functions

#### From `db_connect.sh`:
- `execute_sql()` - Executes SQL with multiple output formats (table, json, csv)
- `detect_db_config()` - Auto-detects DB connection parameters
- `test_db_connection()` - Tests DB connectivity with retries
- `get_pane_context()` - Gets current pane context for database operations

#### From `task_api.sh`:
- `get_my_tasks()` - Gets tasks for current pane (API first, DB fallback)
- `get_shared_tasks()` - Gets shared tasks for current pane
- `get_parent_task()` - Gets parent task information
- `get_sibling_tasks()` - Gets sibling tasks
- `get_task_hierarchy()` - Gets full task hierarchy using recursive CTE
- `get_progress_stats()` - Gets progress statistics
- `update_task_status()` - Updates task status with propagation
- `create_task_report()` - Creates detailed task reports

### Current Database Schema Features
- Hierarchical task structure with parent-child relationships
- Task sharing mechanism between panes
- Progress tracking with detailed statistics
- ULID-based primary keys for distributed systems
- Rich metadata and status tracking

## Proposed Enhanced Integration Patterns

### 1. Database-Aware Manager Prompt Structure

```go
type DatabaseAwareManagerPrompt struct {
    BasePrompt       string
    CurrentContext   *TaskContext
    AvailableCommands map[string]DBCommand
    RealTimeData     *RealTimeTaskData
}

type TaskContext struct {
    CurrentPaneID    string
    ParentTaskID     *string
    ActiveTasks      []*Task
    SharedTasks      []*Task
    ProgressStats    *ProgressStats
    TaskHierarchy    *TaskHierarchy
}

type DBCommand struct {
    Name         string
    Description  string
    ShellScript  string
    OutputFormat string
    Parameters   []string
}
```

### 2. Enhanced buildManagerPrompt Method

```go
func (m *AIManager) buildManagerPrompt() string {
    // Initialize database context
    dbContext := m.initializeDatabaseContext()
    
    // Build prompt with real-time data
    prompt := m.buildBasePrompt()
    prompt += m.buildTaskContextSection(dbContext)
    prompt += m.buildAvailableCommandsSection()
    prompt += m.buildWorkflowTemplatesSection(dbContext)
    
    return prompt
}

func (m *AIManager) initializeDatabaseContext() *TaskContext {
    context := &TaskContext{
        CurrentPaneID: m.getCurrentPaneID(),
    }
    
    // Load current tasks
    context.ActiveTasks = m.loadActiveTasks()
    context.SharedTasks = m.loadSharedTasks()
    context.ProgressStats = m.loadProgressStats()
    
    // Load hierarchy if parent task exists
    if m.taskTracker.MainTask.ID != "" {
        context.ParentTaskID = &m.taskTracker.MainTask.ID
        context.TaskHierarchy = m.loadTaskHierarchy(m.taskTracker.MainTask.ID)
    }
    
    return context
}
```

### 3. Database Command Integration

#### Core Database Commands for Manager AI

```bash
# Task Query Commands
DB_COMMANDS=(
    "get_current_pane_tasks|Get all tasks for current pane|scripts/task_api.sh my-tasks"
    "get_shared_tasks|Get tasks shared with current pane|scripts/task_api.sh shared-tasks"
    "get_progress_overview|Get progress statistics|scripts/task_api.sh progress"
    "get_task_hierarchy|Get full task hierarchy|scripts/task_api.sh hierarchy"
    "get_sibling_tasks|Get sibling tasks|scripts/task_api.sh siblings"
    "get_parent_task|Get parent task info|scripts/task_api.sh parent"
)

# Task Management Commands  
DB_MANAGEMENT_COMMANDS=(
    "update_task_status|Update task status|scripts/task_api.sh update-status"
    "share_with_siblings|Share task with siblings|scripts/task_api.sh share-siblings"
    "create_detailed_report|Create detailed task report|scripts/task_api.sh report detailed"
    "get_overdue_tasks|Get overdue tasks|scripts/db_connect.sh execute_sql 'SELECT * FROM tasks WHERE due_date < NOW()'"
)

# Analytics Commands
DB_ANALYTICS_COMMANDS=(
    "get_completion_rate|Get task completion rate|scripts/db_connect.sh execute_sql 'SELECT status, COUNT(*) FROM tasks GROUP BY status'"
    "get_workload_distribution|Get workload by pane|scripts/db_connect.sh execute_sql 'SELECT pane_id, COUNT(*) FROM tasks GROUP BY pane_id'"
    "get_priority_breakdown|Get tasks by priority|scripts/db_connect.sh execute_sql 'SELECT priority, COUNT(*) FROM tasks GROUP BY priority'"
)
```

### 4. Real-Time Context Integration

#### Enhanced Prompt Template with Database Context

```markdown
# Database-Aware Manager Prompt Template

## Current Task Context (Real-Time)
{{if .TaskContext.ParentTaskID}}
**Parent Task**: {{.TaskContext.ParentTaskID}}
- Status: {{.ParentTask.Status}}
- Progress: {{.ParentTask.Progress}}%
{{end}}

**Current Pane**: {{.TaskContext.CurrentPaneID}}
**Active Tasks**: {{len .TaskContext.ActiveTasks}}
**Shared Tasks**: {{len .TaskContext.SharedTasks}}

### Task Progress Overview
{{range .TaskContext.ProgressStats}}
- {{.Status}}: {{.Count}} tasks ({{.Percentage}}%)
{{end}}

### Task Hierarchy
{{if .TaskContext.TaskHierarchy}}
{{range .TaskContext.TaskHierarchy.Children}}
- Level {{.Level}}: {{.Description}} ({{.Status}})
{{end}}
{{end}}

## Available Database Commands

### Query Commands
{{range .DBCommands.Query}}
- `{{.Name}}`: {{.Description}}
  Usage: {{.ShellScript}}
{{end}}

### Management Commands  
{{range .DBCommands.Management}}
- `{{.Name}}`: {{.Description}}
  Usage: {{.ShellScript}}
{{end}}

### Analytics Commands
{{range .DBCommands.Analytics}}
- `{{.Name}}`: {{.Description}}
  Usage: {{.ShellScript}}
{{end}}
```

### 5. Intelligent Database Integration Workflows

#### Smart Task Assignment Workflow

```bash
# Enhanced task assignment with database awareness
smart_task_assignment() {
    local task_description="$1"
    
    # Get current workload distribution
    workload=$(scripts/db_connect.sh execute_sql "
        SELECT pane_id, COUNT(*) as task_count, 
               AVG(CASE WHEN status = 'in_progress' THEN 1 ELSE 0 END) as load_factor
        FROM tasks 
        WHERE status IN ('pending', 'in_progress')
        GROUP BY pane_id
        ORDER BY load_factor ASC, task_count ASC
        LIMIT 1
    " "table")
    
    # Find optimal pane for assignment
    optimal_pane=$(echo "$workload" | tail -1 | awk '{print $1}')
    
    # Check if task requires specific expertise
    if [[ "$task_description" =~ (database|sql|schema) ]]; then
        # Assign to database-capable pane
        optimal_pane=$(find_database_expert_pane)
    elif [[ "$task_description" =~ (frontend|ui|react) ]]; then
        # Assign to frontend-capable pane
        optimal_pane=$(find_frontend_expert_pane)
    fi
    
    echo "$optimal_pane"
}
```

#### Progress Monitoring with Database Insights

```bash
# Enhanced progress monitoring
enhanced_progress_monitoring() {
    echo "📊 === COMPREHENSIVE PROGRESS ANALYSIS ==="
    
    # Overall progress
    scripts/task_api.sh progress
    
    # Bottleneck analysis
    echo ""
    echo "🚨 === BOTTLENECK ANALYSIS ==="
    scripts/db_connect.sh execute_sql "
        SELECT pane_id, status, COUNT(*) as count,
               AVG(EXTRACT(EPOCH FROM (NOW() - created_at))/3600) as avg_hours
        FROM tasks 
        WHERE status = 'in_progress'
        GROUP BY pane_id, status
        HAVING AVG(EXTRACT(EPOCH FROM (NOW() - created_at))/3600) > 2
        ORDER BY avg_hours DESC
    " "table"
    
    # Completion velocity
    echo ""
    echo "⚡ === COMPLETION VELOCITY ==="
    scripts/db_connect.sh execute_sql "
        SELECT DATE(completed_at) as date, COUNT(*) as completed_tasks
        FROM tasks 
        WHERE completed_at >= NOW() - INTERVAL '7 days'
        GROUP BY DATE(completed_at)
        ORDER BY date DESC
    " "table"
    
    # Risk assessment
    echo ""
    echo "⚠️ === RISK ASSESSMENT ==="
    scripts/db_connect.sh execute_sql "
        SELECT id, description, pane_id,
               EXTRACT(EPOCH FROM (NOW() - created_at))/3600 as hours_since_creation
        FROM tasks 
        WHERE status = 'in_progress' 
        AND EXTRACT(EPOCH FROM (NOW() - created_at))/3600 > 4
        ORDER BY hours_since_creation DESC
    " "table"
}
```

### 6. Database-Driven Decision Making

#### Intelligent Task Decomposition

```bash
# Database-aware task decomposition
intelligent_task_decomposition() {
    local main_task="$1"
    
    # Analyze similar historical tasks
    echo "🧠 === INTELLIGENT TASK ANALYSIS ==="
    scripts/db_connect.sh execute_sql "
        SELECT description, 
               COUNT(*) as subtask_count,
               AVG(EXTRACT(EPOCH FROM (completed_at - created_at))/3600) as avg_completion_hours
        FROM tasks 
        WHERE parent_id IS NOT NULL
        AND description ILIKE '%$(echo "$main_task" | cut -d' ' -f1-2)%'
        GROUP BY description
        ORDER BY subtask_count DESC
        LIMIT 5
    " "table"
    
    # Get optimal team size based on task complexity
    optimal_team_size=$(estimate_team_size_from_history "$main_task")
    
    # Suggest decomposition strategy
    echo ""
    echo "💡 === DECOMPOSITION STRATEGY ==="
    echo "Recommended team size: $optimal_team_size panes"
    echo "Based on historical data for similar tasks"
}
```

### 7. Enhanced Quality Management

#### Database-Driven Quality Assurance

```bash
# Comprehensive quality management with database insights
database_driven_quality_assurance() {
    echo "🔍 === DATABASE-DRIVEN QUALITY ANALYSIS ==="
    
    # Code quality metrics
    scripts/db_connect.sh execute_sql "
        SELECT pane_id,
               COUNT(*) as total_tasks,
               COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed,
               COUNT(CASE WHEN result LIKE '%error%' OR result LIKE '%failed%' THEN 1 END) as error_count,
               ROUND(
                   COUNT(CASE WHEN status = 'completed' THEN 1 END) * 100.0 / COUNT(*), 2
               ) as success_rate
        FROM tasks
        WHERE created_at >= NOW() - INTERVAL '1 day'
        GROUP BY pane_id
        ORDER BY success_rate DESC
    " "table"
    
    # Integration test recommendations
    echo ""
    echo "🧪 === INTEGRATION TEST RECOMMENDATIONS ==="
    scripts/db_connect.sh execute_sql "
        SELECT DISTINCT pane_id
        FROM tasks 
        WHERE status = 'completed' 
        AND created_at >= NOW() - INTERVAL '1 hour'
        AND description ILIKE '%implement%'
    " "table"
}
```

### 8. Real-Time Reporting and Analytics

#### Comprehensive Reporting System

```bash
# Generate comprehensive manager dashboard
generate_manager_dashboard() {
    echo "📊 === CLAUDE COMPANY MANAGER DASHBOARD ==="
    echo "Generated at: $(date)"
    echo ""
    
    # Executive summary
    echo "📈 === EXECUTIVE SUMMARY ==="
    scripts/task_api.sh progress
    
    # Team performance
    echo ""
    echo "👥 === TEAM PERFORMANCE ==="
    scripts/db_connect.sh execute_sql "
        SELECT pane_id as 'Pane',
               COUNT(*) as 'Total Tasks',
               COUNT(CASE WHEN status = 'completed' THEN 1 END) as 'Completed',
               COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as 'In Progress',
               ROUND(AVG(priority), 1) as 'Avg Priority'
        FROM tasks
        GROUP BY pane_id
        ORDER BY COUNT(*) DESC
    " "table"
    
    # Recent activity
    echo ""
    echo "🕐 === RECENT ACTIVITY (Last 2 hours) ==="
    scripts/db_connect.sh execute_sql "
        SELECT LEFT(description, 50) as 'Task',
               pane_id as 'Pane',
               status as 'Status',
               TO_CHAR(updated_at, 'HH24:MI') as 'Time'
        FROM tasks
        WHERE updated_at >= NOW() - INTERVAL '2 hours'
        ORDER BY updated_at DESC
        LIMIT 10
    " "table"
    
    # Bottlenecks and recommendations
    echo ""
    echo "⚠️ === BOTTLENECKS & RECOMMENDATIONS ==="
    identify_bottlenecks_and_recommendations
}
```

## Implementation Roadmap

### Phase 1: Core Integration (Week 1)
1. ✅ Integrate `db_connect.sh` functions into manager prompt
2. ✅ Add real-time task context loading
3. ✅ Implement basic database command integration

### Phase 2: Enhanced Analytics (Week 2)
1. ✅ Add progress monitoring with database insights
2. ✅ Implement intelligent task assignment
3. ✅ Create database-driven quality assurance

### Phase 3: Advanced Features (Week 3)
1. ✅ Build comprehensive reporting system
2. ✅ Add predictive analytics for task completion
3. ✅ Implement risk assessment and mitigation

### Phase 4: Optimization (Week 4)
1. ✅ Performance optimization for database queries
2. ✅ Caching layer for frequently accessed data
3. ✅ Advanced ML-based recommendations

## Key Benefits

1. **Real-Time Awareness**: Manager AI has complete visibility into current task state
2. **Data-Driven Decisions**: Task assignments based on actual workload and performance data
3. **Predictive Management**: Identify bottlenecks and risks before they become critical
4. **Quality Assurance**: Continuous monitoring and quality metrics
5. **Historical Learning**: Learn from past task patterns to improve future planning
6. **Automated Reporting**: Generate comprehensive status reports automatically

## Integration Points

### In `manager.go`:
```go
func (m *AIManager) buildManagerPrompt() string {
    // Load real-time database context
    dbContext := m.loadDatabaseContext()
    
    // Build enhanced prompt with database integration
    return m.buildDatabaseAwarePrompt(dbContext)
}
```

### Shell Script Integration:
```bash
# Initialize database context in prompt
source scripts/db_connect.sh
source scripts/task_api.sh

# Detect database configuration
detect_db_config

# Test connection
if test_db_connection; then
    # Load current context into prompt
    export DB_CONTEXT_READY=true
fi
```

This enhanced database integration pattern transforms the manager AI from a static prompt generator into an intelligent, data-driven project management system with real-time awareness and predictive capabilities.