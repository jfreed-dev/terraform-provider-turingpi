// Package helm provides Helm client functionality for chart deployment.
package helm

import (
	"context"
	"fmt"
	"os"
	"time"

	helmclient "github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
)

// Client interface for Helm operations - allows mocking in tests
type Client interface {
	AddRepository(name, url string) error
	UpdateRepositories() error
	InstallOrUpgradeChart(ctx context.Context, spec *ChartSpec) (*release.Release, error)
	UninstallRelease(name string) error
	GetRelease(name string) (*release.Release, error)
	ListReleases() ([]*release.Release, error)
}

// ChartSpec defines a Helm chart deployment specification
type ChartSpec struct {
	ReleaseName     string                 // Name of the Helm release
	ChartName       string                 // Chart name (e.g., "metallb/metallb" or local path)
	Namespace       string                 // Target namespace
	Version         string                 // Chart version (optional, empty = latest)
	Values          map[string]interface{} // Inline values
	ValuesYaml      string                 // YAML string of values
	CreateNamespace bool                   // Create namespace if it doesn't exist
	Wait            bool                   // Wait for resources to be ready
	Timeout         time.Duration          // Timeout for wait operations
	Atomic          bool                   // Rollback on failure
}

// RealClient implements Client using mittwald/go-helm-client
type RealClient struct {
	client    helmclient.Client
	namespace string
}

// NewClient creates a new Helm client from a kubeconfig file path
func NewClient(kubeconfigPath, namespace string) (Client, error) {
	kubeconfig, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read kubeconfig: %w", err)
	}

	return NewClientFromBytes(kubeconfig, namespace)
}

// NewClientFromBytes creates a new Helm client from kubeconfig bytes
func NewClientFromBytes(kubeconfig []byte, namespace string) (Client, error) {
	if namespace == "" {
		namespace = "default"
	}

	opt := &helmclient.KubeConfClientOptions{
		Options: &helmclient.Options{
			Namespace:        namespace,
			RepositoryCache:  "/tmp/.helmcache",
			RepositoryConfig: "/tmp/.helmrepo",
			Debug:            false,
			Linting:          false,
		},
		KubeConfig: kubeconfig,
	}

	client, err := helmclient.NewClientFromKubeConf(opt)
	if err != nil {
		return nil, fmt.Errorf("failed to create Helm client: %w", err)
	}

	return &RealClient{
		client:    client,
		namespace: namespace,
	}, nil
}

// AddRepository adds or updates a Helm chart repository
func (c *RealClient) AddRepository(name, url string) error {
	chartRepo := repo.Entry{
		Name: name,
		URL:  url,
	}

	if err := c.client.AddOrUpdateChartRepo(chartRepo); err != nil {
		return fmt.Errorf("failed to add repository %s: %w", name, err)
	}

	return nil
}

// UpdateRepositories updates all configured Helm repositories
func (c *RealClient) UpdateRepositories() error {
	if err := c.client.UpdateChartRepos(); err != nil {
		return fmt.Errorf("failed to update repositories: %w", err)
	}
	return nil
}

// InstallOrUpgradeChart installs or upgrades a Helm chart
func (c *RealClient) InstallOrUpgradeChart(ctx context.Context, spec *ChartSpec) (*release.Release, error) {
	if spec.Timeout == 0 {
		spec.Timeout = 5 * time.Minute
	}

	chartSpec := helmclient.ChartSpec{
		ReleaseName:     spec.ReleaseName,
		ChartName:       spec.ChartName,
		Namespace:       spec.Namespace,
		Version:         spec.Version,
		ValuesYaml:      spec.ValuesYaml,
		CreateNamespace: spec.CreateNamespace,
		Wait:            spec.Wait,
		Timeout:         spec.Timeout,
		Atomic:          spec.Atomic,
		CleanupOnFail:   spec.Atomic, // Clean up on failure if atomic
	}

	rel, err := c.client.InstallOrUpgradeChart(ctx, &chartSpec, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to install/upgrade chart %s: %w", spec.ChartName, err)
	}

	return rel, nil
}

// UninstallRelease uninstalls a Helm release
func (c *RealClient) UninstallRelease(name string) error {
	if err := c.client.UninstallReleaseByName(name); err != nil {
		return fmt.Errorf("failed to uninstall release %s: %w", name, err)
	}
	return nil
}

// GetRelease returns information about an installed release
func (c *RealClient) GetRelease(name string) (*release.Release, error) {
	rel, err := c.client.GetRelease(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get release %s: %w", name, err)
	}
	return rel, nil
}

// ListReleases lists all releases in the configured namespace
func (c *RealClient) ListReleases() ([]*release.Release, error) {
	releases, err := c.client.ListDeployedReleases()
	if err != nil {
		return nil, fmt.Errorf("failed to list releases: %w", err)
	}
	return releases, nil
}

// DeployChart is a high-level convenience function for deploying a chart
func DeployChart(ctx context.Context, kubeconfigPath string, spec *ChartSpec) error {
	client, err := NewClient(kubeconfigPath, spec.Namespace)
	if err != nil {
		return err
	}

	_, err = client.InstallOrUpgradeChart(ctx, spec)
	return err
}

// DeployChartWithClient deploys a chart using a provided client (for testing)
func DeployChartWithClient(ctx context.Context, client Client, spec *ChartSpec) error {
	_, err := client.InstallOrUpgradeChart(ctx, spec)
	return err
}

// DeployFromRepository adds a repo and deploys a chart in one call
func DeployFromRepository(ctx context.Context, kubeconfigPath, repoName, repoURL string, spec *ChartSpec) error {
	client, err := NewClient(kubeconfigPath, spec.Namespace)
	if err != nil {
		return err
	}

	if err := client.AddRepository(repoName, repoURL); err != nil {
		return fmt.Errorf("failed to add repository: %w", err)
	}

	_, err = client.InstallOrUpgradeChart(ctx, spec)
	return err
}

// DeployFromRepositoryWithClient adds a repo and deploys using a provided client (for testing)
func DeployFromRepositoryWithClient(ctx context.Context, client Client, repoName, repoURL string, spec *ChartSpec) error {
	if err := client.AddRepository(repoName, repoURL); err != nil {
		return fmt.Errorf("failed to add repository: %w", err)
	}

	_, err := client.InstallOrUpgradeChart(ctx, spec)
	return err
}

// WaitForRelease waits for a release to reach deployed status
func WaitForRelease(kubeconfigPath, name, namespace string, timeout time.Duration) error {
	client, err := NewClient(kubeconfigPath, namespace)
	if err != nil {
		return err
	}

	return WaitForReleaseWithClient(client, name, timeout)
}

// WaitForReleaseWithClient waits for a release using a provided client (for testing)
func WaitForReleaseWithClient(client Client, name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		rel, err := client.GetRelease(name)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		switch rel.Info.Status {
		case release.StatusDeployed:
			return nil
		case release.StatusFailed:
			return fmt.Errorf("release %s failed: %s", name, rel.Info.Description)
		case release.StatusPendingInstall, release.StatusPendingUpgrade, release.StatusPendingRollback:
			// Still in progress, keep waiting
		default:
			// Unknown status, keep waiting
		}

		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("timeout waiting for release %s after %v", name, timeout)
}
