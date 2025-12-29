package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceAbout(t *testing.T) {
	d := dataSourceAbout()
	if err := d.InternalValidate(nil, false); err != nil {
		t.Fatalf("data source internal validation failed: %s", err)
	}
}

func TestDataSourceAbout_Schema(t *testing.T) {
	d := dataSourceAbout()

	expectedFields := []string{
		"api_version",
		"daemon_version",
		"buildroot_version",
		"firmware_version",
		"build_time",
	}

	for _, field := range expectedFields {
		if _, ok := d.Schema[field]; !ok {
			t.Errorf("schema missing '%s' field", field)
		}
	}
}

func TestDataSourceAbout_SchemaTypes(t *testing.T) {
	d := dataSourceAbout()

	tests := []struct {
		field    string
		expected schema.ValueType
	}{
		{"api_version", schema.TypeString},
		{"daemon_version", schema.TypeString},
		{"buildroot_version", schema.TypeString},
		{"firmware_version", schema.TypeString},
		{"build_time", schema.TypeString},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			if d.Schema[tt.field].Type != tt.expected {
				t.Errorf("expected %s to be type %v, got %v", tt.field, tt.expected, d.Schema[tt.field].Type)
			}
		})
	}
}

func TestDataSourceAbout_AllFieldsComputed(t *testing.T) {
	d := dataSourceAbout()

	for name, s := range d.Schema {
		if !s.Computed {
			t.Errorf("field %s should be computed", name)
		}
	}
}

func TestDataSourceAbout_HasReadFunction(t *testing.T) {
	d := dataSourceAbout()

	if d.ReadContext == nil {
		t.Error("data source should have ReadContext function")
	}
}

func TestDataSourceAboutRead_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("type") == "about" {
			response := map[string]interface{}{
				"response": [][]interface{}{
					{"api", "1.0.0"},
					{"version", "2.0.5"},
					{"buildroot", "2023.02.1"},
					{"firmware", "1.1.0"},
					{"buildtime", "2024-01-15T10:30:00Z"},
				},
			}
			_ = json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	d := dataSourceAbout()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := dataSourceAboutRead(context.Background(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	// Verify ID is set
	if rd.Id() != "turingpi-about" {
		t.Errorf("expected ID 'turingpi-about', got '%s'", rd.Id())
	}

	// Verify all fields
	if v := rd.Get("api_version").(string); v != "1.0.0" {
		t.Errorf("expected api_version '1.0.0', got '%s'", v)
	}
	if v := rd.Get("daemon_version").(string); v != "2.0.5" {
		t.Errorf("expected daemon_version '2.0.5', got '%s'", v)
	}
	if v := rd.Get("buildroot_version").(string); v != "2023.02.1" {
		t.Errorf("expected buildroot_version '2023.02.1', got '%s'", v)
	}
	if v := rd.Get("firmware_version").(string); v != "1.1.0" {
		t.Errorf("expected firmware_version '1.1.0', got '%s'", v)
	}
	if v := rd.Get("build_time").(string); v != "2024-01-15T10:30:00Z" {
		t.Errorf("expected build_time '2024-01-15T10:30:00Z', got '%s'", v)
	}
}

func TestDataSourceAboutRead_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	d := dataSourceAbout()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := dataSourceAboutRead(context.Background(), rd, config)
	if !diags.HasError() {
		t.Error("expected error for API failure")
	}
}

func TestDataSourceAboutRead_PartialResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return only some fields
		response := map[string]interface{}{
			"response": [][]interface{}{
				{"api", "1.0.0"},
				{"version", "2.0.5"},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	d := dataSourceAbout()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := dataSourceAboutRead(context.Background(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	// Verify available fields
	if v := rd.Get("api_version").(string); v != "1.0.0" {
		t.Errorf("expected api_version '1.0.0', got '%s'", v)
	}
	if v := rd.Get("daemon_version").(string); v != "2.0.5" {
		t.Errorf("expected daemon_version '2.0.5', got '%s'", v)
	}

	// Missing fields should be empty strings
	if v := rd.Get("buildroot_version").(string); v != "" {
		t.Errorf("expected buildroot_version '', got '%s'", v)
	}
}

func TestDataSourceAboutRead_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"response": [][]interface{}{},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	d := dataSourceAbout()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := dataSourceAboutRead(context.Background(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error for empty response: %v", diags)
	}

	// ID should still be set
	if rd.Id() != "turingpi-about" {
		t.Errorf("expected ID 'turingpi-about', got '%s'", rd.Id())
	}
}

func TestDataSourceAboutRead_InvalidResponseFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return malformed data - items with only one element
		response := map[string]interface{}{
			"response": [][]interface{}{
				{"api"}, // Missing value
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	d := dataSourceAbout()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := dataSourceAboutRead(context.Background(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error for malformed response: %v", diags)
	}

	// Should not set anything for malformed items
	if v := rd.Get("api_version").(string); v != "" {
		t.Errorf("expected api_version '', got '%s'", v)
	}
}
