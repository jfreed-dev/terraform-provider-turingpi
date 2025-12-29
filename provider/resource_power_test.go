package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourcePower(t *testing.T) {
	r := resourcePower()
	if err := r.InternalValidate(nil, true); err != nil {
		t.Fatalf("resource internal validation failed: %s", err)
	}
}

func TestResourcePower_Schema(t *testing.T) {
	r := resourcePower()

	expectedFields := []string{"node", "state", "current_state"}

	for _, field := range expectedFields {
		if _, ok := r.Schema[field]; !ok {
			t.Errorf("schema missing '%s' field", field)
		}
	}
}

func TestResourcePower_SchemaTypes(t *testing.T) {
	r := resourcePower()

	tests := []struct {
		field    string
		expected schema.ValueType
	}{
		{"node", schema.TypeInt},
		{"state", schema.TypeString},
		{"current_state", schema.TypeBool},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			if r.Schema[tt.field].Type != tt.expected {
				t.Errorf("expected %s to be type %v, got %v", tt.field, tt.expected, r.Schema[tt.field].Type)
			}
		})
	}
}

func TestResourcePower_RequiredFields(t *testing.T) {
	r := resourcePower()

	if !r.Schema["node"].Required {
		t.Error("node should be required")
	}

	if !r.Schema["state"].Required {
		t.Error("state should be required")
	}
}

func TestResourcePower_ComputedFields(t *testing.T) {
	r := resourcePower()

	if !r.Schema["current_state"].Computed {
		t.Error("current_state should be computed")
	}
}

func TestResourcePower_HasCRUDFunctions(t *testing.T) {
	r := resourcePower()

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

func TestResourcePower_HasImporter(t *testing.T) {
	r := resourcePower()

	if r.Importer == nil {
		t.Error("resource should have Importer")
	}
}

func TestResourcePowerCreate_PowerOn(t *testing.T) {
	var capturedURL string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "opt=set") && strings.Contains(r.URL.String(), "type=power") {
			capturedURL = r.URL.String()
			w.WriteHeader(http.StatusOK)
			return
		}
		// GET power status for Read
		response := map[string]interface{}{
			"response": [][]interface{}{
				{"node1", float64(1)},
				{"node2", float64(0)},
				{"node3", float64(0)},
				{"node4", float64(0)},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	r := resourcePower()
	d := r.TestResourceData()

	_ = d.Set("node", 1)
	_ = d.Set("state", "on")

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := resourcePowerCreate(context.Background(), d, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "power-node-1" {
		t.Errorf("expected ID 'power-node-1', got '%s'", d.Id())
	}

	if !strings.Contains(capturedURL, "node1=1") {
		t.Errorf("expected URL to contain 'node1=1', got '%s'", capturedURL)
	}
}

func TestResourcePowerCreate_PowerOff(t *testing.T) {
	var capturedURL string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "opt=set") && strings.Contains(r.URL.String(), "type=power") {
			capturedURL = r.URL.String()
			w.WriteHeader(http.StatusOK)
			return
		}
		response := map[string]interface{}{
			"response": [][]interface{}{
				{"node1", float64(0)},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	r := resourcePower()
	d := r.TestResourceData()

	_ = d.Set("node", 1)
	_ = d.Set("state", "off")

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := resourcePowerCreate(context.Background(), d, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !strings.Contains(capturedURL, "node1=0") {
		t.Errorf("expected URL to contain 'node1=0', got '%s'", capturedURL)
	}
}

func TestResourcePowerCreate_Reset(t *testing.T) {
	var capturedURL string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "type=reset") {
			capturedURL = r.URL.String()
			w.WriteHeader(http.StatusOK)
			return
		}
		response := map[string]interface{}{
			"response": [][]interface{}{
				{"node1", float64(1)},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	r := resourcePower()
	d := r.TestResourceData()

	_ = d.Set("node", 2)
	_ = d.Set("state", "reset")

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := resourcePowerCreate(context.Background(), d, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	// Node 2 should be sent as node=1 (0-indexed) for reset
	if !strings.Contains(capturedURL, "node=1") {
		t.Errorf("expected URL to contain 'node=1' (0-indexed), got '%s'", capturedURL)
	}
}

func TestResourcePowerCreate_DifferentNodes(t *testing.T) {
	tests := []struct {
		node       int
		expectedID string
	}{
		{1, "power-node-1"},
		{2, "power-node-2"},
		{3, "power-node-3"},
		{4, "power-node-4"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedID, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.String(), "opt=set") && strings.Contains(r.URL.String(), "type=power") {
					w.WriteHeader(http.StatusOK)
					return
				}
				response := map[string]interface{}{
					"response": [][]interface{}{
						{"node1", float64(1)},
						{"node2", float64(1)},
						{"node3", float64(1)},
						{"node4", float64(1)},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			r := resourcePower()
			d := r.TestResourceData()

			_ = d.Set("node", tt.node)
			_ = d.Set("state", "on")

			config := &ProviderConfig{
				Token:    "test-token",
				Endpoint: server.URL,
			}

			diags := resourcePowerCreate(context.Background(), d, config)
			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}

			if d.Id() != tt.expectedID {
				t.Errorf("expected ID '%s', got '%s'", tt.expectedID, d.Id())
			}
		})
	}
}

func TestResourcePowerRead_SetsCurrentState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"response": [][]interface{}{
				{"node1", float64(1)},
				{"node2", float64(0)},
				{"node3", float64(1)},
				{"node4", float64(0)},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	r := resourcePower()
	d := r.TestResourceData()

	_ = d.Set("node", 1)
	d.SetId("power-node-1")

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := resourcePowerRead(context.Background(), d, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !d.Get("current_state").(bool) {
		t.Error("expected current_state to be true for node1")
	}
}

func TestResourcePowerRead_NodeOff(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"response": [][]interface{}{
				{"node1", float64(0)},
				{"node2", float64(0)},
				{"node3", float64(0)},
				{"node4", float64(0)},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	r := resourcePower()
	d := r.TestResourceData()

	_ = d.Set("node", 2)
	d.SetId("power-node-2")

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := resourcePowerRead(context.Background(), d, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Get("current_state").(bool) {
		t.Error("expected current_state to be false for node2")
	}
}

func TestResourcePowerDelete_PowersOffNode(t *testing.T) {
	var capturedURL string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	r := resourcePower()
	d := r.TestResourceData()

	_ = d.Set("node", 1)
	d.SetId("power-node-1")

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := resourcePowerDelete(context.Background(), d, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !strings.Contains(capturedURL, "node1=0") {
		t.Errorf("expected URL to contain 'node1=0' for power off, got '%s'", capturedURL)
	}

	if d.Id() != "" {
		t.Error("expected ID to be cleared after delete")
	}
}

func TestResourcePowerImport_ValidNode(t *testing.T) {
	r := resourcePower()
	d := r.TestResourceData()
	d.SetId("2")

	results, err := resourcePowerImport(context.Background(), d, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Get("node").(int) != 2 {
		t.Errorf("expected node 2, got %d", results[0].Get("node").(int))
	}

	if results[0].Id() != "power-node-2" {
		t.Errorf("expected ID 'power-node-2', got '%s'", results[0].Id())
	}
}

func TestResourcePowerImport_InvalidNode(t *testing.T) {
	tests := []struct {
		id          string
		expectError bool
	}{
		{"0", true},
		{"5", true},
		{"abc", true},
		{"1", false},
		{"4", false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			r := resourcePower()
			d := r.TestResourceData()
			d.SetId(tt.id)

			_, err := resourcePowerImport(context.Background(), d, nil)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestSetNodePower_Success(t *testing.T) {
	var capturedURL string
	var capturedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		capturedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	err := setNodePower(server.URL, "test-token", 3, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(capturedURL, "node3=1") {
		t.Errorf("expected URL to contain 'node3=1', got '%s'", capturedURL)
	}

	if capturedAuth != "Bearer test-token" {
		t.Errorf("expected Authorization 'Bearer test-token', got '%s'", capturedAuth)
	}
}

func TestSetNodePower_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server error"))
	}))
	defer server.Close()

	err := setNodePower(server.URL, "test-token", 1, true)
	if err == nil {
		t.Error("expected error for API failure")
	}
}

func TestResetNode_Success(t *testing.T) {
	var capturedURL string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	err := resetNode(server.URL, "test-token", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Node 3 should be sent as node=2 (0-indexed)
	if !strings.Contains(capturedURL, "node=2") {
		t.Errorf("expected URL to contain 'node=2' (0-indexed), got '%s'", capturedURL)
	}

	if !strings.Contains(capturedURL, "type=reset") {
		t.Errorf("expected URL to contain 'type=reset', got '%s'", capturedURL)
	}
}

func TestResetNode_AllNodes(t *testing.T) {
	tests := []struct {
		inputNode     int
		expectedParam string
	}{
		{1, "node=0"},
		{2, "node=1"},
		{3, "node=2"},
		{4, "node=3"},
	}

	for _, tt := range tests {
		t.Run(string(rune('0'+tt.inputNode)), func(t *testing.T) {
			var capturedURL string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedURL = r.URL.String()
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			err := resetNode(server.URL, "test-token", tt.inputNode)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(capturedURL, tt.expectedParam) {
				t.Errorf("expected URL to contain '%s', got '%s'", tt.expectedParam, capturedURL)
			}
		})
	}
}

func TestResetNode_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	err := resetNode(server.URL, "test-token", 1)
	if err == nil {
		t.Error("expected error for API failure")
	}
}

func TestSetPowerState_AllStates(t *testing.T) {
	tests := []struct {
		state       string
		expectReset bool
	}{
		{"on", false},
		{"off", false},
		{"reset", true},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			var sawReset, sawPower bool

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.String(), "type=reset") {
					sawReset = true
				}
				if strings.Contains(r.URL.String(), "type=power") {
					sawPower = true
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			err := setPowerState(server.URL, "test-token", 1, tt.state)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectReset && !sawReset {
				t.Error("expected reset API call")
			}
			if !tt.expectReset && !sawPower {
				t.Error("expected power API call")
			}
		})
	}
}

func TestSetPowerState_InvalidState(t *testing.T) {
	err := setPowerState("http://localhost", "token", 1, "invalid")
	if err == nil {
		t.Error("expected error for invalid state")
	}
}
