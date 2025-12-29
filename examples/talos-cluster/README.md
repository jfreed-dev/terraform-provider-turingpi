# Talos Cluster Example

This example deploys a Talos Kubernetes cluster on pre-flashed Turing Pi nodes.

## Prerequisites

1. `talosctl` must be installed and in PATH
2. Nodes must be flashed with Talos Linux image
3. Nodes must be powered on and network-accessible

### Installing talosctl

```bash
# macOS
brew install siderolabs/tap/talosctl

# Linux
curl -sL https://talos.dev/install | sh
```

### Preparing Talos Image

For Turing RK1 nodes, create a custom image via Image Factory:

```bash
# Create schematic
cat > talos-schematic.yaml << 'EOF'
overlay:
  name: turingrk1
  image: siderolabs/sbc-rockchip
customization:
  systemExtensions:
    officialExtensions:
      - siderolabs/iscsi-tools
      - siderolabs/util-linux-tools
EOF

# Get schematic ID
SCHEMATIC_ID=$(curl -X POST --data-binary @talos-schematic.yaml \
  https://factory.talos.dev/schematics | jq -r '.id')

# Download image
curl -L -o metal-arm64.raw.xz \
  "https://factory.talos.dev/image/${SCHEMATIC_ID}/v1.9.1/metal-arm64.raw.xz"
xz -d metal-arm64.raw.xz
```

Then flash nodes using the Turing Pi BMC.

## Usage

1. Update the variables or create a `terraform.tfvars` file:

```hcl
control_plane_ip = "10.10.88.73"
worker_ips       = ["10.10.88.74", "10.10.88.75", "10.10.88.76"]
cluster_name     = "my-talos-cluster"
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

4. Or use talosctl:

```bash
export TALOSCONFIG=$(pwd)/talosconfig
talosctl health
talosctl dashboard
```

## What Gets Deployed

- Talos control plane on the first node
- Talos workers on remaining nodes
- MetalLB for LoadBalancer service support
- NGINX Ingress controller for HTTP/HTTPS routing

## Outputs

| Output | Description |
|--------|-------------|
| `kubeconfig` | Cluster kubeconfig content (sensitive) |
| `talosconfig` | Talosctl config content (sensitive) |
| `api_endpoint` | Kubernetes API server URL |
| `cluster_status` | Current cluster status |
| `kubeconfig_path` | Local path to kubeconfig file |
| `talosconfig_path` | Local path to talosconfig file |

## Files Generated

- `kubeconfig` - Kubernetes cluster access
- `talosconfig` - Talos CLI configuration
- `secrets.yaml` - Cluster PKI secrets (keep secure!)

## NPU Limitation

Talos uses a mainline kernel without Rockchip NPU drivers. For NPU workloads, use `turingpi_k3s_cluster` with Armbian instead.

## Troubleshooting

### Cluster not bootstrapping

Check node status:
```bash
talosctl --talosconfig ./talosconfig health --nodes 10.10.88.73
```

### Reset a node

```bash
talosctl --talosconfig ./talosconfig reset \
  --nodes 10.10.88.73 --graceful=false --reboot
```
