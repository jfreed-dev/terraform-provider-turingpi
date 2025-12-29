package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceBMCReload_Schema(t *testing.T) {
	r := resourceBMCReload()

	expectedFields := []string{
		"triggers",
		"wait_for_ready",
		"ready_timeout",
		"last_reload",
	}

	for _, field := range expectedFields {
		if _, ok := r.Schema[field]; !ok {
			t.Errorf("schema missing '%s' field", field)
		}
	}
}

func TestResourceBMCReload_SchemaTypes(t *testing.T) {
	r := resourceBMCReload()

	tests := []struct {
		field    string
		expected schema.ValueType
	}{
		{"triggers", schema.TypeMap},
		{"wait_for_ready", schema.TypeBool},
		{"ready_timeout", schema.TypeInt},
		{"last_reload", schema.TypeString},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			if r.Schema[tt.field].Type != tt.expected {
				t.Errorf("expected %s to be type %v, got %v", tt.field, tt.expected, r.Schema[tt.field].Type)
			}
		})
	}
}

func TestResourceBMCReload_Defaults(t *testing.T) {
	r := resourceBMCReload()

	if r.Schema["wait_for_ready"].Default != true {
		t.Error("wait_for_ready should default to true")
	}
	if r.Schema["ready_timeout"].Default != 30 {
		t.Error("ready_timeout should default to 30")
	}
}

func TestResourceBMCReload_LastReloadComputed(t *testing.T) {
	r := resourceBMCReload()

	if !r.Schema["last_reload"].Computed {
		t.Error("last_reload field should be computed")
	}
}

func TestResourceBMCReload_HasCRUDFunctions(t *testing.T) {
	r := resourceBMCReload()

	if r.CreateContext == nil {
		t.Error("resource should have CreateContext function")
	}
	if r.ReadContext == nil {
		t.Error("resource should have ReadContext function")
	}
	if r.UpdateContext == nil {
		t.Error("resource should have UpdateContext function")
	}
	if r.DeleteContext == nil {
		t.Error("resource should have DeleteContext function")
	}
}

func TestResourceBMCReloadCreate_Success(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		query := r.URL.Query()

		switch query.Get("type") {
		case "reload":
			response := map[string]interface{}{
				"response": []interface{}{},
			}
			_ = json.NewEncoder(w).Encode(response)
		case "about":
			// Health check endpoint
			response := map[string]interface{}{
				"response": [][]interface{}{
					{"api", "1.0"},
				},
			}
			_ = json.NewEncoder(w).Encode(response)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	r := resourceBMCReload()
	rd := r.TestResourceData()
	_ = rd.Set("wait_for_ready", false) // Disable waiting for faster test

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := resourceBMCReloadCreate(context.TODO(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if rd.Id() != "bmc-reload" {
		t.Errorf("expected ID 'bmc-reload', got '%s'", rd.Id())
	}

	if rd.Get("last_reload").(string) == "" {
		t.Error("expected last_reload to be set")
	}
}

func TestResourceBMCReloadCreate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	r := resourceBMCReload()
	rd := r.TestResourceData()
	_ = rd.Set("wait_for_ready", false)

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := resourceBMCReloadCreate(context.TODO(), rd, config)
	if !diags.HasError() {
		t.Error("expected error for API failure")
	}
}

func TestResourceBMCReloadRead(t *testing.T) {
	r := resourceBMCReload()
	rd := r.TestResourceData()
	rd.SetId("bmc-reload")

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: "http://localhost",
	}

	diags := resourceBMCReloadRead(context.TODO(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
}

func TestResourceBMCReloadUpdate_TriggersChanged(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"response": []interface{}{},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	r := resourceBMCReload()
	rd := r.TestResourceData()
	rd.SetId("bmc-reload")
	_ = rd.Set("wait_for_ready", false)

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := resourceBMCReloadUpdate(context.TODO(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
}

func TestResourceBMCReloadDelete(t *testing.T) {
	r := resourceBMCReload()
	rd := r.TestResourceData()
	rd.SetId("bmc-reload")

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: "http://localhost",
	}

	diags := resourceBMCReloadDelete(context.TODO(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if rd.Id() != "" {
		t.Errorf("expected ID to be cleared, got '%s'", rd.Id())
	}
}

func TestReloadBMCDaemon_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("expected Authorization 'Bearer test-token', got '%s'", auth)
		}

		query := r.URL.Query()
		if query.Get("type") != "reload" {
			t.Errorf("expected type 'reload', got '%s'", query.Get("type"))
		}
		if query.Get("opt") != "set" {
			t.Errorf("expected opt 'set', got '%s'", query.Get("opt"))
		}

		response := map[string]interface{}{
			"response": []interface{}{},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	err := reloadBMCDaemon(server.URL, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReloadBMCDaemon_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	err := reloadBMCDaemon(server.URL, "test-token")
	if err == nil {
		t.Error("expected error for API failure")
	}
}
