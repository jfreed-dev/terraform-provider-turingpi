# TODO: Future Implementation Tasks

This file tracks planned features and implementation tasks for the Terraform Turing Pi provider.

## Milestone: v1.1.2 - Foundation

### Cluster Helper Infrastructure
- [x] Create `provider/cluster_helpers.go` with shared utilities
- [x] Create `provider/ssh_client.go` with SSHClient interface (key-based auth)
- [x] Add `WaitForSSH()` function with configurable timeout
- [x] Add `WaitForKubeAPI()` function for cluster readiness checks
- [x] Create `provider/kubeconfig.go` with parsing and validation
- [x] Add unit tests with mock SSH client

### Helm Integration
- [x] Create `provider/helm_client.go` with HelmClient interface (mittwald/go-helm-client)
- [x] Implement `DeployHelmChart()` and `DeployFromRepository()` functions
- [x] Add Helm repo management (AddRepository, UpdateRepositories)
- [x] Support custom values via ValuesYaml and Values map
- [x] Add `WaitForHelmRelease()` with configurable timeout
- [x] Add MockHelmClient for testing

### Testing Infrastructure
- [x] Add mock SSH client for testing (MockSSHClient in cluster_helpers_test.go)
- [ ] Add mock Kubernetes API for testing
- [ ] Create cluster integration test framework

---

## Milestone: v1.1.3 - K3s Cluster Module (MVP Complete)

### Resource: `turingpi_k3s_cluster`

#### Schema Definition
- [x] Define resource schema in `provider/resource_k3s_cluster.go`
- [x] Add cluster identity fields (name)
- [ ] Add image configuration (path, source type) - deferred to v1.1.5
- [x] Add control plane node configuration
- [x] Add worker nodes configuration (list)
- [x] Add K3s version and network settings
- [x] Add addon toggles (metallb, ingress)
- [ ] Add NPU/RKNN configuration options - deferred to v1.1.5
- [x] Add timeout configurations

#### Node Provisioning (Assumes Pre-flashed Nodes)
- [ ] Integrate with existing `turingpi_node` flash logic - deferred to v1.1.5
- [ ] Implement parallel node flashing - deferred to v1.1.5
- [ ] Add boot verification using UART monitoring - deferred to v1.1.5
- [x] Implement SSH-based OS bootstrap (swap disable)

#### Storage Setup - Deferred to v1.1.5
- [ ] Detect NVMe devices on nodes
- [ ] Partition and format NVMe for Longhorn
- [ ] Create mount points and symlinks
- [ ] Configure iSCSI for Longhorn

#### K3s Installation
- [x] Implement K3s server installation on control plane
- [x] Extract node-token after server start
- [x] Implement K3s agent installation on workers
- [x] Configure agent with server URL and token
- [x] Wait for all nodes to reach Ready state

#### Addon Deployment
- [x] Deploy MetalLB with IPAddressPool configuration
- [x] Deploy NGINX Ingress with LoadBalancer service
- [ ] Deploy Longhorn with NVMe storage class - deferred to v1.1.5
- [ ] Deploy Prometheus stack (optional) - deferred to v1.1.5
- [ ] Deploy Portainer agent (optional) - deferred to v1.1.5
- [ ] Create Ingress resources for addon UIs - deferred to v1.1.5

#### NPU Support (RK3588) - Deferred to v1.1.5
- [ ] Detect vendor kernel (6.1.x) for NPU compatibility
- [ ] Install RKNN-Toolkit2 runtime
- [ ] Install RKNN-LLM library
- [ ] Deploy rkllama service
- [ ] Download and configure AI models (DeepSeek, etc.)
- [ ] Verify NPU functionality (/sys/kernel/debug/rknpu/version)

#### State Management
- [x] Track cluster state (bootstrapping, ready, degraded)
- [ ] Handle partial failures gracefully - basic implementation
- [ ] Implement cluster update logic - deferred to v1.1.5
- [x] Implement cluster destroy with cleanup

#### Testing
- [x] Unit tests for schema validation
- [x] Mock tests for provisioning logic
- [ ] Integration tests with real hardware (manual)

#### Documentation
- [x] Create `docs/resources/k3s_cluster.md`
- [x] Create `examples/k3s-cluster/` with full example
- [ ] Add K3s deployment guide

---

## Milestone: v1.1.4 - Talos Cluster Module

### Resource: `turingpi_talos_cluster`

#### Talos Image Factory Integration
- [ ] Create `provider/talos_factory_client.go`
- [ ] Implement schematic submission to factory.talos.dev
- [ ] Parse schematic response for image URL
- [ ] Download and cache Talos images
- [ ] Support custom extensions (iscsi-tools, util-linux-tools)
- [ ] Support sbc-rockchip overlay

#### Schema Definition
- [ ] Define resource schema in `provider/resource_talos_cluster.go`
- [ ] Add cluster identity and endpoint fields
- [ ] Add Talos version configuration
- [ ] Add image source options (factory, local, url)
- [ ] Add control plane node configuration
- [ ] Add worker nodes with NVMe configuration
- [ ] Add Kubernetes network settings
- [ ] Add addon toggles

#### Machine Config Generation
- [ ] Shell out to `talosctl gen secrets`
- [ ] Shell out to `talosctl gen config`
- [ ] Implement patch generation for:
  - [ ] Hostnames
  - [ ] Network interfaces
  - [ ] NVMe disk partitioning
  - [ ] Kubelet extra mounts
  - [ ] Scheduling on control plane
- [ ] Apply patches to base configs

#### Cluster Bootstrap
- [ ] Flash Talos image to all nodes
- [ ] Wait for nodes to enter maintenance mode
- [ ] Apply machine configs using `talosctl apply-config --insecure`
- [ ] Configure talosctl endpoints and nodes
- [ ] Execute `talosctl bootstrap` (one-time)
- [ ] Monitor cluster health with `talosctl health`
- [ ] Extract kubeconfig with `talosctl kubeconfig`

#### Addon Deployment
- [ ] Wait for Kubernetes API availability
- [ ] Deploy MetalLB
- [ ] Deploy NGINX Ingress
- [ ] Deploy Longhorn (with privileged namespace)
- [ ] Deploy Prometheus stack (optional)

#### State Management
- [ ] Detect if cluster already bootstrapped (prevent re-bootstrap)
- [ ] Track machine config versions
- [ ] Implement config update/upgrade logic
- [ ] Implement cluster destroy

#### NPU Limitation Handling
- [ ] Add warning in schema about NPU unavailability
- [ ] Document NPU limitation in resource docs
- [ ] Suggest K3s module for NPU workloads

#### Testing
- [ ] Unit tests for schema validation
- [ ] Unit tests for factory client
- [ ] Mock tests for talosctl interactions
- [ ] Integration tests with real hardware (manual)

#### Documentation
- [ ] Create `docs/resources/talos_cluster.md`
- [ ] Create `examples/talos-cluster/` with full example
- [ ] Add Talos deployment guide
- [ ] Document NPU limitations clearly

---

## Milestone: v1.1.5 - Polish & Advanced Features

### Cluster Operations
- [ ] Implement cluster upgrade support (K3s version bumps)
- [ ] Implement Talos upgrade support
- [ ] Add node add/remove operations
- [ ] Add cluster backup functionality
- [ ] Add cluster restore functionality

### Multi-Cluster Support
- [ ] Support managing multiple clusters
- [ ] Add cluster federation options
- [ ] Cross-cluster service mesh (optional)

### Observability
- [ ] Add cluster health data source
- [ ] Add node metrics data source
- [ ] Add addon status data source

### Documentation
- [ ] Comprehensive deployment guides
- [ ] Troubleshooting guide
- [ ] Best practices guide
- [ ] Video tutorials

---

## Research Tasks

### Ansible Integration Options
- [ ] Evaluate `hashicorp/terraform-provider-ansible`
- [ ] Evaluate embedded Ansible runner in Go
- [ ] Evaluate pure SSH-based provisioning
- [ ] Document pros/cons of each approach
- [ ] Make final recommendation

### Talos Integration Options
- [ ] Evaluate `siderolabs/terraform-provider-talos`
- [ ] Evaluate shelling out to talosctl
- [ ] Evaluate direct Talos API integration
- [ ] Document pros/cons of each approach
- [ ] Make final recommendation

### NPU Support Timeline
- [ ] Monitor mainline kernel NPU driver progress
- [ ] Track Rockchip open-source driver efforts
- [ ] Evaluate custom kernel builds for Talos
- [ ] Update documentation when status changes

---

## Notes

### Source Repositories
- K3s automation: `~/Code/turing-ansible-cluster`
- Talos configs: `~/Code/turing-rk1-clust`

### Key Differences: K3s vs Talos

| Feature | K3s (Armbian) | Talos |
|---------|---------------|-------|
| Base OS | Armbian (Debian) | Talos Linux |
| Management | SSH + Ansible | talosctl API |
| Mutability | Mutable (install packages) | Immutable |
| NPU Support | ✅ Yes (vendor kernel) | ❌ No (mainline kernel) |
| GPU Support | Limited | ❌ No |
| Complexity | Higher (more config) | Lower (opinionated) |
| Recovery | SSH access | API only |

### Network Configuration (Default)

| Component | IP/Range |
|-----------|----------|
| BMC | 10.10.88.70 |
| Control Plane | 10.10.88.73 |
| Workers | 10.10.88.74-76 |
| MetalLB Pool | 10.10.88.80-89 |
| Pod Network | 10.244.0.0/16 |
| Service Network | 10.96.0.0/12 |

### Hardware Reference
- Board: Turing Pi 2.5
- Compute: 4x RK1 (RK3588 SoC)
- CPU: 8-core ARM64 (4x A76 + 4x A55) per node
- RAM: 16-32GB per node
- Storage: 32GB eMMC + 500GB NVMe per worker
- NPU: 6 TOPS per node (K3s only)
