---
page_title: "turingpi_network_reset Resource - Turing Pi"
subcategory: ""
description: |-
  Triggers a reset of the Turing Pi network switch.
---

# turingpi_network_reset (Resource)

Triggers a reset of the Turing Pi BMC's network switch. This is a "trigger" resource that performs a network switch reset when created or when its triggers change.

Use this resource when you need to reset the network switch after configuration changes or to resolve network connectivity issues between nodes.

## Example Usage

### Basic Usage

```hcl
resource "turingpi_network_reset" "switch" {}
```

### Reset After Node Power Changes

```hcl
resource "turingpi_power" "nodes" {
  for_each = toset(["1", "2", "3", "4"])

  node  = tonumber(each.key)
  state = "on"
}

resource "turingpi_network_reset" "switch" {
  triggers = {
    # Reset network when any node power state changes
    node_states = join(",", [for k, v in turingpi_power.nodes : "${k}:${v.current_state}"])
  }
}
```

### Reset After Firmware Flash

```hcl
resource "turingpi_flash" "node1" {
  node          = 1
  firmware_file = "/path/to/firmware.img"
}

resource "turingpi_network_reset" "switch" {
  triggers = {
    # Reset network after flashing node 1
    flash_id = turingpi_flash.node1.id
  }
}
```

### Scheduled Reset with Time-Based Trigger

```hcl
resource "turingpi_network_reset" "daily" {
  triggers = {
    # Force reset once per day by including the date
    date = formatdate("YYYY-MM-DD", timestamp())
  }
}
```

### Reset Only When Explicitly Needed

```hcl
variable "force_network_reset" {
  type    = bool
  default = false
}

resource "turingpi_network_reset" "switch" {
  triggers = {
    force = var.force_network_reset ? timestamp() : "stable"
  }
}
```

## Argument Reference

- `triggers` - (Optional, Map of String) A map of values that, when changed, will trigger a network reset. This can be used to force a reset based on other resource changes or external conditions.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - Always set to `network-reset`.
- `last_reset` - (String) Timestamp (RFC3339 format) of the last network reset operation.

## Behavior Notes

- **Create**: Creating this resource triggers a network switch reset.
- **Update**: If the `triggers` map changes, a network reset is performed.
- **Read**: This is a trigger resource with no server-side state to read.
- **Delete**: Deleting this resource does not perform any action on the BMC.

## Use Cases

1. **Post-flash network recovery**: Reset the switch after flashing firmware to ensure network connectivity is restored.
2. **Node power cycle recovery**: Reset after powering nodes on/off to clear any stuck network state.
3. **Troubleshooting**: Force a network reset when experiencing connectivity issues between nodes.
4. **Scheduled maintenance**: Use time-based triggers to periodically reset the network switch.

## API Endpoint Used

| Endpoint | Purpose |
|----------|---------|
| `GET /api/bmc?opt=set&type=network` | Trigger network switch reset |
