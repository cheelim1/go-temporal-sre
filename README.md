# Go Temporal SRE

[![Go Report Card](https://goreportcard.com/badge/github.com/cheelim1/go-temporal-sre)](https://goreportcard.com/report/github.com/cheelim1/go-temporal-sre)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A collection of Site Reliability Engineering (SRE) patterns and practices implemented using Temporal.io's workflow orchestration platform. This project demonstrates how Temporal can be used to solve common SRE challenges and build reliable, scalable systems.

## 🚀 Features

- **Idempotent Script Execution**: Make non-idempotent scripts reliable and idempotent
- **Batch Processing**: Handle large-scale batch operations with reliability
- **Data Enrichment**: Orchestrate complex data enrichment workflows
- **Cron Job Management**: Schedule and manage cron jobs with Temporal
- **PostgreSQL Sharding**: Implement sharding patterns with workflow orchestration

## 🎯 Use Cases

This project showcases how Temporal can be used to solve real-world SRE challenges:

1. **Making Scripts Reliable**
   - Convert non-idempotent scripts into reliable workflows
   - Handle retries and failures gracefully
   - Ensure exactly-once execution

2. **Batch Processing**
   - Process large datasets reliably
   - Handle partial failures
   - Implement idempotent operations

3. **Data Enrichment**
   - Orchestrate complex enrichment workflows
   - Handle external API calls
   - Manage data consistency

4. **Cron Job Management**
   - Schedule and monitor jobs
   - Handle job failures
   - Ensure job completion

## 🛠️ Getting Started

### Prerequisites

- Go 1.24.1 or later
- Temporal server running locally
- PostgreSQL (for some features)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/cheelim1/go-temporal-sre.git
   cd go-temporal-sre
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Start Temporal server:
   ```bash
   make start-temporal
   ```

### Running Demos

The project includes several demo applications that showcase different use cases:

```bash
# Run the basic demo
make run-basic-demo

# Run the batch processing demo
make run-batch-demo

# Run the data enrichment demo
make run-enrichment-demo
```

## 📚 Documentation

- [Architecture Overview](docs/architecture/README.md)
- [Feature Documentation](docs/features/README.md)
- [API Reference](docs/api/README.md)

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details on:
- Project structure
- Code style
- Testing requirements
- Pull request process
