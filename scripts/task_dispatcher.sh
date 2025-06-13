#!/bin/bash

# Subtask Automatic Dispatching System for Claude Company
# Automatically distributes subtasks to appropriate child panes

set -euo pipefail

# Configuration
DISPATCH_TIMEOUT="${DISPATCH_TIMEOUT:-10}"
MAX_RETRIES="${MAX_RETRIES:-3}"
RETRY_DELAY="${RETRY_DELAY:-2}"

# Source required scripts
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/pane_info.sh"

# Classify task type based on content
classify_task_type() {
    local task="$1"
    local task_lower=$(echo "$task" | tr '[:upper:]' '[:lower:]')
    
    # Documentation tasks (check first for specific patterns)
    if [[ "$task_lower" =~ readme.*書|書.*readme|ドキュメント|文書|説明.*書|書.*説明|doc|documentation|explain|document|manual ]]; then
        echo "documentation"
        return 0
    fi
    
    # Review/Analysis tasks
    if [[ "$task_lower" =~ レビュー|分析|チェック|検査|調査|review|analyze|investigate|examine|inspect ]]; then
        echo "review"
        return 0
    fi
    
    # Testing tasks
    if [[ "$task_lower" =~ テスト|試験|検証|確認|動作|test|verify|check|validation|run ]]; then
        echo "testing"
        return 0
    fi
    
    # Database tasks
    if [[ "$task_lower" =~ データベース|db|sql|クエリ|データ|database|query|data|migration|schema ]]; then
        echo "database"
        return 0
    fi
    
    # Build/Deployment tasks
    if [[ "$task_lower" =~ ビルド|デプロイ|リリース|パッケージ|build|deploy|release|package|compile ]]; then
        echo "deployment"
        return 0
    fi
    
    # Configuration/Setup tasks
    if [[ "$task_lower" =~ 設定|構成|セットアップ|インストール|config|setup|install|configure ]]; then
        echo "configuration"
        return 0
    fi
    
    # Implementation tasks (check last as it's most general)
    if [[ "$task_lower" =~ 作成|実装|書いて|プログラム|コード|関数|クラス|機能|新しい|追加|build|create|implement|write|code|function|class|feature|new|add ]]; then
        echo "implementation"
        return 0
    fi
    
    # Default to general
    echo "general"
}

# Find available worker panes (excluding parent and Claude-running panes)
find_available_worker_panes() {
    local session_name="${CURRENT_SESSION:-$(tmux display-message -p '#{session_name}')}"
    local current_pane="${CURRENT_PANE_ID:-$(tmux display-message -p '#{pane_id}')}"
    local available_panes=()
    
    echo "🔍 Searching for available worker panes..." >&2
    
    # Get all panes in session
    while IFS='|' read -r pane_id pane_index current_cmd title active; do
        # Skip current pane (likely the parent)
        if [ "$pane_id" = "$current_pane" ]; then
            continue
        fi
        
        # Skip panes running Claude (to avoid interference)
        if [[ "$current_cmd" =~ claude|python.*claude|node.*claude ]]; then
            echo "⚠️ Skipping Claude pane: $pane_id ($current_cmd)" >&2
            continue
        fi
        
        # Check if pane is responsive
        if tmux send-keys -t "$pane_id" "" 2>/dev/null; then
            available_panes+=("$pane_id:$pane_index:$current_cmd")
            echo "✅ Available pane found: $pane_id (index: $pane_index, cmd: $current_cmd)" >&2
        else
            echo "❌ Pane not responsive: $pane_id" >&2
        fi
    done < <(tmux list-panes -s -t "$session_name" -F "#{pane_id}|#{pane_index}|#{pane_current_command}|#{pane_title}|#{pane_active}")
    
    if [ ${#available_panes[@]} -eq 0 ]; then
        echo "❌ No available worker panes found" >&2
        return 1
    fi
    
    # Return available panes (one per line)
    printf '%s\n' "${available_panes[@]}"
}

# Select best pane for task type
select_best_pane() {
    local task_type="$1"
    local available_panes=("${@:2}")
    
    # Preference mapping for task types
    case "$task_type" in
        "implementation")
            # Prefer shell/editor panes for implementation
            for pane_info in "${available_panes[@]}"; do
                local cmd=$(echo "$pane_info" | cut -d: -f3)
                if [[ "$cmd" =~ (bash|zsh|sh|vim|nvim|code|editor) ]]; then
                    echo "$pane_info"
                    return 0
                fi
            done
            ;;
        "testing")
            # Prefer shell panes for testing
            for pane_info in "${available_panes[@]}"; do
                local cmd=$(echo "$pane_info" | cut -d: -f3)
                if [[ "$cmd" =~ (bash|zsh|sh) ]]; then
                    echo "$pane_info"
                    return 0
                fi
            done
            ;;
        "database")
            # Prefer DB client panes
            for pane_info in "${available_panes[@]}"; do
                local cmd=$(echo "$pane_info" | cut -d: -f3)
                if [[ "$cmd" =~ (psql|mysql|sqlite|mongo) ]]; then
                    echo "$pane_info"
                    return 0
                fi
            done
            ;;
    esac
    
    # Return first available pane if no preference match
    echo "${available_panes[0]}"
}

# Dispatch task to selected pane
dispatch_to_pane() {
    local task="$1"
    local target_pane="$2"
    local retry_count=0
    
    echo "📤 Dispatching task to pane $target_pane..." >&2
    
    while [ $retry_count -lt $MAX_RETRIES ]; do
        # Prepare the task command
        local dispatch_cmd="echo '📋 New task: $task'"
        
        # Try to send the task
        if tmux send-keys -t "$target_pane" "$dispatch_cmd" Enter 2>/dev/null; then
            echo "✅ Task dispatched successfully to $target_pane" >&2
            
            # Optional: Send the actual task as a comment or command
            sleep 1
            if tmux send-keys -t "$target_pane" "# Task: $task" Enter 2>/dev/null; then
                echo "✅ Task details sent to $target_pane" >&2
            fi
            
            return 0
        else
            ((retry_count++))
            if [ $retry_count -lt $MAX_RETRIES ]; then
                echo "⚠️ Dispatch failed, retrying... ($retry_count/$MAX_RETRIES)" >&2
                sleep "$RETRY_DELAY"
            fi
        fi
    done
    
    echo "❌ Failed to dispatch task after $MAX_RETRIES attempts" >&2
    return 1
}

# Enhanced dispatch with Claude integration
dispatch_with_claude() {
    local task="$1"
    local target_pane="$2"
    
    echo "🤖 Dispatching Claude-compatible task to $target_pane..." >&2
    
    # Create a more sophisticated command for Claude
    local claude_prompt="Please help with the following task: $task"
    
    # Check if target pane might be running Claude
    local pane_cmd=$(tmux display-message -t "$target_pane" -p "#{pane_current_command}" 2>/dev/null || echo "unknown")
    
    if [[ "$pane_cmd" =~ claude ]]; then
        # Direct Claude interaction
        if tmux send-keys -t "$target_pane" "$claude_prompt" Enter 2>/dev/null; then
            echo "✅ Claude task dispatched successfully" >&2
            return 0
        fi
    else
        # Regular shell command
        local shell_cmd="echo '🤖 Claude Task: $task' && echo 'Run: claude code or appropriate Claude command'"
        if tmux send-keys -t "$target_pane" "$shell_cmd" Enter 2>/dev/null; then
            echo "✅ Task dispatched to shell pane" >&2
            return 0
        fi
    fi
    
    echo "❌ Failed to dispatch Claude task" >&2
    return 1
}

# Log dispatch activity
log_dispatch() {
    local task="$1"
    local task_type="$2"
    local target_pane="$3"
    local status="$4"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    local log_file="/tmp/task_dispatch.log"
    echo "[$timestamp] $status: Task '$task' (type: $task_type) -> Pane $target_pane" >> "$log_file"
}

# Main dispatch function
dispatch_task() {
    local task="$1"
    local force_pane="${2:-}"
    
    echo "🚀 === Task Dispatcher Starting ===" >&2
    echo "Task: $task" >&2
    echo "" >&2
    
    # Initialize pane context
    get_full_pane_context >/dev/null
    
    # Classify the task
    local task_type=$(classify_task_type "$task")
    echo "🏷️ Task classified as: $task_type" >&2
    
    # Find available panes
    local available_panes_output
    if ! available_panes_output=$(find_available_worker_panes); then
        echo "❌ No available worker panes found" >&2
        log_dispatch "$task" "$task_type" "none" "FAILED_NO_PANES"
        return 1
    fi
    
    # Convert output to array (compatible with older bash)
    local available_panes=()
    while IFS= read -r line; do
        available_panes+=("$line")
    done <<< "$available_panes_output"
    echo "📊 Found ${#available_panes[@]} available panes" >&2
    
    # Select target pane
    local target_pane_info
    if [ -n "$force_pane" ]; then
        # Use forced pane if specified
        for pane_info in "${available_panes[@]}"; do
            local pane_id=$(echo "$pane_info" | cut -d: -f1)
            if [ "$pane_id" = "$force_pane" ]; then
                target_pane_info="$pane_info"
                break
            fi
        done
        if [ -z "$target_pane_info" ]; then
            echo "❌ Forced pane $force_pane not available" >&2
            return 1
        fi
    else
        target_pane_info=$(select_best_pane "$task_type" "${available_panes[@]}")
    fi
    
    local target_pane=$(echo "$target_pane_info" | cut -d: -f1)
    local target_index=$(echo "$target_pane_info" | cut -d: -f2)
    local target_cmd=$(echo "$target_pane_info" | cut -d: -f3)
    
    echo "🎯 Selected target: Pane $target_pane (index: $target_index, cmd: $target_cmd)" >&2
    
    # Dispatch the task
    if dispatch_to_pane "$task" "$target_pane"; then
        echo "✅ Task dispatched successfully!" >&2
        log_dispatch "$task" "$task_type" "$target_pane" "SUCCESS"
        
        # Show success summary
        echo "" >&2
        echo "📋 === Dispatch Summary ===" >&2
        echo "Task: $task" >&2
        echo "Type: $task_type" >&2
        echo "Target: Pane $target_pane (index $target_index)" >&2
        echo "Status: ✅ SUCCESS" >&2
        
        return 0
    else
        echo "❌ Failed to dispatch task" >&2
        log_dispatch "$task" "$task_type" "$target_pane" "FAILED_DISPATCH"
        return 1
    fi
}

# Show dispatch statistics
show_dispatch_stats() {
    local log_file="/tmp/task_dispatch.log"
    
    if [ ! -f "$log_file" ]; then
        echo "📊 No dispatch history found"
        return 0
    fi
    
    echo "📊 === Dispatch Statistics ==="
    echo "Total dispatches: $(wc -l < "$log_file")"
    echo "Successful: $(grep -c "SUCCESS" "$log_file" || echo 0)"
    echo "Failed: $(grep -c "FAILED" "$log_file" || echo 0)"
    echo ""
    echo "Recent dispatches:"
    tail -5 "$log_file" 2>/dev/null || echo "No recent dispatches"
}

# Interactive mode for testing
interactive_mode() {
    echo "🎮 === Interactive Task Dispatcher ==="
    echo "Enter tasks to dispatch (type 'quit' to exit):"
    echo ""
    
    while true; do
        read -p "Task> " task
        
        case "$task" in
            "quit"|"exit"|"q")
                echo "👋 Goodbye!"
                break
                ;;
            "stats"|"status")
                show_dispatch_stats
                ;;
            "panes")
                find_available_worker_panes
                ;;
            "")
                continue
                ;;
            *)
                echo ""
                dispatch_task "$task"
                echo ""
                ;;
        esac
    done
}

# Usage examples and help
show_usage() {
    cat << 'EOF'
Claude Company Task Dispatcher

USAGE:
    ./task_dispatcher.sh <task>                    # Dispatch task automatically
    ./task_dispatcher.sh <task> <pane_id>          # Dispatch to specific pane
    ./task_dispatcher.sh --interactive             # Interactive mode
    ./task_dispatcher.sh --stats                   # Show statistics
    ./task_dispatcher.sh --test                    # Run test cases

EXAMPLES:
    ./task_dispatcher.sh "ファイルXを作成してください"
    ./task_dispatcher.sh "Run the test suite" %2
    ./task_dispatcher.sh "データベースをセットアップしてください"
    
TASK TYPES DETECTED:
    - implementation: 作成, 実装, プログラム, create, implement, code
    - testing: テスト, 試験, test, verify, check
    - database: データベース, db, sql, query, data
    - documentation: ドキュメント, doc, readme, explain
    - deployment: ビルド, デプロイ, build, deploy, release
    - review: レビュー, 分析, review, analyze, investigate
    - configuration: 設定, セットアップ, config, setup, install
    - general: (default for unclassified tasks)

OPTIONS:
    --interactive, -i    Interactive mode
    --stats, -s         Show dispatch statistics
    --test, -t          Run test cases
    --help, -h          Show this help
EOF
}

# Test cases
run_tests() {
    echo "🧪 === Running Test Cases ==="
    echo ""
    
    # Test task classification
    echo "📝 Testing task classification:"
    local test_tasks=(
        "ファイルを作成してください:implementation"
        "テストを実行してください:testing"
        "データベースをセットアップ:database"
        "READMEを書いてください:documentation"
        "アプリをデプロイしてください:deployment"
        "コードをレビューしてください:review"
        "設定ファイルを更新:configuration"
        "その他のタスク:general"
    )
    
    for test_case in "${test_tasks[@]}"; do
        local task=$(echo "$test_case" | cut -d: -f1)
        local expected=$(echo "$test_case" | cut -d: -f2)
        local result=$(classify_task_type "$task")
        
        if [ "$result" = "$expected" ]; then
            echo "✅ '$task' → $result"
        else
            echo "❌ '$task' → $result (expected: $expected)"
        fi
    done
    
    echo ""
    echo "🔍 Testing pane discovery:"
    if find_available_worker_panes >/dev/null 2>&1; then
        echo "✅ Pane discovery working"
    else
        echo "❌ Pane discovery failed"
    fi
    
    echo ""
    echo "✅ Test cases completed"
}

# Main function
main() {
    local command="${1:-help}"
    
    case $command in
        "--interactive"|"-i")
            interactive_mode
            ;;
        "--stats"|"-s")
            show_dispatch_stats
            ;;
        "--test"|"-t")
            run_tests
            ;;
        "--help"|"-h"|"help")
            show_usage
            ;;
        *)
            if [ $# -eq 0 ]; then
                show_usage
            else
                dispatch_task "$@"
            fi
            ;;
    esac
}

# Execute if run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi