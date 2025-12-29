package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceNodeToMSD_Schema(t *testing.T) {
	r := resourceNodeToMSD()

	expectedFields := []string{
		"node",
		"triggers",
		"last_triggered",
	}

	for _, field := range expectedFields {
		if _, ok := r.Schema[field]; !ok {
			t.Errorf("schema missing '%s' field", field)
		}
	}
}

func TestResourceNodeToMSD_SchemaTypes(t *testing.T) {
	r := resourceNodeToMSD()

	tests := []struct {
		field    string
		expected schema.ValueType
	}{
		{"node", schema.TypeInt},
		{"triggers", schema.TypeMap},
		{"last_triggered", schema.TypeString},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			if r.Schema[tt.field].Type != tt.expected {
				t.Errorf("expected %s to be type %v, got %v", tt.field, tt.expected, r.Schema[tt.field].Type)
			}
		})
	}
}

func TestResourceNodeToMSD_NodeRequired(t *testing.T) {
	r := resourceNodeToMSD()

	if !r.Schema["node"].Required {
		t.Error("node field should be required")
	}
}

func TestResourceNodeToMSD_LastTriggeredComputed(t *testing.T) {
	r := resourceNodeToMSD()

	if !r.Schema["last_triggered"].Computed {
		t.Error("last_triggered field should be computed")
	}
}

func TestResourceNodeToMSD_HasCRUDFunctions(t *testing.T) {
	r := resourceNodeToMSD()

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

func TestResourceNodeToMSDCreate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("type") == "node_to_msd" && query.Get("node") == "1" {
			response := map[string]interface{}{
				"response": []interface{}{},
			}
			_ = json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	r := resourceNodeToMSD()
	rd := r.TestResourceData()
	_ = rd.Set("node", 1)

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := resourceNodeToMSDCreate(context.TODO(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if rd.Id() != "node-to-msd-1" {
		t.Errorf("expected ID 'node-to-msd-1', got '%s'", rd.Id())
	}

	if rd.Get("last_triggered").(string) == "" {
		t.Error("expected last_triggered to be set")
	}
}

func TestResourceNodeToMSDCreate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	r := resourceNodeToMSD()
	rd := r.TestResourceData()
	_ = rd.Set("node", 1)

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := resourceNodeToMSDCreate(context.TODO(), rd, config)
	if !diags.HasError() {
		t.Error("expected error for API failure")
	}
}

func TestResourceNodeToMSDRead(t *testing.T) {
	r := resourceNodeToMSD()
	rd := r.TestResourceData()
	rd.SetId("node-to-msd-1")

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: "http://localhost",
	}

	diags := resourceNodeToMSDRead(context.TODO(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
}

func TestResourceNodeToMSDUpdate_TriggersChanged(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"response": []interface{}{},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	r := resourceNodeToMSD()
	rd := r.TestResourceData()
	rd.SetId("node-to-msd-1")
	_ = rd.Set("node", 1)

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := resourceNodeToMSDUpdate(context.TODO(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
}

func TestResourceNodeToMSDDelete(t *testing.T) {
	r := resourceNodeToMSD()
	rd := r.TestResourceData()
	rd.SetId("node-to-msd-1")
	_ = rd.Set("node", 1)

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: "http://localhost",
	}

	diags := resourceNodeToMSDDelete(context.TODO(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if rd.Id() != "" {
		t.Errorf("expected ID to be cleared, got '%s'", rd.Id())
	}
}

func TestNodeToMSD_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("expected Authorization 'Bearer test-token', got '%s'", auth)
		}

		query := r.URL.Query()
		if query.Get("type") != "node_to_msd" {
			t.Errorf("expected type 'node_to_msd', got '%s'", query.Get("type"))
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

	err := nodeToMSD(server.URL, "test-token", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNodeToMSD_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	err := nodeToMSD(server.URL, "test-token", 1)
	if err == nil {
		t.Error("expected error for API failure")
	}
}
