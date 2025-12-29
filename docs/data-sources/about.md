---
page_title: "turingpi_about Data Source - Turing Pi"
subcategory: ""
description: |-
  Retrieves detailed version and build information about the Turing Pi BMC daemon.
---

# turingpi_about (Data Source)

Retrieves detailed version and build information about the Turing Pi BMC daemon (bmcd), including API version, daemon version, buildroot version, firmware version, and build timestamp.

This data source is useful for:
- Displaying BMC version information
- Making decisions based on firmware capabilities
- Debugging and troubleshooting
- Documentation and inventory management

## Example Usage

### Basic Usage

```hcl
data "turingpi_about" "bmc" {}

output "bmc_versions" {
  value = {
    api       = data.turingpi_about.bmc.api_version
    daemon    = data.turingpi_about.bmc.daemon_version
    firmware  = data.turingpi_about.bmc.firmware_version
    buildroot = data.turingpi_about.bmc.buildroot_version
    built     = data.turingpi_about.bmc.build_time
  }
}
```

### Version-Based Feature Detection

```hcl
data "turingpi_about" "bmc" {}

locals {
  # Parse major version number
  api_major = tonumber(split(".", data.turingpi_about.bmc.api_version)[0])

  # API v2+ supports authentication
  requires_auth = local.api_major >= 2
}

output "api_info" {
  value = {
    version       = data.turingpi_about.bmc.api_version
    requires_auth = local.requires_auth
  }
}
```

### Display Build Information

```hcl
data "turingpi_about" "bmc" {}

output "bmc_build_info" {
  value = format(
    "BMC Firmware v%s (API %s) - Built %s with Buildroot %s",
    data.turingpi_about.bmc.firmware_version,
    data.turingpi_about.bmc.api_version,
    data.turingpi_about.bmc.build_time,
    data.turingpi_about.bmc.buildroot_version
  )
}
```

### Conditional Resource Based on Version

```hcl
data "turingpi_about" "bmc" {}

locals {
  # Some features only available in newer firmware
  supports_uart_write = tonumber(
    split(".", data.turingpi_about.bmc.daemon_version)[0]
  ) >= 2
}

resource "turingpi_uart" "command" {
  count = local.supports_uart_write ? 1 : 0

  node    = 1
  command = "echo 'Hello from Terraform'"
}
```

### Comparison: turingpi_about vs turingpi_info

```hcl
# turingpi_about - Focused on version information
data "turingpi_about" "versions" {}

# turingpi_info - Comprehensive BMC information
data "turingpi_info" "full" {}

output "comparison" {
  value = {
    # Both provide version info
    about_firmware = data.turingpi_about.versions.firmware_version
    info_firmware  = data.turingpi_info.full.firmware_version

    # Only turingpi_info provides these
    network_interfaces = data.turingpi_info.full.network_interfaces
    storage_devices    = data.turingpi_info.full.storage_devices
    node_power_status  = data.turingpi_info.full.nodes
  }
}
```

### Inventory Documentation

```hcl
data "turingpi_about" "bmc" {}

output "inventory_record" {
  value = {
    device_type      = "Turing Pi 2.5 BMC"
    api_version      = data.turingpi_about.bmc.api_version
    daemon_version   = data.turingpi_about.bmc.daemon_version
    firmware_version = data.turingpi_about.bmc.firmware_version
    buildroot        = data.turingpi_about.bmc.buildroot_version
    build_timestamp  = data.turingpi_about.bmc.build_time
    queried_at       = timestamp()
  }
}
```

## Attribute Reference

- `api_version` - (String) BMC API version (e.g., "1.0.0", "2.0.3").
- `daemon_version` - (String) BMC daemon (bmcd) version.
- `buildroot_version` - (String) Buildroot version used to build the BMC firmware.
- `firmware_version` - (String) BMC firmware version.
- `build_time` - (String) Timestamp when the BMC firmware was built.

## Notes

1. **Lightweight Query**: This data source only queries version information, making it faster than `turingpi_info` which queries multiple endpoints.

2. **Use turingpi_info for More Data**: If you also need network, storage, or power status information, use `turingpi_info` instead to reduce API calls.

3. **Version Format**: Version strings follow semantic versioning in most cases, but the exact format may vary between firmware releases.

4. **Build Time Format**: The build time format depends on the firmware build system and may vary.

## API Endpoint Used

| Endpoint | Purpose |
|----------|---------|
| `GET /api/bmc?opt=get&type=about` | Get daemon version information |
