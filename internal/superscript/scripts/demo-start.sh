#!/usr/bin/env bash
#
# Script to start the SuperScript application
#
# Purpose: Start the SuperScript application for demos
# Author: Enterprise Team
# Created: 2025-03-30
#

# Enable strict mode
set -euo pipefail
IFS=$'\n\t'

# Get script directory (safer approach for sourced scripts)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Define application paths and configuration
readonly APP_BIN="/Users/leow/GOMOD/go-temporal-sre/bin/superscript"
readonly APP_LOG="/Users/leow/GOMOD/go-temporal-sre/superscript.log"
readonly APP_PORT=8080
readonly TEMPORAL_PORT=7233
readonly MAX_RETRIES=5
readonly INIT_TIMEOUT=10 # seconds

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

# Function to display a simple spinner during waits
display_spinner() {
    local pid=$1
    local message="$2"
    local delay=0.1
    local spinstr='|/-\\'
    printf "%s " "$message"

    while ps -p "$pid" > /dev/null; do
        local temp=${spinstr#?}
        printf "\r%s %c" "$message" "${spinstr:0:1}"
        spinstr="$temp${spinstr:0:1}"
        sleep "$delay"
    done
    printf "\r%s Done.\n" "$message"
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if the application is running
is_app_running() {
    pgrep -f "bin/superscript" > /dev/null
    return $?
}

# Function to check if Temporal server is running
is_temporal_running() {
    if command_exists nc; then
        nc -z localhost "${TEMPORAL_PORT}" > /dev/null 2>&1
        return $?
    elif command_exists curl; then
        curl -s "http://localhost:${TEMPORAL_PORT}" -m 1 > /dev/null 2>&1
        return $?
    else
        # Fallback method if neither nc nor curl is available
        (</dev/tcp/localhost/${TEMPORAL_PORT}) > /dev/null 2>&1
        return $?
    fi
}

# Function to verify application is responding to HTTP requests
verify_app_responding() {
    # Try to connect to the application's HTTP endpoint
    if command_exists curl; then
        curl -s "http://localhost:${APP_PORT}/health" -m 2 > /dev/null 2>&1
        return $?
    else
        # Fallback method if curl is not available
        (</dev/tcp/localhost/${APP_PORT}) > /dev/null 2>&1
        return $?
    fi
}

# Function to ask for user confirmation
ask_for_confirmation() {
    local prompt="$1"
    local default="${2:-n}" # Default is 'n' unless specified

    local yn
    while true; do
        printf "%s %s: " "${prompt}" "($(if [[ "${default}" = "y" ]]; then echo "Y/n"; else echo "y/N"; fi))"
        read -r -n 1 yn </dev/tty

        case "$yn" in
            "" ) # Enter pressed - use default
                if [[ "${default}" = "y" ]]; then
                    return 0
                else
                    return 1
                fi
                ;;
            [Yy]* ) return 0 ;;
            [Nn]* ) return 1 ;;
            * ) print_warning "Please answer yes or no." ;;
        esac
    done
}

# Function to start the application
start_application() {
    print_message "${YELLOW}" "Starting SuperScript application in background"

    # Check if the application binary exists
    if [[ ! -x "${APP_BIN}" ]]; then
        print_error "Application binary not found or not executable: ${APP_BIN}"
        return 1
    fi

    # Create log directory if it doesn't exist
    local log_dir="$(dirname "${APP_LOG}")"
    mkdir -p "${log_dir}" 2>/dev/null || true

    # Start the application
    nohup "${APP_BIN}" > "${APP_LOG}" 2>&1 &
    local app_pid=$!

    # Wait for application to initialize
    print_message "${YELLOW}" "Waiting for application to initialize..."

    # Wait for the application to start responding or timeout
    local counter=0
    while ! verify_app_responding && ((counter < INIT_TIMEOUT)); do
        printf "."
        sleep 1
        ((counter++))
    done
    printf "\n"

    # Verify the application is running
    if is_app_running; then
        if verify_app_responding; then
            print_success "SuperScript is now running successfully!"
            return 0
        else
            print_warning "SuperScript process is running but not responding to HTTP requests"
            print_warning "You may need to check ${APP_LOG} for errors"
            return 0
        fi
    else
        print_error "Failed to start SuperScript. Check ${APP_LOG} for details."
        return 1
    fi
}

# Cleanup function that runs on script exit
cleanup() {
    local exit_code=$?
    # Add any cleanup tasks here
    # Do not modify exit code
    return ${exit_code}
}

# Set trap for cleanup on EXIT and common signals
trap cleanup EXIT HUP INT TERM

# Main script execution
main() {
    print_message "${GREEN}${BOLD}" "=== Starting SuperScript Application ==="

    # Check for required commands
    if ! command_exists pgrep; then
        print_error "Required command 'pgrep' not found"
        exit 1
    fi

    # Check if application is already running
    if is_app_running; then
        print_message "${YELLOW}" "SuperScript appears to be already running"
    else
        print_message "${YELLOW}" "Checking if Temporal server is running"

        # Check if Temporal server is running
        if is_temporal_running; then
            print_success "Temporal server is running on port ${TEMPORAL_PORT}"
        else
            print_warning "Temporal server is not running"
            printf "Please start Temporal server in another terminal with:\n"
            printf "  make start-temporal\n\n"

            if ask_for_confirmation "Do you want to continue anyway?"; then
                printf "Continuing without verified Temporal server...\n"
            else
                print_error "Aborted. Please start Temporal first with 'make start-temporal'"
                exit 1
            fi
        fi

        # Start the application
        if ! start_application; then
            exit 1
        fi
    fi

    # Print final success message with usage instructions
    print_message "${GREEN}${BOLD}" "SuperScript is ready!"
    printf "HTTP server is running at http://localhost:${APP_PORT}\n"
    printf "You can now run the demo scripts:\n"
    printf "  - make superscript-demo-2 # Traditional approach\n"
    printf "  - make superscript-demo-3 # Single payment workflow\n"
    printf "  - make superscript-demo-4 # Orchestrator workflow\n\n"
    print_message "${YELLOW}" "When done, stop the application with: make superscript-stop"
}

# Execute main function
main
