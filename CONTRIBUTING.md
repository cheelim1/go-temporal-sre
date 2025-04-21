# Contributing to go-temporal-sre

This document outlines the project structure and guidelines for contributing to this project.

## Project Structure

```
go-temporal-sre/
├── cmd/                    # Main application entry points
│   ├── worker/            # Centralized Temporal worker
│   └── server/            # HTTP server for demos
├── internal/              # Private application code
│   ├── worker/           # Shared worker implementation
│   │   ├── config/       # Worker configuration
│   │   ├── activities/   # Shared activities
│   │   └── workflows/    # Shared workflows
│   ├── common/           # Shared utilities
│   │   ├── logger/       # Logging utilities
│   │   ├── metrics/      # Metrics collection
│   │   └── utils/        # Common utilities
│   └── features/         # Feature-specific implementations
│       ├── batch/        # Batch processing
│       ├── enrichment/   # Data enrichment
│       └── superscript/  # Script execution
├── docs/                 # Documentation
│   ├── architecture/    # Architecture diagrams
│   └── guides/          # How-to guides
└── test/                # Test utilities and fixtures
    ├── fixtures/        # Test data
    └── utils/           # Test utilities
```

## Key Components

### Centralized Worker
The project uses a centralized Temporal worker implementation located in `internal/worker/`. This worker:
- Can be configured for different use cases
- Supports custom activities and workflows
- Handles common concerns like logging and metrics

### Shared Utilities
Common utilities are located in `internal/common/`:
- Logging utilities
- Metrics collection
- Common helper functions

### Feature Implementation
Feature-specific code is organized in `internal/features/`:
- Each feature is self-contained
- Can be easily added or removed
- Shares common utilities and worker

### Demos
Demo applications are organized in `demos/`:
- Each demo is self-contained
- Uses shared worker and utilities
- Includes documentation and examples

## Adding New Features

1. Create a new directory in `internal/features/`
2. Implement your feature-specific code
3. Register activities and workflows with the shared worker
4. Add tests in the corresponding test directory
5. Create a demo if applicable

## Configuration

Configuration is managed through:
- Environment variables
- Configuration files
- Command-line flags

See `pkg/config/` for details.

## Testing

- Unit tests: `*_test.go` files alongside implementation
- Integration tests: `test/integration/`
- E2E tests: `test/e2e/`

## Documentation

- Architecture: `docs/architecture/`
- Guides: `docs/guides/`
- API documentation: Generated from code comments

## Contributing Guidelines

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Update documentation
6. Submit a pull request

## Code Style

- Follow Go standard formatting
- Use meaningful variable names
- Add comments for complex logic
- Write tests for new features 