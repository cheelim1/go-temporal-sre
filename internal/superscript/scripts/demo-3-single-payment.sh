#!/bin/bash
# Demo 3: Test single payment idempotent workflow using Temporal
set -e

echo -e "\033[1;32m=== Demo 3: Testing Single Payment Idempotent Workflow ===\033[0m"
echo -e "\033[1;33mThis demo shows how Temporal makes script execution IDEMPOTENT\033[0m"
echo -e "Running the workflow multiple times with the same OrderID will execute the script only ONCE!"
echo -e "\033[1;32mTemporally uses WorkflowIDReusePolicy.REJECT_DUPLICATE to ensure idempotency\033[0m"

# First check if the SuperScript application is running
if ! pgrep -f "bin/superscript" >/dev/null; then
    echo -e "\033[1;31mError: SuperScript application is not running\033[0m"
    echo -e "Please run 'make superscript-demo-1' first"
    exit 1
fi

# Function to run the single payment workflow and display output
run_single_payment() {
    echo -e "\n\033[1;33mRunning single payment workflow (Run $1 of 3)...\033[0m"
    response=$(curl -s -X POST http://localhost:8080/run/single -H "Content-Type: application/json" -d '{"order_id":"ORD-DEMO-123"}')
    echo -e "\n\033[1;36mResponse:\033[0m"
    echo "$response" | python3 -m json.tool 2>/dev/null || echo "$response"
    echo ""
    sleep 1
}

echo -e "\nWe'll run the single payment workflow 3 times in succession with the same OrderID."
echo -e "This demonstrates how Temporal ensures each payment is processed exactly once.\n"

# Run the workflow multiple times to show idempotent behavior
for i in {1..3}; do
    run_single_payment $i
done

echo -e "\n\033[1;32mObserve how only the first request executes the script\033[0m"
echo -e "The subsequent requests are handled idempotently, returning results from the first execution"
echo -e "This prevents duplicate processing even with concurrent requests\n"

echo -e "\033[1;33mKey points about the implementation:\033[0m"
echo -e "1. We're using WorkflowIDReusePolicy.REJECT_DUPLICATE in Temporal"
echo -e "2. The workflow ID is derived from the OrderID to ensure uniqueness"
echo -e "3. Temporal guarantees exactly-once execution semantics"
echo -e "4. All concurrent calls get the same results without duplicate processing\n"

echo -e "\033[1;32mTo see how we can orchestrate multiple payments while maintaining idempotency,\033[0m"
echo -e "\033[1;32mrun 'make superscript-demo-4' next\033[0m"
