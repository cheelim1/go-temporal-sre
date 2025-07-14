# go-temporal-sre

**SRE day to day powered by Temporal**

A collection of SRE (Site Reliability Engineering) tools and demos built on [Temporal](https://temporal.io/), showcasing real-world automation scenarios with centralized worker architecture.

## 🏗️ Architecture

This project uses a **centralized Temporal worker** architecture that supports multiple features and demos:

- **Centralized Worker**: Single worker that can handle multiple features
- **Feature-based Organization**: Each feature is self-contained and can be enabled/disabled
- **Configuration-driven**: Environment variables control which features are active
- **Shared Infrastructure**: Common logging, metrics, and client management

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- Temporal CLI installed
- Make (for convenience targets)

### 1. Start Temporal Server

```bash
make start-temporal
```

### 2. Run the Centralized Worker

```bash
# Start worker with all features
make start-worker

# Or start worker with specific features
make start-worker-features FEATURES=kilcron,superscript
```

### 3. Run Individual Demos

```bash
# Kilcron demo (payment collection scheduler)
make kilcron-demo

# SuperScript demo (script execution orchestrator)
make superscript-demo

# JIT Access demo (just-in-time access management)
make jit-fe-setup
make jit-demo
make jit-fe
```

## 📁 Project Structure

```
go-temporal-sre/
├── cmd/
│   ├── worker/                # Centralized Temporal worker
│   └── demos/                 # Demo applications
│       ├── kilcron/          # Kilcron demo
│       └── superscript/      # SuperScript demo
├── internal/
│   ├── worker/               # Shared worker implementation
│   │   ├── config/          # Worker configuration
│   │   ├── registry.go      # Workflow/activity registry
│   │   └── worker.go        # Centralized worker logic
│   ├── features/            # Feature implementations
│   │   ├── kilcron/         # Kilcron feature
│   │   └── superscript/     # SuperScript feature
│   └── [legacy features]/   # Original feature implementations
└── docs/                    # Documentation
```

## 🎯 Features

### Kilcron - Scheduled Payment Collection
- **Purpose**: Automated payment collection with retry logic
- **Use Case**: SRE automation for billing systems
- **Demo**: `make kilcron-demo`

### SuperScript - Script Execution Orchestrator
- **Purpose**: Orchestrate script execution with error handling
- **Use Case**: Infrastructure automation and maintenance tasks
- **Demo**: `make superscript-demo`

### JIT Access - Just-In-Time Access Management
- **Purpose**: Temporary access provisioning with automatic cleanup
- **Use Case**: Security-focused access management  
- **Demo**: `make jit-demo`

## 🔧 Configuration

Configuration is managed through environment variables:

```bash
# Temporal connection
TEMPORAL_HOST=localhost:7233
TEMPORAL_namespace=default

# Worker settings
MAX_CONCURRENT_ACTIVITIES=10
MAX_CONCURRENT_WORKFLOWS=10

# Feature enablement
ENABLED_FEATURES=kilcron,superscript,jit

# Atlas/MongoDB settings (for JIT feature)
ATLAS_PUBLIC_KEY=your_atlas_public_key
ATLAS_PRIVATE_KEY=your_atlas_private_key
ATLAS_PROJECT_ID=your_atlas_project_id

# Logging
LOG_LEVEL=INFO

# HTTP settings
HTTP_PORT=8080
HTTP_HOST=localhost
```

## 📚 Documentation

- [Contributing Guide](CONTRIBUTING.md) - How to contribute to the project
- [CHANGELOG](CHANGELOG.md) - Project change history
- [Architecture Deep Dive](docs/architecture/) - Detailed architecture documentation

## 🛠️ Development

### Adding a New Feature

1. Create feature implementation in `internal/features/yourfeature/`
2. Implement the `FeatureRegistrar` interface
3. Register the feature in `cmd/worker/main.go`
4. Create a demo in `cmd/demos/yourfeature/`
5. Add Makefile targets for your feature

### Testing

```bash
# Run all tests
make test

# Run specific feature tests
go test ./internal/features/kilcron/...
```

## 🌟 Why Centralized Worker For this Project?

- **Simplified Deployment**: One binary to deploy and manage
- **Resource Efficiency**: Shared connections and resources
- **Consistent Configuration**: Single source of truth for settings
- **Easy Feature Management**: Enable/disable features without code changes
- **Better Observability**: Centralized logging and metrics

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

- [Temporal Documentation](https://docs.temporal.io/)
- [Project Wiki](https://deepwiki.com/cheelim1/go-temporal-sre/1-overview)
- [SRE Best Practices](https://sre.google/)
