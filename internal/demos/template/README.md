# [Demo Name] Template

This is a template for creating new Temporal demos. Replace this text with your demo's description.

## Architecture

The demo consists of the following components:

1. **Temporal Worker**: Runs the workflow and activities
2. **Backend API**: HTTP server that interfaces with Temporal (if needed)
3. **Frontend**: User interface (if needed)

## Directory Structure

```
internal/demos/[demo-name]/
├── api/           # API types and handlers (if needed)
├── frontend/      # Frontend implementation (if needed)
├── workflow.go    # Workflow implementation
├── activities.go  # Activities implementation
└── README.md      # Documentation

cmd/demos/[demo-name]/
├── backend/       # Backend service (if needed)
└── frontend/      # Frontend service (if needed)
```

## Implementation Steps

1. **Create Workflow and Activities**:
   - Implement workflow in `workflow.go`
   - Implement activities in `activities.go`
   - Add tests in `workflow_test.go` and `activities_test.go`

2. **Create API (if needed)**:
   - Define types in `api/types.go`
   - Implement handlers in `api/handlers.go`

3. **Create Frontend (if needed)**:
   - Implement UI components
   - Add API integration

4. **Register with Central Worker**:
   - Add registration in `cmd/worker/main.go`

5. **Add Documentation**:
   - Update this README
   - Add architecture diagrams
   - Document API endpoints
   - Add usage examples

## Testing

1. Unit tests for workflow and activities
2. Integration tests for API (if applicable)
3. End-to-end tests for the complete flow

## Deployment

1. Build and run the central worker
2. Build and run the backend service (if applicable)
3. Build and run the frontend service (if applicable)

## Best Practices

1. Use shared utilities from `internal/shared/`
2. Follow consistent error handling patterns
3. Add proper logging and metrics
4. Document all public APIs
5. Include example usage 