package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceUSBBoot_Schema(t *testing.T) {
	r := resourceUSBBoot()

	expectedFields := []string{
		"node",
		"triggers",
		"last_enabled",
	}

	for _, field := range expectedFields {
		if _, ok := r.Schema[field]; !ok {
			t.Errorf("schema missing '%s' field", field)
		}
	}
}

func TestResourceUSBBoot_SchemaTypes(t *testing.T) {
	r := resourceUSBBoot()

	tests := []struct {
		field    string
		expected schema.ValueType
	}{
		{"node", schema.TypeInt},
		{"triggers", schema.TypeMap},
		{"last_enabled", schema.TypeString},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			if r.Schema[tt.field].Type != tt.expected {
				t.Errorf("expected %s to be type %v, got %v", tt.field, tt.expected, r.Schema[tt.field].Type)
			}
		})
	}
}

func TestResourceUSBBoot_NodeRequired(t *testing.T) {
	r := resourceUSBBoot()

	if !r.Schema["node"].Required {
		t.Error("node field should be required")
	}
}

func TestResourceUSBBoot_LastEnabledComputed(t *testing.T) {
	r := resourceUSBBoot()

	if !r.Schema["last_enabled"].Computed {
		t.Error("last_enabled field should be computed")
	}
}

func TestResourceUSBBoot_HasCRUDFunctions(t *testing.T) {
	r := resourceUSBBoot()

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

func TestResourceUSBBootCreate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("type") == "usb_boot" && query.Get("node") == "1" {
			response := map[string]interface{}{
				"response": []interface{}{},
			}
			_ = json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	r := resourceUSBBoot()
	rd := r.TestResourceData()
	_ = rd.Set("node", 1)

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := resourceUSBBootCreate(context.TODO(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if rd.Id() != "usb-boot-node-1" {
		t.Errorf("expected ID 'usb-boot-node-1', got '%s'", rd.Id())
	}

	if rd.Get("last_enabled").(string) == "" {
		t.Error("expected last_enabled to be set")
	}
}

func TestResourceUSBBootCreate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	r := resourceUSBBoot()
	rd := r.TestResourceData()
	_ = rd.Set("node", 1)

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := resourceUSBBootCreate(context.TODO(), rd, config)
	if !diags.HasError() {
		t.Error("expected error for API failure")
	}
}

func TestResourceUSBBootRead(t *testing.T) {
	r := resourceUSBBoot()
	rd := r.TestResourceData()
	rd.SetId("usb-boot-node-1")

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: "http://localhost",
	}

	diags := resourceUSBBootRead(context.TODO(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
}

func TestResourceUSBBootUpdate_TriggersChanged(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"response": []interface{}{},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	r := resourceUSBBoot()
	rd := r.TestResourceData()
	rd.SetId("usb-boot-node-1")
	_ = rd.Set("node", 1)

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := resourceUSBBootUpdate(context.TODO(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
}

func TestResourceUSBBootDelete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("type") == "clear_usb_boot" {
			response := map[string]interface{}{
				"response": []interface{}{},
			}
			_ = json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	r := resourceUSBBoot()
	rd := r.TestResourceData()
	rd.SetId("usb-boot-node-1")
	_ = rd.Set("node", 1)

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := resourceUSBBootDelete(context.TODO(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if rd.Id() != "" {
		t.Errorf("expected ID to be cleared, got '%s'", rd.Id())
	}
}

func TestEnableUSBBoot_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("expected Authorization 'Bearer test-token', got '%s'", auth)
		}

		query := r.URL.Query()
		if query.Get("type") != "usb_boot" {
			t.Errorf("expected type 'usb_boot', got '%s'", query.Get("type"))
		}
		if query.Get("node") != "2" {
			t.Errorf("expected node '2', got '%s'", query.Get("node"))
		}

		response := map[string]interface{}{
			"response": []interface{}{},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	err := enableUSBBoot(server.URL, "test-token", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClearUSBBoot_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("type") != "clear_usb_boot" {
			t.Errorf("expected type 'clear_usb_boot', got '%s'", query.Get("type"))
		}
		if query.Get("node") != "3" {
			t.Errorf("expected node '3', got '%s'", query.Get("node"))
		}

		response := map[string]interface{}{
			"response": []interface{}{},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	err := clearUSBBoot(server.URL, "test-token", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
