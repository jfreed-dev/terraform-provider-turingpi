package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDataSourceUARTSchema(t *testing.T) {
	ds := dataSourceUART()

	// Check required fields
	if ds.Schema["node"] == nil {
		t.Error("expected node field in schema")
	}
	if !ds.Schema["node"].Required {
		t.Error("node should be required")
	}

	// Check optional fields
	if ds.Schema["encoding"] == nil {
		t.Error("expected encoding field in schema")
	}
	if !ds.Schema["encoding"].Optional {
		t.Error("encoding should be optional")
	}
	if ds.Schema["encoding"].Default != "utf8" {
		t.Error("encoding should default to utf8")
	}

	// Check computed fields
	if ds.Schema["output"] == nil {
		t.Error("expected output field in schema")
	}
	if !ds.Schema["output"].Computed {
		t.Error("output should be computed")
	}

	if ds.Schema["has_output"] == nil {
		t.Error("expected has_output field in schema")
	}
	if !ds.Schema["has_output"].Computed {
		t.Error("has_output should be computed")
	}
}

func TestReadUART(t *testing.T) {
	tests := []struct {
		name           string
		node           int
		encoding       string
		serverResponse string
		wantOutput     string
		wantErr        bool
	}{
		{
			name:           "successful read with output",
			node:           1,
			encoding:       "utf8",
			serverResponse: `{"response":[["uart","Hello from node 1\n"]]}`,
			wantOutput:     "Hello from node 1\n",
			wantErr:        false,
		},
		{
			name:           "empty buffer",
			node:           2,
			encoding:       "utf8",
			serverResponse: `{"response":[]}`,
			wantOutput:     "",
			wantErr:        false,
		},
		{
			name:           "output key variant",
			node:           3,
			encoding:       "utf8",
			serverResponse: `{"response":[["output","Boot complete"]]}`,
			wantOutput:     "Boot complete",
			wantErr:        false,
		},
		{
			name:           "data key variant",
			node:           4,
			encoding:       "utf8",
			serverResponse: `{"response":[["data","login: "]]}`,
			wantOutput:     "login: ",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request parameters
				if !strings.Contains(r.URL.String(), "opt=get") {
					t.Errorf("expected opt=get in URL: %s", r.URL.String())
				}
				if !strings.Contains(r.URL.String(), "type=uart") {
					t.Errorf("expected type=uart in URL: %s", r.URL.String())
				}
				// Check node parameter (0-indexed)
				expectedNode := tt.node - 1
				if !strings.Contains(r.URL.String(), "node="+string(rune('0'+expectedNode))) {
					t.Errorf("expected node=%d in URL: %s", expectedNode, r.URL.String())
				}
				if !strings.Contains(r.URL.String(), "encoding="+tt.encoding) {
					t.Errorf("expected encoding=%s in URL: %s", tt.encoding, r.URL.String())
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			originalClient := HTTPClient
			HTTPClient = server.Client()
			defer func() { HTTPClient = originalClient }()

			output, err := readUART(server.URL, "test-token", tt.node, tt.encoding)

			if (err != nil) != tt.wantErr {
				t.Errorf("readUART() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if output != tt.wantOutput {
				t.Errorf("readUART() = %q, want %q", output, tt.wantOutput)
			}
		})
	}
}

func TestReadUART_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal error"}`))
	}))
	defer server.Close()

	originalClient := HTTPClient
	HTTPClient = server.Client()
	defer func() { HTTPClient = originalClient }()

	_, err := readUART(server.URL, "test-token", 1, "utf8")
	if err == nil {
		t.Error("expected error for server error response")
	}
}

func TestExtractUARTOutput(t *testing.T) {
	tests := []struct {
		name     string
		response uartReadResponse
		want     string
	}{
		{
			name: "uart key",
			response: uartReadResponse{
				Response: [][]interface{}{{"uart", "test output"}},
			},
			want: "test output",
		},
		{
			name: "output key",
			response: uartReadResponse{
				Response: [][]interface{}{{"output", "boot messages"}},
			},
			want: "boot messages",
		},
		{
			name: "data key",
			response: uartReadResponse{
				Response: [][]interface{}{{"data", "serial data"}},
			},
			want: "serial data",
		},
		{
			name: "empty response",
			response: uartReadResponse{
				Response: [][]interface{}{},
			},
			want: "",
		},
		{
			name: "fallback to first string",
			response: uartReadResponse{
				Response: [][]interface{}{{"raw output here"}},
			},
			want: "raw output here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractUARTOutput(tt.response)
			if got != tt.want {
				t.Errorf("extractUARTOutput() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDataSourceUARTRead(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"response":[["uart","Node 1 boot log\nlogin: "]]}`))
	}))
	defer server.Close()

	originalClient := HTTPClient
	HTTPClient = server.Client()
	defer func() { HTTPClient = originalClient }()

	config := &ProviderConfig{
		Endpoint: server.URL,
		Token:    "test-token",
	}

	ds := dataSourceUART()
	d := ds.TestResourceData()

	if err := d.Set("node", 1); err != nil {
		t.Fatalf("failed to set node: %v", err)
	}
	if err := d.Set("encoding", "utf8"); err != nil {
		t.Fatalf("failed to set encoding: %v", err)
	}

	diags := dataSourceUARTRead(context.TODO(), d, config)
	if diags.HasError() {
		t.Errorf("Read returned error: %v", diags)
	}

	output := d.Get("output").(string)
	if output != "Node 1 boot log\nlogin: " {
		t.Errorf("expected output 'Node 1 boot log\\nlogin: ', got %q", output)
	}

	hasOutput := d.Get("has_output").(bool)
	if !hasOutput {
		t.Error("expected has_output to be true")
	}

	if d.Id() != "uart-node-1" {
		t.Errorf("expected ID 'uart-node-1', got '%s'", d.Id())
	}
}

func TestDataSourceUARTRead_EmptyBuffer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"response":[]}`))
	}))
	defer server.Close()

	originalClient := HTTPClient
	HTTPClient = server.Client()
	defer func() { HTTPClient = originalClient }()

	config := &ProviderConfig{
		Endpoint: server.URL,
		Token:    "test-token",
	}

	ds := dataSourceUART()
	d := ds.TestResourceData()

	if err := d.Set("node", 2); err != nil {
		t.Fatalf("failed to set node: %v", err)
	}

	diags := dataSourceUARTRead(context.TODO(), d, config)
	if diags.HasError() {
		t.Errorf("Read returned error: %v", diags)
	}

	hasOutput := d.Get("has_output").(bool)
	if hasOutput {
		t.Error("expected has_output to be false for empty buffer")
	}
}
