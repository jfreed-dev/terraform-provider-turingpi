package kubeconfig

import (
	"os"
	"path/filepath"
	"testing"
)

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

// Test Load with valid file
func TestLoad(t *testing.T) {
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

	config, err := Load(kubeconfigPath)
	if err != nil {
		t.Fatalf("failed to load kubeconfig: %v", err)
	}

	if config.Host != "https://192.168.1.100:6443" {
		t.Errorf("expected host 'https://192.168.1.100:6443', got %q", config.Host)
	}
}
