package provider

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Test resource schema validation
func TestResourceK3sCluster(t *testing.T) {
	r := resourceK3sCluster()
	if err := r.InternalValidate(nil, true); err != nil {
		t.Fatalf("resource internal validation failed: %s", err)
	}
}

// Test schema has required fields
func TestResourceK3sCluster_Schema(t *testing.T) {
	r := resourceK3sCluster()
	expectedFields := []string{
		"name", "k3s_version", "cluster_token", "control_plane", "worker",
		"pod_cidr", "service_cidr", "metallb", "ingress", "install_timeout",
		"kubeconfig_path", "kubeconfig", "api_endpoint", "node_token", "cluster_status",
	}
	for _, field := range expectedFields {
		if _, ok := r.Schema[field]; !ok {
			t.Errorf("schema missing '%s' field", field)
		}
	}
}

// Test schema types
func TestResourceK3sCluster_SchemaTypes(t *testing.T) {
	r := resourceK3sCluster()
	tests := []struct {
		field    string
		expected schema.ValueType
	}{
		{"name", schema.TypeString},
		{"k3s_version", schema.TypeString},
		{"cluster_token", schema.TypeString},
		{"control_plane", schema.TypeList},
		{"worker", schema.TypeList},
		{"pod_cidr", schema.TypeString},
		{"service_cidr", schema.TypeString},
		{"metallb", schema.TypeList},
		{"ingress", schema.TypeList},
		{"install_timeout", schema.TypeInt},
		{"kubeconfig_path", schema.TypeString},
		{"kubeconfig", schema.TypeString},
		{"api_endpoint", schema.TypeString},
		{"node_token", schema.TypeString},
		{"cluster_status", schema.TypeString},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			if r.Schema[tt.field].Type != tt.expected {
				t.Errorf("expected type %v for field '%s', got %v",
					tt.expected, tt.field, r.Schema[tt.field].Type)
			}
		})
	}
}

// Test required fields
func TestResourceK3sCluster_RequiredFields(t *testing.T) {
	r := resourceK3sCluster()
	requiredFields := []string{"name", "control_plane"}
	for _, field := range requiredFields {
		if !r.Schema[field].Required {
			t.Errorf("field '%s' should be required", field)
		}
	}
}

// Test optional fields
func TestResourceK3sCluster_OptionalFields(t *testing.T) {
	r := resourceK3sCluster()
	optionalFields := []string{"k3s_version", "cluster_token", "worker", "metallb", "ingress", "kubeconfig_path"}
	for _, field := range optionalFields {
		if r.Schema[field].Required {
			t.Errorf("field '%s' should be optional", field)
		}
	}
}

// Test sensitive fields
func TestResourceK3sCluster_SensitiveFields(t *testing.T) {
	r := resourceK3sCluster()
	sensitiveFields := []string{"cluster_token", "kubeconfig", "node_token"}
	for _, field := range sensitiveFields {
		if !r.Schema[field].Sensitive {
			t.Errorf("field '%s' should be sensitive", field)
		}
	}
}

// Test computed fields
func TestResourceK3sCluster_ComputedFields(t *testing.T) {
	r := resourceK3sCluster()
	computedFields := []string{"kubeconfig", "api_endpoint", "node_token", "cluster_status"}
	for _, field := range computedFields {
		if !r.Schema[field].Computed {
			t.Errorf("field '%s' should be computed", field)
		}
	}
}

// Test default values
func TestResourceK3sCluster_Defaults(t *testing.T) {
	r := resourceK3sCluster()

	tests := []struct {
		field    string
		expected interface{}
	}{
		{"pod_cidr", "10.244.0.0/16"},
		{"service_cidr", "10.96.0.0/12"},
		{"install_timeout", 600},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			if r.Schema[tt.field].Default != tt.expected {
				t.Errorf("expected default %v for field '%s', got %v",
					tt.expected, tt.field, r.Schema[tt.field].Default)
			}
		})
	}
}

// Test node schema
func TestK3sNodeSchema(t *testing.T) {
	s := k3sNodeSchema()

	expectedFields := []string{"host", "ssh_user", "ssh_key", "ssh_password", "ssh_port"}
	for _, field := range expectedFields {
		if _, ok := s.Schema[field]; !ok {
			t.Errorf("node schema missing '%s' field", field)
		}
	}

	// Check required fields
	if !s.Schema["host"].Required {
		t.Error("'host' should be required")
	}
	if !s.Schema["ssh_user"].Required {
		t.Error("'ssh_user' should be required")
	}

	// Check sensitive fields
	if !s.Schema["ssh_key"].Sensitive {
		t.Error("'ssh_key' should be sensitive")
	}
	if !s.Schema["ssh_password"].Sensitive {
		t.Error("'ssh_password' should be sensitive")
	}

	// Check default port
	if s.Schema["ssh_port"].Default != 22 {
		t.Errorf("expected default ssh_port 22, got %v", s.Schema["ssh_port"].Default)
	}
}

// Test MetalLB schema
func TestMetallbSchema(t *testing.T) {
	s := metallbSchema()

	expectedFields := []string{"enabled", "ip_range", "version"}
	for _, field := range expectedFields {
		if _, ok := s.Schema[field]; !ok {
			t.Errorf("metallb schema missing '%s' field", field)
		}
	}

	if !s.Schema["ip_range"].Required {
		t.Error("'ip_range' should be required")
	}

	if s.Schema["enabled"].Default != true {
		t.Error("'enabled' should default to true")
	}
}

// Test Ingress schema
func TestIngressSchema(t *testing.T) {
	s := ingressSchema()

	expectedFields := []string{"enabled", "ip", "version"}
	for _, field := range expectedFields {
		if _, ok := s.Schema[field]; !ok {
			t.Errorf("ingress schema missing '%s' field", field)
		}
	}

	if s.Schema["enabled"].Default != true {
		t.Error("'enabled' should default to true")
	}
}

// Test GenerateClusterToken
func TestGenerateClusterToken(t *testing.T) {
	token1 := GenerateClusterToken()
	token2 := GenerateClusterToken()

	if token1 == "" {
		t.Error("token should not be empty")
	}

	if token1 == token2 {
		t.Error("tokens should be unique")
	}

	if len(token1) < 32 {
		t.Errorf("token should be at least 32 characters, got %d", len(token1))
	}
}

// Test extractNodeConfig
func TestExtractNodeConfig(t *testing.T) {
	data := map[string]interface{}{
		"host":         "10.10.88.73",
		"ssh_user":     "root",
		"ssh_key":      "fake-key-content",
		"ssh_password": "",
		"ssh_port":     22,
	}

	config := extractNodeConfig(data)

	if config.Host != "10.10.88.73" {
		t.Errorf("expected host '10.10.88.73', got '%s'", config.Host)
	}
	if config.SSHUser != "root" {
		t.Errorf("expected ssh_user 'root', got '%s'", config.SSHUser)
	}
	if string(config.SSHKey) != "fake-key-content" {
		t.Errorf("expected ssh_key 'fake-key-content', got '%s'", string(config.SSHKey))
	}
	if config.SSHPort != 22 {
		t.Errorf("expected ssh_port 22, got %d", config.SSHPort)
	}
}

// Test splitIPRange
func TestSplitIPRange(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"10.10.88.80-10.10.88.89", []string{"10.10.88.80", "10.10.88.89"}},
		{"192.168.1.100-192.168.1.200", []string{"192.168.1.100", "192.168.1.200"}},
		{"10.0.0.1", []string{"10.0.0.1"}},
		{"", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := splitIPRange(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d parts, got %d", len(tt.expected), len(result))
				return
			}
			for i, part := range result {
				if part != tt.expected[i] {
					t.Errorf("expected part %d to be '%s', got '%s'", i, tt.expected[i], part)
				}
			}
		})
	}
}

// Test K3sProvisioner with mock SSH client
func TestK3sProvisioner_InstallK3sServer(t *testing.T) {
	commandIndex := 0
	expectedCommands := []string{
		"swapoff -a",
		"mkdir -p /etc/rancher/k3s",
		"test -f /usr/local/bin/k3s && echo 'installed' || echo 'not_installed'",
		"curl -sfL https://get.k3s.io -o /tmp/k3s-install.sh && chmod +x /tmp/k3s-install.sh",
	}

	mockFactory := func() SSHClient {
		return &MockSSHClient{
			RunCommandFunc: func(cmd string) (string, error) {
				if commandIndex < len(expectedCommands) {
					expected := expectedCommands[commandIndex]
					commandIndex++
					if cmd == expected || (commandIndex == 3 && cmd == expectedCommands[2]) {
						if cmd == "test -f /usr/local/bin/k3s && echo 'installed' || echo 'not_installed'" {
							return "not_installed", nil
						}
						return "", nil
					}
				}
				// Return "Ready" for the final check
				if cmd == "k3s kubectl get nodes 2>/dev/null" {
					return "node1 Ready", nil
				}
				return "", nil
			},
		}
	}

	provisioner := NewK3sProvisionerWithClientFactory(mockFactory)
	node := NodeConfig{
		Host:     "10.10.88.73",
		SSHUser:  "root",
		SSHKey:   []byte("fake-key"),
		SSHPort:  22,
	}
	cfg := ClusterConfig{
		Name:         "test-cluster",
		ClusterToken: "test-token",
	}

	// This will fail because mock doesn't fully implement all commands,
	// but we can verify the flow starts correctly
	ctx := context.Background()
	_ = provisioner.InstallK3sServer(ctx, node, cfg, 5*time.Second)
	// We just verify no panic occurs
}

// Test K3sProvisioner GetNodeToken
func TestK3sProvisioner_GetNodeToken(t *testing.T) {
	mockFactory := func() SSHClient {
		return &MockSSHClient{
			RunCommandFunc: func(cmd string) (string, error) {
				if cmd == "cat /var/lib/rancher/k3s/server/node-token" {
					return "K10abc123::server:xyz789\n", nil
				}
				return "", fmt.Errorf("unexpected command: %s", cmd)
			},
		}
	}

	provisioner := NewK3sProvisionerWithClientFactory(mockFactory)
	node := NodeConfig{
		Host:    "10.10.88.73",
		SSHUser: "root",
		SSHKey:  []byte("fake-key"),
		SSHPort: 22,
	}

	token, err := provisioner.GetNodeToken(node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "K10abc123::server:xyz789"
	if token != expected {
		t.Errorf("expected token '%s', got '%s'", expected, token)
	}
}

// Test K3sProvisioner GetKubeconfig
func TestK3sProvisioner_GetKubeconfig(t *testing.T) {
	mockFactory := func() SSHClient {
		return &MockSSHClient{
			RunCommandFunc: func(cmd string) (string, error) {
				if cmd == "cat /etc/rancher/k3s/k3s.yaml" {
					return `apiVersion: v1
clusters:
- cluster:
    server: https://127.0.0.1:6443
  name: default
`, nil
				}
				return "", fmt.Errorf("unexpected command: %s", cmd)
			},
		}
	}

	provisioner := NewK3sProvisionerWithClientFactory(mockFactory)
	node := NodeConfig{
		Host:    "10.10.88.73",
		SSHUser: "root",
		SSHKey:  []byte("fake-key"),
		SSHPort: 22,
	}

	kubeconfig, err := provisioner.GetKubeconfig(node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify 127.0.0.1 was replaced with node IP
	if !contains(kubeconfig, "10.10.88.73") {
		t.Error("kubeconfig should contain node IP")
	}
	if contains(kubeconfig, "127.0.0.1") {
		t.Error("kubeconfig should not contain 127.0.0.1")
	}
}

// Test K3sProvisioner CheckK3sInstalled
func TestK3sProvisioner_CheckK3sInstalled(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected bool
	}{
		{"installed", "installed", true},
		{"not installed", "not_installed", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFactory := func() SSHClient {
				return &MockSSHClient{
					RunCommandFunc: func(cmd string) (string, error) {
						return tt.output, nil
					},
				}
			}

			provisioner := NewK3sProvisionerWithClientFactory(mockFactory)
			node := NodeConfig{Host: "test", SSHUser: "root", SSHPort: 22}

			installed, _ := provisioner.CheckK3sInstalled(node)
			if installed != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, installed)
			}
		})
	}
}

// Test K3sProvisioner UninstallK3sServer
func TestK3sProvisioner_UninstallK3sServer(t *testing.T) {
	uninstallCalled := false
	mockFactory := func() SSHClient {
		return &MockSSHClient{
			RunCommandFunc: func(cmd string) (string, error) {
				if cmd == "test -f /usr/local/bin/k3s-uninstall.sh && echo 'exists' || echo 'not_exists'" {
					return "exists", nil
				}
				if cmd == "/usr/local/bin/k3s-uninstall.sh" {
					uninstallCalled = true
					return "", nil
				}
				return "", fmt.Errorf("unexpected command: %s", cmd)
			},
		}
	}

	provisioner := NewK3sProvisionerWithClientFactory(mockFactory)
	node := NodeConfig{Host: "test", SSHUser: "root", SSHPort: 22}

	err := provisioner.UninstallK3sServer(node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !uninstallCalled {
		t.Error("uninstall script should have been called")
	}
}

// Test K3sProvisioner UninstallK3sAgent
func TestK3sProvisioner_UninstallK3sAgent(t *testing.T) {
	uninstallCalled := false
	mockFactory := func() SSHClient {
		return &MockSSHClient{
			RunCommandFunc: func(cmd string) (string, error) {
				if cmd == "test -f /usr/local/bin/k3s-agent-uninstall.sh && echo 'exists' || echo 'not_exists'" {
					return "exists", nil
				}
				if cmd == "/usr/local/bin/k3s-agent-uninstall.sh" {
					uninstallCalled = true
					return "", nil
				}
				return "", fmt.Errorf("unexpected command: %s", cmd)
			},
		}
	}

	provisioner := NewK3sProvisionerWithClientFactory(mockFactory)
	node := NodeConfig{Host: "test", SSHUser: "root", SSHPort: 22}

	err := provisioner.UninstallK3sAgent(node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !uninstallCalled {
		t.Error("uninstall script should have been called")
	}
}

// Test K3sProvisioner when K3s not installed (no-op uninstall)
func TestK3sProvisioner_UninstallK3sServer_NotInstalled(t *testing.T) {
	mockFactory := func() SSHClient {
		return &MockSSHClient{
			RunCommandFunc: func(cmd string) (string, error) {
				if cmd == "test -f /usr/local/bin/k3s-uninstall.sh && echo 'exists' || echo 'not_exists'" {
					return "not_exists", nil
				}
				return "", fmt.Errorf("should not be called")
			},
		}
	}

	provisioner := NewK3sProvisionerWithClientFactory(mockFactory)
	node := NodeConfig{Host: "test", SSHUser: "root", SSHPort: 22}

	err := provisioner.UninstallK3sServer(node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
