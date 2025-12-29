package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceSDCard(t *testing.T) {
	d := dataSourceSDCard()
	if err := d.InternalValidate(nil, false); err != nil {
		t.Fatalf("data source internal validation failed: %s", err)
	}
}

func TestDataSourceSDCard_Schema(t *testing.T) {
	d := dataSourceSDCard()

	expectedFields := []string{
		"total_bytes",
		"used_bytes",
		"free_bytes",
		"total_gb",
		"used_gb",
		"free_gb",
		"used_percent",
	}

	for _, field := range expectedFields {
		if _, ok := d.Schema[field]; !ok {
			t.Errorf("schema missing '%s' field", field)
		}
	}
}

func TestDataSourceSDCard_SchemaTypes(t *testing.T) {
	d := dataSourceSDCard()

	tests := []struct {
		field    string
		expected schema.ValueType
	}{
		{"total_bytes", schema.TypeInt},
		{"used_bytes", schema.TypeInt},
		{"free_bytes", schema.TypeInt},
		{"total_gb", schema.TypeFloat},
		{"used_gb", schema.TypeFloat},
		{"free_gb", schema.TypeFloat},
		{"used_percent", schema.TypeFloat},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			if d.Schema[tt.field].Type != tt.expected {
				t.Errorf("expected %s to be type %v, got %v", tt.field, tt.expected, d.Schema[tt.field].Type)
			}
		})
	}
}

func TestDataSourceSDCard_AllFieldsComputed(t *testing.T) {
	d := dataSourceSDCard()

	for name, s := range d.Schema {
		if !s.Computed {
			t.Errorf("field %s should be computed", name)
		}
	}
}

func TestDataSourceSDCard_HasReadFunction(t *testing.T) {
	d := dataSourceSDCard()

	if d.ReadContext == nil {
		t.Error("data source should have ReadContext function")
	}
}

func TestDataSourceSDCardRead_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("type") == "sdcard" {
			// 32GB card with 8GB used
			response := map[string]interface{}{
				"response": []map[string]interface{}{
					{
						"total": int64(32 * 1024 * 1024 * 1024), // 32GB
						"use":   int64(8 * 1024 * 1024 * 1024),  // 8GB
						"free":  int64(24 * 1024 * 1024 * 1024), // 24GB
					},
				},
			}
			_ = json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	d := dataSourceSDCard()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := dataSourceSDCardRead(context.Background(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	// Verify ID is set
	if rd.Id() != "turingpi-sdcard" {
		t.Errorf("expected ID 'turingpi-sdcard', got '%s'", rd.Id())
	}

	// Verify byte values
	expectedTotal := int64(32 * 1024 * 1024 * 1024)
	if v := rd.Get("total_bytes").(int); int64(v) != expectedTotal {
		t.Errorf("expected total_bytes %d, got %d", expectedTotal, v)
	}

	expectedUsed := int64(8 * 1024 * 1024 * 1024)
	if v := rd.Get("used_bytes").(int); int64(v) != expectedUsed {
		t.Errorf("expected used_bytes %d, got %d", expectedUsed, v)
	}

	expectedFree := int64(24 * 1024 * 1024 * 1024)
	if v := rd.Get("free_bytes").(int); int64(v) != expectedFree {
		t.Errorf("expected free_bytes %d, got %d", expectedFree, v)
	}

	// Verify GB values (approximate due to floating point)
	totalGB := rd.Get("total_gb").(float64)
	if totalGB < 31.9 || totalGB > 32.1 {
		t.Errorf("expected total_gb ~32, got %f", totalGB)
	}

	usedGB := rd.Get("used_gb").(float64)
	if usedGB < 7.9 || usedGB > 8.1 {
		t.Errorf("expected used_gb ~8, got %f", usedGB)
	}

	freeGB := rd.Get("free_gb").(float64)
	if freeGB < 23.9 || freeGB > 24.1 {
		t.Errorf("expected free_gb ~24, got %f", freeGB)
	}

	// Verify percentage (8/32 = 25%)
	usedPercent := rd.Get("used_percent").(float64)
	if usedPercent < 24.9 || usedPercent > 25.1 {
		t.Errorf("expected used_percent ~25, got %f", usedPercent)
	}
}

func TestDataSourceSDCardRead_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	d := dataSourceSDCard()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := dataSourceSDCardRead(context.Background(), rd, config)
	if !diags.HasError() {
		t.Error("expected error for API failure")
	}
}

func TestDataSourceSDCardRead_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"response": []map[string]interface{}{},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	d := dataSourceSDCard()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := dataSourceSDCardRead(context.Background(), rd, config)
	if !diags.HasError() {
		t.Error("expected error for empty response")
	}
}

func TestFetchSDCardInfo_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("expected Authorization 'Bearer test-token', got '%s'", auth)
		}

		query := r.URL.Query()
		if query.Get("type") != "sdcard" {
			t.Errorf("expected type 'sdcard', got '%s'", query.Get("type"))
		}

		response := map[string]interface{}{
			"response": []map[string]interface{}{
				{"total": 1073741824, "use": 536870912, "free": 536870912},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	result, err := fetchSDCardInfo(server.URL, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Response) != 1 {
		t.Errorf("expected 1 response item, got %d", len(result.Response))
	}

	if result.Response[0].Total != 1073741824 {
		t.Errorf("expected total 1073741824, got %d", result.Response[0].Total)
	}
}

func TestFetchSDCardInfo_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	_, err := fetchSDCardInfo(server.URL, "test-token")
	if err == nil {
		t.Error("expected error for API failure")
	}
}

func TestDataSourceSDCardRead_ZeroTotal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"response": []map[string]interface{}{
				{"total": 0, "use": 0, "free": 0},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	d := dataSourceSDCard()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := dataSourceSDCardRead(context.Background(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	// Verify percentage is 0 when total is 0 (avoid division by zero)
	usedPercent := rd.Get("used_percent").(float64)
	if usedPercent != 0 {
		t.Errorf("expected used_percent 0 when total is 0, got %f", usedPercent)
	}
}
