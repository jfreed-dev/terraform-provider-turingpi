package provider

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHConfig holds configuration for SSH connections
type SSHConfig struct {
	User           string        // SSH username
	Password       string        // Password authentication (fallback)
	PrivateKey     []byte        // Private key bytes (preferred)
	PrivateKeyPath string        // Path to private key file
	Timeout        time.Duration // Connection timeout (default 30s)
	HostKeyCheck   bool          // Verify host keys (default false for cluster provisioning)
}

// SSHClient interface for SSH operations - allows mocking in tests
type SSHClient interface {
	Connect(host string, port int, config *SSHConfig) error
	RunCommand(cmd string) (string, error)
	Close() error
}

// RealSSHClient implements SSHClient using golang.org/x/crypto/ssh
type RealSSHClient struct {
	client *ssh.Client
}

// NewSSHClient creates a new SSH client instance
func NewSSHClient() SSHClient {
	return &RealSSHClient{}
}

// Connect establishes an SSH connection to the specified host
func (c *RealSSHClient) Connect(host string, port int, config *SSHConfig) error {
	if c.client != nil {
		return fmt.Errorf("client already connected")
	}

	// Build authentication methods
	var authMethods []ssh.AuthMethod

	// Try key-based auth first (preferred)
	if config.PrivateKey != nil {
		signer, err := ssh.ParsePrivateKey(config.PrivateKey)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	} else if config.PrivateKeyPath != "" {
		keyData, err := os.ReadFile(config.PrivateKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read private key file: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(keyData)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// Add password auth as fallback
	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}

	if len(authMethods) == 0 {
		return fmt.Errorf("no authentication method provided (need private key or password)")
	}

	// Set default timeout
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// Build SSH client config
	var hostKeyCallback ssh.HostKeyCallback
	if config.HostKeyCheck {
		// In production, you'd use ssh.FixedHostKey or a known_hosts file
		// For now, we don't support strict host key checking
		return fmt.Errorf("host key checking not yet implemented - set HostKeyCheck to false")
	}
	hostKeyCallback = ssh.InsecureIgnoreHostKey()

	sshConfig := &ssh.ClientConfig{
		User:            config.User,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         timeout,
	}

	// Connect to SSH server
	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	c.client = client
	return nil
}

// RunCommand executes a command on the remote host and returns combined output
func (c *RealSSHClient) RunCommand(cmd string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("not connected")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer func() { _ = session.Close() }()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}

// Close closes the SSH connection
func (c *RealSSHClient) Close() error {
	if c.client == nil {
		return nil
	}

	err := c.client.Close()
	c.client = nil
	return err
}
