#!/bin/bash

# Pane Difference Detection System for Claude Company
# Detects newly created panes by comparing before/after states
# Specifically designed to distinguish between parent and child panes

set -euo pipefail

# Global variables for state storage
PANES_BEFORE_FILE="${TMPDIR:-/tmp}/before_panes.txt"
PANES_AFTER_FILE="${TMPDIR:-/tmp}/after_panes.txt"
PARENT_PANES_FILE="${TMPDIR:-/tmp}/parent_panes.txt"
CHILD_PANES_FILE="${TMPDIR:-/tmp}/child_panes.txt"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Cleanup function
cleanup() {
    # Only clean up process-specific temp files on exit
    # Keep the main before/after files for debugging
    rm -f "/tmp/panes_before_$$.tmp" "/tmp/panes_after_$$.tmp"
}
trap cleanup EXIT

# Get all panes in the current session and store in specified file
get_panes_state() {
    local output_file="$1"
    local session_name="${2:-$(tmux display-message -p '#{session_name}' 2>/dev/null || echo '')}"
    
    if [ -z "$session_name" ]; then
        echo -e "${RED}Error: Not in a tmux session${NC}" >&2
        return 1
    fi
    
    # Get panes with detailed info: ID|index|command|title|active|pid
    tmux list-panes -s -t "$session_name" -F "#{pane_id}|#{pane_index}|#{pane_current_command}|#{pane_title}|#{pane_active}|#{pane_pid}" 2>/dev/null > "$output_file" || {
        echo -e "${RED}Error: Failed to get pane list for session '$session_name'${NC}" >&2
        return 1
    }
    
    return 0
}

# Record panes before operation
get_panes_before() {
    local session_name="${1:-$(tmux display-message -p '#{session_name}' 2>/dev/null || echo '')}"
    
    echo -e "${BLUE}📸 Recording panes before operation...${NC}"
    
    if get_panes_state "$PANES_BEFORE_FILE" "$session_name"; then
        local pane_count=$(wc -l < "$PANES_BEFORE_FILE")
        echo -e "${GREEN}✅ Recorded $pane_count panes in session '$session_name'${NC}"
        
        # Show current panes for reference
        echo -e "${YELLOW}Current panes:${NC}"
        while IFS='|' read -r pane_id pane_index current_cmd title active pid; do
            local status_icon="⚪"
            [ "$active" = "1" ] && status_icon="🔵"
            echo "   $status_icon Pane $pane_index ($pane_id): $current_cmd"
        done < "$PANES_BEFORE_FILE"
        
        return 0
    else
        return 1
    fi
}

# Record panes after operation
get_panes_after() {
    local session_name="${1:-$(tmux display-message -p '#{session_name}' 2>/dev/null || echo '')}"
    
    echo -e "${BLUE}📸 Recording panes after operation...${NC}"
    
    if get_panes_state "$PANES_AFTER_FILE" "$session_name"; then
        local pane_count=$(wc -l < "$PANES_AFTER_FILE")
        echo -e "${GREEN}✅ Recorded $pane_count panes in session '$session_name'${NC}"
        return 0
    else
        return 1
    fi
}

# Validate pane exists and is accessible
validate_pane() {
    local pane_id="$1"
    
    # Check if pane exists
    if ! tmux list-panes -F "#{pane_id}" 2>/dev/null | grep -q "^$pane_id$"; then
        return 1
    fi
    
    # Check if pane is responsive (try to get its PID)
    local pane_pid=$(tmux display-message -t "$pane_id" -p "#{pane_pid}" 2>/dev/null || echo "")
    if [ -z "$pane_pid" ] || [ "$pane_pid" = "0" ]; then
        return 1
    fi
    
    # Check if the process is still running
    if ! kill -0 "$pane_pid" 2>/dev/null; then
        return 1
    fi
    
    return 0
}

# Detect newly created panes and classify them
detect_new_pane() {
    echo -e "${BLUE}🔍 Detecting pane differences...${NC}" >&2
    echo "" >&2
    
    # Check if both files exist
    if [ ! -f "$PANES_BEFORE_FILE" ]; then
        echo -e "${RED}Error: Before-state file not found at: $PANES_BEFORE_FILE${NC}" >&2
        echo -e "${YELLOW}Run '$0 before' first to record initial state.${NC}" >&2
        return 1
    fi
    
    if [ ! -f "$PANES_AFTER_FILE" ]; then
        echo -e "${RED}Error: After-state file not found at: $PANES_AFTER_FILE${NC}" >&2
        echo -e "${YELLOW}Run '$0 after' to record post-creation state.${NC}" >&2
        return 1
    fi
    
    # Clear previous classification files
    > "$PARENT_PANES_FILE"
    > "$CHILD_PANES_FILE"
    
    # Get pane IDs from before and after
    local panes_before=($(cut -d'|' -f1 "$PANES_BEFORE_FILE"))
    local panes_after=($(cut -d'|' -f1 "$PANES_AFTER_FILE"))
    
    echo -e "${YELLOW}=== Pane Classification ===${NC}" >&2
    echo "" >&2
    
    # Classify parent panes (existing before operation)
    echo -e "${BLUE}📁 Parent Panes (existing before operation):${NC}" >&2
    local parent_count=0
    for pane_before in "${panes_before[@]}"; do
        # Get detailed info about the parent pane
        local pane_info=$(grep "^$pane_before|" "$PANES_BEFORE_FILE" || echo "$pane_before|unknown|unknown|unknown|0|unknown")
        IFS='|' read -r pane_id pane_index current_cmd title active pid <<< "$pane_info"
        
        echo -e "   🔵 $pane_id (index: $pane_index, cmd: $current_cmd)" >&2
        echo "$pane_id" >> "$PARENT_PANES_FILE"
        ((parent_count++))
    done
    
    # Special identification of primary parent panes (%0 and %1)
    if [ ${#panes_before[@]} -ge 1 ]; then
        local primary_parent="${panes_before[0]}"
        echo -e "   ${YELLOW}→ Primary parent: $primary_parent${NC}" >&2
    fi
    
    echo "" >&2
    
    # Find and classify child panes (created during operation)
    echo -e "${GREEN}📄 Child Panes (newly created):${NC}" >&2
    local new_panes=()
    local child_count=0
    
    for pane_after in "${panes_after[@]}"; do
        local found=false
        for pane_before in "${panes_before[@]}"; do
            if [ "$pane_after" = "$pane_before" ]; then
                found=true
                break
            fi
        done
        
        if [ "$found" = false ]; then
            # Validate the new pane before adding to results
            if validate_pane "$pane_after"; then
                new_panes+=("$pane_after")
                
                # Get detailed info about the new pane
                local pane_info=$(grep "^$pane_after|" "$PANES_AFTER_FILE" || echo "$pane_after|unknown|unknown|unknown|0|unknown")
                IFS='|' read -r pane_id pane_index current_cmd title active pid <<< "$pane_info"
                
                echo -e "   🆕 $pane_id (index: $pane_index, cmd: $current_cmd, pid: $pid)" >&2
                echo "$pane_id" >> "$CHILD_PANES_FILE"
                ((child_count++))
            else
                echo -e "${YELLOW}   ⚠️ Skipping invalid/inaccessible pane: $pane_after${NC}" >&2
            fi
        fi
    done
    
    if [ $child_count -eq 0 ]; then
        echo -e "   ${YELLOW}(No new panes detected)${NC}" >&2
    fi
    
    echo "" >&2
    echo -e "${YELLOW}=== Summary ===${NC}" >&2
    echo -e "Parent panes: $parent_count" >&2
    echo -e "Child panes:  $child_count" >&2
    echo -e "Total panes:  ${#panes_after[@]}" >&2
    
    # Save last detected child pane for easy access
    if [ ${#new_panes[@]} -gt 0 ]; then
        echo "${new_panes[0]}" > "${TMPDIR:-/tmp}/last_child_pane.txt"
        echo "" >&2
        echo -e "${GREEN}✅ Child pane IDs saved to:${NC}" >&2
        echo -e "   - $CHILD_PANES_FILE" >&2
        echo -e "   - Last child: ${new_panes[0]}" >&2
    fi
    
    # Report results
    if [ ${#new_panes[@]} -eq 0 ]; then
        # No new panes detected, return empty string
        echo ""
        return 1
    else
        # Return only the last new pane ID to stdout
        echo "${new_panes[0]}"
        return 0
    fi
}

# Complete workflow: before -> operation -> after -> detect
detect_pane_diff() {
    local operation_command="$1"
    local session_name="${2:-$(tmux display-message -p '#{session_name}' 2>/dev/null || echo '')}"
    
    echo -e "${BLUE}🔄 Starting pane difference detection workflow${NC}"
    echo -e "${BLUE}Session: $session_name${NC}"
    echo -e "${BLUE}Operation: $operation_command${NC}"
    echo ""
    
    # Step 1: Record before state
    if ! get_panes_before "$session_name"; then
        echo -e "${RED}❌ Failed to record before state${NC}"
        return 1
    fi
    echo ""
    
    # Step 2: Execute the operation
    echo -e "${BLUE}⚙️ Executing operation: $operation_command${NC}"
    if eval "$operation_command"; then
        echo -e "${GREEN}✅ Operation completed successfully${NC}"
    else
        echo -e "${YELLOW}⚠️ Operation completed with non-zero exit code${NC}"
    fi
    echo ""
    
    # Step 3: Small delay to ensure pane is fully created
    sleep 0.5
    
    # Step 4: Record after state
    if ! get_panes_after "$session_name"; then
        echo -e "${RED}❌ Failed to record after state${NC}"
        return 1
    fi
    echo ""
    
    # Step 5: Detect differences
    detect_new_pane
}

# Show usage examples and help
show_help() {
    cat << 'EOF'
Pane Difference Detection System - Usage Guide

DESCRIPTION:
    Detects newly created tmux panes by comparing before/after states.
    Distinguishes between parent panes (existing) and child panes (newly created).
    Specifically designed to identify panes like %0, %1 as parents and new panes as children.

USAGE:
    pane_diff_detector.sh <command> [arguments]

COMMANDS:
    before [session]          Record current panes state (before operation)
    after [session]           Record current panes state (after operation)  
    detect                    Compare states and classify parent/child panes
    workflow <command>        Complete workflow: before -> run command -> after -> detect
    validate <pane_id>        Check if a pane is valid and accessible
    clean                     Clean up all state files
    help                      Show this help message

EXAMPLES:
    # Manual workflow
    ./pane_diff_detector.sh before
    tmux split-window -h 'sleep 10'  # Your operation here
    ./pane_diff_detector.sh after
    ./pane_diff_detector.sh detect

    # Automatic workflow
    ./pane_diff_detector.sh workflow "tmux split-window -h 'sleep 10'"
    
    # Validate a specific pane
    ./pane_diff_detector.sh validate "%1"
    
    # Clean up state files
    ./pane_diff_detector.sh clean

FEATURES:
    ✅ Parent/Child pane classification
    ✅ Identifies primary parent panes (%0, %1)
    ✅ Automatic session detection
    ✅ Pane validation and error handling
    ✅ Invalid/dead pane filtering
    ✅ Detailed pane information display
    ✅ Color-coded output
    ✅ Comprehensive error reporting

OUTPUT:
    Returns new pane IDs on success (one per line)
    Exit code 0: New panes detected
    Exit code 1: No new panes or error

FILES:
    ${TMPDIR}/before_panes.txt     Pane state before operation
    ${TMPDIR}/after_panes.txt      Pane state after operation
    ${TMPDIR}/parent_panes.txt     List of parent pane IDs
    ${TMPDIR}/child_panes.txt      List of child pane IDs
    ${TMPDIR}/last_child_pane.txt  ID of the last created child pane

EOF
}

# Run test cases
run_tests() {
    echo -e "${BLUE}🧪 Running Pane Diff Detector Tests${NC}"
    echo ""
    
    local test_session="pane_diff_test_$$"
    
    # Test 1: Create test session
    echo -e "${YELLOW}Test 1: Creating test session${NC}"
    if tmux new-session -d -s "$test_session" 'sleep 60'; then
        echo -e "${GREEN}✅ Test session created: $test_session${NC}"
    else
        echo -e "${RED}❌ Failed to create test session${NC}"
        return 1
    fi
    
    # Test 2: Record before state
    echo -e "${YELLOW}Test 2: Recording before state${NC}"
    if get_panes_before "$test_session"; then
        echo -e "${GREEN}✅ Before state recorded successfully${NC}"
    else
        echo -e "${RED}❌ Failed to record before state${NC}"
        tmux kill-session -t "$test_session" 2>/dev/null
        return 1
    fi
    
    # Test 3: Create new pane
    echo -e "${YELLOW}Test 3: Creating new pane${NC}"
    if tmux split-window -t "$test_session" -h 'sleep 30'; then
        echo -e "${GREEN}✅ New pane created${NC}"
    else
        echo -e "${RED}❌ Failed to create new pane${NC}"
        tmux kill-session -t "$test_session" 2>/dev/null
        return 1
    fi
    
    # Test 4: Record after state
    echo -e "${YELLOW}Test 4: Recording after state${NC}"
    if get_panes_after "$test_session"; then
        echo -e "${GREEN}✅ After state recorded successfully${NC}"
    else
        echo -e "${RED}❌ Failed to record after state${NC}"
        tmux kill-session -t "$test_session" 2>/dev/null
        return 1
    fi
    
    # Test 5: Detect new panes
    echo -e "${YELLOW}Test 5: Detecting new panes${NC}"
    local new_panes_output
    if new_panes_output=$(detect_new_pane); then
        echo -e "${GREEN}✅ New pane detection successful${NC}"
        echo "Detected panes: $new_panes_output"
    else
        echo -e "${RED}❌ Failed to detect new panes${NC}"
        tmux kill-session -t "$test_session" 2>/dev/null
        return 1
    fi
    
    # Cleanup
    tmux kill-session -t "$test_session" 2>/dev/null
    echo ""
    echo -e "${GREEN}🎉 All tests passed successfully!${NC}"
}

# Clean up all state files
clean_state_files() {
    echo -e "${YELLOW}🧹 Cleaning up state files...${NC}"
    
    local files=(
        "$PANES_BEFORE_FILE"
        "$PANES_AFTER_FILE"
        "$PARENT_PANES_FILE"
        "$CHILD_PANES_FILE"
        "${TMPDIR:-/tmp}/last_child_pane.txt"
    )
    
    local cleaned=0
    for file in "${files[@]}"; do
        if [ -f "$file" ]; then
            rm -f "$file"
            echo -e "   ${GREEN}✓ Removed: $file${NC}"
            ((cleaned++))
        fi
    done
    
    if [ $cleaned -eq 0 ]; then
        echo -e "   ${YELLOW}No state files found to clean${NC}"
    else
        echo -e "${GREEN}✅ Cleaned $cleaned file(s)${NC}"
    fi
}

# Main function
main() {
    local command="${1:-help}"
    
    case $command in
        "before")
            get_panes_before "$2"
            ;;
        "after")
            get_panes_after "$2"
            ;;
        "detect")
            detect_new_pane
            ;;
        "workflow")
            [ $# -ge 2 ] || { 
                echo -e "${RED}Error: workflow command requires an operation${NC}" >&2
                echo "Usage: $0 workflow '<command>'" >&2
                exit 1
            }
            detect_pane_diff "$2" "$3"
            ;;
        "validate")
            [ $# -ge 2 ] || { 
                echo -e "${RED}Error: validate command requires a pane ID${NC}" >&2
                echo "Usage: $0 validate <pane_id>" >&2
                exit 1
            }
            if validate_pane "$2"; then
                echo -e "${GREEN}✅ Pane $2 is valid and accessible${NC}"
                exit 0
            else
                echo -e "${RED}❌ Pane $2 is invalid or inaccessible${NC}"
                exit 1
            fi
            ;;
        "clean")
            clean_state_files
            ;;
        "test")
            run_tests
            ;;
        "help"|*)
            show_help
            ;;
    esac
}

# Execute if run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi