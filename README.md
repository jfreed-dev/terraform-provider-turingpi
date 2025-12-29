# Terraform Provider for Turing Pi 2.5

[![Terraform Registry](https://img.shields.io/badge/terraform-registry-blueviolet)](https://registry.terraform.io/providers/jfreed-dev/turingpi)
[![Go](https://github.com/jfreed-dev/terraform-provider-turingpi/actions/workflows/go.yml/badge.svg)](https://github.com/jfreed-dev/terraform-provider-turingpi/actions/workflows/go.yml)
[![Security](https://github.com/jfreed-dev/terraform-provider-turingpi/actions/workflows/security.yml/badge.svg)](https://github.com/jfreed-dev/terraform-provider-turingpi/actions/workflows/security.yml)
[![Release](https://img.shields.io/github/v/release/jfreed-dev/terraform-provider-turingpi)](https://github.com/jfreed-dev/terraform-provider-turingpi/releases/latest)
[![License](https://img.shields.io/github/license/jfreed-dev/terraform-provider-turingpi)](LICENSE)

A Terraform provider for managing Turing Pi's Baseboard Management Controller (BMC), enabling power management, firmware flashing, and node provisioning.

## Features

- **Power Management** - Control power state of individual compute nodes (1-4)
- **Firmware Flashing** - Flash firmware images to nodes with automatic resource recreation
- **BMC Firmware Upgrade** - Upgrade BMC firmware with file upload or local file support
- **BMC Reboot** - Trigger BMC reboot with readiness monitoring
- **UART Access** - Read and write to node serial consoles for boot monitoring and command execution
- **Boot Verification** - Monitor UART output with configurable patterns to verify successful boot
- **USB Routing** - Configure USB routing between nodes and USB-A connector or BMC
- **Network Reset** - Trigger network switch reset for recovery after configuration changes
- **Talos Linux Support** - Built-in boot detection for Talos Linux clusters
- **TLS Flexibility** - Skip certificate verification for self-signed or expired BMC certificates
- **Environment Variables** - Configure provider via environment variables for CI/CD pipelines

## Documentation

- **[Architecture](docs/ARCHITECTURE.md)** - System diagrams, data flows, and component interactions
- **[Terraform Registry](https://registry.terraform.io/providers/jfreed-dev/turingpi)** - Provider documentation

## Installation

The provider is available on the [Terraform Registry](https://registry.terraform.io/providers/jfreed-dev/turingpi). Terraform will automatically download it when you run `terraform init`.

## Usage

```hcl
terraform {
  required_providers {
    turingpi = {
      source  = "jfreed-dev/turingpi"
      version = "1.1.0"
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

## Data Sources

### turingpi_info

Retrieve BMC information including version, network, storage, and node power status.

```hcl
data "turingpi_info" "bmc" {}

output "firmware_version" {
  value = data.turingpi_info.bmc.firmware_version
}

output "node_power_status" {
  value = data.turingpi_info.bmc.nodes
}
```

### turingpi_usb

Retrieve current USB routing configuration.

```hcl
data "turingpi_usb" "current" {}

output "usb_config" {
  value = {
    mode  = data.turingpi_usb.current.mode   # "host" or "device"
    node  = data.turingpi_usb.current.node   # 1-4
    route = data.turingpi_usb.current.route  # "usb-a" or "bmc"
  }
}
```

### turingpi_power

Retrieve current power status of all nodes.

```hcl
data "turingpi_power" "status" {}

output "power_status" {
  value = {
    node1      = data.turingpi_power.status.node1  # true/false
    nodes_on   = data.turingpi_power.status.powered_on_count
    nodes_off  = data.turingpi_power.status.powered_off_count
  }
}
```

### turingpi_uart

Read buffered UART (serial console) output from a node. Reading clears the buffer.

```hcl
data "turingpi_uart" "node1" {
  node = 1
}

output "node1_output" {
  value = data.turingpi_uart.node1.output
}
```

## Resources

### turingpi_power

Control node power state with on, off, and reset support.

```hcl
resource "turingpi_power" "node1" {
  node  = 1       # Node ID (1-4)
  state = "on"    # "on", "off", or "reset"
}

# Reset (reboot) a node
resource "turingpi_power" "node2_reset" {
  node  = 2
  state = "reset"
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

### turingpi_usb

Manage USB routing between nodes and the USB-A connector or BMC.

```hcl
resource "turingpi_usb" "node1" {
  node  = 1           # Node ID (1-4)
  mode  = "host"      # "host" or "device"
  route = "usb-a"     # "usb-a" or "bmc" (default: "usb-a")
}
```

### turingpi_network_reset

Trigger a network switch reset. Useful for recovering network connectivity after node power changes or firmware updates.

```hcl
resource "turingpi_network_reset" "switch" {
  triggers = {
    # Reset network when node power changes
    node_states = join(",", [for k, v in turingpi_power.nodes : "${k}:${v.current_state}"])
  }
}
```

### turingpi_bmc_firmware

Upgrade the BMC firmware. Supports uploading from Terraform host or using a file on the BMC filesystem.

```hcl
resource "turingpi_bmc_firmware" "upgrade" {
  firmware_file = "/path/to/bmc-firmware.swu"
  timeout       = 300
}

# Or use a file already on the BMC
resource "turingpi_bmc_firmware" "upgrade_local" {
  firmware_file = "/tmp/firmware.swu"
  bmc_local     = true
}
```

### turingpi_uart

Write commands to a node's UART (serial console).

```hcl
resource "turingpi_uart" "node1_cmd" {
  node    = 1
  command = "echo 'Hello from Terraform'\n"
}
```

### turingpi_bmc_reboot

Trigger a BMC reboot with optional readiness wait.

```hcl
resource "turingpi_bmc_reboot" "maintenance" {
  wait_for_ready = true
  ready_timeout  = 120
}
```

### turingpi_node

Comprehensive node management: power control, firmware flashing, and boot verification.

```hcl
resource "turingpi_node" "node1" {
  node                 = 1                              # Node ID (1-4)
  power_state          = "on"                           # "on" or "off" (default: "on")
  firmware_file        = "/path/to/firmware.img"        # optional
  boot_check           = true                           # Monitor UART for boot pattern (default: false)
  boot_check_pattern   = "login:"                       # Pattern to detect (default: "login:")
  login_prompt_timeout = 120                            # Timeout in seconds (default: 60)
}

# For Talos Linux, use the appropriate boot pattern:
resource "turingpi_node" "talos" {
  node               = 2
  boot_check         = true
  boot_check_pattern = "machine is running and ready"
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
