---
page_title: "turingpi_usb Data Source - Turing Pi"
subcategory: ""
description: |-
  Retrieves the current USB routing configuration from the Turing Pi BMC.
---

# turingpi_usb (Data Source)

Retrieves the current USB routing configuration from the Turing Pi BMC, including:

- Current USB mode (host or device)
- Which node USB is currently routed to
- USB routing destination (USB-A connector or BMC)

This data source is useful for:
- Checking current USB configuration before making changes
- Conditional logic based on USB routing
- Displaying USB status in Terraform outputs

## Example Usage

### Basic Usage

```hcl
data "turingpi_usb" "current" {}

output "usb_status" {
  value = {
    mode  = data.turingpi_usb.current.mode
    node  = data.turingpi_usb.current.node
    route = data.turingpi_usb.current.route
  }
}
```

### Conditional Resource Creation

Only create a resource if USB is routed to a specific node:

```hcl
data "turingpi_usb" "current" {}

resource "turingpi_node" "node1" {
  count = data.turingpi_usb.current.node == 1 ? 1 : 0

  node        = 1
  power_state = "on"
  boot_check  = true
}
```

### Check USB Before Installation

Verify USB is properly configured before proceeding with installation:

```hcl
data "turingpi_usb" "current" {}

locals {
  usb_ready_for_install = (
    data.turingpi_usb.current.mode == "host" &&
    data.turingpi_usb.current.route == "usb-a"
  )
}

output "usb_ready" {
  value = local.usb_ready_for_install
}
```

### Display Full USB Configuration

```hcl
data "turingpi_usb" "current" {}

output "usb_configuration" {
  value = <<-EOT
    USB Configuration:
      Mode:  ${data.turingpi_usb.current.mode}
      Node:  ${data.turingpi_usb.current.node}
      Route: ${data.turingpi_usb.current.route}
  EOT
}
```

## Attribute Reference

- `mode` - (String) Current USB mode. Values:
  - `"host"` - Node acts as USB host (can connect USB devices)
  - `"device"` - Node acts as USB device (appears as device to connected host)
- `node` - (Integer) Node ID that USB is currently routed to (1-4).
- `route` - (String) Current USB routing destination. Values:
  - `"usb-a"` - Routed through external USB-A connector
  - `"bmc"` - Routed through BMC chip

## API Endpoint Used

| Endpoint | Purpose |
|----------|---------|
| `GET /api/bmc?opt=get&type=usb` | Retrieve current USB configuration |
