# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Centralized Temporal worker architecture
- Consolidated worker configuration and registration system
- Restructured project layout for better organization and contribution
- CHANGELOG.md for tracking project changes
- Improved documentation structure

### Changed
- Migrated individual workers to centralized worker pattern
- Reorganized features from cmd/ to internal/features/
- Updated demo applications to use centralized worker
- Restructured project directories to follow standard Go layout

### Removed
- Individual worker implementations in cmd/kilcron/worker.go
- Individual worker implementations in cmd/superscript/worker.go
- Scattered worker configurations

## [0.1.0] - Initial Release

### Added
- Initial project structure with multiple Temporal demos
- Kilcron demo for scheduled payment collection
- Superscript demo for script execution orchestration
- JIT (Just-In-Time) access demo
- Batch processing demo
- Data enrichment demo
- Basic project documentation 