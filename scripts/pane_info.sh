#!/bin/bash

# Inter-pane Information Collection Commands for Claude Company
# Comprehensive pane structure analysis and status monitoring

set -euo pipefail

# Get comprehensive pane context
get_full_pane_context() {
    echo "🔍 === Pane Context Analysis ==="
    
    # Basic tmux context
    local session_name=$(tmux display-message -p "#{session_name}" 2>/dev/null || echo "unknown")
    local pane_id=$(tmux display-message -p "#{pane_id}" 2>/dev/null || echo "unknown")
    local pane_index=$(tmux display-message -p "#{pane_index}" 2>/dev/null || echo "unknown")
    local window_name=$(tmux display-message -p "#{window_name}" 2>/dev/null || echo "unknown")
    local pane_title=$(tmux display-message -p "#{pane_title}" 2>/dev/null || echo "unknown")
    local pane_current_command=$(tmux display-message -p "#{pane_current_command}" 2>/dev/null || echo "unknown")
    
    echo "📍 Current Context:"
    echo "   Session: $session_name"
    echo "   Pane ID: $pane_id"
    echo "   Pane Index: $pane_index"
    echo "   Window: $window_name"
    echo "   Title: $pane_title"
    echo "   Command: $pane_current_command"
    echo ""
    
    # Export for other scripts
    export CURRENT_SESSION="$session_name"
    export CURRENT_PANE_ID="$pane_id"
    export CURRENT_PANE_INDEX="$pane_index"
    export CURRENT_WINDOW="$window_name"
}

# Discover all panes in session
discover_session_panes() {
    local session_name="${CURRENT_SESSION:-$(tmux display-message -p '#{session_name}')}"
    echo "🌐 === Session Pane Discovery ==="
    echo "Session: $session_name"
    echo ""
    
    # Get all panes with detailed info
    echo "📊 All Panes in Session:"
    tmux list-panes -s -t "$session_name" -F "#{pane_id}|#{pane_index}|#{pane_current_command}|#{pane_title}|#{pane_active}" | \
    while IFS='|' read -r pane_id pane_index current_cmd title active; do
        local status_icon="⚪"
        [ "$active" = "1" ] && status_icon="🔵"
        echo "   $status_icon Pane $pane_index ($pane_id): $current_cmd"
        echo "      Title: $title"
    done
    echo ""
}

# Identify parent and sibling panes with title-based detection
identify_pane_relationships() {
    local current_pane="${CURRENT_PANE_ID:-$(tmux display-message -p '#{pane_id}')}"
    echo "👥 === Pane Relationships (Title-Based Detection) ==="
    
    # Get all panes and try to determine relationships
    local session_name="${CURRENT_SESSION:-$(tmux display-message -p '#{session_name}')}"
    local panes=($(tmux list-panes -s -t "$session_name" -F "#{pane_id}"))
    
    echo "🔗 Relationship Analysis:"
    echo "   Current Pane: $current_pane"
    
    # Categorize panes by title tags
    local manager_panes=()
    local worker_panes=()
    local other_panes=()
    
    echo "   Pane Classification:"
    for pane in "${panes[@]}"; do
        local pane_title=$(tmux display-message -t "$pane" -p "#{pane_title}" 2>/dev/null || echo "unknown")
        local pane_cmd=$(tmux display-message -t "$pane" -p "#{pane_current_command}" 2>/dev/null || echo "unknown")
        
        if [[ "$pane_title" == *"[MANAGER]"* ]]; then
            manager_panes+=("$pane")
            echo "     👔 Manager: $pane (title: $pane_title, cmd: $pane_cmd)"
        elif [[ "$pane_title" == *"[WORKER]"* ]]; then
            worker_panes+=("$pane")
            echo "     🔧 Worker: $pane (title: $pane_title, cmd: $pane_cmd)"
        else
            other_panes+=("$pane")
            echo "     ❓ Other: $pane (title: $pane_title, cmd: $pane_cmd)"
        fi
    done
    
    echo ""
    echo "   Summary:"
    echo "     Manager Panes: ${#manager_panes[@]}"
    echo "     Worker Panes: ${#worker_panes[@]}"
    echo "     Other Panes: ${#other_panes[@]}"
    echo ""
}

# Monitor pane activity
monitor_pane_activity() {
    local session_name="${CURRENT_SESSION:-$(tmux display-message -p '#{session_name}')}"
    echo "⚡ === Pane Activity Monitor ==="
    
    # Get activity status for all panes
    tmux list-panes -s -t "$session_name" -F "#{pane_id}|#{pane_index}|#{pane_current_command}|#{pane_activity}|#{pane_last_activity}" | \
    while IFS='|' read -r pane_id pane_index current_cmd activity last_activity; do
        local activity_icon="💤"
        [ "$activity" = "1" ] && activity_icon="⚡"
        
        echo "   $activity_icon Pane $pane_index ($pane_id): $current_cmd"
        echo "      Last Activity: $(date -d "@$last_activity" 2>/dev/null || echo "unknown")"
    done
    echo ""
}

# Capture pane content for analysis
capture_pane_content() {
    local target_pane="${1:-$(tmux display-message -p '#{pane_id}')}"
    local lines="${2:-20}"
    
    echo "📸 === Pane Content Capture ==="
    echo "Pane: $target_pane (last $lines lines)"
    echo "----------------------------------------"
    
    tmux capture-pane -t "$target_pane" -p -S "-$lines" 2>/dev/null || echo "❌ Failed to capture pane content"
    echo "----------------------------------------"
    echo ""
}

# Detect Claude instances with role identification
detect_claude_instances() {
    local session_name="${CURRENT_SESSION:-$(tmux display-message -p '#{session_name}')}"
    echo "🤖 === Claude Instance Detection (with Role Identification) ==="
    
    local claude_panes=()
    local manager_claude=0
    local worker_claude=0
    
    tmux list-panes -s -t "$session_name" -F "#{pane_id}|#{pane_current_command}|#{pane_title}" | \
    while IFS='|' read -r pane_id current_cmd pane_title; do
        if [[ "$current_cmd" =~ claude|python.*claude|node.*claude ]]; then
            claude_panes+=("$pane_id")
            
            # Determine role based on title
            local role="Unknown"
            if [[ "$pane_title" == *"[MANAGER]"* ]]; then
                role="Manager"
                manager_claude=$((manager_claude + 1))
            elif [[ "$pane_title" == *"[WORKER]"* ]]; then
                role="Worker"
                worker_claude=$((worker_claude + 1))
            else
                role="Generic"
            fi
            
            echo "   ✅ Claude detected in $pane_id: $current_cmd"
            echo "      Role: $role (title: $pane_title)"
        fi
    done
    
    if [ ${#claude_panes[@]} -eq 0 ]; then
        echo "   ⚠️ No Claude instances detected"
    else
        echo "   📊 Summary:"
        echo "     Total Claude instances: ${#claude_panes[@]}"
        echo "     Manager Claude instances: $manager_claude"
        echo "     Worker Claude instances: $worker_claude"
    fi
    echo ""
}

# Check for running processes in panes
check_pane_processes() {
    local session_name="${CURRENT_SESSION:-$(tmux display-message -p '#{session_name}')}"
    echo "⚙️ === Pane Process Analysis ==="
    
    tmux list-panes -s -t "$session_name" -F "#{pane_id}|#{pane_pid}|#{pane_current_command}" | \
    while IFS='|' read -r pane_id pane_pid current_cmd; do
        echo "   📋 Pane $pane_id (PID: $pane_pid): $current_cmd"
        
        # Try to get more detailed process info
        if command -v ps >/dev/null 2>&1; then
            local child_processes=$(ps --ppid "$pane_pid" -o pid,cmd --no-headers 2>/dev/null | head -3)
            if [ -n "$child_processes" ]; then
                echo "      Child processes:"
                echo "$child_processes" | sed 's/^/        /'
            fi
        fi
    done
    echo ""
}

# Send test message to pane
send_test_message() {
    local target_pane="$1"
    local message="${2:-echo 'Test message from pane info script'}"
    
    echo "📤 Sending test message to pane $target_pane"
    if tmux send-keys -t "$target_pane" "$message" Enter 2>/dev/null; then
        echo "✅ Message sent successfully"
    else
        echo "❌ Failed to send message"
    fi
}

# Generate comprehensive pane report
generate_pane_report() {
    local output_file="${1:-/tmp/pane_report_$(date +%Y%m%d_%H%M%S).txt}"
    
    echo "📋 === Generating Comprehensive Pane Report ==="
    echo "Output: $output_file"
    echo ""
    
    {
        echo "# Claude Company Pane Analysis Report"
        echo "Generated: $(date)"
        echo "Session: ${CURRENT_SESSION:-$(tmux display-message -p '#{session_name}')}"
        echo ""
        
        get_full_pane_context
        discover_session_panes
        identify_pane_relationships
        monitor_pane_activity
        detect_claude_instances
        check_pane_processes
        
    } | tee "$output_file"
    
    echo "✅ Report saved to: $output_file"
}

# Watch pane changes in real-time
watch_pane_changes() {
    local interval="${1:-5}"
    echo "👀 === Watching Pane Changes (every ${interval}s) ==="
    echo "Press Ctrl+C to stop"
    echo ""
    
    while true; do
        clear
        echo "🕐 $(date)"
        discover_session_panes
        monitor_pane_activity
        sleep "$interval"
    done
}

# Identify pane roles specifically by title tags
identify_pane_roles() {
    local session_name="${CURRENT_SESSION:-$(tmux display-message -p '#{session_name}')}"
    echo "🏷️ === Pane Role Identification by Title Tags ==="
    echo "Session: $session_name"
    echo ""
    
    # Get all panes with title information
    tmux list-panes -s -t "$session_name" -F "#{pane_id}|#{pane_index}|#{pane_title}|#{pane_current_command}" | \
    while IFS='|' read -r pane_id pane_index pane_title current_cmd; do
        local role_icon="❓"
        local role_name="Unidentified"
        
        if [[ "$pane_title" == *"[MANAGER]"* ]]; then
            role_icon="👔"
            role_name="Manager"
        elif [[ "$pane_title" == *"[WORKER]"* ]]; then
            role_icon="🔧"
            role_name="Worker"
        elif [[ "$pane_title" == *"コンソールペイン"* ]]; then
            role_icon="💻"
            role_name="Console"
        elif [[ "$pane_title" == *"マネージャーペイン"* ]]; then
            role_icon="📋"
            role_name="Manager"
        fi
        
        echo "   $role_icon Pane $pane_index ($pane_id): $role_name"
        echo "      Title: $pane_title"
        echo "      Command: $current_cmd"
        echo ""
    done
}

# Identify console and manager panes specifically
identify_console_manager_panes() {
    local session_name="${CURRENT_SESSION:-$(tmux display-message -p '#{session_name}')}"
    echo "🖥️ === Console/Manager Pane Identification ==="
    echo "Session: $session_name"
    echo ""
    
    local console_pane=""
    local manager_pane=""
    local other_count=0
    
    # Get all panes with title information
    while IFS='|' read -r pane_id pane_index pane_title current_cmd; do
        if [[ "$pane_title" == *"コンソールペイン"* ]]; then
            console_pane="$pane_id"
            echo "   💻 Console Pane: $pane_index ($pane_id)"
            echo "      Title: $pane_title"
            echo "      Command: $current_cmd"
        elif [[ "$pane_title" == *"マネージャーペイン"* ]]; then
            manager_pane="$pane_id"
            echo "   📋 Manager Pane: $pane_index ($pane_id)"
            echo "      Title: $pane_title"
            echo "      Command: $current_cmd"
        else
            ((other_count++))
            echo "   ❓ Other Pane: $pane_index ($pane_id)"
            echo "      Title: $pane_title"
            echo "      Command: $current_cmd"
        fi
        echo ""
    done < <(tmux list-panes -s -t "$session_name" -F "#{pane_id}|#{pane_index}|#{pane_title}|#{pane_current_command}")
    
    # Export for other scripts to use
    export CONSOLE_PANE_ID="$console_pane"
    export MANAGER_PANE_ID="$manager_pane"
    
    echo "Summary:"
    [ -n "$console_pane" ] && echo "   Console Pane ID: $console_pane" || echo "   Console Pane: Not found"
    [ -n "$manager_pane" ] && echo "   Manager Pane ID: $manager_pane" || echo "   Manager Pane: Not found"
    echo "   Other Panes: $other_count"
    echo ""
}

# Get console pane ID
get_console_pane_id() {
    local session_name="${CURRENT_SESSION:-$(tmux display-message -p '#{session_name}')}"
    tmux list-panes -s -t "$session_name" -F "#{pane_id}|#{pane_title}" | \
    while IFS='|' read -r pane_id pane_title; do
        if [[ "$pane_title" == *"コンソールペイン"* ]]; then
            echo "$pane_id"
            return
        fi
    done
}

# Get manager pane ID
get_manager_pane_id() {
    local session_name="${CURRENT_SESSION:-$(tmux display-message -p '#{session_name}')}"
    tmux list-panes -s -t "$session_name" -F "#{pane_id}|#{pane_title}" | \
    while IFS='|' read -r pane_id pane_title; do
        if [[ "$pane_title" == *"マネージャーペイン"* ]]; then
            echo "$pane_id"
            return
        fi
    done
}

# Check if current pane is console pane
is_console_pane() {
    local current_pane="${TMUX_PANE:-$(tmux display-message -p '#{pane_id}')}"
    local console_pane=$(get_console_pane_id)
    
    if [ "$current_pane" = "$console_pane" ]; then
        return 0
    else
        return 1
    fi
}

# Check if current pane is manager pane  
is_manager_pane() {
    local current_pane="${TMUX_PANE:-$(tmux display-message -p '#{pane_id}')}"
    local manager_pane=$(get_manager_pane_id)
    
    if [ "$current_pane" = "$manager_pane" ]; then
        return 0
    else
        return 1
    fi
}

# Get pane network information
get_pane_network_info() {
    echo "🌐 === Pane Network Information ==="
    
    # Check if API server is running
    if curl -s --max-time 2 "http://localhost:8080/api/health" >/dev/null 2>&1; then
        echo "✅ API Server: Running (localhost:8080)"
    else
        echo "❌ API Server: Not responding"
    fi
    
    # Check database connectivity
    if command -v pg_isready >/dev/null 2>&1; then
        if pg_isready -h localhost -p 5432 -d claude_company >/dev/null 2>&1; then
            echo "✅ Database: Accessible (localhost:5432)"
        else
            echo "❌ Database: Not accessible"
        fi
    else
        echo "⚠️ Database: pg_isready not available"
    fi
    echo ""
}

# Main interactive function
main() {
    local command="${1:-help}"
    
    case $command in
        "context")
            get_full_pane_context
            ;;
        "discover")
            discover_session_panes
            ;;
        "relationships")
            identify_pane_relationships
            ;;
        "activity")
            monitor_pane_activity
            ;;
        "capture")
            capture_pane_content "${2:-$(tmux display-message -p '#{pane_id}')}" "${3:-20}"
            ;;
        "claude")
            detect_claude_instances
            ;;
        "processes")
            check_pane_processes
            ;;
        "test-send")
            [ $# -ge 2 ] || { echo "Usage: $0 test-send <pane_id> [message]"; exit 1; }
            send_test_message "$2" "${3:-echo 'Test message'}"
            ;;
        "report")
            generate_pane_report "$2"
            ;;
        "watch")
            watch_pane_changes "${2:-5}"
            ;;
        "network")
            get_pane_network_info
            ;;
        "identify-roles")
            identify_pane_roles
            ;;
        "identify-console-manager")
            identify_console_manager_panes
            ;;
        "get-console-pane")
            get_console_pane_id
            ;;
        "get-manager-pane")
            get_manager_pane_id
            ;;
        "is-console")
            if is_console_pane; then
                echo "Current pane is console pane"
                exit 0
            else
                echo "Current pane is not console pane"
                exit 1
            fi
            ;;
        "is-manager")
            if is_manager_pane; then
                echo "Current pane is manager pane"
                exit 0
            else
                echo "Current pane is not manager pane"
                exit 1
            fi
            ;;
        "full")
            get_full_pane_context
            discover_session_panes
            identify_pane_relationships
            detect_claude_instances
            identify_console_manager_panes
            get_pane_network_info
            ;;
        "help"|*)
            echo "Claude Company Pane Information Tool"
            echo "Usage: $0 <command> [args...]"
            echo ""
            echo "Commands:"
            echo "  context           Get current pane context"
            echo "  discover          Discover all panes in session"
            echo "  relationships     Identify parent/sibling relationships"
            echo "  activity          Monitor pane activity"
            echo "  capture [pane] [lines]  Capture pane content"
            echo "  claude            Detect Claude instances"
            echo "  processes         Check running processes"
            echo "  test-send <pane> [msg]  Send test message to pane"
            echo "  report [file]     Generate comprehensive report"
            echo "  watch [interval]  Watch pane changes in real-time"
            echo "  network           Check network connectivity"
            echo "  identify-roles    Identify pane roles by title tags"
            echo "  identify-console-manager  Identify console/manager panes"
            echo "  get-console-pane  Get console pane ID"
            echo "  get-manager-pane  Get manager pane ID"
            echo "  is-console        Check if current pane is console"
            echo "  is-manager        Check if current pane is manager"
            echo "  full              Run full analysis"
            echo "  help              Show this help"
            ;;
    esac
}

# Execute if run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi