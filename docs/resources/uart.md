---
page_title: "turingpi_uart Resource - Turing Pi"
subcategory: ""
description: |-
  Writes data to the UART of a Turing Pi compute node.
---

# turingpi_uart (Resource)

Writes data (commands) to the UART (serial console) of a Turing Pi compute node. Use this to send commands or data to a node's serial console, such as login credentials, shell commands, or configuration data.

## Example Usage

### Send a Simple Command

```hcl
resource "turingpi_uart" "node1_cmd" {
  node    = 1
  command = "echo 'Hello from Terraform'\n"
}
```

### Login to a Node

```hcl
resource "turingpi_uart" "node1_login" {
  node    = 1
  command = "root\n"
}

resource "turingpi_uart" "node1_password" {
  node    = 1
  command = "password123\n"

  depends_on = [turingpi_uart.node1_login]
}
```

### Execute Command After Boot

```hcl
resource "turingpi_power" "node1" {
  node  = 1
  state = "on"
}

resource "time_sleep" "wait_for_boot" {
  depends_on = [turingpi_power.node1]
  create_duration = "60s"
}

resource "turingpi_uart" "node1_command" {
  node    = 1
  command = "hostname\n"

  depends_on = [time_sleep.wait_for_boot]
}
```

### Trigger Command Resend

```hcl
variable "command_version" {
  type    = string
  default = "v1"
}

resource "turingpi_uart" "node1_cmd" {
  node    = 1
  command = "systemctl restart myservice\n"

  triggers = {
    version = var.command_version
  }
}
```

### Send Commands to Multiple Nodes

```hcl
resource "turingpi_uart" "nodes" {
  for_each = toset(["1", "2", "3", "4"])

  node    = tonumber(each.key)
  command = "uptime\n"
}
```

### Interactive Script Execution

```hcl
locals {
  script_commands = [
    "#!/bin/bash",
    "echo 'Starting setup...'",
    "apt-get update",
    "apt-get install -y curl",
    "echo 'Setup complete!'",
  ]
}

resource "turingpi_uart" "node1_script" {
  node    = 1
  command = join("\n", local.script_commands)
}
```

### Conditional Command Based on Power State

```hcl
data "turingpi_power" "status" {}

resource "turingpi_uart" "node1_cmd" {
  count = data.turingpi_power.status.node1 ? 1 : 0

  node    = 1
  command = "date\n"
}
```

## Argument Reference

- `node` - (Required, Integer) Node ID to write UART data to (1-4).

- `command` - (Required, String) The command or data to write to the UART. Will be URL-encoded automatically. Include `\n` for newlines/enter key.

- `triggers` - (Optional, Map of String) A map of values that, when changed, will trigger resending the command.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The resource identifier in the format `uart-write-node-{node}`.

- `last_sent` - (String) Timestamp (RFC3339 format) of when the command was last sent.

## Behavior Notes

- **Create**: Creates this resource sends the command to the UART.
- **Update**: If `command`, `node`, or `triggers` change, the command is resent.
- **Read**: This is a trigger resource with no server-side state to read.
- **Delete**: Deleting this resource does not send any data to the UART.

## Tips for Reliable Serial Communication

1. **Include Newlines**: Most commands require a newline (`\n`) at the end to be executed.

2. **Use Delays**: Serial communication is slow. Use `time_sleep` or `depends_on` to ensure proper timing between commands.

3. **Check Power State**: Ensure the node is powered on before sending commands.

4. **Wait for Boot**: Use boot detection (via `turingpi_node` resource or UART data source) before sending login commands.

5. **Buffer Considerations**: The UART has limited buffer space. Very long commands may need to be split.

## Example: Full Node Setup Sequence

```hcl
# Power on the node
resource "turingpi_power" "node1" {
  node  = 1
  state = "on"
}

# Wait for boot
resource "time_sleep" "boot_wait" {
  depends_on      = [turingpi_power.node1]
  create_duration = "90s"
}

# Send login username
resource "turingpi_uart" "login_user" {
  node       = 1
  command    = "root\n"
  depends_on = [time_sleep.boot_wait]
}

# Wait for password prompt
resource "time_sleep" "password_wait" {
  depends_on      = [turingpi_uart.login_user]
  create_duration = "2s"
}

# Send password
resource "turingpi_uart" "login_pass" {
  node       = 1
  command    = "turing\n"
  depends_on = [time_sleep.password_wait]
}

# Wait for shell prompt
resource "time_sleep" "shell_wait" {
  depends_on      = [turingpi_uart.login_pass]
  create_duration = "2s"
}

# Execute setup command
resource "turingpi_uart" "setup" {
  node       = 1
  command    = "curl -sSL https://example.com/setup.sh | bash\n"
  depends_on = [time_sleep.shell_wait]
}
```

## API Endpoint Used

| Endpoint | Purpose |
|----------|---------|
| `GET /api/bmc?opt=set&type=uart&node={n}&cmd={command}` | Write data to UART |
