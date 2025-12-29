package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourcePower(t *testing.T) {
	d := dataSourcePower()
	if err := d.InternalValidate(nil, false); err != nil {
		t.Fatalf("data source internal validation failed: %s", err)
	}
}

func TestDataSourcePower_Schema(t *testing.T) {
	d := dataSourcePower()

	expectedFields := []string{
		"node1",
		"node2",
		"node3",
		"node4",
		"nodes",
		"powered_on_count",
		"powered_off_count",
	}

	for _, field := range expectedFields {
		if _, ok := d.Schema[field]; !ok {
			t.Errorf("schema missing '%s' field", field)
		}
	}
}

func TestDataSourcePower_SchemaTypes(t *testing.T) {
	d := dataSourcePower()

	tests := []struct {
		field    string
		expected schema.ValueType
	}{
		{"node1", schema.TypeBool},
		{"node2", schema.TypeBool},
		{"node3", schema.TypeBool},
		{"node4", schema.TypeBool},
		{"nodes", schema.TypeMap},
		{"powered_on_count", schema.TypeInt},
		{"powered_off_count", schema.TypeInt},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			if d.Schema[tt.field].Type != tt.expected {
				t.Errorf("expected %s to be type %v, got %v", tt.field, tt.expected, d.Schema[tt.field].Type)
			}
		})
	}
}

func TestDataSourcePower_AllFieldsComputed(t *testing.T) {
	d := dataSourcePower()

	for name, s := range d.Schema {
		if !s.Computed {
			t.Errorf("field %s should be computed", name)
		}
	}
}

func TestDataSourcePower_HasReadFunction(t *testing.T) {
	d := dataSourcePower()

	if d.ReadContext == nil {
		t.Error("data source should have ReadContext function")
	}
}

func TestDataSourcePowerRead_AllNodesOn(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	d := dataSourcePower()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := dataSourcePowerRead(context.Background(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	// Verify ID
	if rd.Id() != "turingpi-power-status" {
		t.Errorf("expected ID 'turingpi-power-status', got '%s'", rd.Id())
	}

	// Verify all nodes are on
	for i := 1; i <= 4; i++ {
		field := "node" + string(rune('0'+i))
		if !rd.Get(field).(bool) {
			t.Errorf("expected %s to be true", field)
		}
	}

	// Verify counts
	if rd.Get("powered_on_count").(int) != 4 {
		t.Errorf("expected powered_on_count 4, got %d", rd.Get("powered_on_count").(int))
	}
	if rd.Get("powered_off_count").(int) != 0 {
		t.Errorf("expected powered_off_count 0, got %d", rd.Get("powered_off_count").(int))
	}
}

func TestDataSourcePowerRead_AllNodesOff(t *testing.T) {
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

	d := dataSourcePower()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := dataSourcePowerRead(context.Background(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	// Verify all nodes are off
	for i := 1; i <= 4; i++ {
		field := "node" + string(rune('0'+i))
		if rd.Get(field).(bool) {
			t.Errorf("expected %s to be false", field)
		}
	}

	// Verify counts
	if rd.Get("powered_on_count").(int) != 0 {
		t.Errorf("expected powered_on_count 0, got %d", rd.Get("powered_on_count").(int))
	}
	if rd.Get("powered_off_count").(int) != 4 {
		t.Errorf("expected powered_off_count 4, got %d", rd.Get("powered_off_count").(int))
	}
}

func TestDataSourcePowerRead_MixedStatus(t *testing.T) {
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

	d := dataSourcePower()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := dataSourcePowerRead(context.Background(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	// Verify mixed status
	if !rd.Get("node1").(bool) {
		t.Error("expected node1 to be true")
	}
	if rd.Get("node2").(bool) {
		t.Error("expected node2 to be false")
	}
	if !rd.Get("node3").(bool) {
		t.Error("expected node3 to be true")
	}
	if rd.Get("node4").(bool) {
		t.Error("expected node4 to be false")
	}

	// Verify counts
	if rd.Get("powered_on_count").(int) != 2 {
		t.Errorf("expected powered_on_count 2, got %d", rd.Get("powered_on_count").(int))
	}
	if rd.Get("powered_off_count").(int) != 2 {
		t.Errorf("expected powered_off_count 2, got %d", rd.Get("powered_off_count").(int))
	}
}

func TestDataSourcePowerRead_BooleanValues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"response": [][]interface{}{
				{"node1", true},
				{"node2", false},
				{"node3", true},
				{"node4", false},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	d := dataSourcePower()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := dataSourcePowerRead(context.Background(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !rd.Get("node1").(bool) {
		t.Error("expected node1 to be true")
	}
	if rd.Get("node2").(bool) {
		t.Error("expected node2 to be false")
	}
}

func TestDataSourcePowerRead_NodesMap(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"response": [][]interface{}{
				{"node1", float64(1)},
				{"node2", float64(0)},
				{"node3", float64(1)},
				{"node4", float64(1)},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	d := dataSourcePower()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := dataSourcePowerRead(context.Background(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	nodes := rd.Get("nodes").(map[string]interface{})
	if len(nodes) != 4 {
		t.Errorf("expected 4 nodes in map, got %d", len(nodes))
	}

	if nodes["node1"] != true {
		t.Error("expected nodes[node1] to be true")
	}
	if nodes["node2"] != false {
		t.Error("expected nodes[node2] to be false")
	}
}

func TestDataSourcePowerRead_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	d := dataSourcePower()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := dataSourcePowerRead(context.Background(), rd, config)
	if !diags.HasError() {
		t.Error("expected error for API failure")
	}
}

func TestDataSourcePowerRead_AuthHeader(t *testing.T) {
	var capturedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
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

	d := dataSourcePower()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "my-auth-token",
		Endpoint: server.URL,
	}

	diags := dataSourcePowerRead(context.Background(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if capturedAuth != "Bearer my-auth-token" {
		t.Errorf("expected Authorization 'Bearer my-auth-token', got '%s'", capturedAuth)
	}
}

func TestGetPowerStatus_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"response": [][]interface{}{
				{"node1", float64(1)},
				{"node2", float64(0)},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	result, err := getPowerStatus(server.URL, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Response) != 2 {
		t.Errorf("expected 2 response items, got %d", len(result.Response))
	}
}

func TestGetPowerStatus_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer server.Close()

	_, err := getPowerStatus(server.URL, "test-token")
	if err == nil {
		t.Error("expected error for API failure")
	}
}

func TestParsePowerStatus(t *testing.T) {
	tests := []struct {
		name     string
		response *powerStatusResponse
		expected map[string]bool
	}{
		{
			name: "all_on",
			response: &powerStatusResponse{
				Response: [][]interface{}{
					{"node1", float64(1)},
					{"node2", float64(1)},
					{"node3", float64(1)},
					{"node4", float64(1)},
				},
			},
			expected: map[string]bool{"node1": true, "node2": true, "node3": true, "node4": true},
		},
		{
			name: "all_off",
			response: &powerStatusResponse{
				Response: [][]interface{}{
					{"node1", float64(0)},
					{"node2", float64(0)},
					{"node3", float64(0)},
					{"node4", float64(0)},
				},
			},
			expected: map[string]bool{"node1": false, "node2": false, "node3": false, "node4": false},
		},
		{
			name: "mixed",
			response: &powerStatusResponse{
				Response: [][]interface{}{
					{"node1", float64(1)},
					{"node2", float64(0)},
					{"node3", float64(1)},
					{"node4", float64(0)},
				},
			},
			expected: map[string]bool{"node1": true, "node2": false, "node3": true, "node4": false},
		},
		{
			name: "boolean_values",
			response: &powerStatusResponse{
				Response: [][]interface{}{
					{"node1", true},
					{"node2", false},
					{"node3", true},
					{"node4", false},
				},
			},
			expected: map[string]bool{"node1": true, "node2": false, "node3": true, "node4": false},
		},
		{
			name: "empty_response",
			response: &powerStatusResponse{
				Response: [][]interface{}{},
			},
			expected: map[string]bool{"node1": false, "node2": false, "node3": false, "node4": false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePowerStatus(tt.response)

			for node, expected := range tt.expected {
				if result[node] != expected {
					t.Errorf("expected %s to be %v, got %v", node, expected, result[node])
				}
			}
		})
	}
}
