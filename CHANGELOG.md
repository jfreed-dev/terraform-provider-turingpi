# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.1.3] - 2025-12-29

### Added
- **New Resource**
  - `turingpi_k3s_cluster` - Deploy K3s Kubernetes clusters on pre-flashed Turing Pi nodes
    - K3s server installation on control plane via SSH
    - K3s agent installation on worker nodes
    - MetalLB load balancer deployment with configurable IP range
    - NGINX Ingress controller deployment
    - Kubeconfig output to file and Terraform state
    - Cluster token generation and management
    - Cluster uninstall on destroy

- **Infrastructure (v1.1.2)**
  - `provider/ssh_client.go` - SSH client interface with key-based and password authentication
  - `provider/cluster_helpers.go` - WaitForSSH, RunSSHCommand utilities
  - `provider/kubeconfig.go` - LoadKubeconfig, WaitForKubeAPI, ExtractClusterEndpoint
  - `provider/helm_client.go` - Helm client using mittwald/go-helm-client for addon deployment
  - `provider/k3s_provisioner.go` - K3s installation logic via SSH

- **Documentation**
  - `docs/resources/k3s_cluster.md` - K3s cluster resource documentation
  - `examples/k3s-cluster/` - Example configuration with MetalLB and Ingress

### Security
- Updated `github.com/containerd/containerd` from v1.7.28 to v1.7.29
  - Fixed local privilege escalation via wide permissions on CRI directory (high)
  - Fixed host memory exhaustion through Attach goroutine leak (medium)

## [1.1.1] - 2025-12-29

### Added
- **New Resources**
  - `turingpi_usb_boot` - Enable USB boot mode for nodes (pulls nRPIBOOT pin low for CM4s)
  - `turingpi_node_to_msd` - Reboot node into USB Mass Storage Device mode
  - `turingpi_clear_usb_boot` - Clear USB boot status for nodes
  - `turingpi_bmc_reload` - Restart BMC daemon (bmcd) with readiness monitoring

- **New Data Sources**
  - `turingpi_sdcard` - MicroSD card info (total/used/free bytes, GB values, usage percentage)
  - `turingpi_about` - BMC version info (API, daemon, buildroot, firmware, build time)

- **Documentation**
  - `docs/FUTURE_MODULES.md` - Comprehensive roadmap for K3s and Talos cluster modules
  - `TODO.md` - Implementation milestones (v1.1.2 - v1.1.5)

### Planned
- v1.1.4: `turingpi_talos_cluster` resource with Talos Image Factory integration
- v1.1.5: K3s enhancements (NPU support, Longhorn storage, cluster upgrades)

## [1.1.0] - 2025-12-29

### Added
- **New Data Sources**
  - `turingpi_info` - BMC version, network interfaces, storage devices, and node power status
  - `turingpi_power` - Current power status of all nodes with aggregated counts
  - `turingpi_usb` - Current USB routing configuration
  - `turingpi_uart` - Read buffered UART output from nodes (clears buffer on read)

- **New Resources**
  - `turingpi_usb` - Configure USB routing between nodes and USB-A/BMC
  - `turingpi_network_reset` - Trigger network switch reset
  - `turingpi_bmc_firmware` - Upgrade BMC firmware (upload or local file)
  - `turingpi_uart` - Write commands to node UART (serial console)
  - `turingpi_bmc_reboot` - Trigger BMC reboot with readiness monitoring

- **Enhanced Resources**
  - `turingpi_power` - Added `reset` state for node reboot, added `current_state` computed attribute

### Changed
- All new resources use Context-aware CRUD functions (CreateContext, etc.)
- Added input validation (ValidateDiagFunc) to all new resources
- Comprehensive unit tests for all new resources and data sources

## [1.0.10] - 2025-12-24

### Fixed
- Fix golangci-lint v2.7.2 errcheck violations (resp.Body.Close, os.Setenv/Unsetenv)

## [1.0.9] - 2025-12-24

### Added
- `make release VERSION=x.y.z` - automated release workflow
- `make release-prep VERSION=x.y.z` - update version in docs/examples only

### Changed
- Update all documentation and examples to v1.0.9

## [1.0.8] - 2025-12-24

### Changed
- Add security workflow badge to README
- Update all documentation examples to v1.0.7
- Update examples/basic, examples/flash-firmware, examples/full-provisioning to v1.0.7

## [1.0.7] - 2025-12-24

### Changed
- Bump `terraform-plugin-sdk/v2` from 2.35.0 to 2.38.1
- Bump `actions/checkout` from 4.3.1 to 6.0.1
- Bump `actions/setup-go` from 5.6.0 to 6.1.0
- Bump `golangci/golangci-lint-action` from 6.5.2 to 9.2.0
- Bump `github/codeql-action` from 3.28.0 to 4.31.9
- Bump `actions/dependency-review-action` from 4.5.0 to 4.8.2

## [1.0.6] - 2025-12-24

### Security
- Pin all GitHub Actions to SHA commits (supply chain protection)
- Add Dependabot for automated security updates (Go modules + Actions)
- Add gosec security scanner with SARIF reporting
- Add dependency-review-action for PR vulnerability scanning
- Enable branch protection (signed commits, required reviews, status checks)

### Added
- `.github/CODEOWNERS` for mandatory code review
- `.github/dependabot.yml` for automated dependency updates
- `.github/workflows/security.yml` for security scanning
- GPG signature verification documentation in SECURITY.md

### Changed
- Workflows now use `go-version-file: go.mod` instead of hardcoded version
- Enhanced SECURITY.md with release verification instructions

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

[Unreleased]: https://github.com/jfreed-dev/terraform-provider-turingpi/compare/v1.1.3...HEAD
[1.1.3]: https://github.com/jfreed-dev/terraform-provider-turingpi/compare/v1.1.1...v1.1.3
[1.1.1]: https://github.com/jfreed-dev/terraform-provider-turingpi/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/jfreed-dev/terraform-provider-turingpi/compare/v1.0.10...v1.1.0
[1.0.10]: https://github.com/jfreed-dev/terraform-provider-turingpi/compare/v1.0.9...v1.0.10
[1.0.9]: https://github.com/jfreed-dev/terraform-provider-turingpi/compare/v1.0.8...v1.0.9
[1.0.8]: https://github.com/jfreed-dev/terraform-provider-turingpi/compare/v1.0.7...v1.0.8
[1.0.7]: https://github.com/jfreed-dev/terraform-provider-turingpi/compare/v1.0.6...v1.0.7
[1.0.6]: https://github.com/jfreed-dev/terraform-provider-turingpi/compare/v1.0.5...v1.0.6
[1.0.5]: https://github.com/jfreed-dev/terraform-provider-turingpi/compare/v1.0.4...v1.0.5
[1.0.4]: https://github.com/jfreed-dev/terraform-provider-turingpi/compare/v1.0.3...v1.0.4
[1.0.3]: https://github.com/jfreed-dev/terraform-provider-turingpi/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/jfreed-dev/terraform-provider-turingpi/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/jfreed-dev/terraform-provider-turingpi/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/jfreed-dev/terraform-provider-turingpi/releases/tag/v1.0.0
