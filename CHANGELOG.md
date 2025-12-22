# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Example Terraform configurations (basic, flash-firmware, full-provisioning)
- Build, release, and license badges to README

### Fixed
- All golangci-lint issues (unchecked errors, deprecated APIs)
- GoReleaser config for unsigned releases

## [1.0.0] - 2024-12-22

### Added
- Initial release
- `turingpi_power` resource for node power control
- `turingpi_flash` resource for firmware flashing
- `turingpi_node` resource for comprehensive node management
- BMC authentication with username/password
- Environment variable support for credentials
- Configurable endpoint URL
- UART monitoring for boot verification
- Comprehensive test suite
- CONTRIBUTING.md with contribution guidelines
- SECURITY.md with security policy
- CODE_OF_CONDUCT.md (Contributor Covenant)
- GitHub issue and PR templates
- Makefile for build automation
- Release automation workflow with GoReleaser
- Multi-platform binaries (linux/darwin/windows, amd64/arm64)

[Unreleased]: https://github.com/jfreed-dev/terraform-provider-turingpi/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/jfreed-dev/terraform-provider-turingpi/releases/tag/v1.0.0
