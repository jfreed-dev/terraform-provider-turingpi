# K3s Cluster Example

This example deploys a K3s Kubernetes cluster on pre-flashed Turing Pi nodes.

## Prerequisites

1. Nodes must be flashed with a compatible Linux distribution (e.g., Armbian)
2. SSH access must be configured (key-based recommended)
3. Network connectivity between nodes

## Usage

1. Update the variables or create a `terraform.tfvars` file:

```hcl
control_plane_ip = "10.10.88.73"
worker_ips       = ["10.10.88.74", "10.10.88.75", "10.10.88.76"]
ssh_user         = "root"
ssh_key_path     = "~/.ssh/id_rsa"
metallb_ip_range = "10.10.88.80-10.10.88.89"
ingress_ip       = "10.10.88.80"
```

2. Initialize and apply:

```bash
terraform init
terraform apply
```

3. Use the generated kubeconfig:

```bash
export KUBECONFIG=$(pwd)/kubeconfig
kubectl get nodes
```

## What Gets Deployed

- K3s server on the control plane node
- K3s agents on all worker nodes
- MetalLB for LoadBalancer service support
- NGINX Ingress controller for HTTP/HTTPS routing

## Outputs

| Output | Description |
|--------|-------------|
| `kubeconfig` | Cluster kubeconfig content (sensitive) |
| `api_endpoint` | Kubernetes API server URL |
| `cluster_status` | Current cluster status |
| `kubeconfig_path` | Local path to kubeconfig file |
