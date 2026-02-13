# AGENTS

## Language

- Use latest Go v1.25.x and its full capabilities
- Use slog structured logging for observability
- Use stdlib as much as possible
- Use first class anonymous function as method replacement technique when testing

## Orchestration

- Use Temporal + Go SDK to handle the workflows
- Use Temporal testsuite as much as possibel to have widest test coverage
- Use advanced techique like registerdelayedcallback to simulate long running system
- Ensure use Temporal Worker version to ensure multiple workflow versions can co-exists

## Runtime

- Whole system should be fully testable standalone with Go binary + temporal testsuite
- Use modern Go capabilities (when needed): generics, structured log, built-in http routing, synctest
- Use techniques like first class anonymous function as method replacement, synctest to ensure all things are deterministic

## Testing

- All new cases should be at least 80% coverage
- Unit tests and inetgration tests MUST be comepleted without needing to spin up any external dependencies
- Only Full End-to-End should need to stand up Temporal Test Server

## Tools

- Use mise to run tasks, set env variables, automate
- Tools available: ripgrep, overmind, air, goreleaser, watchexec 
- Use overmind to start temporal-cli + air

## Specification
<todo>
