---
page_title: "turingpi_power Resource - Turing Pi"
subcategory: ""
description: |-
  Controls the power state of a Turing Pi compute node.
---

# turingpi_power (Resource)

Controls the power state of a Turing Pi compute node.

## Example Usage

```hcl
# Power on node 1
resource "turingpi_power" "node1" {
  node  = 1
  state = true
}

# Power off node 4
resource "turingpi_power" "node4" {
  node  = 4
  state = false
}
```

## Argument Reference

- `node` - (Required, Integer) The node ID (1-4).
- `state` - (Required, Boolean) The desired power state. `true` for on, `false` for off.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The resource identifier in the format `power-{node}`.

## Import

Power resources can be imported using the node ID:

```shell
terraform import turingpi_power.node1 1
```
