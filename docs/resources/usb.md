---
page_title: "turingpi_usb Resource - Turing Pi"
subcategory: ""
description: |-
  Manages USB routing configuration on the Turing Pi BMC.
---

# turingpi_usb (Resource)

Manages USB routing configuration on the Turing Pi BMC. The USB bus can only be routed to one node at a time.

This resource is useful for:
- Routing USB storage to a specific node for OS installation
- Switching between USB host and device modes
- Routing USB through the BMC for firmware operations

## Example Usage

### Basic USB Host Mode

Route USB-A port to node 1 in host mode (node can use USB devices):

```hcl
resource "turingpi_usb" "node1" {
  node  = 1
  mode  = "host"
  route = "usb-a"
}
```

### USB Device Mode

Configure node 2 as a USB device (appears as a USB device to connected host):

```hcl
resource "turingpi_usb" "node2_device" {
  node  = 2
  mode  = "device"
  route = "usb-a"
}
```

### Route USB to BMC

Route USB through the BMC chip instead of the USB-A connector:

```hcl
resource "turingpi_usb" "bmc_route" {
  node  = 1
  mode  = "host"
  route = "bmc"
}
```

### USB for OS Installation

Temporarily route USB to a node for installing an operating system:

```hcl
# Route USB to node for installation
resource "turingpi_usb" "install" {
  node  = 1
  mode  = "host"
  route = "usb-a"
}

# Power on the node after USB is configured
resource "turingpi_node" "node1" {
  node        = 1
  power_state = "on"

  depends_on = [turingpi_usb.install]
}
```

### Display Current USB Status

```hcl
resource "turingpi_usb" "current" {
  node  = 1
  mode  = "host"
}

output "usb_status" {
  value = {
    configured_node  = turingpi_usb.current.node
    configured_mode  = turingpi_usb.current.mode
    configured_route = turingpi_usb.current.route
    current_node     = turingpi_usb.current.current_node
    current_mode     = turingpi_usb.current.current_mode
    current_route    = turingpi_usb.current.current_route
  }
}
```

## Argument Reference

- `node` - (Required, Integer) The node ID to route USB to (1-4).
- `mode` - (Required, String) USB mode. Valid values:
  - `"host"` - Node acts as USB host (can connect USB devices to the node)
  - `"device"` - Node acts as USB device (node appears as a USB device to connected host)
- `route` - (Optional, String) USB routing destination. Valid values:
  - `"usb-a"` - Route through external USB-A connector (default)
  - `"bmc"` - Route through BMC chip

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The resource identifier in the format `usb-node-{node}`.
- `current_mode` - (String) Current USB mode as reported by BMC.
- `current_node` - (Integer) Current node that USB is routed to.
- `current_route` - (String) Current USB routing destination.

## USB Modes Explained

| Mode | Description | Use Case |
|------|-------------|----------|
| `host` | Node acts as USB host | Connect USB drives, keyboards, etc. to the node |
| `device` | Node acts as USB device | Node appears as storage device to connected computer |

## Routing Options

| Route | Description |
|-------|-------------|
| `usb-a` | USB traffic goes through the external USB-A connector on the Turing Pi board |
| `bmc` | USB traffic is routed through the BMC chip |

## Important Notes

- **Single Node Routing**: The USB bus can only be routed to one node at a time. Creating a new `turingpi_usb` resource will change the routing away from any previously configured node.
- **Persistent Configuration**: USB routing persists on the BMC. Deleting this resource from Terraform state does not reset the USB configuration.
- **Node Indexing**: The provider uses 1-indexed node IDs (1-4), matching the physical labels on the Turing Pi board.

## Import

USB resources can be imported using the node ID:

```shell
terraform import turingpi_usb.node1 usb-node-1
```
