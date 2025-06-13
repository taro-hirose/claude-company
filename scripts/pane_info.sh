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

# Identify parent and sibling panes
identify_pane_relationships() {
    local current_pane="${CURRENT_PANE_ID:-$(tmux display-message -p '#{pane_id}')}"
    echo "👥 === Pane Relationships ==="
    
    # Get all panes and try to determine relationships
    local session_name="${CURRENT_SESSION:-$(tmux display-message -p '#{session_name}')}"
    local panes=($(tmux list-panes -s -t "$session_name" -F "#{pane_id}"))
    
    echo "🔗 Relationship Analysis:"
    echo "   Current Pane: $current_pane"
    
    # Identify likely parent (first pane in session)
    if [ ${#panes[@]} -gt 0 ]; then
        local likely_parent="${panes[0]}"
        echo "   Likely Parent: $likely_parent"
        
        # Identify siblings (other panes)
        echo "   Sibling Panes:"
        for pane in "${panes[@]}"; do
            if [ "$pane" != "$current_pane" ] && [ "$pane" != "$likely_parent" ]; then
                local pane_cmd=$(tmux display-message -t "$pane" -p "#{pane_current_command}" 2>/dev/null || echo "unknown")
                echo "     - $pane ($pane_cmd)"
            fi
        done
    fi
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

# Detect Claude instances
detect_claude_instances() {
    local session_name="${CURRENT_SESSION:-$(tmux display-message -p '#{session_name}')}"
    echo "🤖 === Claude Instance Detection ==="
    
    local claude_panes=()
    tmux list-panes -s -t "$session_name" -F "#{pane_id}|#{pane_current_command}" | \
    while IFS='|' read -r pane_id current_cmd; do
        if [[ "$current_cmd" =~ claude|python.*claude|node.*claude ]]; then
            echo "   ✅ Claude detected in $pane_id: $current_cmd"
            claude_panes+=("$pane_id")
        fi
    done
    
    if [ ${#claude_panes[@]} -eq 0 ]; then
        echo "   ⚠️ No Claude instances detected"
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
        "full")
            get_full_pane_context
            discover_session_panes
            identify_pane_relationships
            detect_claude_instances
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
            echo "  full              Run full analysis"
            echo "  help              Show this help"
            ;;
    esac
}

# Execute if run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi