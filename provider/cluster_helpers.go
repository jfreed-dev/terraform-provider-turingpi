package provider

import (
	"fmt"
	"time"
)

// WaitForSSH polls until SSH is available on a host
// Returns nil when SSH connection succeeds, or error on timeout
func WaitForSSH(host string, port int, config *SSHConfig, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	var lastErr error
	for time.Now().Before(deadline) {
		client := NewSSHClient()
		err := client.Connect(host, port, config)
		if err == nil {
			_ = client.Close()
			return nil
		}
		lastErr = err
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("timeout waiting for SSH on %s:%d after %v: %w", host, port, timeout, lastErr)
}

// WaitForSSHWithClient polls until SSH is available using a custom client factory
// Useful for testing with mock clients
func WaitForSSHWithClient(host string, port int, config *SSHConfig, timeout time.Duration, clientFactory func() SSHClient) error {
	deadline := time.Now().Add(timeout)

	var lastErr error
	for time.Now().Before(deadline) {
		client := clientFactory()
		err := client.Connect(host, port, config)
		if err == nil {
			_ = client.Close()
			return nil
		}
		lastErr = err
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("timeout waiting for SSH on %s:%d after %v: %w", host, port, timeout, lastErr)
}

// RunSSHCommand executes a command over SSH and returns output
func RunSSHCommand(host string, port int, config *SSHConfig, command string) (string, error) {
	client := NewSSHClient()
	if err := client.Connect(host, port, config); err != nil {
		return "", fmt.Errorf("SSH connection failed: %w", err)
	}
	defer func() { _ = client.Close() }()

	output, err := client.RunCommand(command)
	if err != nil {
		return output, fmt.Errorf("command execution failed: %w", err)
	}

	return output, nil
}

// RunSSHCommandWithClient executes a command using a custom client
// Useful for testing with mock clients
func RunSSHCommandWithClient(host string, port int, config *SSHConfig, command string, client SSHClient) (string, error) {
	if err := client.Connect(host, port, config); err != nil {
		return "", fmt.Errorf("SSH connection failed: %w", err)
	}
	defer func() { _ = client.Close() }()

	output, err := client.RunCommand(command)
	if err != nil {
		return output, fmt.Errorf("command execution failed: %w", err)
	}

	return output, nil
}

// CheckSSHConnectivity tests if SSH is available (single attempt, no retry)
func CheckSSHConnectivity(host string, port int, config *SSHConfig) bool {
	client := NewSSHClient()
	err := client.Connect(host, port, config)
	if err != nil {
		return false
	}
	_ = client.Close()
	return true
}
