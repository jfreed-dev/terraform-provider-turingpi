package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResetNetwork(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse int
		serverBody     string
		wantErr        bool
	}{
		{
			name:           "successful reset",
			serverResponse: http.StatusOK,
			serverBody:     `{"response":[[{"result":"ok"}]]}`,
			wantErr:        false,
		},
		{
			name:           "server error",
			serverResponse: http.StatusInternalServerError,
			serverBody:     `{"error":"internal error"}`,
			wantErr:        true,
		},
		{
			name:           "unauthorized",
			serverResponse: http.StatusUnauthorized,
			serverBody:     `{"error":"unauthorized"}`,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the request
				if !strings.Contains(r.URL.String(), "/api/bmc") {
					t.Errorf("unexpected URL: %s", r.URL.String())
				}
				if !strings.Contains(r.URL.String(), "opt=set") {
					t.Errorf("expected opt=set in URL: %s", r.URL.String())
				}
				if !strings.Contains(r.URL.String(), "type=network") {
					t.Errorf("expected type=network in URL: %s", r.URL.String())
				}

				// Check authorization header
				auth := r.Header.Get("Authorization")
				if auth != "Bearer test-token" {
					t.Errorf("expected Bearer test-token, got %s", auth)
				}

				w.WriteHeader(tt.serverResponse)
				_, _ = w.Write([]byte(tt.serverBody))
			}))
			defer server.Close()

			// Override the HTTP client
			originalClient := HTTPClient
			HTTPClient = server.Client()
			defer func() { HTTPClient = originalClient }()

			// Test the function
			err := resetNetwork(server.URL, "test-token")

			if (err != nil) != tt.wantErr {
				t.Errorf("resetNetwork() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResourceNetworkResetSchema(t *testing.T) {
	resource := resourceNetworkReset()

	// Check that required fields exist
	if resource.Schema["triggers"] == nil {
		t.Error("expected triggers field in schema")
	}
	if resource.Schema["last_reset"] == nil {
		t.Error("expected last_reset field in schema")
	}

	// Check triggers field properties
	triggers := resource.Schema["triggers"]
	if triggers.Required {
		t.Error("triggers should be optional")
	}
	if !triggers.Optional {
		t.Error("triggers should be optional")
	}

	// Check last_reset field properties
	lastReset := resource.Schema["last_reset"]
	if !lastReset.Computed {
		t.Error("last_reset should be computed")
	}
}

func TestResourceNetworkResetCRUD(t *testing.T) {
	resetCalled := 0

	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "type=network") && strings.Contains(r.URL.String(), "opt=set") {
			resetCalled++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"response":[[{"result":"ok"}]]}`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Override the HTTP client
	originalClient := HTTPClient
	HTTPClient = server.Client()
	defer func() { HTTPClient = originalClient }()

	// Test Create
	config := &ProviderConfig{
		Endpoint: server.URL,
		Token:    "test-token",
	}

	resource := resourceNetworkReset()

	// Create a mock ResourceData
	d := resource.TestResourceData()

	// Test Create
	diags := resourceNetworkResetCreate(context.TODO(), d, config)
	if diags.HasError() {
		t.Errorf("Create returned error: %v", diags)
	}
	if d.Id() != "network-reset" {
		t.Errorf("expected ID 'network-reset', got '%s'", d.Id())
	}
	if resetCalled != 1 {
		t.Errorf("expected reset to be called once, got %d", resetCalled)
	}
	lastReset := d.Get("last_reset").(string)
	if lastReset == "" {
		t.Error("expected last_reset to be set")
	}

	// Test Read (should be a no-op)
	diags = resourceNetworkResetRead(context.TODO(), d, config)
	if diags.HasError() {
		t.Errorf("Read returned error: %v", diags)
	}
	// Reset count should not change
	if resetCalled != 1 {
		t.Errorf("expected reset count to remain 1, got %d", resetCalled)
	}

	// Test Update without trigger change (no reset should happen)
	diags = resourceNetworkResetUpdate(context.TODO(), d, config)
	if diags.HasError() {
		t.Errorf("Update returned error: %v", diags)
	}
	if resetCalled != 1 {
		t.Errorf("expected reset count to remain 1, got %d", resetCalled)
	}

	// Test Delete
	diags = resourceNetworkResetDelete(context.TODO(), d, config)
	if diags.HasError() {
		t.Errorf("Delete returned error: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected empty ID after delete, got '%s'", d.Id())
	}
}
