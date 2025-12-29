package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceUSB(t *testing.T) {
	d := dataSourceUSB()
	if err := d.InternalValidate(nil, false); err != nil {
		t.Fatalf("data source internal validation failed: %s", err)
	}
}

func TestDataSourceUSB_Schema(t *testing.T) {
	d := dataSourceUSB()

	expectedFields := []string{
		"mode",
		"node",
		"route",
	}

	for _, field := range expectedFields {
		if _, ok := d.Schema[field]; !ok {
			t.Errorf("schema missing '%s' field", field)
		}
	}
}

func TestDataSourceUSB_SchemaTypes(t *testing.T) {
	d := dataSourceUSB()

	tests := []struct {
		field    string
		expected schema.ValueType
	}{
		{"mode", schema.TypeString},
		{"node", schema.TypeInt},
		{"route", schema.TypeString},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			if d.Schema[tt.field].Type != tt.expected {
				t.Errorf("expected %s to be type %v, got %v", tt.field, tt.expected, d.Schema[tt.field].Type)
			}
		})
	}
}

func TestDataSourceUSB_AllFieldsComputed(t *testing.T) {
	d := dataSourceUSB()

	for name, s := range d.Schema {
		if !s.Computed {
			t.Errorf("field %s should be computed", name)
		}
	}
}

func TestDataSourceUSB_HasReadFunction(t *testing.T) {
	d := dataSourceUSB()

	if d.ReadContext == nil {
		t.Error("data source should have ReadContext function")
	}
}

func TestDataSourceUSBRead_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"response": [][]interface{}{
				{"mode", "Host"},
				{"node", float64(0)},
				{"route", "USB-A"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	d := dataSourceUSB()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := dataSourceUSBRead(context.Background(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	// Verify ID is set
	if rd.Id() != "turingpi-usb-status" {
		t.Errorf("expected ID 'turingpi-usb-status', got '%s'", rd.Id())
	}

	// Verify values
	if v := rd.Get("mode").(string); v != "host" {
		t.Errorf("expected mode 'host', got '%s'", v)
	}
	if v := rd.Get("node").(int); v != 1 {
		t.Errorf("expected node 1, got %d", v)
	}
	if v := rd.Get("route").(string); v != "usb-a" {
		t.Errorf("expected route 'usb-a', got '%s'", v)
	}
}

func TestDataSourceUSBRead_DeviceMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"response": [][]interface{}{
				{"mode", "Device"},
				{"node", float64(1)},
				{"route", "BMC"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	d := dataSourceUSB()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := dataSourceUSBRead(context.Background(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := rd.Get("mode").(string); v != "device" {
		t.Errorf("expected mode 'device', got '%s'", v)
	}
	if v := rd.Get("node").(int); v != 2 {
		t.Errorf("expected node 2, got %d", v)
	}
	if v := rd.Get("route").(string); v != "bmc" {
		t.Errorf("expected route 'bmc', got '%s'", v)
	}
}

func TestDataSourceUSBRead_AllNodes(t *testing.T) {
	tests := []struct {
		apiNode      float64
		expectedNode int
	}{
		{0, 1},
		{1, 2},
		{2, 3},
		{3, 4},
	}

	for _, tt := range tests {
		t.Run(string(rune('0'+tt.expectedNode)), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := map[string]interface{}{
					"response": [][]interface{}{
						{"mode", "Host"},
						{"node", tt.apiNode},
						{"route", "USB-A"},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			d := dataSourceUSB()
			rd := d.TestResourceData()

			config := &ProviderConfig{
				Token:    "test-token",
				Endpoint: server.URL,
			}

			diags := dataSourceUSBRead(context.Background(), rd, config)
			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}

			if v := rd.Get("node").(int); v != tt.expectedNode {
				t.Errorf("expected node %d, got %d", tt.expectedNode, v)
			}
		})
	}
}

func TestDataSourceUSBRead_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	d := dataSourceUSB()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := dataSourceUSBRead(context.Background(), rd, config)
	if !diags.HasError() {
		t.Error("expected error for API failure")
	}
}

func TestDataSourceUSBRead_AuthHeader(t *testing.T) {
	var capturedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		response := map[string]interface{}{
			"response": [][]interface{}{
				{"mode", "Host"},
				{"node", float64(0)},
				{"route", "USB-A"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	d := dataSourceUSB()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "my-secret-token",
		Endpoint: server.URL,
	}

	diags := dataSourceUSBRead(context.Background(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if capturedAuth != "Bearer my-secret-token" {
		t.Errorf("expected Authorization 'Bearer my-secret-token', got '%s'", capturedAuth)
	}
}
