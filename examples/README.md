# Examples

This directory contains example Terraform configurations for the Turing Pi provider.

## Prerequisites

1. Install the provider to your local Terraform plugins directory
2. Set environment variables for authentication:

```bash
export TURINGPI_USERNAME=root
export TURINGPI_PASSWORD=turing
export TURINGPI_ENDPOINT=https://turingpi.local  # optional
```

## Examples

| Directory | Description |
|-----------|-------------|
| [basic](./basic) | Simple power control for nodes |
| [flash-firmware](./flash-firmware) | Flash firmware to nodes |
| [full-provisioning](./full-provisioning) | Complete node management with boot verification |

## Running Examples

```bash
cd examples/<example-name>
terraform init
terraform plan
terraform apply
```

## Cleanup

To power off nodes and clean up:

```bash
terraform destroy
```
