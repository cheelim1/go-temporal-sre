# Breakglass Emergency Actions Demo

This demo showcases how to use Temporal for emergency breakglass scenarios, such as service restarts, scaling, and rollbacks.

## Architecture

The demo consists of three main components:

1. **Temporal Worker**: Runs the workflow and activities
2. **Backend API**: HTTP server that interfaces with Temporal
3. **Streamlit Frontend**: User interface for executing breakglass actions

## Running the Demo

### Prerequisites

- Go 1.16+
- Python 3.8+
- Temporal server running locally
- Streamlit (`pip install streamlit`)

### Starting the Components

1. Start the Temporal worker:
```bash
go run cmd/worker/main.go
```

2. Start the backend API:
```bash
go run cmd/demos/breakglass/backend/main.go
```

3. Start the Streamlit frontend:
```bash
streamlit run internal/demos/breakglass/frontend/app.py
```

## Using the Demo

1. Open the Streamlit interface in your browser (default: http://localhost:8501)
2. Fill in the service details:
   - Service ID
   - Action (restart, scale, or rollback)
   - Action-specific parameters
   - Your email (for audit purposes)
3. Click "Execute Action" to start the workflow
4. Monitor the workflow status in the interface

## API Endpoints

### POST /api/breakglass
Start a breakglass action workflow.

Request body:
```json
{
  "service_id": "service-123",
  "action": "restart",
  "parameters": {},
  "requested_by": "admin@example.com"
}
```

### GET /api/breakglass/status
Get the status of a workflow.

Query parameters:
- `workflow_id`: The ID of the workflow to check

## Workflow Details

The workflow handles three types of emergency actions:

1. **Restart**: Restarts a service
2. **Scale**: Changes the number of service replicas
3. **Rollback**: Reverts a service to a previous version

Each action is implemented as a separate activity, and the workflow ensures proper error handling and status reporting. 