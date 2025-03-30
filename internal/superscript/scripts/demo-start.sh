#!/bin/bash
# Script to start the SuperScript application
set -e

echo -e "\033[1;32m=== Starting SuperScript Application ===\033[0m"

# Check if application is already running
if pgrep -f "bin/superscript" >/dev/null; then
    echo -e "\033[1;33mSuperScript appears to be already running\033[0m"
else
    echo -e "\033[1;33mChecking if Temporal server is running\033[0m"
    # Check if Temporal server is running
    if nc -z localhost 7233 >/dev/null 2>&1; then
        echo -e "\033[1;32mTemporal server is running on port 7233\033[0m"
    else
        echo -e "\033[1;31mWarning: Temporal server is not running\033[0m"
        echo -e "Please start Temporal server in another terminal with:"
        echo -e "  make start-temporal"
        echo -e "\nDo you want to continue anyway? [y/N]: "
        read -n 1 answer
        if [[ "$answer" != "y" && "$answer" != "Y" ]]; then
            echo -e "\nAborted. Please start Temporal first with 'make start-temporal'"
            exit 1
        fi
        echo -e "\nContinuing without verified Temporal server..."
    fi

    echo -e "\033[1;33mStarting SuperScript application in background\033[0m"
    nohup /Users/leow/GOMOD/go-temporal-sre/bin/superscript > /Users/leow/GOMOD/go-temporal-sre/superscript.log 2>&1 &
    
    echo -e "Waiting for application to initialize..."
    sleep 3
    
    if pgrep -f "bin/superscript" >/dev/null; then
        echo -e "\033[1;32mSuperScript is now running successfully!\033[0m"
    else
        echo -e "\033[1;31mFailed to start SuperScript. Check superscript.log for details.\033[0m"
        exit 1
    fi
fi

echo -e "\033[1;32mSuperScript is ready!\033[0m"
echo -e "HTTP server is running at http://localhost:8080"
echo -e "You can now run the demo scripts (superscript-demo-2, superscript-demo-3, etc.)"
echo -e "When done, stop the application with 'make superscript-stop'"
