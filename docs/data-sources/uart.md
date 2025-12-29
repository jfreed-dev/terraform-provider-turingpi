---
page_title: "turingpi_uart Data Source - Turing Pi"
subcategory: ""
description: |-
  Reads buffered UART output from a Turing Pi compute node.
---

# turingpi_uart (Data Source)

Reads buffered UART (serial console) output from a Turing Pi compute node. This is useful for monitoring boot progress, capturing log output, or detecting specific patterns in the serial output.

**Important**: Reading the UART output clears the buffer. Subsequent reads will only return new data that arrived after the previous read.

## Example Usage

### Basic Usage

```hcl
data "turingpi_uart" "node1" {
  node = 1
}

output "node1_uart" {
  value = data.turingpi_uart.node1.output
}
```

### Check for Boot Completion

```hcl
data "turingpi_uart" "node1" {
  node = 1
}

output "node1_booted" {
  value = strcontains(data.turingpi_uart.node1.output, "login:")
}
```

### Read with Different Encoding

```hcl
data "turingpi_uart" "node1_utf16" {
  node     = 1
  encoding = "utf16"
}
```

### Conditional Processing Based on Output

```hcl
data "turingpi_uart" "node1" {
  node = 1
}

locals {
  boot_complete = data.turingpi_uart.node1.has_output && strcontains(data.turingpi_uart.node1.output, "login:")
}

resource "null_resource" "post_boot" {
  count = local.boot_complete ? 1 : 0

  provisioner "local-exec" {
    command = "echo 'Node 1 has booted successfully'"
  }
}
```

### Monitor Multiple Nodes

```hcl
data "turingpi_uart" "nodes" {
  for_each = toset(["1", "2", "3", "4"])
  node     = tonumber(each.key)
}

output "uart_outputs" {
  value = {
    for k, v in data.turingpi_uart.nodes : "node${k}" => {
      has_output = v.has_output
      output     = v.output
    }
  }
}
```

## Argument Reference

- `node` - (Required, Integer) Node ID to read UART output from (1-4).

- `encoding` - (Optional, String) Character encoding for UART output. Default: `utf8`. Valid values:
  - `utf8` (default)
  - `utf16`
  - `utf16le`
  - `utf16be`
  - `utf32`
  - `utf32le`
  - `utf32be`

## Attribute Reference

- `id` - The data source identifier in the format `uart-node-{node}`.

- `output` - (String) The buffered UART output from the node. Empty string if no data in buffer.

- `has_output` - (Boolean) `true` if there was any output in the UART buffer, `false` otherwise.

## Use Cases

1. **Boot Monitoring**: Check if a node has completed booting by looking for login prompts or specific boot messages.

2. **Log Capture**: Capture serial console output for debugging or logging purposes.

3. **Pattern Detection**: Look for specific patterns in UART output to trigger subsequent operations.

4. **Talos Linux Detection**: Detect Talos Linux boot completion by looking for "machine is running and ready".

## Important Notes

- **Buffer Clearing**: Reading UART output clears the buffer. Each read returns only new data since the last read.

- **Timing**: UART output may not be immediately available. Consider using `depends_on` or timing resources if you need to wait for specific output.

- **Buffer Size**: The BMC has a limited UART buffer. If output is generated faster than it's read, some data may be lost.

## API Endpoint Used

| Endpoint | Purpose |
|----------|---------|
| `GET /api/bmc?opt=get&type=uart&node={n}&encoding={enc}` | Read buffered UART output |
