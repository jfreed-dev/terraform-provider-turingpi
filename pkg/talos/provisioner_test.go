package talos

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Test NewProvisioner creates valid provisioner (requires talosctl in PATH)
func TestNewProvisioner_RequiresTalosctl(t *testing.T) {
	// Skip if talosctl is not installed
	_, err := exec.LookPath("talosctl")
	if err != nil {
		t.Skip("talosctl not installed, skipping test")
	}

	p, err := NewProvisioner()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = p.Cleanup() }()

	if p == nil {
		t.Fatal("expected non-nil provisioner")
	}
	if p.talosctlPath == "" {
		t.Fatal("expected non-empty talosctl path")
	}
	if p.workDir == "" {
		t.Fatal("expected non-empty work directory")
	}
	if p.execCommand == nil {
		t.Fatal("expected non-nil execCommand")
	}
}

// Test NewProvisionerWithExec creates provisioner with custom exec
func TestNewProvisionerWithExec(t *testing.T) {
	customExec := func(name string, args ...string) *exec.Cmd {
		return exec.Command("echo", "test")
	}

	p := NewProvisionerWithExec(customExec)
	defer func() { _ = p.Cleanup() }()

	if p == nil {
		t.Fatal("expected non-nil provisioner")
	}
	if p.talosctlPath != "talosctl" {
		t.Errorf("expected talosctlPath 'talosctl', got %q", p.talosctlPath)
	}
}

// Test NodeConfig fields
func TestNodeConfig_Fields(t *testing.T) {
	node := NodeConfig{
		Host:     "192.168.1.100",
		Hostname: "node1",
	}

	if node.Host != "192.168.1.100" {
		t.Errorf("expected host '192.168.1.100', got %q", node.Host)
	}
	if node.Hostname != "node1" {
		t.Errorf("expected hostname 'node1', got %q", node.Hostname)
	}
}

// Test ClusterConfig fields
func TestClusterConfig_Fields(t *testing.T) {
	cfg := ClusterConfig{
		Name:              "test-cluster",
		ClusterEndpoint:   "https://192.168.1.100:6443",
		KubernetesVersion: "v1.31.0",
		InstallDisk:       "/dev/sda",
		ControlPlanes: []NodeConfig{
			{Host: "192.168.1.100", Hostname: "cp1"},
		},
		Workers: []NodeConfig{
			{Host: "192.168.1.101", Hostname: "worker1"},
			{Host: "192.168.1.102", Hostname: "worker2"},
		},
		AllowSchedulingOnCP: true,
	}

	if cfg.Name != "test-cluster" {
		t.Errorf("expected name 'test-cluster', got %q", cfg.Name)
	}
	if len(cfg.ControlPlanes) != 1 {
		t.Errorf("expected 1 control plane node, got %d", len(cfg.ControlPlanes))
	}
	if len(cfg.Workers) != 2 {
		t.Errorf("expected 2 worker nodes, got %d", len(cfg.Workers))
	}
	if !cfg.AllowSchedulingOnCP {
		t.Error("expected AllowSchedulingOnCP to be true")
	}
}

// Test ClusterState fields
func TestClusterState_Fields(t *testing.T) {
	state := ClusterState{
		SecretsYAML:     "secrets-content",
		Talosconfig:     "talosconfig-content",
		Kubeconfig:      "kubeconfig-content",
		APIEndpoint:     "https://192.168.1.100:6443",
		ClusterStatus:   "ready",
		ControlPlaneIPs: []string{"192.168.1.100"},
		WorkerIPs:       []string{"192.168.1.101", "192.168.1.102"},
	}

	if state.ClusterStatus != "ready" {
		t.Errorf("expected status 'ready', got %q", state.ClusterStatus)
	}
	if len(state.ControlPlaneIPs) != 1 {
		t.Errorf("expected 1 control plane IP, got %d", len(state.ControlPlaneIPs))
	}
	if len(state.WorkerIPs) != 2 {
		t.Errorf("expected 2 worker IPs, got %d", len(state.WorkerIPs))
	}
}

// Test Provisioner WorkDir returns correct path
func TestProvisioner_WorkDir(t *testing.T) {
	p := NewProvisionerWithExec(exec.Command)
	defer func() { _ = p.Cleanup() }()

	workDir := p.WorkDir()
	if workDir == "" {
		t.Error("expected non-empty work directory")
	}

	// Verify directory exists
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		t.Error("work directory does not exist")
	}
}

// Test Provisioner Cleanup removes work directory
func TestProvisioner_Cleanup(t *testing.T) {
	p := NewProvisionerWithExec(exec.Command)

	workDir := p.WorkDir()

	// Verify directory exists
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		t.Fatal("work directory should exist before cleanup")
	}

	// Cleanup
	err := p.Cleanup()
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	// Verify directory is removed
	if _, err := os.Stat(workDir); !os.IsNotExist(err) {
		t.Error("work directory should not exist after cleanup")
	}
}

// Test ReadTalosconfig reads file content
func TestProvisioner_ReadTalosconfig(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewProvisionerWithExec(exec.Command)
	defer func() { _ = p.Cleanup() }()

	expectedContent := "context: test\ncontexts:\n  test:\n    endpoints:\n      - 192.168.1.100\n"
	filePath := filepath.Join(tmpDir, "talosconfig")
	err := os.WriteFile(filePath, []byte(expectedContent), 0600)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	content, err := p.ReadTalosconfig(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != expectedContent {
		t.Errorf("expected content %q, got %q", expectedContent, content)
	}
}

// Test ReadSecrets reads file content
func TestProvisioner_ReadSecrets(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewProvisionerWithExec(exec.Command)
	defer func() { _ = p.Cleanup() }()

	expectedContent := "secrets:\n  bootstrap_token: abc123\n"
	filePath := filepath.Join(tmpDir, "secrets.yaml")
	err := os.WriteFile(filePath, []byte(expectedContent), 0600)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	content, err := p.ReadSecrets(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != expectedContent {
		t.Errorf("expected content %q, got %q", expectedContent, content)
	}
}

// Test ReadTalosconfig file not found
func TestProvisioner_ReadTalosconfig_NotFound(t *testing.T) {
	p := NewProvisionerWithExec(exec.Command)
	defer func() { _ = p.Cleanup() }()

	_, err := p.ReadTalosconfig("/nonexistent/file")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

// Test ClusterConfig validation
func TestClusterConfig_Validation(t *testing.T) {
	// Valid config
	cfg := ClusterConfig{
		Name:            "test",
		ClusterEndpoint: "https://192.168.1.100:6443",
		ControlPlanes: []NodeConfig{
			{Host: "192.168.1.100"},
		},
	}

	if cfg.Name == "" {
		t.Error("name should not be empty")
	}
	if len(cfg.ControlPlanes) < 1 {
		t.Error("need at least one control plane node")
	}
}

// Test generatePatchYAML helper
func TestGeneratePatchYAML(t *testing.T) {
	// Test control plane with scheduling allowed
	patch, err := generatePatchYAML("cp1", true, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if patch == "" {
		t.Error("expected non-empty patch")
	}

	// Test worker (no scheduling option)
	patch, err = generatePatchYAML("worker1", false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if patch == "" {
		t.Error("expected non-empty patch")
	}
}

// Test with mock exec that simulates talosctl commands
func TestProvisioner_GenerateSecrets_MockExec(t *testing.T) {
	tmpDir := t.TempDir()

	mockExec := func(name string, args ...string) *exec.Cmd {
		// Simulate generating secrets
		secretsFile := filepath.Join(tmpDir, "secrets.yaml")
		os.WriteFile(secretsFile, []byte("secrets: test"), 0600)
		return exec.Command("true")
	}

	p := NewProvisionerWithExec(mockExec)
	// Manually set workDir for this test
	p.workDir = tmpDir
	defer func() { p.workDir = ""; _ = p.Cleanup() }()

	secretsPath := filepath.Join(tmpDir, "secrets.yaml")
	err := p.GenerateSecrets(secretsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(secretsPath); os.IsNotExist(err) {
		t.Error("expected secrets.yaml to be created")
	}
}

// Test runTalosctl error handling
func TestProvisioner_runTalosctl_Error(t *testing.T) {
	mockExec := func(name string, args ...string) *exec.Cmd {
		return exec.Command("false") // Always exits with error
	}

	p := NewProvisionerWithExec(mockExec)
	defer func() { _ = p.Cleanup() }()

	_, err := p.runTalosctl("test", "command")
	if err == nil {
		t.Fatal("expected error from failed command")
	}
}
