# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- CONTRIBUTING.md with contribution guidelines
- SECURITY.md with security policy
- CODE_OF_CONDUCT.md (Contributor Covenant)
- Issue and PR templates
- Makefile for build automation
- Release automation workflow

## [1.0.0] - 2024-12-21

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

[Unreleased]: https://github.com/jfreed-dev/terraform-provider-turingpi/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/jfreed-dev/terraform-provider-turingpi/releases/tag/v1.0.0
