# Terraform Provider for Turing Pi 2.5

[![Go](https://github.com/jfreed-dev/terraform-provider-turingpi/actions/workflows/go.yml/badge.svg)](https://github.com/jfreed-dev/terraform-provider-turingpi/actions/workflows/go.yml)
[![Release](https://img.shields.io/github/v/release/jfreed-dev/terraform-provider-turingpi)](https://github.com/jfreed-dev/terraform-provider-turingpi/releases/latest)
[![License](https://img.shields.io/github/license/jfreed-dev/terraform-provider-turingpi)](LICENSE)

A Terraform provider for managing Turing Pi's Baseboard Management Controller (BMC), enabling power management, firmware flashing, and node provisioning.

## Prerequisites

- [Go 1.23+](https://go.dev/)
- [Terraform](https://www.terraform.io/)

## Building & Installation

```bash
# Clone and build
git clone https://github.com/jfreed-dev/terraform-provider-turingpi.git
cd terraform-provider-turingpi
go build -o terraform-provider-turingpi

# Install (Linux/macOS)
mkdir -p ~/.terraform.d/plugins/local/turingpi/turingpi/1.0.1/linux_amd64/
mv terraform-provider-turingpi ~/.terraform.d/plugins/local/turingpi/turingpi/1.0.1/linux_amd64/

# Install (Windows)
mkdir %APPDATA%\terraform.d\plugins\local\turingpi\turingpi\1.0.1\windows_amd64\
move terraform-provider-turingpi.exe %APPDATA%\terraform.d\plugins\local\turingpi\turingpi\1.0.1\windows_amd64\
```

## Provider Configuration

```hcl
terraform {
  required_providers {
    turingpi = {
      source  = "local/turingpi/turingpi"
      version = "1.0.1"
    }
  }
}

provider "turingpi" {
  username = "root"                      # or TURINGPI_USERNAME env var
  password = "turing"                    # or TURINGPI_PASSWORD env var
  endpoint = "https://turingpi.local"    # or TURINGPI_ENDPOINT env var (optional)
}
```

Using environment variables:

```bash
export TURINGPI_USERNAME=root
export TURINGPI_PASSWORD=turing
export TURINGPI_ENDPOINT=https://192.168.1.100  # optional
```

```hcl
provider "turingpi" {}
```

## Resources

### turingpi_power

Control node power state.

```hcl
resource "turingpi_power" "node1" {
  node  = 1       # Node ID (1-4)
  state = true    # true = on, false = off
}
```

### turingpi_flash

Flash firmware to a node. Changes to `node` or `firmware_file` trigger resource recreation.

```hcl
resource "turingpi_flash" "node1" {
  node          = 1
  firmware_file = "/path/to/firmware.img"
}
```

### turingpi_node

Comprehensive node management: power control, firmware flashing, and boot verification.

```hcl
resource "turingpi_node" "node1" {
  node                 = 1                        # Node ID (1-4)
  power_state          = "on"                     # "on" or "off" (default: "on")
  firmware_file        = "/path/to/firmware.img"  # optional
  boot_check           = true                     # Monitor UART for login prompt (default: false)
  login_prompt_timeout = 120                      # Timeout in seconds (default: 60)
}
```

## Examples

See the [examples](./examples) directory for complete configurations:

- [basic](./examples/basic) - Simple power control
- [flash-firmware](./examples/flash-firmware) - Firmware flashing
- [full-provisioning](./examples/full-provisioning) - Complete node management with boot verification

## Development

```bash
# Run tests
go test -v ./...

# Build
go build -o terraform-provider-turingpi

# Enable debug logging
export TF_LOG=DEBUG
terraform apply
```

## License

See [LICENSE](LICENSE) file.
