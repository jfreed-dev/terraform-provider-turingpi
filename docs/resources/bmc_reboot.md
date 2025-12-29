---
page_title: "turingpi_bmc_reboot Resource - Turing Pi"
subcategory: ""
description: |-
  Triggers a reboot of the Turing Pi BMC.
---

# turingpi_bmc_reboot (Resource)

Triggers a reboot of the Turing Pi BMC (Baseboard Management Controller). The BMC will be temporarily unavailable during the reboot process.

This is a "trigger" resource that reboots the BMC when created or when its triggers change.

## Example Usage

### Basic Usage

```hcl
resource "turingpi_bmc_reboot" "maintenance" {}
```

### Reboot After Firmware Upgrade

```hcl
resource "turingpi_bmc_firmware" "upgrade" {
  firmware_file = "/path/to/firmware.swu"
}

resource "turingpi_bmc_reboot" "post_upgrade" {
  triggers = {
    firmware_upgrade = turingpi_bmc_firmware.upgrade.last_upgrade
  }

  depends_on = [turingpi_bmc_firmware.upgrade]
}
```

### Scheduled Maintenance Reboot

```hcl
variable "maintenance_window" {
  type    = string
  default = "2024-01-15"
}

resource "turingpi_bmc_reboot" "maintenance" {
  triggers = {
    maintenance = var.maintenance_window
  }
}
```

### Reboot with Custom Timeout

```hcl
resource "turingpi_bmc_reboot" "slow_network" {
  wait_for_ready = true
  ready_timeout  = 300  # 5 minutes
}
```

### Skip Wait for Ready

```hcl
resource "turingpi_bmc_reboot" "quick" {
  wait_for_ready = false
}

# Subsequent resources must handle BMC unavailability
resource "time_sleep" "wait_for_bmc" {
  depends_on      = [turingpi_bmc_reboot.quick]
  create_duration = "120s"
}
```

### Conditional Reboot

```hcl
variable "reboot_bmc" {
  type    = bool
  default = false
}

resource "turingpi_bmc_reboot" "conditional" {
  count = var.reboot_bmc ? 1 : 0

  triggers = {
    force = timestamp()
  }
}
```

### Reboot After Configuration Changes

```hcl
resource "turingpi_usb" "node1" {
  node  = 1
  mode  = "host"
  route = "usb-a"
}

resource "turingpi_network_reset" "switch" {
  triggers = {
    usb_config = turingpi_usb.node1.id
  }
}

resource "turingpi_bmc_reboot" "apply_changes" {
  triggers = {
    usb_change     = turingpi_usb.node1.id
    network_reset  = turingpi_network_reset.switch.last_reset
  }

  depends_on = [turingpi_network_reset.switch]
}
```

## Argument Reference

- `triggers` - (Optional, Map of String) A map of values that, when changed, will trigger a BMC reboot.

- `wait_for_ready` - (Optional, Boolean) Wait for the BMC to become available again after reboot. Default: `true`.

- `ready_timeout` - (Optional, Integer) Timeout in seconds to wait for BMC to become ready after reboot. Default: `120` (2 minutes). Only applies when `wait_for_ready = true`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - Always set to `bmc-reboot`.
- `last_reboot` - (String) Timestamp (RFC3339 format) of the last BMC reboot operation.

## Behavior Notes

- **Create**: Creating this resource triggers a BMC reboot.
- **Update**: If the `triggers` map changes, a BMC reboot is performed.
- **Read**: This is a trigger resource with no server-side state to read.
- **Delete**: Deleting this resource does not perform any action.

## Important Considerations

1. **Service Disruption**: Rebooting the BMC will temporarily make all BMC APIs unavailable. Node power state is typically preserved during BMC reboot.

2. **Wait for Ready**: The `wait_for_ready` option (default: true) ensures Terraform waits until the BMC is responding again before continuing. Disable this only if you have other mechanisms to handle BMC unavailability.

3. **Timeout**: The default timeout of 120 seconds should be sufficient for most cases. Increase `ready_timeout` if you have a slow network or the BMC takes longer to boot.

4. **Dependencies**: Use `depends_on` to ensure this resource runs after other BMC configuration changes.

5. **Node Power**: Compute nodes typically remain in their current power state during a BMC reboot, though you cannot control them until the BMC comes back online.

## Readiness Check

When `wait_for_ready = true`, the provider polls the BMC's `/api/bmc?opt=get&type=about` endpoint until it responds with HTTP 200, indicating the BMC is ready to accept commands.

## API Endpoint Used

| Endpoint | Purpose |
|----------|---------|
| `GET /api/bmc?opt=set&type=reboot` | Trigger BMC reboot |
| `GET /api/bmc?opt=get&type=about` | Check BMC readiness |
