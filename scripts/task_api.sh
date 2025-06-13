#!/bin/bash

# Task API Client for Claude Company
# Handles shared task retrieval and status management via REST API

set -euo pipefail

# API Configuration
API_BASE_URL="${API_BASE_URL:-http://localhost:8080/api}"
API_TIMEOUT="${API_TIMEOUT:-5}"
MAX_RETRIES="${MAX_RETRIES:-3}"

# Source DB connection script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/db_connect.sh"

# HTTP request with retry logic
api_request() {
    local method="$1"
    local endpoint="$2"
    local data="${3:-}"
    local retry_count=0
    
    while [ $retry_count -lt $MAX_RETRIES ]; do
        local curl_cmd="curl -s --max-time $API_TIMEOUT -X $method"
        
        if [ -n "$data" ]; then
            curl_cmd="$curl_cmd -H 'Content-Type: application/json' -d '$data'"
        fi
        
        local response
        if response=$(eval "$curl_cmd '$API_BASE_URL$endpoint'" 2>/dev/null); then
            echo "$response"
            return 0
        else
            ((retry_count++))
            if [ $retry_count -lt $MAX_RETRIES ]; then
                echo "⚠️ API request failed, retrying... ($retry_count/$MAX_RETRIES)" >&2
                sleep 1
            fi
        fi
    done
    
    echo "❌ API request failed after $MAX_RETRIES attempts" >&2
    return 1
}

# Get tasks for current pane
get_my_tasks() {
    local pane_id="${CURRENT_PANE_ID:-$(tmux display-message -p '#{pane_id}')}"
    echo "📋 Fetching tasks for pane: $pane_id"
    
    # Try API first, fallback to direct DB
    if api_request "GET" "/tasks?pane_id=$pane_id" 2>/dev/null; then
        return 0
    else
        echo "⚠️ API unavailable, querying database directly..." >&2
        execute_sql "SELECT id, description, status, priority, created_at FROM tasks WHERE pane_id = '$pane_id' ORDER BY created_at DESC;" "table"
    fi
}

# Get shared tasks for current pane
get_shared_tasks() {
    local pane_id="${CURRENT_PANE_ID:-$(tmux display-message -p '#{pane_id}')}"
    echo "🔗 Fetching shared tasks for pane: $pane_id"
    
    # Try API first, fallback to direct DB
    if api_request "GET" "/tasks/shared?pane_id=$pane_id" 2>/dev/null; then
        return 0
    else
        echo "⚠️ API unavailable, querying database directly..." >&2
        execute_sql "
            SELECT t.id, t.description, t.status, t.priority, t.pane_id, ts.permission
            FROM tasks t
            INNER JOIN task_shares ts ON t.id = ts.task_id
            WHERE ts.shared_with_pane_id = '$pane_id'
            ORDER BY t.created_at DESC;
        " "table"
    fi
}

# Get parent task information
get_parent_task() {
    local task_id="$1"
    echo "⬆️ Fetching parent task for: $task_id"
    
    if api_request "GET" "/tasks/$task_id" 2>/dev/null; then
        return 0
    else
        execute_sql "
            SELECT p.id, p.description, p.status, p.pane_id
            FROM tasks t
            INNER JOIN tasks p ON t.parent_id = p.id
            WHERE t.id = '$task_id';
        " "table"
    fi
}

# Get sibling tasks
get_sibling_tasks() {
    local task_id="$1"
    echo "👥 Fetching sibling tasks for: $task_id"
    
    execute_sql "
        SELECT s.id, s.description, s.status, s.pane_id, s.priority
        FROM tasks t
        INNER JOIN tasks s ON t.parent_id = s.parent_id
        WHERE t.id = '$task_id' AND s.id != '$task_id'
        ORDER BY s.created_at ASC;
    " "table"
}

# Get task hierarchy
get_task_hierarchy() {
    local task_id="$1"
    echo "🌳 Fetching task hierarchy for: $task_id"
    
    if api_request "GET" "/tasks/$task_id/hierarchy" 2>/dev/null; then
        return 0
    else
        execute_sql "
            WITH RECURSIVE task_tree AS (
                SELECT id, parent_id, description, status, pane_id, 0 as level
                FROM tasks WHERE id = '$task_id'
                UNION ALL
                SELECT t.id, t.parent_id, t.description, t.status, t.pane_id, tt.level + 1
                FROM tasks t
                INNER JOIN task_tree tt ON t.parent_id = tt.id
            )
            SELECT id, description, status, pane_id, level
            FROM task_tree
            ORDER BY level, id;
        " "table"
    fi
}

# Update task status with propagation
update_task_status() {
    local task_id="$1"
    local new_status="$2"
    local result="${3:-}"
    
    echo "📝 Updating task $task_id status to: $new_status"
    
    local data="{\"status\": \"$new_status\""
    if [ -n "$result" ]; then
        data="$data, \"result\": \"$result\""
    fi
    data="$data}"
    
    if api_request "PUT" "/tasks/$task_id/status/$new_status" "$data" 2>/dev/null; then
        echo "✅ Task status updated successfully"
        return 0
    else
        echo "⚠️ API unavailable, updating database directly..." >&2
        local update_sql="UPDATE tasks SET status = '$new_status', updated_at = NOW()"
        if [ -n "$result" ]; then
            update_sql="$update_sql, result = '$result'"
        fi
        if [ "$new_status" = "completed" ]; then
            update_sql="$update_sql, completed_at = NOW()"
        fi
        update_sql="$update_sql WHERE id = '$task_id';"
        execute_sql "$update_sql"
    fi
}

# Share task with siblings
share_with_siblings() {
    local task_id="$1"
    echo "🔗 Sharing task $task_id with siblings"
    
    if api_request "POST" "/tasks/$task_id/share/siblings" 2>/dev/null; then
        echo "✅ Task shared with siblings successfully"
    else
        echo "⚠️ Failed to share task with siblings via API" >&2
        return 1
    fi
}

# Get progress statistics
get_progress_stats() {
    local pane_id="${CURRENT_PANE_ID:-$(tmux display-message -p '#{pane_id}')}"
    echo "📊 Fetching progress statistics for pane: $pane_id"
    
    if api_request "GET" "/progress?pane_id=$pane_id" 2>/dev/null; then
        return 0
    else
        execute_sql "
            SELECT 
                status,
                COUNT(*) as count,
                ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 2) as percentage
            FROM tasks 
            WHERE pane_id = '$pane_id'
            GROUP BY status
            ORDER BY status;
        " "table"
    fi
}

# Create task report
create_task_report() {
    local report_type="${1:-summary}"
    local pane_id="${CURRENT_PANE_ID:-$(tmux display-message -p '#{pane_id}')}"
    
    echo "📋 === Task Report for Pane $pane_id ==="
    echo "Generated at: $(date)"
    echo ""
    
    case $report_type in
        "detailed")
            echo "📋 My Tasks:"
            get_my_tasks
            echo ""
            echo "🔗 Shared Tasks:"
            get_shared_tasks
            echo ""
            echo "📊 Progress Statistics:"
            get_progress_stats
            ;;
        "summary"|*)
            get_progress_stats
            ;;
    esac
}

# Main function for interactive use
main() {
    local command="${1:-help}"
    
    case $command in
        "my-tasks")
            get_my_tasks
            ;;
        "shared-tasks")
            get_shared_tasks
            ;;
        "parent")
            [ $# -ge 2 ] || { echo "Usage: $0 parent <task_id>"; exit 1; }
            get_parent_task "$2"
            ;;
        "siblings")
            [ $# -ge 2 ] || { echo "Usage: $0 siblings <task_id>"; exit 1; }
            get_sibling_tasks "$2"
            ;;
        "hierarchy")
            [ $# -ge 2 ] || { echo "Usage: $0 hierarchy <task_id>"; exit 1; }
            get_task_hierarchy "$2"
            ;;
        "update-status")
            [ $# -ge 3 ] || { echo "Usage: $0 update-status <task_id> <status> [result]"; exit 1; }
            update_task_status "$2" "$3" "${4:-}"
            ;;
        "share-siblings")
            [ $# -ge 2 ] || { echo "Usage: $0 share-siblings <task_id>"; exit 1; }
            share_with_siblings "$2"
            ;;
        "progress")
            get_progress_stats
            ;;
        "report")
            create_task_report "${2:-summary}"
            ;;
        "help"|*)
            echo "Claude Company Task API Client"
            echo "Usage: $0 <command> [args...]"
            echo ""
            echo "Commands:"
            echo "  my-tasks              Get tasks for current pane"
            echo "  shared-tasks          Get shared tasks"
            echo "  parent <task_id>      Get parent task"
            echo "  siblings <task_id>    Get sibling tasks"
            echo "  hierarchy <task_id>   Get task hierarchy"
            echo "  update-status <id> <status> [result]  Update task status"
            echo "  share-siblings <id>   Share task with siblings"
            echo "  progress              Get progress statistics"
            echo "  report [detailed]     Create task report"
            echo "  help                  Show this help"
            ;;
    esac
}

# Execute if run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi