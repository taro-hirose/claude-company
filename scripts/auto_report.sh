#!/bin/bash

# Automated Reporting System for Claude Company
# Handles standardized progress reports and parent pane communication

set -euo pipefail

# Source required scripts
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/db_connect.sh"
source "${SCRIPT_DIR}/task_api.sh"
source "${SCRIPT_DIR}/pane_info.sh"

# Report templates and formatting
REPORT_FORMAT_COMPLETION="実装完了：%s - %s"
REPORT_FORMAT_PROGRESS="進捗報告：%s%% - %s"
REPORT_FORMAT_ERROR="エラー報告：%s - %s"
REPORT_FORMAT_STATUS="ステータス更新：%s → %s"

# Get parent pane ID (usually the first pane in session)
get_parent_pane() {
    local session_name="${CURRENT_SESSION:-$(tmux display-message -p '#{session_name}')}"
    local current_pane="${CURRENT_PANE_ID:-$(tmux display-message -p '#{pane_id}')}"
    
    # Get all panes and assume first pane is parent
    local panes=($(tmux list-panes -s -t "$session_name" -F "#{pane_id}"))
    
    for pane in "${panes[@]}"; do
        if [ "$pane" != "$current_pane" ]; then
            echo "$pane"
            return 0
        fi
    done
    
    echo "${panes[0]}"  # Fallback to first pane
}

# Send report to parent pane
send_to_parent() {
    local message="$1"
    local parent_pane="$(get_parent_pane)"
    local timestamp="$(date '+%H:%M:%S')"
    
    echo "📤 Sending report to parent pane $parent_pane: $message"
    
    # Format message with timestamp and pane identification
    local formatted_message="[$timestamp] Pane ${CURRENT_PANE_ID}: $message"
    
    if tmux send-keys -t "$parent_pane" "$formatted_message" Enter 2>/dev/null; then
        echo "✅ Report sent successfully"
        return 0
    else
        echo "❌ Failed to send report to parent pane"
        return 1
    fi
}

# Report task completion
report_completion() {
    local file_path="$1"
    local description="$2"
    local task_id="${3:-}"
    
    local message=$(printf "$REPORT_FORMAT_COMPLETION" "$file_path" "$description")
    
    # Update task status if task_id provided
    if [ -n "$task_id" ]; then
        update_task_status "$task_id" "completed" "$description" 2>/dev/null || true
    fi
    
    send_to_parent "$message"
}

# Report progress status
report_progress() {
    local progress_percent="$1"
    local current_work="$2"
    local task_id="${3:-}"
    
    local message=$(printf "$REPORT_FORMAT_PROGRESS" "$progress_percent" "$current_work")
    
    # Update task status if task_id provided
    if [ -n "$task_id" ]; then
        update_task_status "$task_id" "in_progress" "Progress: $progress_percent% - $current_work" 2>/dev/null || true
    fi
    
    send_to_parent "$message"
}

# Report error
report_error() {
    local error_description="$1"
    local assistance_request="$2"
    local task_id="${3:-}"
    
    local message=$(printf "$REPORT_FORMAT_ERROR" "$error_description" "$assistance_request")
    
    # Update task status if task_id provided
    if [ -n "$task_id" ]; then
        update_task_status "$task_id" "failed" "Error: $error_description" 2>/dev/null || true
    fi
    
    send_to_parent "$message"
}

# Report status change
report_status_change() {
    local old_status="$1"
    local new_status="$2"
    local task_id="${3:-}"
    local details="${4:-}"
    
    local message=$(printf "$REPORT_FORMAT_STATUS" "$old_status" "$new_status")
    if [ -n "$details" ]; then
        message="$message - $details"
    fi
    
    # Update task status if task_id provided
    if [ -n "$task_id" ]; then
        update_task_status "$task_id" "$new_status" "$details" 2>/dev/null || true
    fi
    
    send_to_parent "$message"
}

# Generate automated progress report
generate_progress_report() {
    local detailed="${1:-false}"
    
    echo "📊 Generating automated progress report..."
    
    # Collect current task information
    local my_tasks_info
    my_tasks_info=$(get_progress_stats 2>/dev/null || echo "No task data available")
    
    # Count tasks by status
    local total_tasks pending_tasks completed_tasks in_progress_tasks
    if command -v jq >/dev/null 2>&1; then
        # If we have JSON response, parse it
        total_tasks=$(echo "$my_tasks_info" | jq -r '.total_tasks // 0' 2>/dev/null || echo "0")
        completed_tasks=$(echo "$my_tasks_info" | jq -r '.completed_tasks // 0' 2>/dev/null || echo "0")
        pending_tasks=$(echo "$my_tasks_info" | jq -r '.pending_tasks // 0' 2>/dev/null || echo "0")
        in_progress_tasks=$(echo "$my_tasks_info" | jq -r '.in_progress_tasks // 0' 2>/dev/null || echo "0")
    else
        # Parse from table output or default values
        total_tasks="0"
        completed_tasks="0"
        pending_tasks="0"
        in_progress_tasks="0"
    fi
    
    # Calculate progress percentage
    local progress_percent=0
    if [ "$total_tasks" -gt 0 ]; then
        progress_percent=$((completed_tasks * 100 / total_tasks))
    fi
    
    # Generate report message
    local report_message="自動進捗報告 - 完了:${completed_tasks}/${total_tasks}(${progress_percent}%) 進行中:${in_progress_tasks} 待機:${pending_tasks}"
    
    if [ "$detailed" = "true" ]; then
        report_message="$report_message - 詳細: $(date '+%Y-%m-%d %H:%M')"
    fi
    
    send_to_parent "$report_message"
}

# Start periodic reporting
start_periodic_reporting() {
    local interval="${1:-1800}"  # Default 30 minutes
    local detailed="${2:-false}"
    
    echo "⏰ Starting periodic reporting (every ${interval} seconds)"
    echo "Press Ctrl+C to stop"
    
    # Create lock file to prevent multiple instances
    local lock_file="/tmp/claude_auto_report_${CURRENT_PANE_ID}.lock"
    
    if [ -f "$lock_file" ]; then
        echo "⚠️ Periodic reporting already running for this pane"
        return 1
    fi
    
    echo $$ > "$lock_file"
    
    # Cleanup function
    cleanup() {
        rm -f "$lock_file"
        echo "🛑 Periodic reporting stopped"
        exit 0
    }
    
    trap cleanup INT TERM
    
    # Initial report
    send_to_parent "🤖 自動レポート開始 - ${interval}秒間隔で進捗を報告します"
    
    while true; do
        sleep "$interval"
        generate_progress_report "$detailed"
    done
}

# Stop periodic reporting
stop_periodic_reporting() {
    local lock_file="/tmp/claude_auto_report_${CURRENT_PANE_ID}.lock"
    
    if [ -f "$lock_file" ]; then
        local pid=$(cat "$lock_file")
        if kill -0 "$pid" 2>/dev/null; then
            kill "$pid"
            echo "✅ Stopped periodic reporting (PID: $pid)"
        else
            echo "⚠️ Periodic reporting process not found"
        fi
        rm -f "$lock_file"
    else
        echo "⚠️ No periodic reporting lock file found"
    fi
}

# Check if periodic reporting is running
check_periodic_status() {
    local lock_file="/tmp/claude_auto_report_${CURRENT_PANE_ID}.lock"
    
    if [ -f "$lock_file" ]; then
        local pid=$(cat "$lock_file")
        if kill -0 "$pid" 2>/dev/null; then
            echo "✅ Periodic reporting is running (PID: $pid)"
        else
            echo "❌ Periodic reporting lock file exists but process is dead"
            rm -f "$lock_file"
        fi
    else
        echo "❌ Periodic reporting is not running"
    fi
}

# Send startup notification
send_startup_notification() {
    local pane_role="${1:-worker}"
    local startup_message="🚀 ${pane_role}ペイン起動完了 - Pane ${CURRENT_PANE_ID} がオンラインになりました"
    
    send_to_parent "$startup_message"
}

# Send shutdown notification
send_shutdown_notification() {
    local pane_role="${1:-worker}"
    local shutdown_message="🛑 ${pane_role}ペイン終了 - Pane ${CURRENT_PANE_ID} がオフラインになります"
    
    send_to_parent "$shutdown_message"
}

# Create report summary for current session
create_session_summary() {
    local output_file="${1:-/tmp/session_summary_$(date +%Y%m%d_%H%M%S).txt}"
    
    echo "📋 Creating session summary..."
    
    {
        echo "# Claude Company Session Summary"
        echo "Generated: $(date)"
        echo "Session: ${CURRENT_SESSION}"
        echo "Reporting Pane: ${CURRENT_PANE_ID}"
        echo ""
        
        echo "## Task Statistics"
        get_progress_stats 2>/dev/null || echo "No task statistics available"
        echo ""
        
        echo "## Pane Information"
        get_full_pane_context 2>/dev/null || echo "No pane context available"
        echo ""
        
        echo "## Recent Activity"
        capture_pane_content "${CURRENT_PANE_ID}" 10 2>/dev/null || echo "No recent activity captured"
        
    } | tee "$output_file"
    
    echo "✅ Session summary saved to: $output_file"
    send_to_parent "📋 セッションサマリー作成完了: $output_file"
}

# Main interactive function
main() {
    local command="${1:-help}"
    
    # Ensure we have pane context
    get_full_pane_context >/dev/null 2>&1 || true
    
    case $command in
        "completion")
            [ $# -ge 3 ] || { echo "Usage: $0 completion <file_path> <description> [task_id]"; exit 1; }
            report_completion "$2" "$3" "${4:-}"
            ;;
        "progress")
            [ $# -ge 3 ] || { echo "Usage: $0 progress <percent> <current_work> [task_id]"; exit 1; }
            report_progress "$2" "$3" "${4:-}"
            ;;
        "error")
            [ $# -ge 3 ] || { echo "Usage: $0 error <error_desc> <assistance_req> [task_id]"; exit 1; }
            report_error "$2" "$3" "${4:-}"
            ;;
        "status")
            [ $# -ge 3 ] || { echo "Usage: $0 status <old_status> <new_status> [task_id] [details]"; exit 1; }
            report_status_change "$2" "$3" "${4:-}" "${5:-}"
            ;;
        "auto-report")
            generate_progress_report "${2:-false}"
            ;;
        "start-periodic")
            start_periodic_reporting "${2:-1800}" "${3:-false}"
            ;;
        "stop-periodic")
            stop_periodic_reporting
            ;;
        "check-periodic")
            check_periodic_status
            ;;
        "startup")
            send_startup_notification "${2:-worker}"
            ;;
        "shutdown")
            send_shutdown_notification "${2:-worker}"
            ;;
        "summary")
            create_session_summary "$2"
            ;;
        "help"|*)
            echo "Claude Company Automated Reporting System"
            echo "Usage: $0 <command> [args...]"
            echo ""
            echo "Commands:"
            echo "  completion <file> <desc> [task_id]     Report task completion"
            echo "  progress <percent> <work> [task_id]    Report progress status"
            echo "  error <error> <help_req> [task_id]     Report error"
            echo "  status <old> <new> [task_id] [details] Report status change"
            echo "  auto-report [detailed]                 Generate progress report"
            echo "  start-periodic [interval] [detailed]   Start periodic reporting"
            echo "  stop-periodic                          Stop periodic reporting"
            echo "  check-periodic                         Check periodic status"
            echo "  startup [role]                         Send startup notification"
            echo "  shutdown [role]                        Send shutdown notification"
            echo "  summary [file]                         Create session summary"
            echo "  help                                   Show this help"
            ;;
    esac
}

# Execute if run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi