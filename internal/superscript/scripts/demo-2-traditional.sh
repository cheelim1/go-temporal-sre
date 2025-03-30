#!/bin/bash
# Demo 2: Run traditional non-idempotent script multiple times
set -e

echo -e "\033[1;32m=== Demo 2: Testing Traditional Non-Idempotent Script ===\033[0m"
echo -e "\033[1;33mThis demo shows the NON-IDEMPOTENT behavior of the traditional script\033[0m"
echo -e "Running the script multiple times will process payments multiple times!"
echo -e "\033[1;31mIn a real-world application, this could cause double-charging customers\033[0m"

# First check if the SuperScript application is running
if ! pgrep -f "bin/superscript" >/dev/null; then
    echo -e "\033[1;31mError: SuperScript application is not running\033[0m"
    echo -e "Please run 'make superscript-demo-1' first"
    exit 1
fi

# Function to run the traditional script and display output
run_traditional() {
    echo -e "\n\033[1;33mRunning traditional script (Run $1 of 3)...\033[0m"
    curl -s http://localhost:8080/run/traditional
    echo -e "\n\033[1;36mScript is running in the background. Check application logs for output.\033[0m"
    sleep 2
}

echo -e "\nWe'll run the traditional script 3 times in quick succession."
echo -e "This demonstrates how multiple requests can lead to duplicate processing.\n"

# Run the script multiple times to show non-idempotent behavior
for i in {1..3}; do
    run_traditional $i
done

echo -e "\n\033[1;31mNote how each script execution processes payments independently\033[0m"
echo -e "This can lead to race conditions and duplicate payments\n"

echo -e "\033[1;32mTo contrast this with the idempotent behavior of Temporal workflows,\033[0m"
echo -e "\033[1;32mrun 'make superscript-demo-3' next\033[0m"
