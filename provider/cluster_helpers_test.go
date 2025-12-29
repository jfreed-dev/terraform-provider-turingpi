package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// MockSSHClient implements SSHClient for testing
type MockSSHClient struct {
	ConnectFunc    func(host string, port int, config *SSHConfig) error
	RunCommandFunc func(cmd string) (string, error)
	CloseFunc      func() error
	connected      bool
}

func (m *MockSSHClient) Connect(host string, port int, config *SSHConfig) error {
	if m.ConnectFunc != nil {
		err := m.ConnectFunc(host, port, config)
		if err == nil {
			m.connected = true
		}
		return err
	}
	m.connected = true
	return nil
}

func (m *MockSSHClient) RunCommand(cmd string) (string, error) {
	if !m.connected {
		return "", fmt.Errorf("not connected")
	}
	if m.RunCommandFunc != nil {
		return m.RunCommandFunc(cmd)
	}
	return "", nil
}

func (m *MockSSHClient) Close() error {
	m.connected = false
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// Test SSHConfig validation
func TestSSHConfig_Defaults(t *testing.T) {
	config := &SSHConfig{
		User:     "testuser",
		Password: "testpass",
	}

	if config.User != "testuser" {
		t.Errorf("expected user 'testuser', got %q", config.User)
	}
	if config.Timeout != 0 {
		t.Errorf("expected zero timeout (will use default), got %v", config.Timeout)
	}
	if config.HostKeyCheck {
		t.Error("expected HostKeyCheck to be false by default")
	}
}

// Test WaitForSSH success case
func TestWaitForSSH_Success(t *testing.T) {
	callCount := 0
	mockFactory := func() SSHClient {
		return &MockSSHClient{
			ConnectFunc: func(host string, port int, config *SSHConfig) error {
				callCount++
				return nil // Success on first try
			},
		}
	}

	config := &SSHConfig{User: "test", Password: "test"}
	err := WaitForSSHWithClient("localhost", 22, config, 10*time.Second, mockFactory)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 connection attempt, got %d", callCount)
	}
}

// Test WaitForSSH with retry
func TestWaitForSSH_SuccessAfterRetry(t *testing.T) {
	callCount := 0
	mockFactory := func() SSHClient {
		return &MockSSHClient{
			ConnectFunc: func(host string, port int, config *SSHConfig) error {
				callCount++
				if callCount < 2 {
					return fmt.Errorf("connection refused")
				}
				return nil // Success on second try
			},
		}
	}

	config := &SSHConfig{User: "test", Password: "test"}
	// Use short timeout since we sleep 5s between retries
	err := WaitForSSHWithClient("localhost", 22, config, 10*time.Second, mockFactory)
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if callCount < 2 {
		t.Errorf("expected at least 2 connection attempts, got %d", callCount)
	}
}

// Test WaitForSSH timeout
func TestWaitForSSH_Timeout(t *testing.T) {
	mockFactory := func() SSHClient {
		return &MockSSHClient{
			ConnectFunc: func(host string, port int, config *SSHConfig) error {
				return fmt.Errorf("connection refused")
			},
		}
	}

	config := &SSHConfig{User: "test", Password: "test"}
	err := WaitForSSHWithClient("localhost", 22, config, 1*time.Second, mockFactory)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

// Test RunSSHCommand success
func TestRunSSHCommand_Success(t *testing.T) {
	mock := &MockSSHClient{
		RunCommandFunc: func(cmd string) (string, error) {
			if cmd == "echo hello" {
				return "hello\n", nil
			}
			return "", fmt.Errorf("unknown command")
		},
	}

	config := &SSHConfig{User: "test", Password: "test"}
	output, err := RunSSHCommandWithClient("localhost", 22, config, "echo hello", mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", output)
	}
}

// Test RunSSHCommand connection failure
func TestRunSSHCommand_ConnectionFailed(t *testing.T) {
	mock := &MockSSHClient{
		ConnectFunc: func(host string, port int, config *SSHConfig) error {
			return fmt.Errorf("connection refused")
		},
	}

	config := &SSHConfig{User: "test", Password: "test"}
	_, err := RunSSHCommandWithClient("localhost", 22, config, "echo hello", mock)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// Test RunSSHCommand command failure
func TestRunSSHCommand_CommandFailed(t *testing.T) {
	mock := &MockSSHClient{
		RunCommandFunc: func(cmd string) (string, error) {
			return "error output", fmt.Errorf("command exited with status 1")
		},
	}

	config := &SSHConfig{User: "test", Password: "test"}
	output, err := RunSSHCommandWithClient("localhost", 22, config, "false", mock)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Output should still be returned even on error
	if output != "error output" {
		t.Errorf("expected error output to be returned, got %q", output)
	}
}

// Test CheckSSHConnectivity
func TestCheckSSHConnectivity(t *testing.T) {
	// We can't easily test this without modifying the function to accept a client factory
	// This is more of an integration test placeholder
	t.Skip("CheckSSHConnectivity requires real SSH server or refactoring to accept mock")
}

// Test kubeconfig parsing with a test fixture
func TestExtractClusterEndpoint(t *testing.T) {
	// Create a temporary kubeconfig file
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "kubeconfig")

	kubeconfig := `apiVersion: v1
kind: Config
current-context: test-context
clusters:
- cluster:
    server: https://192.168.1.100:6443
    certificate-authority-data: dGVzdA==
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
users:
- name: test-user
  user:
    token: test-token
`

	err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0600)
	if err != nil {
		t.Fatalf("failed to write test kubeconfig: %v", err)
	}

	endpoint, err := ExtractClusterEndpoint(kubeconfigPath)
	if err != nil {
		t.Fatalf("failed to extract endpoint: %v", err)
	}

	expected := "https://192.168.1.100:6443"
	if endpoint != expected {
		t.Errorf("expected endpoint %q, got %q", expected, endpoint)
	}
}

// Test kubeconfig with missing context
func TestExtractClusterEndpoint_NoContext(t *testing.T) {
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "kubeconfig")

	kubeconfig := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://192.168.1.100:6443
  name: test-cluster
`

	err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0600)
	if err != nil {
		t.Fatalf("failed to write test kubeconfig: %v", err)
	}

	_, err = ExtractClusterEndpoint(kubeconfigPath)
	if err == nil {
		t.Fatal("expected error for kubeconfig without current context")
	}
}

// Test kubeconfig file not found
func TestExtractClusterEndpoint_FileNotFound(t *testing.T) {
	_, err := ExtractClusterEndpoint("/nonexistent/kubeconfig")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

// Test LoadKubeconfig with valid file
func TestLoadKubeconfig(t *testing.T) {
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "kubeconfig")

	kubeconfig := `apiVersion: v1
kind: Config
current-context: test-context
clusters:
- cluster:
    server: https://192.168.1.100:6443
    insecure-skip-tls-verify: true
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
users:
- name: test-user
  user:
    token: test-token
`

	err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0600)
	if err != nil {
		t.Fatalf("failed to write test kubeconfig: %v", err)
	}

	config, err := LoadKubeconfig(kubeconfigPath)
	if err != nil {
		t.Fatalf("failed to load kubeconfig: %v", err)
	}

	if config.Host != "https://192.168.1.100:6443" {
		t.Errorf("expected host 'https://192.168.1.100:6443', got %q", config.Host)
	}
}

// Test SSHClient interface validation
func TestRealSSHClient_ImplementsInterface(t *testing.T) {
	// Compile-time check that RealSSHClient implements SSHClient
	var _ SSHClient = (*RealSSHClient)(nil)
}

// Test MockSSHClient implements interface
func TestMockSSHClient_ImplementsInterface(t *testing.T) {
	var _ SSHClient = (*MockSSHClient)(nil)
}
