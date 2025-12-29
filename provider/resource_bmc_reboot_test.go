package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResourceBMCRebootSchema(t *testing.T) {
	resource := resourceBMCReboot()

	// Check optional fields
	if resource.Schema["triggers"] == nil {
		t.Error("expected triggers field in schema")
	}
	if !resource.Schema["triggers"].Optional {
		t.Error("triggers should be optional")
	}

	if resource.Schema["wait_for_ready"] == nil {
		t.Error("expected wait_for_ready field in schema")
	}
	if !resource.Schema["wait_for_ready"].Optional {
		t.Error("wait_for_ready should be optional")
	}
	if resource.Schema["wait_for_ready"].Default != true {
		t.Error("wait_for_ready should default to true")
	}

	if resource.Schema["ready_timeout"] == nil {
		t.Error("expected ready_timeout field in schema")
	}
	if !resource.Schema["ready_timeout"].Optional {
		t.Error("ready_timeout should be optional")
	}
	if resource.Schema["ready_timeout"].Default != 120 {
		t.Error("ready_timeout should default to 120")
	}

	// Check computed fields
	if resource.Schema["last_reboot"] == nil {
		t.Error("expected last_reboot field in schema")
	}
	if !resource.Schema["last_reboot"].Computed {
		t.Error("last_reboot should be computed")
	}
}

func TestRebootBMC(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse int
		wantErr        bool
	}{
		{
			name:           "successful reboot",
			serverResponse: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "server error",
			serverResponse: http.StatusInternalServerError,
			wantErr:        true,
		},
		{
			name:           "unauthorized",
			serverResponse: http.StatusUnauthorized,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				if !strings.Contains(r.URL.String(), "opt=set") {
					t.Errorf("expected opt=set in URL: %s", r.URL.String())
				}
				if !strings.Contains(r.URL.String(), "type=reboot") {
					t.Errorf("expected type=reboot in URL: %s", r.URL.String())
				}

				auth := r.Header.Get("Authorization")
				if auth != "Bearer test-token" {
					t.Errorf("expected Bearer test-token, got %s", auth)
				}

				w.WriteHeader(tt.serverResponse)
				_, _ = w.Write([]byte(`{"response":[["result","ok"]]}`))
			}))
			defer server.Close()

			originalClient := HTTPClient
			HTTPClient = server.Client()
			defer func() { HTTPClient = originalClient }()

			err := rebootBMC(server.URL, "test-token")

			if (err != nil) != tt.wantErr {
				t.Errorf("rebootBMC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckBMCReady(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse int
		wantReady      bool
	}{
		{
			name:           "BMC ready",
			serverResponse: http.StatusOK,
			wantReady:      true,
		},
		{
			name:           "BMC not ready - 500",
			serverResponse: http.StatusInternalServerError,
			wantReady:      false,
		},
		{
			name:           "BMC not ready - 503",
			serverResponse: http.StatusServiceUnavailable,
			wantReady:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify it's checking the about endpoint
				if !strings.Contains(r.URL.String(), "type=about") {
					t.Errorf("expected type=about in URL: %s", r.URL.String())
				}

				w.WriteHeader(tt.serverResponse)
				_, _ = w.Write([]byte(`{"response":[["api","1.0"]]}`))
			}))
			defer server.Close()

			originalClient := HTTPClient
			HTTPClient = server.Client()
			defer func() { HTTPClient = originalClient }()

			ready := checkBMCReady(server.URL, "test-token")

			if ready != tt.wantReady {
				t.Errorf("checkBMCReady() = %v, want %v", ready, tt.wantReady)
			}
		})
	}
}

func TestResourceBMCRebootCRUD(t *testing.T) {
	rebootCalled := 0
	aboutCalled := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "type=reboot") && strings.Contains(r.URL.String(), "opt=set") {
			rebootCalled++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"response":[["result","ok"]]}`))
			return
		}

		if strings.Contains(r.URL.String(), "type=about") {
			aboutCalled++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"response":[["api","1.0"]]}`))
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

	resource := resourceBMCReboot()
	d := resource.TestResourceData()

	// Disable wait_for_ready for faster tests
	if err := d.Set("wait_for_ready", false); err != nil {
		t.Fatalf("failed to set wait_for_ready: %v", err)
	}

	// Test Create
	diags := resourceBMCRebootCreate(context.TODO(), d, config)
	if diags.HasError() {
		t.Errorf("Create returned error: %v", diags)
	}

	if d.Id() != "bmc-reboot" {
		t.Errorf("expected ID 'bmc-reboot', got '%s'", d.Id())
	}

	if rebootCalled != 1 {
		t.Errorf("expected reboot to be called once, got %d", rebootCalled)
	}

	lastReboot := d.Get("last_reboot").(string)
	if lastReboot == "" {
		t.Error("expected last_reboot to be set")
	}

	// Test Read (should be no-op)
	diags = resourceBMCRebootRead(context.TODO(), d, config)
	if diags.HasError() {
		t.Errorf("Read returned error: %v", diags)
	}
	if rebootCalled != 1 {
		t.Errorf("expected reboot count to remain 1, got %d", rebootCalled)
	}

	// Test Update without changes (should not reboot)
	diags = resourceBMCRebootUpdate(context.TODO(), d, config)
	if diags.HasError() {
		t.Errorf("Update returned error: %v", diags)
	}
	if rebootCalled != 1 {
		t.Errorf("expected reboot count to remain 1, got %d", rebootCalled)
	}

	// Test Delete
	diags = resourceBMCRebootDelete(context.TODO(), d, config)
	if diags.HasError() {
		t.Errorf("Delete returned error: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected empty ID after delete, got '%s'", d.Id())
	}
}

func TestResourceBMCRebootCreate_WithWait(t *testing.T) {
	rebootCalled := false
	aboutCallCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "type=reboot") {
			rebootCalled = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"response":[["result","ok"]]}`))
			return
		}

		if strings.Contains(r.URL.String(), "type=about") {
			aboutCallCount++
			// Simulate BMC becoming ready after reboot
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"response":[["api","1.0"]]}`))
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

	resource := resourceBMCReboot()
	d := resource.TestResourceData()

	// Enable wait_for_ready with short timeout
	if err := d.Set("wait_for_ready", true); err != nil {
		t.Fatalf("failed to set wait_for_ready: %v", err)
	}
	if err := d.Set("ready_timeout", 10); err != nil {
		t.Fatalf("failed to set ready_timeout: %v", err)
	}

	diags := resourceBMCRebootCreate(context.TODO(), d, config)
	if diags.HasError() {
		t.Errorf("Create returned error: %v", diags)
	}

	if !rebootCalled {
		t.Error("expected reboot to be called")
	}

	if aboutCallCount == 0 {
		t.Error("expected about endpoint to be called for readiness check")
	}
}
