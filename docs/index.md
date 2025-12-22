---
page_title: "Turing Pi Provider"
subcategory: ""
description: |-
  Terraform provider for managing Turing Pi 2.5 BMC (Baseboard Management Controller).
---

# Turing Pi Provider

The Turing Pi provider enables Terraform-based management of [Turing Pi 2.5](https://turingpi.com/) clusters through the BMC (Baseboard Management Controller) API.

## Features

- **Power Management** - Control power state of individual compute nodes
- **Firmware Flashing** - Flash firmware images to nodes
- **Boot Verification** - Monitor UART output to verify successful boot
- **Node Provisioning** - Combined resource for complete node management

## Example Usage

```hcl
terraform {
  required_providers {
    turingpi = {
      source  = "jfreed-dev/turingpi"
      version = "1.0.5"
    }
  }
}

provider "turingpi" {
  username = "root"
  password = "turing"
  endpoint = "https://turingpi.local"
}

resource "turingpi_power" "node1" {
  node  = 1
  state = true
}
```

## Authentication

The provider requires BMC credentials to authenticate with the Turing Pi board.

### Configuration Options

- `username` - (Required) BMC username. Can also be set via `TURINGPI_USERNAME` environment variable.
- `password` - (Required) BMC password. Can also be set via `TURINGPI_PASSWORD` environment variable.
- `endpoint` - (Optional) BMC API endpoint URL. Defaults to `https://turingpi.local`. Can also be set via `TURINGPI_ENDPOINT` environment variable.
- `insecure` - (Optional) Skip TLS certificate verification. Useful for self-signed or expired certificates. Defaults to `false`. Can also be set via `TURINGPI_INSECURE` environment variable.

### Using Environment Variables

```bash
export TURINGPI_USERNAME=root
export TURINGPI_PASSWORD=turing
export TURINGPI_ENDPOINT=https://192.168.1.100
export TURINGPI_INSECURE=true  # optional, for self-signed/expired certs
```

```hcl
provider "turingpi" {}
```

## Resources

- [turingpi_power](resources/power.md) - Control node power state
- [turingpi_flash](resources/flash.md) - Flash firmware to a node
- [turingpi_node](resources/node.md) - Comprehensive node management
