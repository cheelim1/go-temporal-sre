#!/usr/bin/env bash
#
# Script to stop the SuperScript application
#
# Purpose: Safely stop the SuperScript application
# Author: Enterprise Team
# Created: 2025-03-30
#

# Enable strict mode
set -euo pipefail
IFS=$'\n\t'

# Get script directory (safer approach for sourced scripts)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Define application configuration
readonly APP_PROCESS_PATTERN="bin/superscript"
readonly APP_PORT=8080
readonly STOP_TIMEOUT=10    # seconds to wait for graceful termination
readonly FORCE_KILL_DELAY=3 # seconds to wait before force kill

# Define color codes using tput (more portable than ANSI escape sequences)
if [[ -t 1 ]]; then  # Check if stdout is a terminal
    readonly BOLD=$(tput bold)
    readonly GREEN=$(tput setaf 2)
    readonly YELLOW=$(tput setaf 3)
    readonly RED=$(tput setaf 1)
    readonly RESET=$(tput sgr0)
else
    readonly BOLD=""
    readonly GREEN=""
    readonly YELLOW=""
    readonly RED=""
    readonly RESET=""
fi

# Print error message
print_error() {
    printf "%s%s%s\n" "${RED}${BOLD}" "ERROR: $1" "${RESET}" >&2
}

# Print warning message
print_warning() {
    printf "%s%s%s\n" "${YELLOW}" "WARNING: $1" "${RESET}" >&2
}

# Print success message
print_success() {
    printf "%s%s%s\n" "${GREEN}" "$1" "${RESET}"
}

# Function to print colored messages
print_message() {
    local color="$1"
    local message="$2"
    printf "%s%s%s\n" "${color}" "${message}" "${RESET}"
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to get the PID of the application
get_app_pid() {
    pgrep -f "${APP_PROCESS_PATTERN}" || echo ""
}

# Function to check if the application is running
is_app_running() {
    [[ -n "$(get_app_pid)" ]]
    return $?
}

# Function to check if the application port is still in use
is_port_in_use() {
    if command_exists nc; then
        nc -z localhost "${APP_PORT}" > /dev/null 2>&1
        return $?
    elif command_exists lsof; then
        lsof -i :"${APP_PORT}" > /dev/null 2>&1
        return $?
    else
        # Fallback - attempt connection (may not work in all environments)
        (</dev/tcp/localhost/${APP_PORT}) > /dev/null 2>&1
        return $?
    fi
}

# Function to stop the application with progressively stronger signals
stop_app() {
    local pid
    pid=$(get_app_pid)

    if [[ -z "${pid}" ]]; then
        return 0
    fi

    # Step 1: Try graceful termination with SIGTERM
    printf "Sending SIGTERM to process %s..." "${pid}"
    kill -15 "${pid}" 2>/dev/null || true

    # Wait for process to terminate
    local counter=0
    while is_app_running && ((counter < STOP_TIMEOUT)); do
        printf "."
        sleep 1
        ((counter++))
    done
    printf "\n"

    # Check if process is still running
    if is_app_running; then
        print_warning "Application did not terminate gracefully. Sending SIGKILL..."
        local pid_updated
        pid_updated=$(get_app_pid)

        if [[ -n "${pid_updated}" ]]; then
            kill -9 "${pid_updated}" 2>/dev/null || true
            sleep "${FORCE_KILL_DELAY}"
        fi
    fi

    # Final check to ensure process is gone
    if is_app_running; then
        return 1
    fi

    return 0
}

# Function to clean up any remaining resources
final_cleanup() {
    # Check if port is still in use despite process being killed
    if is_port_in_use; then
        print_warning "Port ${APP_PORT} is still in use despite stopping the application"
        print_warning "You may need to manually check for other processes using this port"
    fi

    # Here you could add additional cleanup steps if needed
    # - Remove temporary files
    # - Reset configurations
    # - Close related services
}

# Cleanup function that runs on script exit
cleanup() {
    local exit_code=$?
    printf "\n"
    if [[ ${exit_code} -ne 0 && ${exit_code} -ne 130 ]]; then
        print_error "Script exited with code: ${exit_code}"
    fi
    return ${exit_code}
}

# Set trap for cleanup on EXIT and common signals
trap cleanup EXIT HUP INT TERM

# Main script execution
main() {
    print_message "${YELLOW}${BOLD}" "=== Stopping SuperScript Application ==="

    # Check for required commands
    if ! command_exists pgrep; then
        print_error "Required command 'pgrep' not found"
        exit 1
    fi

    # Start timing for performance tracking
    local start_time=$(date +%s)

    # Check if application is running
    if is_app_running; then
        print_message "${YELLOW}" "SuperScript is running, attempting to stop..."

        # Attempt to stop the application
        if stop_app; then
            print_success "SuperScript has been stopped successfully"

            # Additional cleanup
            final_cleanup
        else
            print_error "Failed to stop SuperScript after multiple attempts"
            print_warning "You may need to investigate and manually kill the process"
            exit 1
        fi
    else
        print_message "${YELLOW}" "SuperScript is not running"
    fi

    # Calculate elapsed time
    local end_time=$(date +%s)
    local elapsed=$((end_time - start_time))

    print_success "Cleanup complete in ${elapsed} seconds!"
}

# Execute main function
main
