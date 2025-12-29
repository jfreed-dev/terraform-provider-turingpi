---
page_title: "turingpi_bmc_firmware Resource - Turing Pi"
subcategory: ""
description: |-
  Upgrades the BMC firmware on the Turing Pi.
---

# turingpi_bmc_firmware (Resource)

Upgrades the BMC (Baseboard Management Controller) firmware on the Turing Pi. The BMC will reboot after a successful firmware update.

This resource supports two modes:
- **Upload mode** (default): Uploads a firmware file from the Terraform host to the BMC
- **Local mode**: Uses a firmware file that already exists on the BMC's filesystem

## Example Usage

### Upload Firmware from Terraform Host

```hcl
resource "turingpi_bmc_firmware" "upgrade" {
  firmware_file = "/path/to/bmc-firmware-2.0.6.swu"
  timeout       = 300
}
```

### Use Firmware File on BMC Filesystem

```hcl
resource "turingpi_bmc_firmware" "upgrade" {
  firmware_file = "/tmp/bmc-firmware.swu"
  bmc_local     = true
  timeout       = 300
}
```

### Trigger Upgrade on Version Change

```hcl
variable "bmc_firmware_version" {
  type    = string
  default = "2.0.6"
}

resource "turingpi_bmc_firmware" "upgrade" {
  firmware_file = "/path/to/bmc-firmware-${var.bmc_firmware_version}.swu"

  triggers = {
    version = var.bmc_firmware_version
  }
}
```

### Upgrade After Downloading Firmware

```hcl
resource "null_resource" "download_firmware" {
  provisioner "local-exec" {
    command = "curl -o /tmp/bmc-firmware.swu https://example.com/firmware/latest.swu"
  }

  triggers = {
    always = timestamp()
  }
}

resource "turingpi_bmc_firmware" "upgrade" {
  firmware_file = "/tmp/bmc-firmware.swu"

  triggers = {
    download_id = null_resource.download_firmware.id
  }

  depends_on = [null_resource.download_firmware]
}
```

### Conditional Upgrade

```hcl
variable "upgrade_bmc" {
  type    = bool
  default = false
}

resource "turingpi_bmc_firmware" "upgrade" {
  count = var.upgrade_bmc ? 1 : 0

  firmware_file = "/path/to/bmc-firmware.swu"
  timeout       = 600
}
```

## Argument Reference

- `firmware_file` - (Required, String) Path to the BMC firmware file. Can be:
  - A local path on the Terraform host (file will be uploaded to BMC)
  - A path on the BMC filesystem when `bmc_local = true`

- `bmc_local` - (Optional, Boolean) If `true`, the `firmware_file` path refers to a file on the BMC's local filesystem. If `false` (default), the file will be uploaded from the Terraform host.

- `triggers` - (Optional, Map of String) A map of values that, when changed, will trigger a firmware upgrade. Use this to force an upgrade based on version changes or other conditions.

- `timeout` - (Optional, Integer) Timeout in seconds for the firmware upgrade operation. Default: `300` (5 minutes). Increase this for slow networks or large firmware files.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - Always set to `bmc-firmware`.
- `last_upgrade` - (String) Timestamp (RFC3339 format) of the last firmware upgrade operation.
- `previous_version` - (String) The firmware version before the upgrade was performed.

## Behavior Notes

- **Create**: Creates this resource triggers a firmware upgrade. The BMC will reboot after successful upgrade.
- **Update**: If `firmware_file`, `bmc_local`, or `triggers` change, a new firmware upgrade is performed.
- **Read**: This is a trigger resource with no server-side state to read.
- **Delete**: Deleting this resource does not affect the BMC firmware.

## Important Considerations

1. **BMC Reboot**: The BMC will reboot after a successful firmware upgrade. This may temporarily disrupt connectivity to all nodes.

2. **Backup First**: Before upgrading, consider backing up your current configuration.

3. **Verify Compatibility**: Ensure the firmware file is compatible with your Turing Pi hardware version.

4. **Network Timeout**: The upgrade process can take several minutes. Adjust the `timeout` parameter if needed.

5. **Recovery**: If the upgrade fails, you may need to use the BMC's recovery mode to restore functionality.

## Checking Current Version

Use the `turingpi_info` data source to check the current firmware version:

```hcl
data "turingpi_info" "bmc" {}

output "current_firmware" {
  value = data.turingpi_info.bmc.firmware_version
}
```

## API Endpoints Used

| Endpoint | Purpose |
|----------|---------|
| `GET /api/bmc?opt=get&type=about` | Get current firmware version |
| `GET /api/bmc?opt=set&type=firmware&length=<bytes>` | Initiate firmware upload |
| `GET /api/bmc?opt=set&type=firmware&local&file=<path>` | Initiate local firmware upgrade |
| `POST /api/bmc/upload/{handle}` | Upload firmware file data |
| `GET /api/bmc/upload/{handle}/cancel` | Cancel firmware upload |
| `GET /api/bmc?opt=get&type=flash` | Check upgrade progress |
