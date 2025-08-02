#!/bin/bash

# Claude Code Stop Hook
# This script captures Claude Code Stop hook data and writes it to .sbs/stop.json
# It reads Stop hook JSON data from stdin and stores it in the project directory

set -euo pipefail

# Function to log errors to stderr
log_error() {
    echo "[claude-code-stop-hook] ERROR: $*" >&2
}

# Function to log info to stderr
log_info() {
    echo "[claude-code-stop-hook] INFO: $*" >&2
}

# Function to detect if running in sandbox
is_sandbox_environment() {
    # Check for sandbox-specific indicators
    if [[ -d "/work" && -f "/work/scripts/claude-code-stop-hook.sh" ]]; then
        return 0  # true - in sandbox
    fi
    return 1  # false - not in sandbox
}

# Main function
main() {
    local project_dir
    local sbs_dir
    local stop_file
    local hook_data
    local timestamp

    # Determine project directory based on environment
    if is_sandbox_environment; then
        project_dir="/work"
        log_info "Running in sandbox environment"
    else
        project_dir="$(pwd)"
        log_info "Running in host environment"
    fi
    
    sbs_dir="${project_dir}/.sbs"
    stop_file="${sbs_dir}/stop.json"

    log_info "Processing Claude Code hook data..."
    log_info "Project directory: ${project_dir}"
    log_info "SBS directory: ${sbs_dir}"

    # Create .sbs directory if it doesn't exist
    if ! mkdir -p "${sbs_dir}"; then
        log_error "Failed to create .sbs directory: ${sbs_dir}"
        exit 1
    fi

    # Read JSON data from stdin
    if ! hook_data=$(cat); then
        log_error "Failed to read hook data from stdin"
        exit 1
    fi

    # Validate that we received some data
    if [[ -z "${hook_data}" ]]; then
        log_error "No hook data received from stdin"
        exit 1
    fi

    # Try to extract session_id to validate this is Stop hook data
    local session_id=""
    if command -v jq >/dev/null 2>&1; then
        session_id=$(echo "${hook_data}" | jq -r '.session_id // empty' 2>/dev/null || echo "")
        if [[ -n "${session_id}" ]]; then
            log_info "Processing Stop hook data for session: ${session_id}"
        else
            log_info "Processing hook data (session_id not found, may not be Stop hook format)"
        fi
    else
        log_info "Processing hook data (jq not available for validation)"
    fi

    # Add timestamp to the hook data
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    # Determine environment context for metadata
    local environment_type
    if is_sandbox_environment; then
        environment_type="sandbox"
    else
        environment_type="host"
    fi

    # Create enhanced JSON with timestamp and metadata
    local enhanced_data
    enhanced_data=$(cat <<EOF
{
  "claude_code_hook": {
    "hook_type": "Stop",
    "timestamp": "${timestamp}",
    "environment": "${environment_type}",
    "project_directory": "${project_dir}",
    "hook_script": "$0",
    "sandbox_detection": $(is_sandbox_environment && echo "true" || echo "false")
  },
  "stop_hook_data": ${hook_data}
}
EOF
)

    # Write the enhanced data to stop.json
    if echo "${enhanced_data}" > "${stop_file}"; then
        log_info "Hook data written to: ${stop_file}"
        
        # Pretty print the JSON for better readability (if jq is available)
        if command -v jq >/dev/null 2>&1; then
            if temp_file=$(mktemp); then
                if jq '.' "${stop_file}" > "${temp_file}" 2>/dev/null; then
                    mv "${temp_file}" "${stop_file}"
                    log_info "JSON formatted with jq"
                else
                    rm -f "${temp_file}"
                    log_info "JSON written (jq formatting failed, using raw format)"
                fi
            fi
        else
            log_info "JSON written (jq not available for formatting)"
        fi
    else
        log_error "Failed to write hook data to: ${stop_file}"
        exit 1
    fi

    log_info "Claude Code hook processing completed successfully"
    return 0
}

# Error handling
trap 'log_error "Script interrupted or failed"' ERR INT TERM

# Run main function
main "$@"