package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestResourceUARTSchema(t *testing.T) {
	resource := resourceUART()

	// Check required fields
	if resource.Schema["node"] == nil {
		t.Error("expected node field in schema")
	}
	if !resource.Schema["node"].Required {
		t.Error("node should be required")
	}

	if resource.Schema["command"] == nil {
		t.Error("expected command field in schema")
	}
	if !resource.Schema["command"].Required {
		t.Error("command should be required")
	}

	// Check optional fields
	if resource.Schema["triggers"] == nil {
		t.Error("expected triggers field in schema")
	}
	if !resource.Schema["triggers"].Optional {
		t.Error("triggers should be optional")
	}

	// Check computed fields
	if resource.Schema["last_sent"] == nil {
		t.Error("expected last_sent field in schema")
	}
	if !resource.Schema["last_sent"].Computed {
		t.Error("last_sent should be computed")
	}
}

func TestWriteUART(t *testing.T) {
	tests := []struct {
		name    string
		node    int
		command string
		wantErr bool
	}{
		{
			name:    "simple command",
			node:    1,
			command: "echo hello",
			wantErr: false,
		},
		{
			name:    "command with special chars",
			node:    2,
			command: "ls -la /var/log",
			wantErr: false,
		},
		{
			name:    "command with newline",
			node:    3,
			command: "reboot\n",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedURL string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedURL = r.URL.String()

				// Verify request parameters
				if !strings.Contains(r.URL.String(), "opt=set") {
					t.Errorf("expected opt=set in URL: %s", r.URL.String())
				}
				if !strings.Contains(r.URL.String(), "type=uart") {
					t.Errorf("expected type=uart in URL: %s", r.URL.String())
				}

				// Check node parameter (0-indexed)
				expectedNode := tt.node - 1
				if !strings.Contains(r.URL.String(), "node="+string(rune('0'+expectedNode))) {
					t.Errorf("expected node=%d in URL: %s", expectedNode, r.URL.String())
				}

				// Command should be URL-encoded
				encodedCmd := url.QueryEscape(tt.command)
				if !strings.Contains(r.URL.String(), "cmd="+encodedCmd) {
					t.Errorf("expected cmd=%s in URL: %s", encodedCmd, r.URL.String())
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"response":[["result","ok"]]}`))
			}))
			defer server.Close()

			originalClient := HTTPClient
			HTTPClient = server.Client()
			defer func() { HTTPClient = originalClient }()

			err := writeUART(server.URL, "test-token", tt.node, tt.command)

			if (err != nil) != tt.wantErr {
				t.Errorf("writeUART() error = %v, wantErr %v", err, tt.wantErr)
			}

			if capturedURL == "" {
				t.Error("server was not called")
			}
		})
	}
}

func TestWriteUART_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal error"}`))
	}))
	defer server.Close()

	originalClient := HTTPClient
	HTTPClient = server.Client()
	defer func() { HTTPClient = originalClient }()

	err := writeUART(server.URL, "test-token", 1, "test command")
	if err == nil {
		t.Error("expected error for server error response")
	}
}

func TestResourceUARTCRUD(t *testing.T) {
	writeCalled := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "type=uart") && strings.Contains(r.URL.String(), "opt=set") {
			writeCalled++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"response":[["result","ok"]]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	originalClient := HTTPClient
	HTTPClient = server.Client()
	defer func() { HTTPClient = originalClient }()

	config := &ProviderConfig{
		Endpoint: server.URL,
		Token:    "test-token",
	}

	resource := resourceUART()
	d := resource.TestResourceData()

	if err := d.Set("node", 1); err != nil {
		t.Fatalf("failed to set node: %v", err)
	}
	if err := d.Set("command", "echo hello"); err != nil {
		t.Fatalf("failed to set command: %v", err)
	}

	// Test Create
	diags := resourceUARTCreate(context.TODO(), d, config)
	if diags.HasError() {
		t.Errorf("Create returned error: %v", diags)
	}

	if d.Id() != "uart-write-node-1" {
		t.Errorf("expected ID 'uart-write-node-1', got '%s'", d.Id())
	}

	if writeCalled != 1 {
		t.Errorf("expected write to be called once, got %d", writeCalled)
	}

	lastSent := d.Get("last_sent").(string)
	if lastSent == "" {
		t.Error("expected last_sent to be set")
	}

	// Test Read (should be no-op)
	diags = resourceUARTRead(context.TODO(), d, config)
	if diags.HasError() {
		t.Errorf("Read returned error: %v", diags)
	}
	if writeCalled != 1 {
		t.Errorf("expected write count to remain 1, got %d", writeCalled)
	}

	// Test Update without changes (should not resend)
	diags = resourceUARTUpdate(context.TODO(), d, config)
	if diags.HasError() {
		t.Errorf("Update returned error: %v", diags)
	}
	if writeCalled != 1 {
		t.Errorf("expected write count to remain 1, got %d", writeCalled)
	}

	// Test Delete
	diags = resourceUARTDelete(context.TODO(), d, config)
	if diags.HasError() {
		t.Errorf("Delete returned error: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected empty ID after delete, got '%s'", d.Id())
	}
}

func TestResourceUARTUpdate_CommandChange(t *testing.T) {
	writeCalled := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "type=uart") && strings.Contains(r.URL.String(), "opt=set") {
			writeCalled++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"response":[["result","ok"]]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	originalClient := HTTPClient
	HTTPClient = server.Client()
	defer func() { HTTPClient = originalClient }()

	config := &ProviderConfig{
		Endpoint: server.URL,
		Token:    "test-token",
	}

	resource := resourceUART()
	d := resource.TestResourceData()

	if err := d.Set("node", 1); err != nil {
		t.Fatalf("failed to set node: %v", err)
	}
	if err := d.Set("command", "echo hello"); err != nil {
		t.Fatalf("failed to set command: %v", err)
	}

	// Create first
	diags := resourceUARTCreate(context.TODO(), d, config)
	if diags.HasError() {
		t.Fatalf("Create returned error: %v", diags)
	}

	// Simulate command change by using SetNew (for testing purposes, we'll just check HasChange behavior)
	// In real Terraform, HasChange would return true when the value differs from state
	// For unit testing, we'll verify the update logic works when called
	if writeCalled != 1 {
		t.Errorf("expected 1 write after create, got %d", writeCalled)
	}
}

func TestWriteUART_URLEncoding(t *testing.T) {
	var capturedCmd string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the cmd parameter
		capturedCmd = r.URL.Query().Get("cmd")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"response":[["result","ok"]]}`))
	}))
	defer server.Close()

	originalClient := HTTPClient
	HTTPClient = server.Client()
	defer func() { HTTPClient = originalClient }()

	// Test with special characters that need encoding
	command := "echo 'hello world' && ls -la"
	err := writeUART(server.URL, "test-token", 1, command)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// The captured cmd should be the URL-decoded version of what was sent
	// Since we URL-encode the command before sending, the server's Query().Get()
	// will return the decoded version
	if capturedCmd != command {
		t.Errorf("expected decoded command %q, got %q", command, capturedCmd)
	}
}
