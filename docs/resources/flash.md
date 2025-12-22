---
page_title: "turingpi_flash Resource - Turing Pi"
subcategory: ""
description: |-
  Flashes firmware to a Turing Pi compute node.
---

# turingpi_flash (Resource)

Flashes firmware to a Turing Pi compute node. Changes to `node` or `firmware_file` will trigger resource recreation (re-flash).

~> **Note:** Flashing firmware is a destructive operation. Ensure you have the correct firmware file for your compute module.

## Example Usage

```hcl
resource "turingpi_flash" "node1" {
  node          = 1
  firmware_file = "/path/to/firmware.img"
}
```

### Using Variables

```hcl
variable "firmware_path" {
  description = "Path to the firmware image file"
  type        = string
}

resource "turingpi_flash" "node1" {
  node          = 1
  firmware_file = var.firmware_path
}

resource "turingpi_flash" "node2" {
  node          = 2
  firmware_file = var.firmware_path
}
```

## Argument Reference

- `node` - (Required, Integer, ForceNew) The node ID (1-4). Changing this forces a new resource.
- `firmware_file` - (Required, String, ForceNew) Path to the firmware image file. Changing this forces a new resource.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The resource identifier in the format `flash-{node}`.

## Import

Flash resources cannot be imported as they represent a one-time operation.
