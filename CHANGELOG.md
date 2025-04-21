# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- Initial restructuring plan and documentation
- New standardized directory structure
- Centralized Temporal worker implementation
  - Worker configuration package (`internal/worker/config`)
  - Main worker implementation (`internal/worker/worker.go`)
- Basic feature implementation
  - Example workflow (`internal/features/basic/workflow.go`)
  - Example activity (`internal/features/basic/activity.go`)
- Basic demo application (`demos/basic/main.go`)
- Documentation
  - CONTRIBUTING.md with project structure and guidelines
  - CHANGELOG.md for tracking changes
- Centralized worker implementation in `internal/worker/worker.go`
- Standardized worker configuration in `internal/worker/config/config.go`
- New worker registration pattern for features

### Changed
- Moved worker implementations to shared location
  - Removed duplicate workers from:
    - `cmd/superscript/worker.go`
    - `cmd/kilcron/worker.go`
    - `demo/jit/demo-be/cmd/worker/main.go`
  - Updated data enrichment feature to use centralized worker
- Reorganized demo structure
- Standardized configuration management
- Improved code organization and reusability
- Removed duplicate worker implementations from:
  - `internal/features/data-enrichment/worker.go`
  - `internal/batch/temporal_idempotency_test.go`
  - `demos/basic/main.go`
- Standardized worker lifecycle management across features

### Removed
- Duplicate worker implementations
- Redundant configuration files
- Outdated documentation

### Fixed
- Fixed worker initialization and cleanup in feature implementations
- Improved error handling in worker startup and shutdown
- Standardized task queue and activity configuration

## Next Steps
1. Migrate remaining features to new structure
   - Batch processing
   - Script execution
2. Update documentation
   - Architecture diagrams
   - API documentation
   - Usage examples
3. Add comprehensive tests
   - Unit tests
   - Integration tests
   - E2E tests
4. Create additional demos
   - Batch processing demo
   - Script execution demo 