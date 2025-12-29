---
page_title: "turingpi_power Resource - Turing Pi"
subcategory: ""
description: |-
  Manages power state of a Turing Pi compute node, including power on, power off, and reset operations.
---

# turingpi_power (Resource)

Manages the power state of a Turing Pi compute node, including power on, power off, and reset (reboot) operations.

## Example Usage

### Power On

```hcl
resource "turingpi_power" "node1" {
  node  = 1
  state = "on"
}
```

### Power Off

```hcl
resource "turingpi_power" "node4" {
  node  = 4
  state = "off"
}
```

### Reset (Reboot)

```hcl
resource "turingpi_power" "node2_reset" {
  node  = 2
  state = "reset"
}
```

### Power On All Nodes

```hcl
resource "turingpi_power" "nodes" {
  for_each = toset(["1", "2", "3", "4"])

  node  = tonumber(each.key)
  state = "on"
}
```

### Conditional Power Control

```hcl
variable "active_nodes" {
  type    = list(number)
  default = [1, 2]
}

resource "turingpi_power" "active" {
  for_each = toset([for n in var.active_nodes : tostring(n)])

  node  = tonumber(each.key)
  state = "on"
}
```

### Display Current Power State

```hcl
resource "turingpi_power" "node1" {
  node  = 1
  state = "on"
}

output "node1_actual_state" {
  value = turingpi_power.node1.current_state ? "powered on" : "powered off"
}
```

## Argument Reference

- `node` - (Required, Integer) The node ID to control (1-4).
- `state` - (Required, String) The desired power state. Valid values:
  - `"on"` - Power on the node
  - `"off"` - Power off the node
  - `"reset"` - Reset (reboot) the node. After reset, the node will be powered on.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The resource identifier in the format `power-node-{node}`.
- `current_state` - (Boolean) The actual power state as reported by the BMC. `true` = powered on, `false` = powered off.

## Power States Explained

| State | Description | After Operation |
|-------|-------------|-----------------|
| `on` | Powers on the node | Node is running |
| `off` | Powers off the node | Node is stopped |
| `reset` | Triggers a reboot | Node restarts and is running |

## Behavior Notes

- **Delete behavior**: When the resource is destroyed, the node is powered off.
- **Reset state**: Setting `state = "reset"` triggers a reboot. The `current_state` will show `true` (on) after the reset completes.
- **Idempotency**: Repeatedly applying `state = "on"` when already on, or `state = "off"` when already off, is safe and idempotent.

## Import

Power resources can be imported using the node ID (1-4):

```shell
terraform import turingpi_power.node1 1
terraform import turingpi_power.node2 2
```

After import, you should set the `state` attribute in your configuration to match the desired state.
