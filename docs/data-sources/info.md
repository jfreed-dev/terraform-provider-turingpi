---
page_title: "turingpi_info Data Source - Turing Pi"
subcategory: ""
description: |-
  Retrieves information about the Turing Pi BMC including version info, network configuration, storage metrics, and node power status.
---

# turingpi_info (Data Source)

Retrieves comprehensive information about the Turing Pi BMC (Baseboard Management Controller), including:

- BMC version information (API, daemon, firmware, buildroot versions)
- Network interface configuration (devices, IPs, MAC addresses)
- Storage metrics (BMC internal storage and microSD card)
- Current power status of all nodes (1-4)

This data source is useful for:
- Displaying cluster information in Terraform outputs
- Making decisions based on firmware versions
- Monitoring storage capacity
- Checking node power status before operations

## Example Usage

### Basic Usage

```hcl
data "turingpi_info" "bmc" {}

output "bmc_firmware_version" {
  value = data.turingpi_info.bmc.firmware_version
}

output "node_power_status" {
  value = data.turingpi_info.bmc.nodes
}
```

### Conditional Logic Based on Firmware Version

```hcl
data "turingpi_info" "bmc" {}

locals {
  # Check if firmware supports a specific feature
  supports_new_api = tonumber(split(".", data.turingpi_info.bmc.api_version)[0]) >= 2
}

resource "turingpi_node" "node1" {
  count = local.supports_new_api ? 1 : 0

  node        = 1
  power_state = "on"
}
```

### Display Cluster Information

```hcl
data "turingpi_info" "bmc" {}

output "cluster_info" {
  value = {
    firmware     = data.turingpi_info.bmc.firmware_version
    api_version  = data.turingpi_info.bmc.api_version
    build_time   = data.turingpi_info.bmc.build_time

    network = [for iface in data.turingpi_info.bmc.network_interfaces : {
      device = iface.device
      ip     = iface.ip
    }]

    storage_free_gb = [for dev in data.turingpi_info.bmc.storage_devices : {
      name    = dev.name
      free_gb = dev.free_bytes / 1073741824
    }]

    nodes_powered_on = [for name, powered in data.turingpi_info.bmc.nodes : name if powered]
  }
}
```

### Check Node Status Before Operations

```hcl
data "turingpi_info" "bmc" {}

locals {
  powered_nodes = [for name, powered in data.turingpi_info.bmc.nodes : name if powered]
}

output "active_nodes" {
  value = local.powered_nodes
}
```

## Attribute Reference

### Version Information

- `api_version` - (String) The BMC API version (e.g., "1.0").
- `daemon_version` - (String) The BMC daemon version (e.g., "2.0.5").
- `buildroot_version` - (String) The Buildroot version used to build the BMC firmware.
- `firmware_version` - (String) The BMC firmware version.
- `build_time` - (String) The timestamp when the BMC firmware was built (RFC 3339 format).

### Network Configuration

- `network_interfaces` - (List of Objects) List of network interfaces on the BMC.
  - `device` - (String) Network interface device name (e.g., "eth0").
  - `ip` - (String) IP address assigned to the interface.
  - `mac` - (String) MAC address of the interface.

### Storage Information

- `storage_devices` - (List of Objects) List of storage devices.
  - `name` - (String) Storage device name (e.g., "bmc", "microSD").
  - `total_bytes` - (Integer) Total storage capacity in bytes.
  - `used_bytes` - (Integer) Used storage in bytes.
  - `free_bytes` - (Integer) Available storage in bytes.

### Node Power Status

- `nodes` - (Map of Boolean) Power status of each node. Keys are node names (e.g., "node1", "node2", "node3", "node4"), values are `true` if powered on, `false` if powered off.

## API Endpoints Used

This data source queries the following BMC API endpoints:

| Endpoint | Purpose |
|----------|---------|
| `/api/bmc?opt=get&type=about` | Version information |
| `/api/bmc?opt=get&type=info` | Network and storage info |
| `/api/bmc?opt=get&type=power` | Node power status |
