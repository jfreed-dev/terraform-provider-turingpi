// Package talos provides Talos cluster provisioning via talosctl.
package talos

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// NodeConfig holds configuration for a Talos node
type NodeConfig struct {
	Host     string
	Hostname string
}

// ClusterConfig holds the Talos cluster configuration
type ClusterConfig struct {
	Name                string
	ClusterEndpoint     string
	KubernetesVersion   string
	InstallDisk         string
	ControlPlanes       []NodeConfig
	Workers             []NodeConfig
	AllowSchedulingOnCP bool
	BootstrapTimeout    time.Duration
}

// ClusterState holds the state of a provisioned cluster
type ClusterState struct {
	SecretsYAML     string
	Talosconfig     string
	Kubeconfig      string
	APIEndpoint     string
	ClusterStatus   string
	ControlPlaneIPs []string
	WorkerIPs       []string
}

// Provisioner handles Talos cluster operations via talosctl
type Provisioner struct {
	talosctlPath string
	workDir      string
	execCommand  func(name string, arg ...string) *exec.Cmd
}

// NewProvisioner creates a new Talos provisioner
func NewProvisioner() (*Provisioner, error) {
	// Find talosctl in PATH
	talosctlPath, err := exec.LookPath("talosctl")
	if err != nil {
		return nil, fmt.Errorf("talosctl not found in PATH: %w", err)
	}

	// Create temp working directory
	workDir, err := os.MkdirTemp("", "talos-provisioner-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return &Provisioner{
		talosctlPath: talosctlPath,
		workDir:      workDir,
		execCommand:  exec.Command,
	}, nil
}

// NewProvisionerWithExec creates a provisioner with custom exec function (for testing)
func NewProvisionerWithExec(execFn func(string, ...string) *exec.Cmd) *Provisioner {
	workDir, _ := os.MkdirTemp("", "talos-provisioner-*")
	return &Provisioner{
		talosctlPath: "talosctl",
		workDir:      workDir,
		execCommand:  execFn,
	}
}

// Cleanup removes the working directory
func (p *Provisioner) Cleanup() error {
	if p.workDir != "" {
		return os.RemoveAll(p.workDir)
	}
	return nil
}

// WorkDir returns the working directory path
func (p *Provisioner) WorkDir() string {
	return p.workDir
}

// runTalosctl executes a talosctl command and returns the output
func (p *Provisioner) runTalosctl(args ...string) (string, error) {
	cmd := p.execCommand(p.talosctlPath, args...)
	cmd.Dir = p.workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("talosctl %s failed: %w\nOutput: %s", strings.Join(args, " "), err, string(output))
	}
	return string(output), nil
}

// runTalosctlWithConfig executes a talosctl command with a specific talosconfig
func (p *Provisioner) runTalosctlWithConfig(talosconfig string, args ...string) (string, error) {
	fullArgs := append([]string{"--talosconfig", talosconfig}, args...)
	return p.runTalosctl(fullArgs...)
}

// GenerateSecrets generates cluster secrets (PKI)
func (p *Provisioner) GenerateSecrets(outputPath string) error {
	_, err := p.runTalosctl("gen", "secrets", "-o", outputPath)
	if err != nil {
		return fmt.Errorf("failed to generate secrets: %w", err)
	}
	return nil
}

// GenerateConfig generates machine configs for the cluster
func (p *Provisioner) GenerateConfig(secretsPath, clusterName, endpoint, installDisk, outputDir string) error {
	args := []string{
		"gen", "config",
		"--with-secrets", secretsPath,
		clusterName,
		endpoint,
		"--install-disk", installDisk,
		"--output-dir", outputDir,
	}

	_, err := p.runTalosctl(args...)
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}
	return nil
}

// generatePatchYAML creates a YAML patch for node configuration
func generatePatchYAML(hostname string, allowSchedulingOnCP bool, isControlPlane bool) (string, error) {
	patch := map[string]interface{}{
		"machine": map[string]interface{}{
			"network": map[string]interface{}{
				"hostname": hostname,
			},
		},
	}

	if isControlPlane && allowSchedulingOnCP {
		patch["cluster"] = map[string]interface{}{
			"allowSchedulingOnControlPlanes": true,
		}
	}

	data, err := yaml.Marshal(patch)
	if err != nil {
		return "", fmt.Errorf("failed to marshal patch: %w", err)
	}
	return string(data), nil
}

// PatchConfig patches a machine config with the given patch file
func (p *Provisioner) PatchConfig(configPath, patchContent, outputPath string) error {
	// Write patch to temp file
	patchFile := filepath.Join(p.workDir, "patch-"+filepath.Base(outputPath))
	if err := os.WriteFile(patchFile, []byte(patchContent), 0600); err != nil {
		return fmt.Errorf("failed to write patch file: %w", err)
	}
	defer func() { _ = os.Remove(patchFile) }()

	args := []string{
		"machineconfig", "patch",
		configPath,
		"--patch", "@" + patchFile,
		"--output", outputPath,
	}

	_, err := p.runTalosctl(args...)
	if err != nil {
		return fmt.Errorf("failed to patch config: %w", err)
	}
	return nil
}

// ApplyConfig applies a machine config to a node
func (p *Provisioner) ApplyConfig(nodeIP, configPath string, insecure bool) error {
	args := []string{
		"apply-config",
		"--nodes", nodeIP,
		"--file", configPath,
	}
	if insecure {
		args = append(args, "--insecure")
	}

	_, err := p.runTalosctl(args...)
	if err != nil {
		return fmt.Errorf("failed to apply config to %s: %w", nodeIP, err)
	}
	return nil
}

// ApplyConfigWithTalosconfig applies config using a specific talosconfig
func (p *Provisioner) ApplyConfigWithTalosconfig(talosconfig, nodeIP, configPath string) error {
	args := []string{
		"apply-config",
		"--nodes", nodeIP,
		"--file", configPath,
	}

	_, err := p.runTalosctlWithConfig(talosconfig, args...)
	if err != nil {
		return fmt.Errorf("failed to apply config to %s: %w", nodeIP, err)
	}
	return nil
}

// IsBootstrapped checks if the cluster is already bootstrapped
func (p *Provisioner) IsBootstrapped(talosconfig, nodeIP string) (bool, error) {
	args := []string{
		"etcd", "status",
		"--nodes", nodeIP,
	}

	output, err := p.runTalosctlWithConfig(talosconfig, args...)
	if err != nil {
		// Error likely means not bootstrapped yet
		return false, nil
	}

	// If we get members in the output, cluster is bootstrapped
	return strings.Contains(output, "MEMBER") || strings.Contains(output, "members"), nil
}

// Bootstrap bootstraps the cluster (ONE TIME ONLY)
func (p *Provisioner) Bootstrap(talosconfig, nodeIP string) error {
	// First check if already bootstrapped
	bootstrapped, err := p.IsBootstrapped(talosconfig, nodeIP)
	if err != nil {
		return fmt.Errorf("failed to check bootstrap status: %w", err)
	}

	if bootstrapped {
		// Already bootstrapped, skip
		return nil
	}

	args := []string{
		"bootstrap",
		"--nodes", nodeIP,
	}

	_, err = p.runTalosctlWithConfig(talosconfig, args...)
	if err != nil {
		return fmt.Errorf("failed to bootstrap cluster: %w", err)
	}
	return nil
}

// GetKubeconfig retrieves the kubeconfig from the cluster
func (p *Provisioner) GetKubeconfig(talosconfig, nodeIP, outputPath string) error {
	args := []string{
		"kubeconfig",
		"--nodes", nodeIP,
		outputPath,
	}

	_, err := p.runTalosctlWithConfig(talosconfig, args...)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig: %w", err)
	}
	return nil
}

// ReadTalosconfig reads the talosconfig file content
func (p *Provisioner) ReadTalosconfig(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read talosconfig: %w", err)
	}
	return string(data), nil
}

// ReadSecrets reads the secrets file content
func (p *Provisioner) ReadSecrets(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read secrets: %w", err)
	}
	return string(data), nil
}

// WaitForHealth waits for the node to be healthy
func (p *Provisioner) WaitForHealth(talosconfig, nodeIP string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		args := []string{
			"health",
			"--nodes", nodeIP,
			"--wait-timeout", "10s",
		}

		_, err := p.runTalosctlWithConfig(talosconfig, args...)
		if err == nil {
			return nil
		}

		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("timeout waiting for node %s to be healthy after %v", nodeIP, timeout)
}

// WaitForAPIServer waits for the Kubernetes API server to be ready
func (p *Provisioner) WaitForAPIServer(talosconfig, nodeIP string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		args := []string{
			"service", "kube-apiserver",
			"--nodes", nodeIP,
		}

		output, err := p.runTalosctlWithConfig(talosconfig, args...)
		if err == nil && strings.Contains(output, "Running") {
			return nil
		}

		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("timeout waiting for API server on %s after %v", nodeIP, timeout)
}

// Reset resets a node (wipes it)
func (p *Provisioner) Reset(talosconfig, nodeIP string, graceful bool) error {
	args := []string{
		"reset",
		"--nodes", nodeIP,
		"--reboot",
	}
	if !graceful {
		args = append(args, "--graceful=false")
	}

	_, err := p.runTalosctlWithConfig(talosconfig, args...)
	if err != nil {
		// Reset might timeout as node reboots - this is expected
		if strings.Contains(err.Error(), "context deadline exceeded") ||
			strings.Contains(err.Error(), "connection refused") {
			return nil
		}
		return fmt.Errorf("failed to reset node %s: %w", nodeIP, err)
	}
	return nil
}

// GetClusterMembers returns the list of etcd cluster members
func (p *Provisioner) GetClusterMembers(talosconfig, nodeIP string) ([]string, error) {
	args := []string{
		"etcd", "members",
		"--nodes", nodeIP,
	}

	output, err := p.runTalosctlWithConfig(talosconfig, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster members: %w", err)
	}

	// Parse member IPs from output
	var members []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "MEMBER") {
			continue
		}
		// Extract IP from member line
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			members = append(members, fields[1])
		}
	}

	return members, nil
}

// ProvisionCluster provisions a complete Talos cluster
func (p *Provisioner) ProvisionCluster(ctx context.Context, cfg ClusterConfig) (*ClusterState, error) {
	state := &ClusterState{
		ClusterStatus: "bootstrapping",
	}

	// 1. Generate secrets
	secretsPath := filepath.Join(p.workDir, "secrets.yaml")
	if err := p.GenerateSecrets(secretsPath); err != nil {
		return nil, err
	}

	// Read secrets for state
	secretsContent, err := p.ReadSecrets(secretsPath)
	if err != nil {
		return nil, err
	}
	state.SecretsYAML = secretsContent

	// 2. Generate base configs
	configDir := filepath.Join(p.workDir, "configs")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := p.GenerateConfig(secretsPath, cfg.Name, cfg.ClusterEndpoint, cfg.InstallDisk, configDir); err != nil {
		return nil, err
	}

	// Read talosconfig
	talosconfigPath := filepath.Join(configDir, "talosconfig")
	talosconfigContent, err := p.ReadTalosconfig(talosconfigPath)
	if err != nil {
		return nil, err
	}
	state.Talosconfig = talosconfigContent

	// 3. Apply configs to control planes
	controlplaneConfig := filepath.Join(configDir, "controlplane.yaml")
	for i, cp := range cfg.ControlPlanes {
		// Generate hostname patch
		hostname := cp.Hostname
		if hostname == "" {
			hostname = fmt.Sprintf("turing-cp-%d", i+1)
		}

		patchContent, err := generatePatchYAML(hostname, cfg.AllowSchedulingOnCP, true)
		if err != nil {
			return nil, err
		}

		// Patch config
		patchedConfig := filepath.Join(p.workDir, fmt.Sprintf("controlplane-%d.yaml", i+1))
		if err := p.PatchConfig(controlplaneConfig, patchContent, patchedConfig); err != nil {
			return nil, err
		}

		// Apply config (insecure for initial setup)
		if err := p.ApplyConfig(cp.Host, patchedConfig, true); err != nil {
			return nil, err
		}

		state.ControlPlaneIPs = append(state.ControlPlaneIPs, cp.Host)
	}

	// 4. Bootstrap the first control plane
	if len(cfg.ControlPlanes) > 0 {
		firstCP := cfg.ControlPlanes[0].Host

		// Wait a bit for the node to be ready for bootstrap
		time.Sleep(10 * time.Second)

		if err := p.Bootstrap(talosconfigPath, firstCP); err != nil {
			return nil, err
		}

		// Wait for API server
		if err := p.WaitForAPIServer(talosconfigPath, firstCP, cfg.BootstrapTimeout); err != nil {
			return nil, err
		}
	}

	// 5. Apply configs to workers
	workerConfig := filepath.Join(configDir, "worker.yaml")
	for i, worker := range cfg.Workers {
		// Generate hostname patch
		hostname := worker.Hostname
		if hostname == "" {
			hostname = fmt.Sprintf("turing-w-%d", i+1)
		}

		patchContent, err := generatePatchYAML(hostname, false, false)
		if err != nil {
			return nil, err
		}

		// Patch config
		patchedConfig := filepath.Join(p.workDir, fmt.Sprintf("worker-%d.yaml", i+1))
		if err := p.PatchConfig(workerConfig, patchContent, patchedConfig); err != nil {
			return nil, err
		}

		// Apply config (insecure for initial setup)
		if err := p.ApplyConfig(worker.Host, patchedConfig, true); err != nil {
			return nil, err
		}

		state.WorkerIPs = append(state.WorkerIPs, worker.Host)
	}

	// 6. Wait for cluster health
	if len(cfg.ControlPlanes) > 0 {
		if err := p.WaitForHealth(talosconfigPath, cfg.ControlPlanes[0].Host, cfg.BootstrapTimeout); err != nil {
			state.ClusterStatus = "degraded"
			// Continue anyway to get kubeconfig if possible
		} else {
			state.ClusterStatus = "ready"
		}
	}

	// 7. Get kubeconfig
	kubeconfigPath := filepath.Join(p.workDir, "kubeconfig")
	if len(cfg.ControlPlanes) > 0 {
		if err := p.GetKubeconfig(talosconfigPath, cfg.ControlPlanes[0].Host, kubeconfigPath); err != nil {
			return nil, err
		}

		kubeconfigContent, err := os.ReadFile(kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read kubeconfig: %w", err)
		}
		state.Kubeconfig = string(kubeconfigContent)
	}

	// Set API endpoint
	state.APIEndpoint = cfg.ClusterEndpoint

	return state, nil
}

// DestroyCluster destroys a Talos cluster by resetting all nodes
func (p *Provisioner) DestroyCluster(talosconfig string, controlPlaneIPs, workerIPs []string) error {
	// Write talosconfig to temp file
	talosconfigPath := filepath.Join(p.workDir, "talosconfig")
	if err := os.WriteFile(talosconfigPath, []byte(talosconfig), 0600); err != nil {
		return fmt.Errorf("failed to write talosconfig: %w", err)
	}

	// Reset workers first
	for _, ip := range workerIPs {
		if err := p.Reset(talosconfigPath, ip, false); err != nil {
			// Log but continue - node might already be reset
			fmt.Printf("Warning: failed to reset worker %s: %v\n", ip, err)
		}
	}

	// Then reset control planes
	for _, ip := range controlPlaneIPs {
		if err := p.Reset(talosconfigPath, ip, false); err != nil {
			// Log but continue
			fmt.Printf("Warning: failed to reset control plane %s: %v\n", ip, err)
		}
	}

	return nil
}

// CheckClusterHealth checks the health status of the cluster
func (p *Provisioner) CheckClusterHealth(talosconfig string, controlPlaneIP string) (string, error) {
	// Write talosconfig to temp file
	talosconfigPath := filepath.Join(p.workDir, "talosconfig")
	if err := os.WriteFile(talosconfigPath, []byte(talosconfig), 0600); err != nil {
		return "unknown", fmt.Errorf("failed to write talosconfig: %w", err)
	}

	args := []string{
		"health",
		"--nodes", controlPlaneIP,
		"--wait-timeout", "10s",
	}

	_, err := p.runTalosctlWithConfig(talosconfigPath, args...)
	if err != nil {
		return "degraded", nil
	}

	return "ready", nil
}
