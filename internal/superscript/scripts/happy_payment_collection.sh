#!/usr/bin/env bash
#
# Demo 1: Setup - Build the application only
#
# Purpose: Prepare the SuperScript application for demos
# Author: Enterprise Team
# Created: 2025-03-30
#
set -euo pipefail
IFS=$'\n\t'

# Global variable to store last error message
LAST_ERROR_MSG=""

# Setup error handling - ensures we clean up properly even if script is terminated early
cleanup() {
    local exit_code=$?
    echo "Cleaning up resources..."

    # Add any cleanup actions here (e.g., removing temp files, releasing locks)

    # Log exit information with the error message if available
    if [[ $exit_code -ne 0 ]]; then
        if [[ -n "$LAST_ERROR_MSG" ]]; then
            echo "ERROR: Script terminated with exit code: $exit_code - $LAST_ERROR_MSG" >&2
        else
            echo "ERROR: Script terminated with exit code: $exit_code" >&2
        fi
    fi

    exit $exit_code
}

# Register the cleanup function for these signals
# Only register for EXIT to avoid duplicate cleanup calls
trap cleanup EXIT

# Source the functions library
#SOURCE_DIR="$(dirname "${BASH_SOURCE[0]}")"
#if ! source "$SOURCE_DIR/func_collect_payment.sh"; then
#    echo "ERROR: Cannot source functions library" >&2
#    exit 1
#fi

# Function 1: Success 100% of the time
# Input: OrderID
# Output: Prints "Step1 <OrderID>" after 1-3s
process_step1() {
    local order_id="$1"
    # Random sleep between 1-3 seconds
    local sleep_time=$(( ( RANDOM % 3 ) + 1 ))
    sleep "$sleep_time"
    echo "Step1 $order_id"
    return 0
}

# Check if OrderID is provided
if [[ $# -lt 1 ]]; then
    echo "ERROR: Missing OrderID parameter" >&2
    echo "Usage: $0 <OrderID>" >&2
    exit 1
fi

# Verify OrderID is a number
ORDER_ID="$1"
if ! [[ "$ORDER_ID" =~ ^[0-9]+$ ]]; then
    echo "ERROR: OrderID must be a number" >&2
    exit 2
fi

echo "Starting payment processing for OrderID: $ORDER_ID"

# Process Step 1
echo "Starting processing step 1..."
# Turn off errexit temporarily to capture the output and return code
set +e
step1_result=$(process_step1 "$ORDER_ID")
step1_code=$?
set -e

if [[ $step1_code -ne 0 ]]; then
    LAST_ERROR_MSG="Step 1 failed: $step1_result"
    echo "$LAST_ERROR_MSG" >&2
    exit $step1_code
fi

echo "Step 1 completed successfully: $step1_result"

# All steps completed successfully
echo "Payment processing completed successfully for OrderID: $ORDER_ID"
exit 0
#
## Enable strict mode
#set -euo pipefail
#IFS=$'\n\t'
#
## Get script directory (safer approach for sourced scripts)
#SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
#
## Define color codes using tput (more portable than ANSI escape sequences)
#if [[ -t 1 ]]; then  # Check if stdout is a terminal
#    readonly BOLD=$(tput bold)
#    readonly GREEN=$(tput setaf 2)
#    readonly YELLOW=$(tput setaf 3)
#    readonly RED=$(tput setaf 1)
#    readonly RESET=$(tput sgr0)
#else
#    readonly BOLD=""
#    readonly GREEN=""
#    readonly YELLOW=""
#    readonly RED=""
#    readonly RESET=""
#fi
#
## Setup error handling - ensures we clean up properly even if script is terminated early
#cleanup() {
#    local exit_code=$?
#    echo "Cleaning up resources..."
#
#    # Add any cleanup actions here (e.g., removing temp files, releasing locks)
#
#    # Log exit information with the error message if available
#    if [[ $exit_code -ne 0 ]]; then
#        if [[ -n "$LAST_ERROR_MSG" ]]; then
#            echo "ERROR: Script terminated with exit code: $exit_code - $LAST_ERROR_MSG" >&2
#        else
#            echo "ERROR: Script terminated with exit code: $exit_code" >&2
#        fi
#    fi
#
#    exit $exit_code
#}
#
## Print error message
#print_error() {
#    printf "%s%s%s\n" "${RED}" "ERROR: $1" "${RESET}" >&2
#}
#
## Print warning message
#print_warning() {
#    printf "%s%s%s\n" "${YELLOW}" "WARNING: $1" "${RESET}" >&2
#}
#
## Function to print colored messages
#print_message() {
#    local color="$1"
#    local message="$2"
#    printf "%s%s%s\n" "${color}" "${message}" "${RESET}"
#}
#
## Function 1: Success 100% of the time
## Input: OrderID
## Output: Prints "Step1 <OrderID>" after 1-3s
#process_step1() {
#    local order_id="$1"
#
#    # 50% chance of failure
##    if (( RANDOM % 2 )); then
##        echo "FAILED: Processing Step 1 for OrderID $order_id"
##        return 1
##    else
#        # Random sleep between 1-3 seconds
#        local sleep_time=$(( ( RANDOM % 3 ) + 1 ))
#        sleep "$sleep_time"
#        echo "Step1 $order_id"
#        return 0
##    fi
#}
#
## Set trap for cleanup on EXIT, HUP, INT, TERM
#trap cleanup EXIT HUP INT TERM
#
## Main script execution
#main() {
#    print_message "${GREEN}${BOLD}" "=== Unit Test: Happy Payment Collection ==="
#    print_message "${YELLOW}" "Lala collect payment in perfect conditions .."
#
#    # Check if OrderID is provided
#    if [[ $# -lt 1 ]]; then
#        echo "ERROR: Missing OrderID parameter" >&2
#        echo "Usage: $0 <OrderID>" >&2
#        exit 1
#    fi
#
#    # Verify OrderID is a number
#    ORDER_ID="$1"
#    if ! [[ "$ORDER_ID" =~ ^[0-9]+$ ]]; then
#        echo "ERROR: OrderID must be a number" >&2
#        exit 2
#    fi
#
#    echo "Starting payment processing for OrderID: $ORDER_ID"
#
#    # Process Step 1
#    echo "Starting processing step 1..."
#    # Turn off errexit temporarily to capture the output and return code
#    set +e
#    step1_result=$(process_step1 "$ORDER_ID")
#    step1_code=$?
#    set -e
#
#    if [[ $step1_code -ne 0 ]]; then
#        LAST_ERROR_MSG="Step 1 failed: $step1_result"
#        echo "$LAST_ERROR_MSG" >&2
#        exit $step1_code
#    fi
#
#    echo "Step 1 completed successfully: $step1_result"
#  # All steps completed successfully
#  echo "Payment processing completed successfully for OrderID: $ORDER_ID"
#  exit 0
#
#}
#
#main