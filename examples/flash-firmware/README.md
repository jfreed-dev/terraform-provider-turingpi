# Flash Firmware Example

This example shows how to flash firmware to Turing Pi nodes.

## Usage

1. Set environment variables:

```bash
export TURINGPI_USERNAME=root
export TURINGPI_PASSWORD=turing
```

2. Initialize and apply with firmware path:

```bash
terraform init
terraform plan -var="firmware_path=/path/to/firmware.img"
terraform apply -var="firmware_path=/path/to/firmware.img"
```

## Notes

- The `turingpi_flash` resource uses `ForceNew` for both `node` and `firmware_file`
- Changing either value will destroy and recreate the resource (re-flash)
- Ensure the firmware file exists and is accessible
