# Full Provisioning Example

This example demonstrates comprehensive node management with the `turingpi_node` resource, including power control, firmware flashing, and boot verification.

## Usage

1. Set environment variables:

```bash
export TURINGPI_USERNAME=root
export TURINGPI_PASSWORD=turing
```

2. Initialize and apply:

```bash
terraform init

# Without firmware flashing
terraform apply

# With firmware flashing
terraform apply -var="firmware_path=/path/to/firmware.img"

# With custom boot timeout
terraform apply -var="boot_timeout=180"
```

## Resources

| Node | Power | Firmware | Boot Check | Description |
|------|-------|----------|------------|-------------|
| 1 | on | optional | yes | Full provisioning with optional firmware |
| 2 | on | no | yes | Power on with boot verification |
| 3 | off | no | no | Kept powered off |
| 4 | on | no | no | Quick power on (no boot check) |

## Boot Verification

When `boot_check = true`, the provider monitors UART output for a login prompt. This ensures the node has fully booted before Terraform continues. The `login_prompt_timeout` controls how long to wait (default: 60 seconds).
