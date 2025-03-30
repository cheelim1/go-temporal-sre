#!/bin/bash
# Script to stop the SuperScript application
set -e

echo -e "\033[1;33m=== Stopping SuperScript Application ===\033[0m"

# Check if application is running
if pgrep -f "bin/superscript" >/dev/null; then
    echo -e "Stopping SuperScript application..."
    pkill -f "bin/superscript"
    sleep 1
    
    # Verify it's stopped
    if pgrep -f "bin/superscript" >/dev/null; then
        echo -e "\033[1;31mFailed to stop SuperScript. Trying again with force...\033[0m"
        pkill -9 -f "bin/superscript"
        sleep 1
    fi
    
    if ! pgrep -f "bin/superscript" >/dev/null; then
        echo -e "\033[1;32mSuperScript has been stopped successfully\033[0m"
    else
        echo -e "\033[1;31mFailed to stop SuperScript\033[0m"
        exit 1
    fi
else
    echo -e "\033[1;33mSuperScript is not running\033[0m"
fi

echo -e "\033[1;32mCleanup complete!\033[0m"
