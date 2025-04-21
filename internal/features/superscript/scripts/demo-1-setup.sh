#!/usr/bin/env bash
#
# Demo 1: Setup - Build the application only
#
# Purpose: Prepare the SuperScript application for demos
# Author: Enterprise Team
# Created: 2025-03-30
#

# Enable strict mode
set -euo pipefail
IFS=$'\n\t'

# Get script directory (safer approach for sourced scripts)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

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

# Cleanup function that runs on script exit
cleanup() {
    # Add any cleanup tasks here
    local exit_code=$?
    if [[ ${exit_code} -ne 0 ]]; then
        print_error "Script exited with code: ${exit_code}"
    fi
    return ${exit_code}
}

# Print error message
print_error() {
    printf "%s%s%s\n" "${RED}" "ERROR: $1" "${RESET}" >&2
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

# Set trap for cleanup on EXIT, HUP, INT, TERM
trap cleanup EXIT HUP INT TERM

# Main script execution
main() {
    print_message "${GREEN}${BOLD}" "=== Demo 1: Building SuperScript ==="
    print_message "${YELLOW}" "Building the SuperScript application"

    # Commented out build steps - uncomment if needed
     cd "${SCRIPT_DIR}/../../.." && go build -o ./bin/superscript ./cmd/superscript/

    print_message "${GREEN}" "Build complete! The binary is now ready."
    print_message "${YELLOW}" "Next steps:"
    printf "  1. Start Temporal server with: make start-temporal\n"
    printf "  2. Start SuperScript with: make superscript-start\n"
    printf "  3. Run demos with: make superscript-demo-2, demo-3, etc.\n"
    printf "  4. When done, stop with: make superscript-stop\n"
}

# Execute main function
main
