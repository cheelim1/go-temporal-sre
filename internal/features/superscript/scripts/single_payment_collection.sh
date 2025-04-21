#!/bin/bash
##
## Single Payment Collection Script
## This script processes a single payment using two processing steps
## It sources functions from func_collect_payment.sh

# Always use strict mode for reliable error handling
# URL: http://redsymbol.net/articles/unofficial-bash-strict-mode/
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
SOURCE_DIR="$(dirname "${BASH_SOURCE[0]}")"
if ! source "$SOURCE_DIR/func_collect_payment.sh"; then
    echo "ERROR: Cannot source functions library" >&2
    exit 1
fi

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

# Process Step 2
echo "Starting processing step 2..."
# Turn off errexit temporarily to capture the output and return code
set +e
step2_result=$(process_step2 "$ORDER_ID")
step2_code=$?
set -e

if [[ $step2_code -ne 0 ]]; then
    LAST_ERROR_MSG="Step 2 failed: $step2_result"
    echo "$LAST_ERROR_MSG" >&2
    exit $step2_code
fi

echo "Step 2 completed successfully: $step2_result"

# All steps completed successfully
echo "Payment processing completed successfully for OrderID: $ORDER_ID"
exit 0
