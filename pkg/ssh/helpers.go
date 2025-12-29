package ssh

import (
	"fmt"
	"time"
)

// WaitForSSH polls until SSH is available on a host
// Returns nil when SSH connection succeeds, or error on timeout
func WaitForSSH(host string, port int, config *Config, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	var lastErr error
	for time.Now().Before(deadline) {
		client := NewClient()
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
func WaitForSSHWithClient(host string, port int, config *Config, timeout time.Duration, clientFactory func() Client) error {
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

// RunCommand executes a command over SSH and returns output
func RunCommand(host string, port int, config *Config, command string) (string, error) {
	client := NewClient()
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

// RunCommandWithClient executes a command using a custom client
// Useful for testing with mock clients
func RunCommandWithClient(host string, port int, config *Config, command string, client Client) (string, error) {
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

// CheckConnectivity tests if SSH is available (single attempt, no retry)
func CheckConnectivity(host string, port int, config *Config) bool {
	client := NewClient()
	err := client.Connect(host, port, config)
	if err != nil {
		return false
	}
	_ = client.Close()
	return true
}
