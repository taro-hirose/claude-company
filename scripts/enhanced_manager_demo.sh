#!/bin/bash

# Enhanced Manager Demo Script
# Demonstrates the database integration capabilities of the enhanced manager prompt

set -euo pipefail

# Source required scripts
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/db_connect.sh"
source "$SCRIPT_DIR/task_api.sh"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Header
echo -e "${PURPLE}========================================${NC}"
echo -e "${PURPLE} Enhanced Database-Aware Manager Demo${NC}"
echo -e "${PURPLE}========================================${NC}"
echo ""

# Demo function: Real-time database context loading
demo_database_context() {
    echo -e "${CYAN}🔍 === Database Context Loading Demo ===${NC}"
    echo ""
    
    # Step 1: Database connection
    echo -e "${YELLOW}Step 1: Testing database connection...${NC}"
    if detect_db_config && test_db_connection; then
        echo -e "${GREEN}✅ Database connection established${NC}"
    else
        echo -e "${RED}❌ Database connection failed${NC}"
        return 1
    fi
    echo ""
    
    # Step 2: Current pane context
    echo -e "${YELLOW}Step 2: Loading current pane context...${NC}"
    get_pane_context
    echo ""
    
    # Step 3: Active tasks
    echo -e "${YELLOW}Step 3: Loading active tasks...${NC}"
    get_my_tasks
    echo ""
    
    # Step 4: Shared tasks
    echo -e "${YELLOW}Step 4: Loading shared tasks...${NC}"
    get_shared_tasks
    echo ""
    
    # Step 5: Progress statistics
    echo -e "${YELLOW}Step 5: Loading progress statistics...${NC}"
    get_progress_stats
    echo ""
}

# Demo function: Workload analysis
demo_workload_analysis() {
    echo -e "${CYAN}📊 === Workload Analysis Demo ===${NC}"
    echo ""
    
    echo -e "${YELLOW}Analyzing workload distribution across panes...${NC}"
    execute_sql "
        SELECT 
            pane_id as 'Pane ID',
            COUNT(*) as 'Total Tasks',
            COUNT(CASE WHEN status = 'pending' THEN 1 END) as 'Pending',
            COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as 'In Progress',
            COUNT(CASE WHEN status = 'completed' THEN 1 END) as 'Completed',
            ROUND(AVG(priority), 1) as 'Avg Priority'
        FROM tasks
        GROUP BY pane_id
        ORDER BY COUNT(*) DESC
    " "table"
    echo ""
    
    echo -e "${YELLOW}Identifying overloaded panes...${NC}"
    execute_sql "
        SELECT 
            pane_id as 'Overloaded Pane',
            COUNT(*) as 'Active Tasks'
        FROM tasks 
        WHERE status IN ('pending', 'in_progress')
        GROUP BY pane_id
        HAVING COUNT(*) > 3
        ORDER BY COUNT(*) DESC
    " "table"
    echo ""
}

# Demo function: Bottleneck detection
demo_bottleneck_detection() {
    echo -e "${CYAN}⚠️ === Bottleneck Detection Demo ===${NC}"
    echo ""
    
    echo -e "${YELLOW}Detecting long-running tasks...${NC}"
    execute_sql "
        SELECT 
            id as 'Task ID',
            LEFT(description, 40) as 'Description',
            pane_id as 'Pane',
            status as 'Status',
            ROUND(EXTRACT(EPOCH FROM (NOW() - created_at))/3600, 1) as 'Hours Running'
        FROM tasks 
        WHERE status = 'in_progress' 
        AND EXTRACT(EPOCH FROM (NOW() - created_at))/3600 > 1
        ORDER BY EXTRACT(EPOCH FROM (NOW() - created_at)) DESC
    " "table"
    echo ""
    
    echo -e "${YELLOW}Pane performance analysis...${NC}"
    execute_sql "
        SELECT 
            pane_id as 'Pane',
            AVG(EXTRACT(EPOCH FROM (completed_at - created_at))/3600) as 'Avg Completion Hours',
            COUNT(*) as 'Completed Tasks'
        FROM tasks 
        WHERE status = 'completed' 
        AND completed_at IS NOT NULL
        GROUP BY pane_id
        ORDER BY AVG(EXTRACT(EPOCH FROM (completed_at - created_at))/3600) ASC
    " "table"
    echo ""
}

# Demo function: Intelligent recommendations
demo_intelligent_recommendations() {
    echo -e "${CYAN}🧠 === Intelligent Recommendations Demo ===${NC}"
    echo ""
    
    # Get current workload stats
    echo -e "${YELLOW}Analyzing current workload for recommendations...${NC}"
    
    # Count active tasks
    active_tasks=$(execute_sql "SELECT COUNT(*) FROM tasks WHERE status IN ('pending', 'in_progress')" "csv" | tail -1)
    echo -e "Active tasks: ${BLUE}$active_tasks${NC}"
    
    # Count completed tasks today
    completed_today=$(execute_sql "SELECT COUNT(*) FROM tasks WHERE status = 'completed' AND DATE(completed_at) = CURRENT_DATE" "csv" | tail -1)
    echo -e "Completed today: ${GREEN}$completed_today${NC}"
    
    # Generate recommendations
    echo ""
    echo -e "${YELLOW}📋 Generated Recommendations:${NC}"
    
    if [ "$active_tasks" -gt 10 ]; then
        echo -e "${RED}🔄 High Load Detected:${NC} Consider creating additional child panes"
        echo "   Suggested action: tmux split-window -h -t claude-squad"
    fi
    
    if [ "$completed_today" -eq 0 ]; then
        echo -e "${YELLOW}📈 Low Activity:${NC} No tasks completed today. Review progress with teams"
        echo "   Suggested action: Run progress checks on all child panes"
    fi
    
    # Check for pane imbalance
    pane_count=$(execute_sql "SELECT COUNT(DISTINCT pane_id) FROM tasks WHERE status IN ('pending', 'in_progress')" "csv" | tail -1)
    if [ "$pane_count" -gt 1 ]; then
        echo -e "${BLUE}⚖️ Load Balancing:${NC} Review workload distribution across $pane_count panes"
        echo "   Suggested action: Check if tasks can be redistributed"
    fi
    
    echo ""
}

# Demo function: Historical analysis
demo_historical_analysis() {
    echo -e "${CYAN}📈 === Historical Analysis Demo ===${NC}"
    echo ""
    
    echo -e "${YELLOW}Task completion trends (last 7 days)...${NC}"
    execute_sql "
        SELECT 
            DATE(completed_at) as 'Date',
            COUNT(*) as 'Completed Tasks',
            ROUND(AVG(EXTRACT(EPOCH FROM (completed_at - created_at))/3600), 1) as 'Avg Hours'
        FROM tasks 
        WHERE completed_at >= NOW() - INTERVAL '7 days'
        AND completed_at IS NOT NULL
        GROUP BY DATE(completed_at)
        ORDER BY DATE(completed_at) DESC
    " "table"
    echo ""
    
    echo -e "${YELLOW}Most productive panes...${NC}"
    execute_sql "
        SELECT 
            pane_id as 'Pane',
            COUNT(*) as 'Completed Tasks',
            ROUND(AVG(priority), 1) as 'Avg Priority',
            COUNT(CASE WHEN result LIKE '%success%' THEN 1 END) as 'Success Count'
        FROM tasks 
        WHERE status = 'completed'
        GROUP BY pane_id
        ORDER BY COUNT(*) DESC
        LIMIT 5
    " "table"
    echo ""
}

# Demo function: Real-time dashboard
demo_realtime_dashboard() {
    echo -e "${CYAN}📊 === Real-time Dashboard Demo ===${NC}"
    echo ""
    
    echo -e "${PURPLE}=== CLAUDE COMPANY REAL-TIME DASHBOARD ===${NC}"
    echo -e "Generated at: $(date)"
    echo ""
    
    # Executive summary
    echo -e "${YELLOW}📈 Executive Summary${NC}"
    get_progress_stats
    echo ""
    
    # Team performance
    echo -e "${YELLOW}👥 Team Performance${NC}"
    execute_sql "
        SELECT 
            pane_id as 'Team',
            COUNT(*) as 'Total',
            COUNT(CASE WHEN status = 'completed' THEN 1 END) as 'Done',
            COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as 'Active',
            ROUND(COUNT(CASE WHEN status = 'completed' THEN 1 END) * 100.0 / COUNT(*), 1) as 'Success %'
        FROM tasks
        GROUP BY pane_id
        ORDER BY COUNT(*) DESC
    " "table"
    echo ""
    
    # Recent activity
    echo -e "${YELLOW}🕐 Recent Activity (Last 2 hours)${NC}"
    execute_sql "
        SELECT 
            LEFT(description, 30) as 'Task',
            pane_id as 'Pane',
            status as 'Status',
            TO_CHAR(updated_at, 'HH24:MI') as 'Time'
        FROM tasks
        WHERE updated_at >= NOW() - INTERVAL '2 hours'
        ORDER BY updated_at DESC
        LIMIT 10
    " "table"
    echo ""
    
    # Alerts
    echo -e "${YELLOW}🚨 Alerts & Warnings${NC}"
    stuck_tasks=$(execute_sql "SELECT COUNT(*) FROM tasks WHERE status = 'in_progress' AND EXTRACT(EPOCH FROM (NOW() - created_at))/3600 > 3" "csv" | tail -1)
    if [ "$stuck_tasks" -gt 0 ]; then
        echo -e "${RED}⚠️ $stuck_tasks tasks have been running for more than 3 hours${NC}"
    fi
    
    pending_tasks=$(execute_sql "SELECT COUNT(*) FROM tasks WHERE status = 'pending' AND EXTRACT(EPOCH FROM (NOW() - created_at))/3600 > 1" "csv" | tail -1)
    if [ "$pending_tasks" -gt 0 ]; then
        echo -e "${YELLOW}⏳ $pending_tasks tasks have been pending for more than 1 hour${NC}"
    fi
    
    if [ "$stuck_tasks" -eq 0 ] && [ "$pending_tasks" -eq 0 ]; then
        echo -e "${GREEN}✅ No alerts - system operating normally${NC}"
    fi
    echo ""
}

# Demo function: Task assignment optimization
demo_task_assignment() {
    echo -e "${CYAN}🎯 === Task Assignment Optimization Demo ===${NC}"
    echo ""
    
    echo -e "${YELLOW}Finding optimal pane for new tasks...${NC}"
    echo ""
    
    # Show current pane capabilities based on history
    echo -e "${BLUE}Pane Specialization Analysis:${NC}"
    execute_sql "
        SELECT 
            pane_id as 'Pane',
            COUNT(CASE WHEN description ILIKE '%database%' OR description ILIKE '%sql%' THEN 1 END) as 'DB Tasks',
            COUNT(CASE WHEN description ILIKE '%frontend%' OR description ILIKE '%ui%' THEN 1 END) as 'Frontend Tasks',
            COUNT(CASE WHEN description ILIKE '%api%' OR description ILIKE '%backend%' THEN 1 END) as 'Backend Tasks',
            COUNT(CASE WHEN description ILIKE '%test%' THEN 1 END) as 'Test Tasks'
        FROM tasks
        WHERE status = 'completed'
        GROUP BY pane_id
        ORDER BY COUNT(*) DESC
    " "table"
    echo ""
    
    # Show current load for assignment decision
    echo -e "${BLUE}Current Load for Assignment Decision:${NC}"
    execute_sql "
        SELECT 
            pane_id as 'Pane',
            COUNT(*) as 'Active Tasks',
            ROUND(AVG(EXTRACT(EPOCH FROM (NOW() - created_at))/3600), 1) as 'Avg Task Age (hours)',
            CASE 
                WHEN COUNT(*) < 3 THEN 'Low Load - Optimal'
                WHEN COUNT(*) < 6 THEN 'Medium Load - Acceptable'
                ELSE 'High Load - Consider alternatives'
            END as 'Recommendation'
        FROM tasks 
        WHERE status IN ('pending', 'in_progress')
        GROUP BY pane_id
        ORDER BY COUNT(*) ASC
    " "table"
    echo ""
}

# Main demo execution
main() {
    echo -e "${GREEN}Starting Enhanced Database Integration Demo...${NC}"
    echo ""
    
    # Run all demo functions
    demo_database_context
    echo -e "${PURPLE}----------------------------------------${NC}"
    
    demo_workload_analysis
    echo -e "${PURPLE}----------------------------------------${NC}"
    
    demo_bottleneck_detection
    echo -e "${PURPLE}----------------------------------------${NC}"
    
    demo_intelligent_recommendations
    echo -e "${PURPLE}----------------------------------------${NC}"
    
    demo_historical_analysis
    echo -e "${PURPLE}----------------------------------------${NC}"
    
    demo_task_assignment
    echo -e "${PURPLE}----------------------------------------${NC}"
    
    demo_realtime_dashboard
    echo -e "${PURPLE}----------------------------------------${NC}"
    
    echo -e "${GREEN}✅ Enhanced Database Integration Demo Complete!${NC}"
    echo ""
    echo -e "${CYAN}The manager AI now has access to all this real-time data for intelligent decision making.${NC}"
    echo ""
}

# Script execution
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi