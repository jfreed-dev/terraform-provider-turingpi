---
page_title: "turingpi_power Data Source - Turing Pi"
subcategory: ""
description: |-
  Retrieves the current power status of all nodes on the Turing Pi BMC.
---

# turingpi_power (Data Source)

Retrieves the current power status of all nodes (1-4) on the Turing Pi BMC.

This data source provides:
- Individual node power status (node1-node4)
- A map of all node statuses for easy iteration
- Counts of powered on/off nodes

## Example Usage

### Basic Usage

```hcl
data "turingpi_power" "status" {}

output "power_status" {
  value = {
    node1 = data.turingpi_power.status.node1
    node2 = data.turingpi_power.status.node2
    node3 = data.turingpi_power.status.node3
    node4 = data.turingpi_power.status.node4
  }
}
```

### Check Cluster Status

```hcl
data "turingpi_power" "status" {}

output "cluster_status" {
  value = {
    nodes_on  = data.turingpi_power.status.powered_on_count
    nodes_off = data.turingpi_power.status.powered_off_count
    all_on    = data.turingpi_power.status.powered_on_count == 4
  }
}
```

### Conditional Logic Based on Power Status

```hcl
data "turingpi_power" "status" {}

# Only proceed if all nodes are powered on
resource "null_resource" "cluster_ready" {
  count = data.turingpi_power.status.powered_on_count == 4 ? 1 : 0

  provisioner "local-exec" {
    command = "echo 'All nodes are powered on!'"
  }
}
```

### Iterate Over Nodes

```hcl
data "turingpi_power" "status" {}

output "powered_nodes" {
  value = [for name, powered in data.turingpi_power.status.nodes : name if powered]
}

output "offline_nodes" {
  value = [for name, powered in data.turingpi_power.status.nodes : name if !powered]
}
```

### Wait for Specific Node

```hcl
data "turingpi_power" "status" {}

locals {
  node1_ready = data.turingpi_power.status.node1
}

resource "turingpi_node" "node2" {
  count = local.node1_ready ? 1 : 0

  node        = 2
  power_state = "on"
}
```

### Display Human-Readable Status

```hcl
data "turingpi_power" "status" {}

output "cluster_power_report" {
  value = <<-EOT
    Turing Pi Cluster Power Status
    ==============================
    Node 1: ${data.turingpi_power.status.node1 ? "ON" : "OFF"}
    Node 2: ${data.turingpi_power.status.node2 ? "ON" : "OFF"}
    Node 3: ${data.turingpi_power.status.node3 ? "ON" : "OFF"}
    Node 4: ${data.turingpi_power.status.node4 ? "ON" : "OFF"}
    ------------------------------
    Total ON:  ${data.turingpi_power.status.powered_on_count}
    Total OFF: ${data.turingpi_power.status.powered_off_count}
  EOT
}
```

## Attribute Reference

### Individual Node Status

- `node1` - (Boolean) Power status of node 1. `true` = powered on, `false` = powered off.
- `node2` - (Boolean) Power status of node 2. `true` = powered on, `false` = powered off.
- `node3` - (Boolean) Power status of node 3. `true` = powered on, `false` = powered off.
- `node4` - (Boolean) Power status of node 4. `true` = powered on, `false` = powered off.

### Aggregated Status

- `nodes` - (Map of Boolean) Power status of all nodes as a map. Keys are `node1`, `node2`, `node3`, `node4`.
- `powered_on_count` - (Integer) Number of nodes currently powered on (0-4).
- `powered_off_count` - (Integer) Number of nodes currently powered off (0-4).

## API Endpoint Used

| Endpoint | Purpose |
|----------|---------|
| `GET /api/bmc?opt=get&type=power` | Retrieve power status of all nodes |
