# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.5] - 2025-12-22

### Added
- `boot_check_pattern` option for turingpi_node resource
- Configurable pattern matching for boot verification (default: `"login:"`)
- Support for Talos Linux boot detection (`"machine is running and ready"`)

## [1.0.4] - 2025-12-22

### Added
- `insecure` provider option to skip TLS certificate verification
- Useful for self-signed or expired BMC certificates
- Environment variable support via `TURINGPI_INSECURE`

### Changed
- Shared HTTP client for all API requests with configurable TLS settings

## [1.0.3] - 2025-12-22

### Added
- Terraform Registry documentation (docs/index.md, docs/resources/)
- Provider overview and authentication docs
- Resource documentation for turingpi_power, turingpi_flash, turingpi_node

## [1.0.2] - 2025-12-22

### Changed
- Provider source updated to `jfreed-dev/turingpi` for Terraform Registry
- Simplified installation instructions (auto-download from registry)
- Consolidated GoReleaser config with GPG signing

### Added
- Terraform Registry badge to README
- Published to public Terraform Registry

## [1.0.1] - 2025-12-22

### Added
- Example Terraform configurations (basic, flash-firmware, full-provisioning)
- Build, release, and license badges to README
- Terraform Registry manifest for registry publishing
- GPG signing support for releases

### Fixed
- All golangci-lint issues (unchecked errors, deprecated APIs)
- GoReleaser config for unsigned releases

## [1.0.0] - 2025-12-22

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

[Unreleased]: https://github.com/jfreed-dev/terraform-provider-turingpi/compare/v1.0.5...HEAD
[1.0.5]: https://github.com/jfreed-dev/terraform-provider-turingpi/compare/v1.0.4...v1.0.5
[1.0.4]: https://github.com/jfreed-dev/terraform-provider-turingpi/compare/v1.0.3...v1.0.4
[1.0.3]: https://github.com/jfreed-dev/terraform-provider-turingpi/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/jfreed-dev/terraform-provider-turingpi/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/jfreed-dev/terraform-provider-turingpi/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/jfreed-dev/terraform-provider-turingpi/releases/tag/v1.0.0
