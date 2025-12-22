# Basic Example

This example shows how to control power state for Turing Pi nodes.

## Usage

1. Set environment variables:

```bash
export TURINGPI_USERNAME=root
export TURINGPI_PASSWORD=turing
export TURINGPI_ENDPOINT=https://turingpi.local  # optional
```

2. Initialize and apply:

```bash
terraform init
terraform plan
terraform apply
```

## Resources

- `turingpi_power.node1` - Powers on node 1
- `turingpi_power.node2` - Powers on node 2
