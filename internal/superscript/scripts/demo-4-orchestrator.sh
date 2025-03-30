#!/usr/bin/env bash
#
# Demo 4: Test the orchestrator workflow that manages multiple payment collections
#
# Purpose: Demonstrate orchestration of multiple workflows with Temporal
# Author: Enterprise Team
# Created: 2025-03-30
#

# Enable strict mode
set -euo pipefail
IFS=$'\n\t'

# Get script directory (safer approach for sourced scripts)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# APP Configuration
APP_PORT=8080
APP_ENDPOINT="batch"
DEMO_RUNS=2

# Define color codes using tput (more portable than ANSI escape sequences)
if [[ -t 1 ]]; then  # Check if stdout is a terminal
    readonly BOLD=$(tput bold)
    readonly GREEN=$(tput setaf 2)
    readonly YELLOW=$(tput setaf 3)
    readonly RED=$(tput setaf 1)
    readonly CYAN=$(tput setaf 6)
    readonly RESET=$(tput sgr0)
else
    readonly BOLD=""
    readonly GREEN=""
    readonly YELLOW=""
    readonly RED=""
    readonly CYAN=""
    readonly RESET=""
fi

# Cleanup function that runs on script exit
cleanup() {
    local exit_code=$?
    # Add any cleanup tasks here
    printf "\n"
    if [[ ${exit_code} -ne 0 && ${exit_code} -ne 130 ]]; then
        print_error "Script exited with code: ${exit_code}"
    fi
    # Don't change the exit code
    return ${exit_code}
}

# Print error message
print_error() {
    printf "%s%s%s\n" "${RED}${BOLD}" "ERROR: $1" "${RESET}" >&2
}

# Print warning message
print_warning() {
    printf "%s%s%s\n" "${YELLOW}" "WARNING: $1" "${RESET}" >&2
}

# Function to print colored messages
print_message() {
    local color="$1"
    local message="$2"
    printf "%s%s%s\n" "${color}" "${message}" "${RESET}"
}

# Function to print a numbered list item
print_numbered_item() {
    local number="$1"
    local text="$2"
    printf "%d. %s\n" "${number}" "${text}"
}

# Function to check if the SuperScript application is running
is_app_running() {
    pgrep -f "bin/superscript" > /dev/null
    return $?
}

# Function to format JSON if possible, otherwise return as-is
format_json() {
    local json_str="$1"
    # Try to use jq if available, fall back to python, then show raw if neither works
    if command -v jq >/dev/null 2>&1; then
        jq . <<< "${json_str}" 2>/dev/null || echo "${json_str}"
    else
        python3 -m json.tool <<< "${json_str}" 2>/dev/null || echo "${json_str}"
    fi
}

# Function to run the orchestrator workflow and display output
run_orchestrator() {
    local run_number=$1
    local api_url="http://localhost:${APP_PORT}/run/${APP_ENDPOINT}"

    print_message "${YELLOW}" "\nRunning orchestrator workflow (Run ${run_number} of ${DEMO_RUNS})..."

    # JSON API request
    local json_payload="{}"

    # Use curl with timeout and better error handling
    local response
    if ! response=$(curl -s --connect-timeout 5 --max-time 10 \
                   -X POST "${api_url}" \
                   -H "Content-Type: application/json" \
                   -d "${json_payload}"); then
        print_error "Failed to connect to the API at ${api_url}"
        return 1
    fi

    print_message "${CYAN}" "\nResponse:"
    # Format JSON response
    format_json "${response}"
    printf "\n"
    sleep 2
    return 0
}

# Function to print key advantages with bullet points
print_advantages() {
    print_message "${YELLOW}${BOLD}" "Key advantages of the Temporal orchestration:"
    print_numbered_item 1 "Automatic retry for failed activities"
    print_numbered_item 2 "Clear visibility into workflow execution status"
    print_numbered_item 3 "Exactly-once execution guarantees for each payment"
    print_numbered_item 4 "Graceful handling of concurrency through Temporal's REJECT_DUPLICATE policy"
    print_numbered_item 5 "Scalable architecture for processing large batches"
    printf "\n"
}

# Set trap for cleanup on exit and common signals
trap cleanup EXIT HUP INT TERM

# Main script execution
main() {
    print_message "${GREEN}${BOLD}" "=== Demo 4: Testing Orchestrator Workflow ==="
    print_message "${YELLOW}" "This demo shows how Temporal orchestrates multiple child workflows while maintaining idempotency"
    printf "The orchestrator replaces the traditional batch script with a proper workflow\n"
    print_message "${GREEN}" "Each child workflow still maintains its own idempotency guarantees"

    # Check if the SuperScript application is running
    if ! is_app_running; then
        print_error "SuperScript application is not running"
        printf "Please run 'make superscript-demo-1' first\n"
        exit 1
    fi

    printf "\nWe'll run the orchestrator workflow ${DEMO_RUNS} times.\n"
    printf "This demonstrates how Temporal manages multiple child workflows with idempotency.\n\n"

    # Run the workflow multiple times
    local success_count=0
    for ((i=1; i<=DEMO_RUNS; i++)); do
        if run_orchestrator "$i"; then
            ((success_count++))
        fi
    done

    print_message "${GREEN}" "\nThe orchestrator creates child workflows for each OrderID"
    printf "Each child workflow has its own WorkflowID based on the OrderID\n"
    printf "This ensures each payment is processed exactly once, even across multiple orchestrator runs\n\n"

    # Print key advantages using the dedicated function
    print_advantages

    print_message "${GREEN}${BOLD}" "This completes the demo of how to make non-idempotent scripts idempotent with Temporal!"
    printf "To stop the SuperScript application, run 'make superscript-stop'\n"
}

# Execute main function
main
