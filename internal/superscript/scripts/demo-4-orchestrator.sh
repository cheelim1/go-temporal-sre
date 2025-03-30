#!/bin/bash
# Demo 4: Test the orchestrator workflow that manages multiple payment collections
set -e

echo -e "\033[1;32m=== Demo 4: Testing Orchestrator Workflow ===\033[0m"
echo -e "\033[1;33mThis demo shows how Temporal orchestrates multiple child workflows while maintaining idempotency\033[0m"
echo -e "The orchestrator replaces the traditional batch script with a proper workflow"
echo -e "\033[1;32mEach child workflow still maintains its own idempotency guarantees\033[0m"

# First check if the SuperScript application is running
if ! pgrep -f "bin/superscript" >/dev/null; then
    echo -e "\033[1;31mError: SuperScript application is not running\033[0m"
    echo -e "Please run 'make superscript-demo-1' first"
    exit 1
fi

# Function to run the orchestrator workflow and display output
run_orchestrator() {
    echo -e "\n\033[1;33mRunning orchestrator workflow (Run $1 of 2)...\033[0m"
    response=$(curl -s -X POST http://localhost:8080/run/batch -H "Content-Type: application/json" -d '{}')
    echo -e "\n\033[1;36mResponse:\033[0m"
    echo "$response" | python3 -m json.tool 2>/dev/null || echo "$response"
    echo ""
    sleep 2
}

echo -e "\nWe'll run the orchestrator workflow twice."
echo -e "This demonstrates how Temporal manages multiple child workflows with idempotency.\n"

# Run the workflow multiple times
for i in {1..2}; do
    run_orchestrator $i
done

echo -e "\n\033[1;32mThe orchestrator creates child workflows for each OrderID\033[0m"
echo -e "Each child workflow has its own WorkflowID based on the OrderID"
echo -e "This ensures each payment is processed exactly once, even across multiple orchestrator runs\n"

echo -e "\033[1;33mKey advantages of the Temporal orchestration:\033[0m"
echo -e "1. Automatic retry for failed activities"
echo -e "2. Clear visibility into workflow execution status"
echo -e "3. Exactly-once execution guarantees for each payment"
echo -e "4. Graceful handling of concurrency through Temporal's REJECT_DUPLICATE policy" 
echo -e "5. Scalable architecture for processing large batches\n"

echo -e "\033[1;32mThis completes the demo of how to make non-idempotent scripts idempotent with Temporal!\033[0m"
echo -e "To stop the SuperScript application, run 'make superscript-stop'"
