# iWF SuperScript Implementation

This directory contains the iWF (Indeed Workflow Framework) implementation of the SuperScript system, which was originally implemented using Temporal directly.

## Overview

The iWF implementation maintains all the same functionality as the original Temporal implementation but uses iWF's state-based workflow model. This provides several advantages:

1. **Simplified State Management**: Uses iWF's explicit state model rather than Temporal's imperative code model
2. **Native Idempotency**: Leverages iWF's built-in idempotency mechanisms
3. **Cleaner Parallel Execution**: Uses iWF's `MultiNextStates` pattern for orchestration
4. **Portability**: Can be run on either Temporal or Cadence backends

## Implementation Details

### Workflows

1. **SinglePaymentWorkflow**
   - State: `COLLECT_PAYMENT` - Executes the payment collection script for a single order
   - Maintains idempotency using WorkflowIDReusePolicy.REJECT_DUPLICATE

2. **OrchestratorWorkflow**
   - States:
     - `START_CHILDREN` - Initiates child workflows for each order ID with concurrency control
     - `AGGREGATE_RESULTS` - Collects and processes results from all child workflows

### Key Differences from Temporal Implementation

1. **State-Based vs. Imperative**: The iWF implementation uses explicit state transitions rather than imperative code
2. **Child Workflow Handling**: Uses iWF's child workflow commands with built-in concurrency control
3. **API Compatibility**: Maintains the same API and result structures for compatibility with existing code

## Usage

### Running the Server

```bash
go run cmd/iwf-superscript/main.go
```

This will start:
- An iWF worker on the `iwf-superscript-task-queue` task queue
- An HTTP server on port 8081 with the same API endpoints as the original implementation

### API Endpoints

- `GET /health` - Health check endpoint
- `POST /run/single` - Run a single payment collection workflow
- `POST /run/batch` - Run the orchestrator workflow for multiple orders
- `GET /run/traditional` - Run the traditional script directly (for comparison)

## Testing

The implementation includes comprehensive tests that verify:
- Single payment workflow execution (success and failure cases)
- Orchestrator workflow with parallel execution
- Idempotency behavior

Run the tests with:

```bash
go test -v ./internal/iwf-superscript
```

## Dependencies

- iWF Golang SDK: `github.com/indeedeng/iwf-golang-sdk`
- Original SuperScript code: `app/internal/superscript`

## Notes

- The iWF implementation runs alongside the original Temporal implementation
- Both implementations use the same activity code and result structures
- The HTTP API is nearly identical, just on a different port (8081 vs 8080)
