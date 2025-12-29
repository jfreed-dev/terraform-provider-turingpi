---
page_title: "turingpi_sdcard Data Source - Turing Pi"
subcategory: ""
description: |-
  Retrieves information about the microSD card in the Turing Pi BMC.
---

# turingpi_sdcard (Data Source)

Retrieves information about the microSD card in the Turing Pi BMC, including total capacity, used space, and free space. Values are provided in both bytes and gigabytes for convenience.

This data source is useful for:
- Checking available storage before flashing firmware
- Monitoring storage usage
- Alerting on low storage conditions

## Example Usage

### Basic Usage

```hcl
data "turingpi_sdcard" "storage" {}

output "sdcard_info" {
  value = {
    total_gb     = data.turingpi_sdcard.storage.total_gb
    used_gb      = data.turingpi_sdcard.storage.used_gb
    free_gb      = data.turingpi_sdcard.storage.free_gb
    used_percent = data.turingpi_sdcard.storage.used_percent
  }
}
```

### Check Storage Before Flashing

```hcl
data "turingpi_sdcard" "storage" {}

locals {
  # Firmware images are typically around 2GB
  has_space_for_firmware = data.turingpi_sdcard.storage.free_gb > 3
}

resource "turingpi_flash" "node1" {
  count = local.has_space_for_firmware ? 1 : 0

  node          = 1
  firmware_file = var.firmware_path
}

output "storage_warning" {
  value = local.has_space_for_firmware ? null : "WARNING: Less than 3GB free on SD card!"
}
```

### Storage Monitoring

```hcl
data "turingpi_sdcard" "storage" {}

output "storage_status" {
  value = {
    status = data.turingpi_sdcard.storage.used_percent > 80 ? "WARNING" : "OK"
    message = format(
      "SD Card: %.1f GB used of %.1f GB (%.1f%% used)",
      data.turingpi_sdcard.storage.used_gb,
      data.turingpi_sdcard.storage.total_gb,
      data.turingpi_sdcard.storage.used_percent
    )
  }
}
```

### Conditional Logic Based on Free Space

```hcl
data "turingpi_sdcard" "storage" {}

variable "firmware_size_gb" {
  type    = number
  default = 2.5
}

locals {
  can_flash = data.turingpi_sdcard.storage.free_gb >= var.firmware_size_gb
}

output "flash_status" {
  value = local.can_flash ? "Ready to flash" : "Insufficient space - need ${var.firmware_size_gb}GB, have ${data.turingpi_sdcard.storage.free_gb}GB"
}
```

### Combined with BMC Info

```hcl
data "turingpi_info" "bmc" {}
data "turingpi_sdcard" "storage" {}

output "bmc_storage_report" {
  value = {
    firmware_version = data.turingpi_info.bmc.firmware_version
    sdcard = {
      total    = "${data.turingpi_sdcard.storage.total_gb} GB"
      used     = "${data.turingpi_sdcard.storage.used_gb} GB"
      free     = "${data.turingpi_sdcard.storage.free_gb} GB"
      percent  = "${data.turingpi_sdcard.storage.used_percent}%"
    }
  }
}
```

## Attribute Reference

### Byte Values

- `total_bytes` - (Integer) Total microSD card capacity in bytes.
- `used_bytes` - (Integer) Used space on the microSD card in bytes.
- `free_bytes` - (Integer) Free space on the microSD card in bytes.

### Gigabyte Values

- `total_gb` - (Float) Total microSD card capacity in gigabytes.
- `used_gb` - (Float) Used space on the microSD card in gigabytes.
- `free_gb` - (Float) Free space on the microSD card in gigabytes.

### Percentage

- `used_percent` - (Float) Percentage of microSD card space used (0-100).

## Notes

1. **SD Card Required**: The BMC requires a microSD card for firmware storage. This data source will return an error if no SD card is present.

2. **Byte Calculation**: Gigabyte values are calculated as bytes / (1024³), representing gibibytes (GiB) commonly displayed by operating systems.

3. **Refresh**: This data source reads current values each time Terraform runs. Use `terraform refresh` to get updated values without making changes.

## API Endpoint Used

| Endpoint | Purpose |
|----------|---------|
| `GET /api/bmc?opt=get&type=sdcard` | Get SD card storage info |
