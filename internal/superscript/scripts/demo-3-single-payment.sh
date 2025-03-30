#!/usr/bin/env bash
#
# Demo 3: Test single payment idempotent workflow using Temporal
#
# Purpose: Demonstrate idempotent script execution with Temporal
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
APP_ENDPOINT="single"
ORDER_ID="ORD-DEMO-123"

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

# Function to run the single payment workflow and display output
run_single_payment() {
    local run_number=$1
    local api_url="http://localhost:${APP_PORT}/run/${APP_ENDPOINT}"

    print_message "${YELLOW}" "\nRunning single payment workflow (Run ${run_number} of 3)..."

    # JSON API request
    local json_payload
    json_payload="{\"order_id\":\"${ORDER_ID}\"}"

    print_message "${CYAN}" "Request:"
    format_json "${json_payload}"

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
    sleep 1
    return 0
}

# Set trap for cleanup on exit and common signals
trap cleanup EXIT HUP INT TERM

# Main script execution
main() {
    print_message "${GREEN}${BOLD}" "=== Demo 3: Testing Single Payment Idempotent Workflow ==="
    print_message "${YELLOW}" "This demo shows how Temporal makes script execution IDEMPOTENT"
    printf "Running the workflow multiple times with the same OrderID will execute the script only ONCE!\n"
    print_message "${GREEN}" "Temporal uses WorkflowIDReusePolicy.REJECT_DUPLICATE to ensure idempotency"

    # Check if the SuperScript application is running
    if ! is_app_running; then
        print_error "SuperScript application is not running"
        printf "Please run 'make superscript-demo-1' first\n"
        exit 1
    fi

    printf "\nWe'll run the single payment workflow 3 times in succession with the same OrderID.\n"
    printf "This demonstrates how Temporal ensures each payment is processed exactly once.\n\n"

    # Run the workflow multiple times to show idempotent behavior
    local success_count=0
    for ((i=1; i<=3; i++)); do
        if run_single_payment "$i"; then
            ((success_count++))
        fi
    done

    print_message "${GREEN}" "\nObserve how only the first request executes the script"
    printf "The subsequent requests are handled idempotently, returning results from the first execution\n"
    printf "This prevents duplicate processing even with concurrent requests\n\n"

    # Display key points with bullet points for better readability
    print_message "${YELLOW}${BOLD}" "Key points about the implementation:"
    printf "• Using WorkflowIDReusePolicy.REJECT_DUPLICATE in Temporal\n"
    printf "• The workflow ID is derived from the OrderID to ensure uniqueness\n"
    printf "• Temporal guarantees exactly-once execution semantics\n"
    printf "• All concurrent calls get the same results without duplicate processing\n\n"

    print_message "${GREEN}" "To see how we can orchestrate multiple payments while maintaining idempotency,"
    print_message "${GREEN}" "run 'make superscript-demo-4' next"
}

# Execute main function
main
