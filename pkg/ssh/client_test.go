package ssh

import (
	"fmt"
	"testing"
	"time"
)

// MockClient implements Client for testing
type MockClient struct {
	ConnectFunc    func(host string, port int, config *Config) error
	RunCommandFunc func(cmd string) (string, error)
	CloseFunc      func() error
	connected      bool
}

func (m *MockClient) Connect(host string, port int, config *Config) error {
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

func (m *MockClient) RunCommand(cmd string) (string, error) {
	if !m.connected {
		return "", fmt.Errorf("not connected")
	}
	if m.RunCommandFunc != nil {
		return m.RunCommandFunc(cmd)
	}
	return "", nil
}

func (m *MockClient) Close() error {
	m.connected = false
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// Test Config validation
func TestConfig_Defaults(t *testing.T) {
	config := &Config{
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
	mockFactory := func() Client {
		return &MockClient{
			ConnectFunc: func(host string, port int, config *Config) error {
				callCount++
				return nil // Success on first try
			},
		}
	}

	config := &Config{User: "test", Password: "test"}
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
	mockFactory := func() Client {
		return &MockClient{
			ConnectFunc: func(host string, port int, config *Config) error {
				callCount++
				if callCount < 2 {
					return fmt.Errorf("connection refused")
				}
				return nil // Success on second try
			},
		}
	}

	config := &Config{User: "test", Password: "test"}
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
	mockFactory := func() Client {
		return &MockClient{
			ConnectFunc: func(host string, port int, config *Config) error {
				return fmt.Errorf("connection refused")
			},
		}
	}

	config := &Config{User: "test", Password: "test"}
	err := WaitForSSHWithClient("localhost", 22, config, 1*time.Second, mockFactory)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

// Test RunCommand success
func TestRunCommand_Success(t *testing.T) {
	mock := &MockClient{
		RunCommandFunc: func(cmd string) (string, error) {
			if cmd == "echo hello" {
				return "hello\n", nil
			}
			return "", fmt.Errorf("unknown command")
		},
	}

	config := &Config{User: "test", Password: "test"}
	output, err := RunCommandWithClient("localhost", 22, config, "echo hello", mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", output)
	}
}

// Test RunCommand connection failure
func TestRunCommand_ConnectionFailed(t *testing.T) {
	mock := &MockClient{
		ConnectFunc: func(host string, port int, config *Config) error {
			return fmt.Errorf("connection refused")
		},
	}

	config := &Config{User: "test", Password: "test"}
	_, err := RunCommandWithClient("localhost", 22, config, "echo hello", mock)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// Test RunCommand command failure
func TestRunCommand_CommandFailed(t *testing.T) {
	mock := &MockClient{
		RunCommandFunc: func(cmd string) (string, error) {
			return "error output", fmt.Errorf("command exited with status 1")
		},
	}

	config := &Config{User: "test", Password: "test"}
	output, err := RunCommandWithClient("localhost", 22, config, "false", mock)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Output should still be returned even on error
	if output != "error output" {
		t.Errorf("expected error output to be returned, got %q", output)
	}
}

// Test Client interface validation
func TestRealClient_ImplementsInterface(t *testing.T) {
	// Compile-time check that RealClient implements Client
	var _ Client = (*RealClient)(nil)
}

// Test MockClient implements interface
func TestMockClient_ImplementsInterface(t *testing.T) {
	var _ Client = (*MockClient)(nil)
}
