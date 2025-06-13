#!/bin/bash

# Enhanced Pane Initialization Script for Claude Company
# Sets up AI worker pane with full autonomous capabilities

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMPLATE_DIR="$(dirname "$SCRIPT_DIR")/templates"
ENHANCED_TEMPLATE="$TEMPLATE_DIR/enhanced_prompt_template.md"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    local missing_tools=()
    local required_tools=("tmux" "curl" "bash")
    
    for tool in "${required_tools[@]}"; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            missing_tools+=("$tool")
        fi
    done
    
    if [ ${#missing_tools[@]} -gt 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        return 1
    fi
    
    # Check if we're in a tmux session
    if [ -z "${TMUX:-}" ]; then
        log_warning "Not in a tmux session - some features may be limited"
    fi
    
    log_success "Prerequisites check passed"
}

# Initialize scripts environment
init_scripts() {
    log_info "Initializing script environment..."
    
    # Make all scripts executable
    chmod +x "$SCRIPT_DIR"/*.sh
    
    # Source error handler first for logging
    if [ -f "$SCRIPT_DIR/error_handler.sh" ]; then
        source "$SCRIPT_DIR/error_handler.sh"
        log_success "Error handler loaded"
    else
        log_error "Error handler script not found"
        return 1
    fi
    
    # Load other essential scripts
    local scripts=("pane_info.sh" "task_api.sh" "auto_report.sh")
    for script in "${scripts[@]}"; do
        if [ -f "$SCRIPT_DIR/$script" ]; then
            source "$SCRIPT_DIR/$script"
            log_success "Loaded $script"
        else
            log_warning "$script not found - some features may be unavailable"
        fi
    done
}

# Collect pane context
collect_context() {
    log_info "=== COLLECTING PANE CONTEXT ==="
    
    # Get basic context
    get_full_pane_context 2>/dev/null || {
        log_warning "Could not collect full pane context"
        export CURRENT_PANE_ID="${TMUX_PANE:-unknown}"
        export CURRENT_SESSION="${TMUX_SESSION:-unknown}"
    }
    
    # Discover session structure
    log_info "Discovering session structure..."
    discover_session_panes 2>/dev/null || log_warning "Could not discover session panes"
    
    # Identify relationships
    log_info "Identifying pane relationships..."
    identify_pane_relationships 2>/dev/null || log_warning "Could not identify pane relationships"
    
    log_success "Context collection completed"
}

# Test connectivity
test_connectivity() {
    log_info "=== TESTING CONNECTIVITY ==="
    
    # Run health check
    if run_health_check >/dev/null 2>&1; then
        log_success "System health check passed"
    else
        log_warning "Some services unavailable - enabling fallback mode"
        run_recovery_operations >/dev/null 2>&1 || log_warning "Recovery operations failed"
    fi
    
    # Test database connection
    if connect_db_with_fallback >/dev/null 2>&1; then
        log_success "Database connectivity verified"
    else
        log_warning "Database unavailable - using fallback mode"
    fi
    
    # Test API connection
    if connect_api_with_fallback "/health" >/dev/null 2>&1; then
        log_success "API connectivity verified"
    else
        log_warning "API unavailable - using fallback mode"
    fi
}

# Load task context
load_task_context() {
    log_info "=== LOADING TASK CONTEXT ==="
    
    echo "📋 My Current Tasks:"
    get_my_tasks 2>/dev/null | head -10 || log_warning "Could not retrieve my tasks"
    echo ""
    
    echo "🔗 Shared Tasks:"
    get_shared_tasks 2>/dev/null | head -5 || log_warning "Could not retrieve shared tasks"
    echo ""
    
    echo "📊 Progress Summary:"
    get_progress_stats 2>/dev/null || log_warning "Could not retrieve progress statistics"
    echo ""
    
    log_success "Task context loaded"
}

# Send startup notification
send_startup() {
    log_info "Sending startup notification..."
    
    if send_startup_notification "AI-Worker" >/dev/null 2>&1; then
        log_success "Startup notification sent to parent pane"
    else
        log_warning "Could not send startup notification"
    fi
}

# Generate dynamic prompt
generate_dynamic_prompt() {
    log_info "Generating dynamic prompt..."
    
    local prompt_file="/tmp/enhanced_prompt_${CURRENT_PANE_ID}_$(date +%s).md"
    
    # Read template and substitute variables
    if [ -f "$ENHANCED_TEMPLATE" ]; then
        sed "s/#{CURRENT_PANE_ID}/${CURRENT_PANE_ID}/g; s/#{CURRENT_SESSION}/${CURRENT_SESSION}/g" "$ENHANCED_TEMPLATE" > "$prompt_file"
        
        # Add current context to prompt
        cat >> "$prompt_file" <<EOF

---

## Current Session Context (Auto-Generated)

**Pane ID**: ${CURRENT_PANE_ID}
**Session**: ${CURRENT_SESSION}
**Timestamp**: $(date)
**Script Path**: ${SCRIPT_DIR}

### Available Functions:
EOF
        
        # Add function documentation
        declare -F | grep -E "(get_|report_|send_)" | sed 's/declare -f /- /' >> "$prompt_file" || true
        
        log_success "Dynamic prompt generated: $prompt_file"
        echo "💡 Enhanced prompt available at: $prompt_file"
    else
        log_error "Enhanced template not found at: $ENHANCED_TEMPLATE"
        return 1
    fi
}

# Setup periodic reporting
setup_periodic_reporting() {
    local enable_periodic="${1:-false}"
    local interval="${2:-1800}"  # 30 minutes default
    
    if [ "$enable_periodic" = "true" ]; then
        log_info "Setting up periodic reporting (every $interval seconds)..."
        
        # Start in background
        nohup bash -c "
            source '$SCRIPT_DIR/auto_report.sh'
            start_periodic_reporting '$interval' 'false'
        " >/dev/null 2>&1 &
        
        log_success "Periodic reporting started in background"
    else
        log_info "Periodic reporting not enabled (use --periodic to enable)"
    fi
}

# Create initialization summary
create_summary() {
    local summary_file="/tmp/init_summary_${CURRENT_PANE_ID}_$(date +%Y%m%d_%H%M%S).txt"
    
    cat > "$summary_file" <<EOF
# Claude Company Enhanced Pane Initialization Summary

**Initialization Time**: $(date)
**Pane ID**: ${CURRENT_PANE_ID}
**Session**: ${CURRENT_SESSION}
**Script Directory**: ${SCRIPT_DIR}

## Loaded Capabilities:
- ✅ Database connection with fallback
- ✅ API client with retry logic
- ✅ Automated progress reporting
- ✅ Inter-pane communication
- ✅ Error handling and recovery
- ✅ Task context awareness

## Available Commands:
- get_my_tasks - Retrieve current tasks
- get_shared_tasks - Get shared tasks
- report_progress - Send progress updates
- report_completion - Report task completion
- report_error - Report issues
- get_sibling_tasks - Check sibling status
- run_health_check - System diagnostics

## Quick Start:
1. Check current tasks: get_my_tasks
2. Start work and report: report_progress "0" "Starting analysis"
3. Update progress: report_progress "50" "Implementation in progress"
4. Complete task: report_completion "file.go" "Feature completed"

## Monitoring:
- Health check: run_health_check
- Error logs: tail -f /tmp/claude_company_errors_$(date +%Y%m%d).log
- Periodic status: check_periodic_status

---
Generated by Claude Company Enhanced Initialization System
EOF
    
    log_success "Initialization summary created: $summary_file"
    echo "📋 Full summary available at: $summary_file"
}

# Main initialization function
main() {
    local enable_periodic="false"
    local periodic_interval="1800"
    local show_help="false"
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --periodic)
                enable_periodic="true"
                shift
                ;;
            --interval)
                periodic_interval="$2"
                shift 2
                ;;
            --help|-h)
                show_help="true"
                shift
                ;;
            *)
                log_warning "Unknown option: $1"
                shift
                ;;
        esac
    done
    
    if [ "$show_help" = "true" ]; then
        cat <<EOF
Claude Company Enhanced Pane Initialization

Usage: $0 [options]

Options:
  --periodic              Enable periodic reporting
  --interval SECONDS      Set periodic reporting interval (default: 1800)
  --help, -h              Show this help message

Description:
  Initializes an AI worker pane with enhanced capabilities including:
  - Automated context collection
  - Database and API connectivity with fallbacks
  - Inter-pane communication setup
  - Task context loading
  - Progress reporting automation

Examples:
  $0                                    # Basic initialization
  $0 --periodic                        # With periodic reporting (30min interval)
  $0 --periodic --interval 900         # With 15-minute reporting interval

EOF
        return 0
    fi
    
    echo "🚀 === CLAUDE COMPANY ENHANCED PANE INITIALIZATION ==="
    echo "Time: $(date)"
    echo "Pane: ${TMUX_PANE:-unknown}"
    echo ""
    
    # Run initialization sequence
    check_prerequisites || return 1
    init_scripts || return 1
    collect_context
    test_connectivity
    load_task_context
    send_startup
    generate_dynamic_prompt
    setup_periodic_reporting "$enable_periodic" "$periodic_interval"
    create_summary
    
    echo ""
    log_success "=== INITIALIZATION COMPLETE ==="
    echo "🤖 Enhanced AI Worker Pane Ready"
    echo "📚 Use 'run_health_check' to verify system status"
    echo "📋 Use 'get_my_tasks' to see current assignments"
    echo "📊 Use 'report_progress' to communicate with parent"
    echo ""
}

# Execute if run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi