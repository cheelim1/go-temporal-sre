# SuperScript Demo

This project demonstrates how to make non-idempotent scripts idempotent using Temporal workflow orchestration. It wraps bash scripts in a way that ensures they are executed exactly once, even when called multiple times.

## Prerequisites

- Go 1.24.1 or later
- Temporal server running locally (default: localhost:7233)
- Bash environment for running scripts

## Setup Instructions

1. Make sure you have a local Temporal server running:
   ```
   # Using Temporal CLI
   temporal server start-dev
   
   # Or using Docker
   docker run --rm -p 7233:7233 temporalio/temporal:latest
   ```

2. Build the SuperScript application:
   ```
   cd /Users/leow/GOMOD/go-temporal-sre
   go build -o bin/superscript ./cmd/superscript/
   ```

3. Run the application:
   ```
   ./bin/superscript
   ```

The application will start both a Temporal worker and an HTTP server on port 8080.

## Demo Instructions

We've created a set of demo scripts that you can run using make targets. These demonstrate the difference between traditional non-idempotent scripts and Temporal-wrapped idempotent workflows.

### Automated Demo Scripts

> **IMPORTANT**: You should have 3 separate terminal windows open:
> 1. **Terminal 1**: For running the Temporal server
> 2. **Terminal 2**: For running the SuperScript application
> 3. **Terminal 3**: For executing the demo scripts

Follow these steps in the corresponding terminals:

#### Terminal 1 (Temporal Server)
```bash
# Start the Temporal server
make start-temporal
```

#### Terminal 2 (SuperScript Application)
```bash
# First build the application
make superscript-demo-1

# Then start the SuperScript application
make superscript-start
```

#### Terminal 3 (Demo Scripts)
```bash
# Run the demo scripts one by one
make superscript-demo-2  # Test traditional non-idempotent script

# After observing the results, run the next demo
make superscript-demo-3  # Test idempotent single payment workflow

# Finally, run the orchestrator demo
make superscript-demo-4  # Test orchestrator workflow
```

#### When done with the demos
```bash
# Stop the SuperScript application (in Terminal 2)
make superscript-stop

# Stop the Temporal server (in Terminal 1)
# Use Ctrl+C to terminate
```

Each demo script provides detailed commentary on what's happening and how the idempotency guarantees work.

### Manual Demo Steps

If you prefer to run the demos manually, you can follow these steps:

#### 1. Demo Traditional Non-Idempotent Script

The traditional script has no idempotency protection, so running it multiple times in parallel can lead to race conditions and duplicate payment processing.

```
# Run the traditional script directly
curl http://localhost:8080/run/traditional

# Check the server logs for output
```

Try running this command multiple times in quick succession to simulate concurrent requests. You'll notice that each execution runs independently, potentially causing duplicate payments.

#### 2. Demo Single Payment Collection (Idempotent)

Using Temporal's workflow ID reuse policy, we ensure that only one execution processes a given OrderID.

```
# Run a single payment workflow with the sample OrderID (ORD-DEMO-123)
curl -X POST http://localhost:8080/run/single -H "Content-Type: application/json" -d '{"order_id":"ORD-DEMO-123"}'

# Run it again immediately
curl -X POST http://localhost:8080/run/single -H "Content-Type: application/json" -d '{"order_id":"ORD-DEMO-123"}'
```

You'll see that the second call is handled idempotently - Temporal will return the same workflow execution, without triggering another script execution.

#### 3. Demo Orchestrator Workflow

The orchestrator workflow runs multiple child workflows, each responsible for a single payment collection.

```
# Run the batch workflow with default order IDs
curl -X POST http://localhost:8080/run/batch -H "Content-Type: application/json" -d '{}'
```

This will start an orchestrator workflow that processes multiple orders in parallel. Check the server logs to see how the orchestrator workflow handles each order.

## Verification

To verify idempotency with Temporal:

1. Open the Temporal Web UI at http://localhost:8088 (if using default Temporal setup)
2. Look at the "default" namespace and find your workflows
3. Notice how:
   - The single payment workflow (with prefix "single-payment-workflow-") will have only one execution per OrderID
   - Child workflows maintain their idempotency even when the parent orchestrator is called multiple times
   - The script is only executed once per OrderID, regardless of how many workflow requests are made

## Understanding the Implementation

- **Single Payment Workflow**: Acts as a wrapper around the non-idempotent script, using Temporal's WorkflowIDReusePolicy.REJECT_DUPLICATE to ensure idempotency.
  
- **Orchestrator Workflow**: Manages multiple child workflows, creating a deterministic WorkflowID for each OrderID, ensuring each child workflow is executed exactly once.
  
- **Script Execution**: The script is executed as an activity within Temporal, allowing for retry policies and proper error handling.

This implementation demonstrates how Temporal can easily wrap existing non-idempotent code to make it idempotent, without having to completely rewrite the underlying business logic.
