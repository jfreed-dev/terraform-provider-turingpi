package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceInfo(t *testing.T) {
	d := dataSourceInfo()
	if err := d.InternalValidate(nil, false); err != nil {
		t.Fatalf("data source internal validation failed: %s", err)
	}
}

func TestDataSourceInfo_Schema(t *testing.T) {
	d := dataSourceInfo()

	expectedFields := []string{
		"api_version",
		"daemon_version",
		"buildroot_version",
		"firmware_version",
		"build_time",
		"network_interfaces",
		"storage_devices",
		"nodes",
	}

	for _, field := range expectedFields {
		if _, ok := d.Schema[field]; !ok {
			t.Errorf("schema missing '%s' field", field)
		}
	}
}

func TestDataSourceInfo_SchemaTypes(t *testing.T) {
	d := dataSourceInfo()

	tests := []struct {
		field    string
		expected schema.ValueType
	}{
		{"api_version", schema.TypeString},
		{"daemon_version", schema.TypeString},
		{"buildroot_version", schema.TypeString},
		{"firmware_version", schema.TypeString},
		{"build_time", schema.TypeString},
		{"network_interfaces", schema.TypeList},
		{"storage_devices", schema.TypeList},
		{"nodes", schema.TypeMap},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			if d.Schema[tt.field].Type != tt.expected {
				t.Errorf("expected %s to be type %v, got %v", tt.field, tt.expected, d.Schema[tt.field].Type)
			}
		})
	}
}

func TestDataSourceInfo_AllFieldsComputed(t *testing.T) {
	d := dataSourceInfo()

	for name, s := range d.Schema {
		if !s.Computed {
			t.Errorf("field %s should be computed", name)
		}
	}
}

func TestDataSourceInfo_HasReadFunction(t *testing.T) {
	d := dataSourceInfo()

	if d.ReadContext == nil {
		t.Error("data source should have ReadContext function")
	}
}

func TestDataSourceInfo_NetworkInterfaceSchema(t *testing.T) {
	d := dataSourceInfo()

	niSchema := d.Schema["network_interfaces"]
	if niSchema.Elem == nil {
		t.Fatal("network_interfaces should have Elem defined")
	}

	elemResource, ok := niSchema.Elem.(*schema.Resource)
	if !ok {
		t.Fatal("network_interfaces Elem should be a *schema.Resource")
	}

	expectedFields := []string{"device", "ip", "mac"}
	for _, field := range expectedFields {
		if _, ok := elemResource.Schema[field]; !ok {
			t.Errorf("network_interfaces element missing '%s' field", field)
		}
	}
}

func TestDataSourceInfo_StorageDeviceSchema(t *testing.T) {
	d := dataSourceInfo()

	sdSchema := d.Schema["storage_devices"]
	if sdSchema.Elem == nil {
		t.Fatal("storage_devices should have Elem defined")
	}

	elemResource, ok := sdSchema.Elem.(*schema.Resource)
	if !ok {
		t.Fatal("storage_devices Elem should be a *schema.Resource")
	}

	expectedFields := []string{"name", "total_bytes", "used_bytes", "free_bytes"}
	for _, field := range expectedFields {
		if _, ok := elemResource.Schema[field]; !ok {
			t.Errorf("storage_devices element missing '%s' field", field)
		}
	}
}

func TestDataSourceInfoRead_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		apiType := query.Get("type")

		w.Header().Set("Content-Type", "application/json")

		switch apiType {
		case "about":
			response := map[string]interface{}{
				"response": [][]interface{}{
					{"api", "1.0"},
					{"version", "2.0.5"},
					{"buildroot", "2023.02"},
					{"firmware", "1.1.0"},
					{"buildtime", "2024-01-15T10:30:00Z"},
				},
			}
			_ = json.NewEncoder(w).Encode(response)
		case "info":
			response := map[string]interface{}{
				"response": map[string]interface{}{
					"network": []map[string]string{
						{"device": "eth0", "ip": "192.168.1.100", "mac": "00:11:22:33:44:55"},
					},
					"storage": []map[string]interface{}{
						{"name": "bmc", "total": 1073741824, "free": 536870912, "use": 536870912},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(response)
		case "power":
			response := map[string]interface{}{
				"response": [][]interface{}{
					{"node1", float64(1)},
					{"node2", float64(0)},
					{"node3", float64(1)},
					{"node4", float64(0)},
				},
			}
			_ = json.NewEncoder(w).Encode(response)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	d := dataSourceInfo()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := dataSourceInfoRead(context.Background(), rd, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	// Verify ID is set
	if rd.Id() != "turingpi-bmc-info" {
		t.Errorf("expected ID 'turingpi-bmc-info', got '%s'", rd.Id())
	}

	// Verify version info
	if v := rd.Get("api_version").(string); v != "1.0" {
		t.Errorf("expected api_version '1.0', got '%s'", v)
	}
	if v := rd.Get("daemon_version").(string); v != "2.0.5" {
		t.Errorf("expected daemon_version '2.0.5', got '%s'", v)
	}
	if v := rd.Get("buildroot_version").(string); v != "2023.02" {
		t.Errorf("expected buildroot_version '2023.02', got '%s'", v)
	}
	if v := rd.Get("firmware_version").(string); v != "1.1.0" {
		t.Errorf("expected firmware_version '1.1.0', got '%s'", v)
	}

	// Verify network interfaces
	networkInterfaces := rd.Get("network_interfaces").([]interface{})
	if len(networkInterfaces) != 1 {
		t.Errorf("expected 1 network interface, got %d", len(networkInterfaces))
	}

	// Verify storage devices
	storageDevices := rd.Get("storage_devices").([]interface{})
	if len(storageDevices) != 1 {
		t.Errorf("expected 1 storage device, got %d", len(storageDevices))
	}

	// Verify nodes power status
	nodes := rd.Get("nodes").(map[string]interface{})
	if len(nodes) != 4 {
		t.Errorf("expected 4 nodes, got %d", len(nodes))
	}
}

func TestDataSourceInfoRead_AboutAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("type") == "about" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"response": {}}`))
	}))
	defer server.Close()

	d := dataSourceInfo()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := dataSourceInfoRead(context.Background(), rd, config)
	if !diags.HasError() {
		t.Error("expected error for API failure")
	}
}

func TestDataSourceInfoRead_InfoAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		switch query.Get("type") {
		case "about":
			response := map[string]interface{}{
				"response": [][]interface{}{},
			}
			_ = json.NewEncoder(w).Encode(response)
		case "info":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"response": {}}`))
		}
	}))
	defer server.Close()

	d := dataSourceInfo()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := dataSourceInfoRead(context.Background(), rd, config)
	if !diags.HasError() {
		t.Error("expected error for API failure")
	}
}

func TestDataSourceInfoRead_PowerAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		switch query.Get("type") {
		case "about":
			response := map[string]interface{}{
				"response": [][]interface{}{},
			}
			_ = json.NewEncoder(w).Encode(response)
		case "info":
			response := map[string]interface{}{
				"response": map[string]interface{}{
					"network": []interface{}{},
					"storage": []interface{}{},
				},
			}
			_ = json.NewEncoder(w).Encode(response)
		case "power":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	d := dataSourceInfo()
	rd := d.TestResourceData()

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := dataSourceInfoRead(context.Background(), rd, config)
	if !diags.HasError() {
		t.Error("expected error for API failure")
	}
}

func TestFetchBMCAbout_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("expected Authorization 'Bearer test-token', got '%s'", auth)
		}

		response := map[string]interface{}{
			"response": [][]interface{}{
				{"api", "1.0"},
				{"version", "2.0.5"},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	result, err := fetchBMCAbout(server.URL, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Response) != 2 {
		t.Errorf("expected 2 response items, got %d", len(result.Response))
	}
}

func TestFetchBMCInfo_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"response": map[string]interface{}{
				"network": []map[string]string{
					{"device": "eth0", "ip": "192.168.1.100", "mac": "00:11:22:33:44:55"},
				},
				"storage": []map[string]interface{}{
					{"name": "bmc", "total": 1073741824, "free": 536870912, "use": 536870912},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	result, err := fetchBMCInfo(server.URL, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Response.Network) != 1 {
		t.Errorf("expected 1 network interface, got %d", len(result.Response.Network))
	}
	if len(result.Response.Storage) != 1 {
		t.Errorf("expected 1 storage device, got %d", len(result.Response.Storage))
	}
}

func TestFetchBMCPower_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"response": [][]interface{}{
				{"node1", float64(1)},
				{"node2", float64(0)},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	result, err := fetchBMCPower(server.URL, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Response) != 2 {
		t.Errorf("expected 2 response items, got %d", len(result.Response))
	}
}

func TestSetAboutData(t *testing.T) {
	d := dataSourceInfo()
	rd := d.TestResourceData()

	aboutData := &bmcAboutResponse{
		Response: [][]interface{}{
			{"api", "1.0"},
			{"version", "2.0.5"},
			{"buildroot", "2023.02"},
			{"firmware", "1.1.0"},
			{"buildtime", "2024-01-15T10:30:00Z"},
		},
	}

	err := setAboutData(rd, aboutData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v := rd.Get("api_version").(string); v != "1.0" {
		t.Errorf("expected api_version '1.0', got '%s'", v)
	}
	if v := rd.Get("daemon_version").(string); v != "2.0.5" {
		t.Errorf("expected daemon_version '2.0.5', got '%s'", v)
	}
	if v := rd.Get("buildroot_version").(string); v != "2023.02" {
		t.Errorf("expected buildroot_version '2023.02', got '%s'", v)
	}
	if v := rd.Get("firmware_version").(string); v != "1.1.0" {
		t.Errorf("expected firmware_version '1.1.0', got '%s'", v)
	}
	if v := rd.Get("build_time").(string); v != "2024-01-15T10:30:00Z" {
		t.Errorf("expected build_time '2024-01-15T10:30:00Z', got '%s'", v)
	}
}

func TestSetPowerData_BoolValues(t *testing.T) {
	d := dataSourceInfo()
	rd := d.TestResourceData()

	powerData := &bmcPowerResponse{
		Response: [][]interface{}{
			{"node1", true},
			{"node2", false},
		},
	}

	err := setPowerData(rd, powerData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nodes := rd.Get("nodes").(map[string]interface{})
	if nodes["node1"] != true {
		t.Errorf("expected node1 to be true")
	}
	if nodes["node2"] != false {
		t.Errorf("expected node2 to be false")
	}
}

func TestSetPowerData_NumericValues(t *testing.T) {
	d := dataSourceInfo()
	rd := d.TestResourceData()

	powerData := &bmcPowerResponse{
		Response: [][]interface{}{
			{"node1", float64(1)},
			{"node2", float64(0)},
		},
	}

	err := setPowerData(rd, powerData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nodes := rd.Get("nodes").(map[string]interface{})
	if nodes["node1"] != true {
		t.Errorf("expected node1 to be true")
	}
	if nodes["node2"] != false {
		t.Errorf("expected node2 to be false")
	}
}
