package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceClearUSBBoot_Schema(t *testing.T) {
	r := resourceClearUSBBoot()

	expectedFields := []string{
		"node",
		"triggers",
		"last_cleared",
	}

	for _, field := range expectedFields {
		if _, ok := r.Schema[field]; !ok {
			t.Errorf("schema missing '%s' field", field)
		}
	}
}

func TestResourceClearUSBBoot_SchemaTypes(t *testing.T) {
	r := resourceClearUSBBoot()

	tests := []struct {
		field    string
		expected schema.ValueType
	}{
		{"node", schema.TypeInt},
		{"triggers", schema.TypeMap},
		{"last_cleared", schema.TypeString},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			if r.Schema[tt.field].Type != tt.expected {
				t.Errorf("expected %s to be type %v, got %v", tt.field, tt.expected, r.Schema[tt.field].Type)
			}
		})
	}
}

func TestResourceClearUSBBoot_NodeRequired(t *testing.T) {
	r := resourceClearUSBBoot()

	if !r.Schema["node"].Required {
		t.Error("node field should be required")
	}
}

func TestResourceClearUSBBoot_LastClearedComputed(t *testing.T) {
	r := resourceClearUSBBoot()

	if !r.Schema["last_cleared"].Computed {
		t.Error("last_cleared field should be computed")
	}
}

func TestResourceClearUSBBoot_HasCRUDFunctions(t *testing.T) {
	r := resourceClearUSBBoot()

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

func TestResourceClearUSBBootCreate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("type") == "clear_usb_boot" && query.Get("node") == "1" {
			response := map[string]interface{}{
				"response": []interface{}{},
			}
			_ = json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	r := resourceClearUSBBoot()
	rd := r.TestResourceData()
	_ = rd.Set("node", 1)

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := resourceClearUSBBootCreate(context.TODO(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if rd.Id() != "clear-usb-boot-node-1" {
		t.Errorf("expected ID 'clear-usb-boot-node-1', got '%s'", rd.Id())
	}

	if rd.Get("last_cleared").(string) == "" {
		t.Error("expected last_cleared to be set")
	}
}

func TestResourceClearUSBBootCreate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	r := resourceClearUSBBoot()
	rd := r.TestResourceData()
	_ = rd.Set("node", 1)

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := resourceClearUSBBootCreate(context.TODO(), rd, config)
	if !diags.HasError() {
		t.Error("expected error for API failure")
	}
}

func TestResourceClearUSBBootRead(t *testing.T) {
	r := resourceClearUSBBoot()
	rd := r.TestResourceData()
	rd.SetId("clear-usb-boot-node-1")

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: "http://localhost",
	}

	diags := resourceClearUSBBootRead(context.TODO(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
}

func TestResourceClearUSBBootUpdate_TriggersChanged(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"response": []interface{}{},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	r := resourceClearUSBBoot()
	rd := r.TestResourceData()
	rd.SetId("clear-usb-boot-node-1")
	_ = rd.Set("node", 1)

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := resourceClearUSBBootUpdate(context.TODO(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
}

func TestResourceClearUSBBootDelete(t *testing.T) {
	r := resourceClearUSBBoot()
	rd := r.TestResourceData()
	rd.SetId("clear-usb-boot-node-1")
	_ = rd.Set("node", 1)

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: "http://localhost",
	}

	diags := resourceClearUSBBootDelete(context.TODO(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if rd.Id() != "" {
		t.Errorf("expected ID to be cleared, got '%s'", rd.Id())
	}
}
