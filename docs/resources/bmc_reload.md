---
page_title: "turingpi_bmc_reload Resource - Turing Pi"
subcategory: ""
description: |-
  Restarts the BMC system management daemon (bmcd).
---

# turingpi_bmc_reload (Resource)

Restarts the BMC system management daemon (bmcd). This is a softer restart than a full BMC reboot (`turingpi_bmc_reboot`), only restarting the daemon process rather than the entire BMC system.

This is useful for:
- Applying configuration changes that require daemon restart
- Recovering from daemon issues without full BMC reboot
- Faster restart cycles during development/testing

## Example Usage

### Basic Usage

```hcl
resource "turingpi_bmc_reload" "daemon" {}
```

### Reload After Configuration Changes

```hcl
resource "turingpi_usb" "node1" {
  node  = 1
  mode  = "host"
  route = "usb-a"
}

resource "turingpi_bmc_reload" "apply_changes" {
  triggers = {
    usb_config = turingpi_usb.node1.id
  }

  depends_on = [turingpi_usb.node1]
}
```

### Triggered Reload

```hcl
variable "reload_daemon" {
  type    = string
  default = ""
}

resource "turingpi_bmc_reload" "manual" {
  triggers = {
    reload = var.reload_daemon
  }
}
```

### Reload with Custom Timeout

```hcl
resource "turingpi_bmc_reload" "slow_network" {
  wait_for_ready = true
  ready_timeout  = 60  # 1 minute
}
```

### Skip Wait for Ready

```hcl
resource "turingpi_bmc_reload" "quick" {
  wait_for_ready = false
}
```

### Conditional Reload

```hcl
variable "reload_bmc_daemon" {
  type    = bool
  default = false
}

resource "turingpi_bmc_reload" "conditional" {
  count = var.reload_bmc_daemon ? 1 : 0

  triggers = {
    force = timestamp()
  }
}
```

## Argument Reference

- `triggers` - (Optional, Map of String) A map of values that, when changed, will trigger a daemon reload.

- `wait_for_ready` - (Optional, Boolean) Wait for the BMC daemon to become available again after reload. Default: `true`.

- `ready_timeout` - (Optional, Integer) Timeout in seconds to wait for daemon to become ready after reload. Default: `30`. Only applies when `wait_for_ready = true`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - Always set to `bmc-reload`.
- `last_reload` - (String) Timestamp (RFC3339 format) of the last daemon reload operation.

## Behavior Notes

- **Create**: Creating this resource triggers a daemon reload.
- **Update**: If the `triggers` map changes, a daemon reload is performed.
- **Read**: This is a trigger resource with no server-side state to read.
- **Delete**: Deleting this resource does not perform any action.

## Comparison with turingpi_bmc_reboot

| Feature | turingpi_bmc_reload | turingpi_bmc_reboot |
|---------|---------------------|---------------------|
| Scope | Daemon only | Entire BMC system |
| Speed | Faster (~5-10 sec) | Slower (~60-120 sec) |
| Impact | Minimal | More disruptive |
| Use Case | Config changes | Firmware updates, recovery |

## Important Considerations

1. **Service Disruption**: During daemon reload, BMC API requests may briefly fail. This is typically much shorter than a full BMC reboot.

2. **Wait for Ready**: The `wait_for_ready` option (default: true) ensures Terraform waits until the daemon is responding again.

3. **Default Timeout**: The default timeout of 30 seconds is usually sufficient for daemon restarts. Increase if needed for slow environments.

4. **Node Power**: Compute nodes are not affected by daemon reload - they maintain their current power state.

## Readiness Check

When `wait_for_ready = true`, the provider polls the BMC's `/api/bmc?opt=get&type=about` endpoint until it responds with HTTP 200, indicating the daemon is ready to accept commands.

## API Endpoint Used

| Endpoint | Purpose |
|----------|---------|
| `GET /api/bmc?opt=set&type=reload` | Trigger daemon reload |
| `GET /api/bmc?opt=get&type=about` | Check daemon readiness |
