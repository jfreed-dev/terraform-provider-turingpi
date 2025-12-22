# Terraform Provider for Turing Pi 2.5

[![Terraform Registry](https://img.shields.io/badge/terraform-registry-blueviolet)](https://registry.terraform.io/providers/jfreed-dev/turingpi)
[![Go](https://github.com/jfreed-dev/terraform-provider-turingpi/actions/workflows/go.yml/badge.svg)](https://github.com/jfreed-dev/terraform-provider-turingpi/actions/workflows/go.yml)
[![Release](https://img.shields.io/github/v/release/jfreed-dev/terraform-provider-turingpi)](https://github.com/jfreed-dev/terraform-provider-turingpi/releases/latest)
[![License](https://img.shields.io/github/license/jfreed-dev/terraform-provider-turingpi)](LICENSE)

A Terraform provider for managing Turing Pi's Baseboard Management Controller (BMC), enabling power management, firmware flashing, and node provisioning.

## Installation

The provider is available on the [Terraform Registry](https://registry.terraform.io/providers/jfreed-dev/turingpi). Terraform will automatically download it when you run `terraform init`.

## Usage

```hcl
terraform {
  required_providers {
    turingpi = {
      source  = "jfreed-dev/turingpi"
      version = "1.0.4"
    }
  }
}

provider "turingpi" {
  username = "root"                      # or TURINGPI_USERNAME env var
  password = "turing"                    # or TURINGPI_PASSWORD env var
  endpoint = "https://turingpi.local"    # or TURINGPI_ENDPOINT env var (optional)
  insecure = false                       # or TURINGPI_INSECURE env var (optional)
}
```

Using environment variables:

```bash
export TURINGPI_USERNAME=root
export TURINGPI_PASSWORD=turing
export TURINGPI_ENDPOINT=https://192.168.1.100  # optional
export TURINGPI_INSECURE=true                   # optional, for self-signed/expired certs
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

Requires [Go 1.23+](https://go.dev/).

```bash
# Clone and build
git clone https://github.com/jfreed-dev/terraform-provider-turingpi.git
cd terraform-provider-turingpi
go build -o terraform-provider-turingpi

# Run tests
go test -v ./...

# Enable debug logging
export TF_LOG=DEBUG
terraform apply
```

## License

See [LICENSE](LICENSE) file.
