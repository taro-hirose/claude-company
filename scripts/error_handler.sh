#!/bin/bash

# Error Handling and Fallback System for Claude Company
# Comprehensive error recovery and alternative execution paths

set -euo pipefail

# Source required scripts
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Error handling configuration
ERROR_LOG_FILE="/tmp/claude_company_errors_$(date +%Y%m%d).log"
FALLBACK_DATA_DIR="/tmp/claude_company_fallback"
MAX_RETRY_ATTEMPTS=3
RETRY_DELAY=2
TIMEOUT_DURATION=10

# Ensure fallback directory exists
mkdir -p "$FALLBACK_DATA_DIR"

# Logging functions
log_error() {
    local error_message="$1"
    local timestamp="$(date '+%Y-%m-%d %H:%M:%S')"
    local pane_id="${CURRENT_PANE_ID:-unknown}"
    
    echo "[$timestamp] ERROR [Pane:$pane_id]: $error_message" | tee -a "$ERROR_LOG_FILE"
}

log_warning() {
    local warning_message="$1"
    local timestamp="$(date '+%Y-%m-%d %H:%M:%S')"
    local pane_id="${CURRENT_PANE_ID:-unknown}"
    
    echo "[$timestamp] WARN [Pane:$pane_id]: $warning_message" | tee -a "$ERROR_LOG_FILE"
}

log_info() {
    local info_message="$1"
    local timestamp="$(date '+%Y-%m-%d %H:%M:%S')"
    local pane_id="${CURRENT_PANE_ID:-unknown}"
    
    echo "[$timestamp] INFO [Pane:$pane_id]: $info_message"
}

# Generic retry mechanism
retry_with_backoff() {
    local command="$1"
    local description="$2"
    local max_attempts="${3:-$MAX_RETRY_ATTEMPTS}"
    local delay="${4:-$RETRY_DELAY}"
    
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        log_info "Attempting $description (attempt $attempt/$max_attempts)"
        
        if eval "$command"; then
            log_info "$description succeeded on attempt $attempt"
            return 0
        else
            local exit_code=$?
            log_warning "$description failed on attempt $attempt (exit code: $exit_code)"
            
            if [ $attempt -lt $max_attempts ]; then
                log_info "Waiting ${delay}s before retry..."
                sleep $delay
                delay=$((delay * 2))  # Exponential backoff
            fi
            
            ((attempt++))
        fi
    done
    
    log_error "$description failed after $max_attempts attempts"
    return 1
}

# Database connection with fallback
connect_db_with_fallback() {
    log_info "Attempting database connection with fallback"
    
    # Try primary connection
    if source "${SCRIPT_DIR}/db_connect.sh" && main >/dev/null 2>&1; then
        log_info "Primary database connection successful"
        return 0
    fi
    
    log_warning "Primary database connection failed, trying alternatives"
    
    # Fallback 1: Try with different connection parameters
    local fallback_hosts=("localhost" "127.0.0.1" "db" "postgres")
    local fallback_ports=("5432" "5433" "5434")
    
    for host in "${fallback_hosts[@]}"; do
        for port in "${fallback_ports[@]}"; do
            log_info "Trying fallback connection: $host:$port"
            
            export DB_HOST="$host"
            export DB_PORT="$port"
            
            if retry_with_backoff "source '${SCRIPT_DIR}/db_connect.sh' && test_db_connection" "fallback DB connection $host:$port" 2 1; then
                log_info "Fallback database connection successful: $host:$port"
                return 0
            fi
        done
    done
    
    # Fallback 2: Create local data cache
    log_warning "All database connections failed, creating local fallback"
    create_local_data_fallback
    return 2  # Special return code for fallback mode
}

# API connection with fallback
connect_api_with_fallback() {
    local endpoint="$1"
    local method="${2:-GET}"
    local data="${3:-}"
    
    log_info "Attempting API connection: $method $endpoint"
    
    # Try primary API
    if source "${SCRIPT_DIR}/task_api.sh" && api_request "$method" "$endpoint" "$data" >/dev/null 2>&1; then
        log_info "Primary API connection successful"
        source "${SCRIPT_DIR}/task_api.sh"
        api_request "$method" "$endpoint" "$data"
        return 0
    fi
    
    log_warning "Primary API connection failed, trying alternatives"
    
    # Fallback 1: Try different ports
    local fallback_ports=("8080" "8081" "3000" "8000")
    local original_url="$API_BASE_URL"
    
    for port in "${fallback_ports[@]}"; do
        export API_BASE_URL="http://localhost:$port/api"
        log_info "Trying fallback API: $API_BASE_URL"
        
        if retry_with_backoff "source '${SCRIPT_DIR}/task_api.sh' && api_request '$method' '$endpoint' '$data'" "fallback API $port" 2 1 >/dev/null 2>&1; then
            log_info "Fallback API connection successful: port $port"
            source "${SCRIPT_DIR}/task_api.sh"
            api_request "$method" "$endpoint" "$data"
            return 0
        fi
    done
    
    # Restore original URL
    export API_BASE_URL="$original_url"
    
    # Fallback 2: Direct database query
    log_warning "All API connections failed, trying direct database"
    if connect_db_with_fallback >/dev/null 2>&1; then
        log_info "Using direct database as API fallback"
        fallback_db_query_for_api "$endpoint" "$method" "$data"
        return 0
    fi
    
    # Fallback 3: Local data
    log_warning "Database also failed, using local fallback data"
    fallback_local_data_for_api "$endpoint"
    return 2
}

# Create local data fallback
create_local_data_fallback() {
    log_info "Creating local data fallback"
    
    local fallback_file="$FALLBACK_DATA_DIR/tasks_$(date +%Y%m%d).json"
    local pane_id="${CURRENT_PANE_ID:-$(tmux display-message -p '#{pane_id}' 2>/dev/null || echo 'unknown')}"
    
    # Create basic task structure
    cat > "$fallback_file" <<EOF
{
  "pane_id": "$pane_id",
  "timestamp": "$(date -Iseconds)",
  "tasks": [
    {
      "id": "fallback-001",
      "description": "Fallback task - Database unavailable",
      "status": "pending",
      "priority": 1,
      "pane_id": "$pane_id",
      "created_at": "$(date -Iseconds)"
    }
  ],
  "shared_tasks": [],
  "progress": {
    "total_tasks": 1,
    "completed_tasks": 0,
    "pending_tasks": 1,
    "progress_percent": 0
  }
}
EOF
    
    log_info "Local fallback data created: $fallback_file"
}

# Database query fallback for API
fallback_db_query_for_api() {
    local endpoint="$1"
    local method="$2"
    local data="$3"
    
    local pane_id="${CURRENT_PANE_ID:-$(tmux display-message -p '#{pane_id}' 2>/dev/null || echo 'unknown')}"
    
    case "$endpoint" in
        "/tasks?pane_id="*)
            log_info "Executing direct DB query for my tasks"
            source "${SCRIPT_DIR}/db_connect.sh"
            execute_sql "SELECT id, description, status, priority, created_at FROM tasks WHERE pane_id = '$pane_id' ORDER BY created_at DESC LIMIT 10;" "table"
            ;;
        "/tasks/shared?pane_id="*)
            log_info "Executing direct DB query for shared tasks"
            source "${SCRIPT_DIR}/db_connect.sh"
            execute_sql "SELECT t.id, t.description, t.status, ts.permission FROM tasks t INNER JOIN task_shares ts ON t.id = ts.task_id WHERE ts.shared_with_pane_id = '$pane_id';" "table"
            ;;
        "/progress?pane_id="*)
            log_info "Executing direct DB query for progress"
            source "${SCRIPT_DIR}/db_connect.sh"
            execute_sql "SELECT status, COUNT(*) as count FROM tasks WHERE pane_id = '$pane_id' GROUP BY status;" "table"
            ;;
        *)
            log_warning "No fallback DB query available for endpoint: $endpoint"
            return 1
            ;;
    esac
}

# Local data fallback for API
fallback_local_data_for_api() {
    local endpoint="$1"
    local fallback_file="$FALLBACK_DATA_DIR/tasks_$(date +%Y%m%d).json"
    
    if [ -f "$fallback_file" ]; then
        log_info "Using local fallback data for: $endpoint"
        
        case "$endpoint" in
            "/tasks?pane_id="*)
                if command -v jq >/dev/null 2>&1; then
                    jq -r '.tasks[] | [.id, .description, .status, .priority] | @csv' "$fallback_file" 2>/dev/null || cat "$fallback_file"
                else
                    cat "$fallback_file"
                fi
                ;;
            "/progress?pane_id="*)
                if command -v jq >/dev/null 2>&1; then
                    jq -r '.progress' "$fallback_file" 2>/dev/null || echo "No progress data available"
                else
                    echo "Local fallback: Progress data in JSON format (jq not available)"
                fi
                ;;
            *)
                cat "$fallback_file"
                ;;
        esac
    else
        log_error "No local fallback data available"
        echo "No fallback data available for: $endpoint"
        return 1
    fi
}

# Tmux operation with fallback
tmux_with_fallback() {
    local tmux_command="$1"
    local description="$2"
    
    log_info "Executing tmux command: $tmux_command"
    
    # Check if tmux is available
    if ! command -v tmux >/dev/null 2>&1; then
        log_error "tmux is not available"
        return 1
    fi
    
    # Check if we're in a tmux session
    if [ -z "${TMUX:-}" ]; then
        log_warning "Not in a tmux session, some operations may fail"
    fi
    
    # Try tmux command with retry
    if retry_with_backoff "$tmux_command" "$description" 2 1; then
        return 0
    else
        log_error "tmux command failed: $tmux_command"
        return 1
    fi
}

# Safe file operations
safe_file_operation() {
    local operation="$1"
    local file_path="$2"
    local backup_path="${file_path}.backup.$(date +%s)"
    
    case "$operation" in
        "read")
            if [ -f "$file_path" ]; then
                cat "$file_path"
            else
                log_error "File not found: $file_path"
                return 1
            fi
            ;;
        "write")
            local content="$3"
            # Create backup if file exists
            if [ -f "$file_path" ]; then
                cp "$file_path" "$backup_path"
                log_info "Backup created: $backup_path"
            fi
            
            # Write with error handling
            if echo "$content" > "$file_path"; then
                log_info "File written successfully: $file_path"
            else
                log_error "Failed to write file: $file_path"
                # Restore backup if available
                if [ -f "$backup_path" ]; then
                    mv "$backup_path" "$file_path"
                    log_info "Backup restored: $file_path"
                fi
                return 1
            fi
            ;;
        *)
            log_error "Unknown file operation: $operation"
            return 1
            ;;
    esac
}

# Health check function
run_health_check() {
    echo "🏥 === System Health Check ==="
    
    local health_status=0
    
    # Check tmux availability
    if command -v tmux >/dev/null 2>&1; then
        echo "✅ tmux: Available"
    else
        echo "❌ tmux: Not available"
        health_status=1
    fi
    
    # Check database connectivity
    if connect_db_with_fallback >/dev/null 2>&1; then
        echo "✅ Database: Connected"
    else
        echo "⚠️ Database: Using fallback"
    fi
    
    # Check API connectivity
    if connect_api_with_fallback "/health" >/dev/null 2>&1; then
        echo "✅ API: Connected"
    else
        echo "⚠️ API: Using fallback"
    fi
    
    # Check required tools
    local tools=("curl" "psql" "jq")
    for tool in "${tools[@]}"; do
        if command -v "$tool" >/dev/null 2>&1; then
            echo "✅ $tool: Available"
        else
            echo "⚠️ $tool: Not available"
        fi
    done
    
    # Check disk space for logs and fallback data
    local available_space=$(df /tmp | awk 'NR==2 {print $4}')
    if [ "$available_space" -gt 100000 ]; then
        echo "✅ Disk Space: Sufficient (/tmp: ${available_space}KB)"
    else
        echo "⚠️ Disk Space: Limited (/tmp: ${available_space}KB)"
    fi
    
    return $health_status
}

# Recovery operations
run_recovery_operations() {
    echo "🔧 === Running Recovery Operations ==="
    
    # Clean old logs
    find /tmp -name "claude_company_errors_*.log" -mtime +7 -delete 2>/dev/null || true
    echo "✅ Cleaned old error logs"
    
    # Clean old fallback data
    find "$FALLBACK_DATA_DIR" -name "tasks_*.json" -mtime +1 -delete 2>/dev/null || true
    echo "✅ Cleaned old fallback data"
    
    # Recreate fallback data
    create_local_data_fallback
    echo "✅ Recreated fallback data"
    
    # Test connections
    log_info "Testing all connections"
    run_health_check >/dev/null 2>&1
    echo "✅ Connection tests completed"
}

# Main function
main() {
    local command="${1:-help}"
    
    case $command in
        "test-db")
            connect_db_with_fallback
            ;;
        "test-api")
            connect_api_with_fallback "${2:-/health}"
            ;;
        "health")
            run_health_check
            ;;
        "recovery")
            run_recovery_operations
            ;;
        "logs")
            if [ -f "$ERROR_LOG_FILE" ]; then
                tail -n "${2:-20}" "$ERROR_LOG_FILE"
            else
                echo "No error log file found"
            fi
            ;;
        "clean")
            rm -f "$ERROR_LOG_FILE"
            rm -rf "$FALLBACK_DATA_DIR"
            echo "✅ Cleaned error logs and fallback data"
            ;;
        "help"|*)
            echo "Claude Company Error Handling System"
            echo "Usage: $0 <command> [args...]"
            echo ""
            echo "Commands:"
            echo "  test-db              Test database connection with fallback"
            echo "  test-api [endpoint]  Test API connection with fallback"
            echo "  health               Run system health check"
            echo "  recovery             Run recovery operations"
            echo "  logs [lines]         Show recent error logs"
            echo "  clean                Clean logs and fallback data"
            echo "  help                 Show this help"
            ;;
    esac
}

# Execute if run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi