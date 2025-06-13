#!/bin/bash

# DB Connection Automation Script for Claude Company
# Auto-detects connection parameters and establishes PostgreSQL connection

set -euo pipefail

# Default connection parameters (matching internal/database/connection.go)
DEFAULT_HOST="localhost"
DEFAULT_PORT="5432"
DEFAULT_USER="claude_user"
DEFAULT_PASSWORD="claude_password"
DEFAULT_DBNAME="claude_company"
DEFAULT_SSLMODE="disable"

# Function to get environment variable or default
get_env_or_default() {
    local var_name="$1"
    local default_value="$2"
    echo "${!var_name:-$default_value}"
}

# Auto-detect DB connection parameters
detect_db_config() {
    export DB_HOST=$(get_env_or_default "DB_HOST" "$DEFAULT_HOST")
    export DB_PORT=$(get_env_or_default "DB_PORT" "$DEFAULT_PORT")
    export DB_USER=$(get_env_or_default "DB_USER" "$DEFAULT_USER")
    export DB_PASSWORD=$(get_env_or_default "DB_PASSWORD" "$DEFAULT_PASSWORD")
    export DB_NAME=$(get_env_or_default "DB_NAME" "$DEFAULT_DBNAME")
    export DB_SSLMODE=$(get_env_or_default "DB_SSLMODE" "$DEFAULT_SSLMODE")
    
    export DATABASE_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"
}

# Test database connection
test_db_connection() {
    local max_retries=3
    local retry_delay=2
    
    for ((i=1; i<=max_retries; i++)); do
        echo "🔍 Testing database connection (attempt $i/$max_retries)..."
        
        if psql "$DATABASE_URL" -c "SELECT 1;" &>/dev/null; then
            echo "✅ Database connection successful"
            return 0
        else
            echo "❌ Connection failed (attempt $i/$max_retries)"
            if [ $i -lt $max_retries ]; then
                echo "⏳ Retrying in $retry_delay seconds..."
                sleep $retry_delay
            fi
        fi
    done
    
    echo "🚨 Database connection failed after $max_retries attempts"
    return 1
}

# Execute SQL query with error handling
execute_sql() {
    local query="$1"
    local output_format="${2:-table}"
    
    case $output_format in
        "json")
            psql "$DATABASE_URL" -t -A -F',' -c "$query" | \
            awk 'BEGIN{print "["} NR>1{print ","} {gsub(/"/,"\\\""); print "\""$0"\""}END{print "]"}'
            ;;
        "csv")
            psql "$DATABASE_URL" -A -F',' -c "$query"
            ;;
        "table"|*)
            psql "$DATABASE_URL" -c "$query"
            ;;
    esac
}

# Get current pane context
get_pane_context() {
    local session_name=$(tmux display-message -p "#{session_name}" 2>/dev/null || echo "unknown")
    local pane_id=$(tmux display-message -p "#{pane_id}" 2>/dev/null || echo "unknown")
    local pane_index=$(tmux display-message -p "#{pane_index}" 2>/dev/null || echo "unknown")
    
    echo "📍 Pane Context: ${session_name}:${pane_id} (index: ${pane_index})"
    export CURRENT_PANE_ID="$pane_id"
    export CURRENT_SESSION="$session_name"
    export CURRENT_PANE_INDEX="$pane_index"
}

# Main execution function
main() {
    echo "🚀 Claude Company DB Connection Automation"
    echo "=========================================="
    
    # Step 1: Detect pane context
    get_pane_context
    
    # Step 2: Detect DB configuration
    echo "🔧 Detecting database configuration..."
    detect_db_config
    echo "   Host: $DB_HOST:$DB_PORT"
    echo "   Database: $DB_NAME"
    echo "   User: $DB_USER"
    
    # Step 3: Test connection
    if test_db_connection; then
        echo "✅ Ready for database operations"
        return 0
    else
        echo "❌ Database connection setup failed"
        return 1
    fi
}

# Script execution
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi