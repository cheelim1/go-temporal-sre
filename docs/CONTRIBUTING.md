# Contributing to go-temporal-sre

Thank you for your interest in contributing to go-temporal-sre! This document provides guidelines for contributing new demos and features.

## Project Structure

```
go-temporal-sre/
├── cmd/
│   ├── worker/                 # Central worker binary
│   └── demos/                  # Demo binaries
├── internal/
│   ├── worker/                # Central worker implementation
│   ├── shared/                # Shared utilities
│   └── demos/                 # Demo implementations
└── docs/                      # Documentation
```

## Adding a New Demo

1. **Create Demo Structure**:
   - Copy the template from `internal/demos/template/`
   - Rename to your demo name
   - Update package names and imports

2. **Implement Core Logic**:
   - Workflow in `workflow.go`
   - Activities in `activities.go`
   - Tests in `*_test.go` files

3. **Add API (if needed)**:
   - Types in `api/types.go`
   - Handlers in `api/handlers.go`

4. **Add Frontend (if needed)**:
   - UI components
   - API integration

5. **Register with Central Worker**:
   - Add registration in `cmd/worker/main.go`

6. **Add Documentation**:
   - Update README
   - Add architecture diagrams
   - Document API endpoints

## Best Practices

1. **Code Organization**:
   - Keep related code together
   - Use shared utilities when possible
   - Follow consistent naming conventions

2. **Error Handling**:
   - Use proper error types
   - Include context in error messages
   - Handle errors at appropriate levels

3. **Testing**:
   - Add unit tests for workflows and activities
   - Add integration tests for API
   - Add end-to-end tests for complete flow

4. **Documentation**:
   - Document all public APIs
   - Include usage examples
   - Keep README up to date

5. **Performance**:
   - Use appropriate timeouts
   - Implement proper retry policies
   - Handle long-running operations

## Development Workflow

1. **Setup**:
   ```bash
   git clone https://github.com/yourusername/go-temporal-sre.git
   cd go-temporal-sre
   go mod download
   ```

2. **Create Branch**:
   ```bash
   git checkout -b feature/your-demo-name
   ```

3. **Implement Changes**:
   - Follow the template structure
   - Add tests
   - Update documentation

4. **Run Tests**:
   ```bash
   go test ./...
   ```

5. **Submit PR**:
   - Create pull request
   - Include description of changes
   - Reference related issues

## Code Review Process

1. **Review Checklist**:
   - [ ] Code follows project structure
   - [ ] Tests are included
   - [ ] Documentation is updated
   - [ ] Error handling is appropriate
   - [ ] Performance considerations are addressed

2. **Review Process**:
   - PR is assigned to reviewers
   - Reviewers provide feedback
   - Author addresses feedback
   - PR is merged when approved

## Questions?

If you have any questions, please:
1. Check the documentation
2. Look at existing demos
3. Open an issue
4. Join our community chat 