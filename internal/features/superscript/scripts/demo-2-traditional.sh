#!/usr/bin/env bash
#
# Demo 2: Run traditional non-idempotent script multiple times
#
# Purpose: Demonstrate the issues with non-idempotent scripts
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
APP_ENDPOINT="traditional"

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

# Function to run the traditional script and display output
run_traditional() {
    local run_number=$1
    local api_url="http://localhost:${APP_PORT}/run/${APP_ENDPOINT}"

    print_message "${YELLOW}" "\nRunning traditional script (Run ${run_number} of 3)..."

    # Use curl with timeout and better error handling
    local response
    if ! response=$(curl -s --connect-timeout 5 --max-time 10 "${api_url}"); then
        print_error "Failed to connect to the API at ${api_url}"
        return 1
    fi

    # Show response if not empty
    if [[ -n "${response}" ]]; then
        echo "${response}"
    fi

    print_message "${CYAN}" "\nScript is running in the background. Check application logs for output."
    sleep 2
}

# Set trap for cleanup on exit and common signals
trap cleanup EXIT HUP INT TERM

# Main script execution
main() {
    print_message "${GREEN}${BOLD}" "=== Demo 2: Testing Traditional Non-Idempotent Script ==="
    print_message "${YELLOW}" "This demo shows the NON-IDEMPOTENT behavior of the traditional script"
    printf "Running the script multiple times will process payments multiple times!\n"
    print_message "${RED}" "In a real-world application, this could cause double-charging customers"

    # Check if the SuperScript application is running <-- not needed
#    if ! is_app_running; then
#        print_error "SuperScript application is not running"
#        printf "Please run 'make superscript-demo-1' first\n"
#        exit 1
#    fi

    printf "\nWe'll run the traditional script 3 times in quick succession.\n"
    printf "This demonstrates how multiple requests can lead to duplicate processing.\n\n"

    set +e
    # Run the script multiple times to show non-idempotent behavior
    local success_count=0
    for ((i=1; i<3; i++)); do
        if run_traditional "$i"; then
            ((success_count++))
        fi
    done
    set -e


    print_message "${RED}" "\nNote how each script execution processes payments independently"
    printf "This can lead to race conditions and duplicate payments\n\n"

    print_message "${GREEN}" "To contrast this with the idempotent behavior of Temporal workflows,"
    print_message "${GREEN}" "run 'make superscript-demo-3' next"
}

# Execute main function
main
