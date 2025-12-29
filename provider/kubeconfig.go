package provider

import (
	"fmt"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// LoadKubeconfig parses a kubeconfig file and returns a rest.Config
func LoadKubeconfig(path string) (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
	}
	return config, nil
}

// ValidateKubeconfig checks if a kubeconfig is valid and can connect to the cluster
func ValidateKubeconfig(path string) error {
	config, err := LoadKubeconfig(path)
	if err != nil {
		return err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	_, err = client.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to cluster: %w", err)
	}

	return nil
}

// ExtractClusterEndpoint returns the server URL from a kubeconfig file
func ExtractClusterEndpoint(path string) (string, error) {
	config, err := clientcmd.LoadFromFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Get current context's cluster
	ctx := config.Contexts[config.CurrentContext]
	if ctx == nil {
		return "", fmt.Errorf("no current context in kubeconfig")
	}

	cluster := config.Clusters[ctx.Cluster]
	if cluster == nil {
		return "", fmt.Errorf("cluster %q not found in kubeconfig", ctx.Cluster)
	}

	return cluster.Server, nil
}

// WaitForKubeAPI polls until Kubernetes API responds
func WaitForKubeAPI(kubeconfigPath string, timeout time.Duration) error {
	config, err := LoadKubeconfig(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		_, err := client.Discovery().ServerVersion()
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("timeout waiting for Kubernetes API after %v: %w", timeout, lastErr)
}

// WaitForKubeAPIWithConfig polls until Kubernetes API responds using a pre-loaded config
func WaitForKubeAPIWithConfig(config *rest.Config, timeout time.Duration) error {
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		_, err := client.Discovery().ServerVersion()
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("timeout waiting for Kubernetes API after %v: %w", timeout, lastErr)
}

// GetKubernetesVersion returns the server version from a kubeconfig
func GetKubernetesVersion(path string) (string, error) {
	config, err := LoadKubeconfig(path)
	if err != nil {
		return "", err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	version, err := client.Discovery().ServerVersion()
	if err != nil {
		return "", fmt.Errorf("failed to get server version: %w", err)
	}

	return version.GitVersion, nil
}
