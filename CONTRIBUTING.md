# Contributing to go-temporal-sre

Thank you for your interest in contributing to go-temporal-sre! This document provides guidelines and information to help you contribute effectively.

## 📋 Table of Contents

1. [Directory Structure](#directory-structure)
2. [Getting Started](#getting-started)
3. [Development Workflow](#development-workflow)
4. [Adding New Features](#adding-new-features)
5. [Testing](#testing)
6. [Documentation](#documentation)
7. [Code Standards](#code-standards)
8. [Pull Request Guidelines](#pull-request-guidelines)

## 🏗️ Directory Structure

This project follows a **centralized worker architecture** with clear separation of concerns:

```
go-temporal-sre/
├── cmd/                        # Main applications
│   ├── worker/                 # ⭐ CENTRALIZED WORKER (Main Entry Point)
│   │   └── main.go             # Single worker handling all features
│   ├── demos/                  # Demo applications
│   │   ├── kilcron/            # Kilcron demo entry point
│   │   ├── superscript/        # SuperScript demo entry point
│   │   └── jit/                # JIT Access demo entry point
│   └── [legacy]/               # Legacy individual commands (deprecated)
├── internal/                   # Private application code
│   ├── worker/                 # Core worker infrastructure
│   │   ├── config/             # Configuration management
│   │   ├── registry.go         # Workflow/activity registration
│   │   └── worker.go           # Centralized worker implementation
│   ├── features/               # Feature implementations
│   │   ├── kilcron/            # Kilcron feature
│   │   ├── superscript/        # SuperScript feature
│   │   └── jit/                # JIT Access feature
│   ├── kilcron/                # Kilcron workflows and activities
│   ├── superscript/            # SuperScript workflows and activities
│   ├── jitaccess/              # JIT Access workflows and activities
│   ├── atlas/                  # MongoDB Atlas integration
│   └── [other features]/       # Other feature implementations
├── demo/                       # Demo frontends and documentation
│   └── jit/                    # JIT Access demo frontend
│       └── demo-fe/            # Python Flask frontend
├── docs/                       # Documentation
├── presentation/               # Presentation materials
└── config.example              # Example configuration file
```

### 🎯 Key Architectural Principles

#### 1. **ONE Centralized Worker**
- **Location**: `cmd/worker/main.go`
- **Purpose**: Single entry point for all Temporal workflows and activities
- **Benefits**: Simplified deployment, shared resources, consistent configuration

#### 2. **Feature-Based Organization**
- **Location**: `internal/features/`
- **Pattern**: Each feature implements the `FeatureRegistrar` interface
- **Structure**: Self-contained features with clear boundaries

#### 3. **Separation of Concerns**
- **Workflows/Activities**: `internal/[feature]/`
- **Feature Integration**: `internal/features/[feature]/`
- **Demos**: `cmd/demos/[feature]/`

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- Temporal CLI installed
- Make (for convenience targets)

### Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/go-temporal-sre.git
   cd go-temporal-sre
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Start Temporal server**
   ```bash
   make start-temporal
   ```

4. **Run the centralized worker**
   ```bash
   make start-worker
   ```

5. **Test with a demo**
   ```bash
   make kilcron-demo
   ```

## 🔄 Development Workflow

### 1. **Create a Feature Branch**
```bash
git checkout -b feature/your-feature-name
```

### 2. **Make Changes**
- Follow the architectural patterns outlined below
- Write tests for new functionality
- Update documentation as needed

### 3. **Test Your Changes**
```bash
make test
make build-all
```

### 4. **Submit Pull Request**
- Follow the PR template
- Include tests and documentation
- Reference relevant issues

## ➕ Adding New Features

### Step 1: Create Feature Implementation

Create your feature in `internal/your-feature/`:

```go
// internal/your-feature/workflow.go
package yourfeature

import (
    "context"
    "go.temporal.io/sdk/workflow"
)

func YourWorkflow(ctx workflow.Context, input YourInput) (YourOutput, error) {
    // Implementation
}
```

### Step 2: Create Feature Registrar

Create `internal/features/your-feature/feature.go`:

```go
package yourfeature

import (
    "app/internal/worker"
    "app/internal/your-feature"
    "go.temporal.io/sdk/log"
)

type Feature struct {
    // Your feature fields
}

func NewFeature(logger log.Logger) *Feature {
    return &Feature{
        // Initialize your feature
    }
}

func (f *Feature) RegisterComponents(registry *worker.Registry, cfg interface{}) error {
    // Register your workflows and activities
    registry.RegisterWorkflow("YourWorkflow", yourfeature.YourWorkflow)
    registry.RegisterActivity("YourActivity", yourfeature.YourActivity)
    return nil
}

func (f *Feature) GetTaskQueues() []string {
    return []string{"your-feature-task-queue"}
}

func (f *Feature) GetFeatureName() string {
    return "your-feature"
}
```

### Step 3: Register in Centralized Worker

Add your feature to `cmd/worker/main.go`:

```go
// Register features
features := []worker.FeatureRegistrar{
    kilcron.NewFeature(logger),
    superscript.NewFeature(logger),
    jit.NewFeature(logger),
    yourfeature.NewFeature(logger), // Add your feature
}
```

### Step 4: Create Demo Application

Create `cmd/demos/your-feature/main.go`:

```go
package main

import (
    "context"
    "log"
    "app/internal/your-feature"
    "go.temporal.io/sdk/client"
)

func main() {
    // Create Temporal client
    // Execute your workflow
    // Handle results
}
```

### Step 5: Add Makefile Targets

Add to `Makefile`:

```makefile
your-feature-demo:
	@echo "Starting Your Feature Demo"
	@go run cmd/demos/your-feature/main.go

build-your-feature:
	@go build -o bin/your-feature-demo cmd/demos/your-feature/main.go
```

### Step 6: Update Configuration

Add feature-specific configuration to `internal/worker/config/config.go`:

```go
type WorkerConfig struct {
    // Existing fields...
    YourFeatureSettings YourFeatureConfig `json:"your_feature_settings"`
}

type YourFeatureConfig struct {
    // Your feature configuration
}
```

## 🧪 Testing

### Unit Tests

Write tests for your workflows and activities:

```go
// internal/your-feature/workflow_test.go
package yourfeature

import (
    "testing"
    "go.temporal.io/sdk/testsuite"
)

func TestYourWorkflow(t *testing.T) {
    testSuite := &testsuite.WorkflowTestSuite{}
    env := testSuite.NewTestWorkflowEnvironment()
    
    env.ExecuteWorkflow(YourWorkflow, YourInput{})
    
    // Assertions
}
```

### Integration Tests

Create integration tests that verify end-to-end functionality:

```go
// internal/your-feature/integration_test.go
package yourfeature

import (
    "testing"
    "go.temporal.io/sdk/testsuite"
)

func TestYourFeatureIntegration(t *testing.T) {
    // Test with real Temporal server
}
```

### Running Tests

```bash
# Run all tests
make test

# Run specific feature tests
go test ./internal/your-feature/...

# Run with coverage
go test -cover ./...
```

## 📚 Documentation

### Code Documentation

- Use Go doc comments for all public functions
- Include usage examples in comments
- Document complex logic with inline comments

### Feature Documentation

Each feature should have:
- `README.md`: Feature overview and usage
- `SCENARIO.md`: Use cases and examples
- `LEARNINGS.md`: Implementation insights (optional)

### Update Project Documentation

When adding features:
1. Update main `README.md`
2. Update `CHANGELOG.md`
3. Add feature to configuration examples

## 📝 Code Standards

### Go Standards

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Use `golint` for linting
- Follow Go naming conventions

### Temporal Standards

- Use descriptive workflow and activity names
- Implement proper error handling
- Use appropriate timeouts and retry policies
- Follow Temporal best practices for determinism

### Project Standards

- **One Worker**: All features must integrate with the centralized worker
- **Feature Interface**: All features must implement `FeatureRegistrar`
- **Configuration**: Use environment variables for configuration
- **Logging**: Use structured logging with consistent format
- **Error Handling**: Implement comprehensive error handling

### File Naming

- Workflows: `workflow.go`
- Activities: `activities.go`
- Feature integration: `feature.go`
- Tests: `*_test.go`

## 🔍 Pull Request Guidelines

### Before Submitting

1. **Test your changes**
   ```bash
   make test
   make build-all
   ```

2. **Check code quality**
   ```bash
   go fmt ./...
   go vet ./...
   ```

3. **Update documentation**
   - Update relevant README files
   - Add/update code comments
   - Update CHANGELOG.md

### PR Requirements

- **Clear Description**: Explain what changes you made and why
- **Test Coverage**: Include tests for new functionality
- **Documentation**: Update relevant documentation
- **Single Responsibility**: One feature/fix per PR
- **Clean History**: Squash commits if necessary

### PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Added unit tests
- [ ] Added integration tests
- [ ] Existing tests pass

## Checklist
- [ ] Code follows project standards
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
```

## 🆘 Getting Help

- **Issues**: Create a GitHub issue for bugs or feature requests
- **Questions**: Use GitHub Discussions for questions
- **Documentation**: Check the project wiki
- **Code Review**: Request reviews from maintainers

## 📊 Project Health

### Key Metrics

- **Build Status**: All builds must pass
- **Test Coverage**: Aim for >80% coverage
- **Code Quality**: Follow linting standards
- **Documentation**: Keep docs up to date

### Performance Considerations

- **Resource Usage**: Monitor worker resource consumption
- **Workflow Efficiency**: Optimize long-running workflows
- **Error Rates**: Monitor and address error patterns

## 🎯 Future Roadmap

- Enhanced monitoring and observability
- Additional SRE automation features
- Web UI for workflow management
- Advanced error handling patterns
- Performance optimization tools

Thank you for contributing to go-temporal-sre! Your contributions help make SRE automation more accessible and reliable. 