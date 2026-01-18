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

func TestResourceUSB(t *testing.T) {
	r := resourceUSB()
	if err := r.InternalValidate(nil, true); err != nil {
		t.Fatalf("resource internal validation failed: %s", err)
	}
}

func TestResourceUSB_Schema(t *testing.T) {
	r := resourceUSB()

	expectedFields := []string{
		"node",
		"mode",
		"route",
		"current_mode",
		"current_node",
		"current_route",
	}

	for _, field := range expectedFields {
		if _, ok := r.Schema[field]; !ok {
			t.Errorf("schema missing '%s' field", field)
		}
	}
}

func TestResourceUSB_SchemaTypes(t *testing.T) {
	r := resourceUSB()

	tests := []struct {
		field    string
		expected schema.ValueType
	}{
		{"node", schema.TypeInt},
		{"mode", schema.TypeString},
		{"route", schema.TypeString},
		{"current_mode", schema.TypeString},
		{"current_node", schema.TypeInt},
		{"current_route", schema.TypeString},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			if r.Schema[tt.field].Type != tt.expected {
				t.Errorf("expected %s to be type %v, got %v", tt.field, tt.expected, r.Schema[tt.field].Type)
			}
		})
	}
}

func TestResourceUSB_RequiredFields(t *testing.T) {
	r := resourceUSB()

	if !r.Schema["node"].Required {
		t.Error("node should be required")
	}
	if !r.Schema["mode"].Required {
		t.Error("mode should be required")
	}
}

func TestResourceUSB_OptionalFields(t *testing.T) {
	r := resourceUSB()

	if !r.Schema["route"].Optional {
		t.Error("route should be optional")
	}
}

func TestResourceUSB_ComputedFields(t *testing.T) {
	r := resourceUSB()

	computedFields := []string{"current_mode", "current_node", "current_route"}
	for _, field := range computedFields {
		if !r.Schema[field].Computed {
			t.Errorf("%s should be computed", field)
		}
	}
}

func TestResourceUSB_DefaultValues(t *testing.T) {
	r := resourceUSB()

	if r.Schema["route"].Default != "usb-a" {
		t.Errorf("route should default to 'usb-a', got %v", r.Schema["route"].Default)
	}
}

func TestResourceUSB_HasCRUDFunctions(t *testing.T) {
	r := resourceUSB()

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

func TestResourceUSB_HasImporter(t *testing.T) {
	r := resourceUSB()

	if r.Importer == nil {
		t.Error("resource should have Importer")
	}
}

func TestGetUSBAPIMode(t *testing.T) {
	tests := []struct {
		mode     string
		route    string
		expected int
	}{
		{"host", "usb-a", 0},
		{"device", "usb-a", 1},
		{"host", "bmc", 4},
		{"device", "bmc", 5},
	}

	for _, tt := range tests {
		t.Run(tt.mode+"_"+tt.route, func(t *testing.T) {
			result := getUSBAPIMode(tt.mode, tt.route)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestParseUSBStatus(t *testing.T) {
	tests := []struct {
		name          string
		responseData  interface{}
		expectedMode  string
		expectedNode  int
		expectedRoute string
	}{
		{
			name: "host_usb-a_node1",
			responseData: [][]interface{}{
				{"mode", "Host"},
				{"node", float64(0)},
				{"route", "USB-A"},
			},
			expectedMode:  "host",
			expectedNode:  1,
			expectedRoute: "usb-a",
		},
		{
			name: "device_bmc_node2",
			responseData: [][]interface{}{
				{"mode", "Device"},
				{"node", float64(1)},
				{"route", "BMC"},
			},
			expectedMode:  "device",
			expectedNode:  2,
			expectedRoute: "bmc",
		},
		{
			name: "lowercase_values",
			responseData: [][]interface{}{
				{"mode", "host"},
				{"node", float64(2)},
				{"route", "bmc"},
			},
			expectedMode:  "host",
			expectedNode:  3,
			expectedRoute: "bmc",
		},
		{
			name: "usb20_route",
			responseData: [][]interface{}{
				{"mode", "Host"},
				{"node", float64(3)},
				{"route", "USB-2.0"},
			},
			expectedMode:  "host",
			expectedNode:  4,
			expectedRoute: "usb-a",
		},
		{
			name:          "empty_response",
			responseData:  [][]interface{}{},
			expectedMode:  "host",
			expectedNode:  1,
			expectedRoute: "usb-a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.responseData)
			response := &usbStatusResponse{
				Response: json.RawMessage(jsonData),
			}
			mode, node, route := parseUSBStatus(response)

			if mode != tt.expectedMode {
				t.Errorf("expected mode %s, got %s", tt.expectedMode, mode)
			}
			if node != tt.expectedNode {
				t.Errorf("expected node %d, got %d", tt.expectedNode, node)
			}
			if route != tt.expectedRoute {
				t.Errorf("expected route %s, got %s", tt.expectedRoute, route)
			}
		})
	}
}

func TestSetUSBMode_Success(t *testing.T) {
	var capturedURL string
	var capturedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		capturedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	err := setUSBMode(server.URL, "test-token", 1, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify URL contains correct parameters (node 1 -> 0 in API)
	if !strings.Contains(capturedURL, "opt=set") {
		t.Error("URL should contain opt=set")
	}
	if !strings.Contains(capturedURL, "type=usb") {
		t.Error("URL should contain type=usb")
	}
	if !strings.Contains(capturedURL, "mode=0") {
		t.Error("URL should contain mode=0")
	}
	if !strings.Contains(capturedURL, "node=0") {
		t.Error("URL should contain node=0 (0-indexed)")
	}

	// Verify authorization header
	if capturedAuth != "Bearer test-token" {
		t.Errorf("expected Authorization 'Bearer test-token', got '%s'", capturedAuth)
	}
}

func TestSetUSBMode_DifferentNodes(t *testing.T) {
	tests := []struct {
		inputNode   int
		expectedAPI string
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

			err := setUSBMode(server.URL, "test-token", tt.inputNode, 0)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(capturedURL, tt.expectedAPI) {
				t.Errorf("expected URL to contain '%s', got '%s'", tt.expectedAPI, capturedURL)
			}
		})
	}
}

func TestSetUSBMode_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	err := setUSBMode(server.URL, "test-token", 1, 0)
	if err == nil {
		t.Error("expected error for API failure")
	}
}

func TestGetUSBStatus_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify it's a GET request to the correct endpoint
		if !strings.Contains(r.URL.String(), "opt=get") {
			t.Error("expected opt=get in URL")
		}
		if !strings.Contains(r.URL.String(), "type=usb") {
			t.Error("expected type=usb in URL")
		}

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

	result, err := getUSBStatus(server.URL, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify we can parse the response
	mode, node, route := parseUSBStatus(result)
	if mode != "host" {
		t.Errorf("expected mode 'host', got '%s'", mode)
	}
	if node != 1 {
		t.Errorf("expected node 1, got %d", node)
	}
	if route != "usb-a" {
		t.Errorf("expected route 'usb-a', got '%s'", route)
	}
}

func TestGetUSBStatus_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	_, err := getUSBStatus(server.URL, "test-token")
	if err == nil {
		t.Error("expected error for API failure")
	}
}

func TestResourceUSBCreate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "opt=set") {
			w.WriteHeader(http.StatusOK)
			return
		}
		// GET request for reading back state
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

	r := resourceUSB()
	d := r.TestResourceData()

	_ = d.Set("node", 1)
	_ = d.Set("mode", "host")
	_ = d.Set("route", "usb-a")

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := resourceUSBCreate(context.Background(), d, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "usb-node-1" {
		t.Errorf("expected ID 'usb-node-1', got '%s'", d.Id())
	}

	// Check computed values were set
	if d.Get("current_mode").(string) != "host" {
		t.Errorf("expected current_mode 'host', got '%s'", d.Get("current_mode").(string))
	}
}

func TestResourceUSBCreate_DifferentModes(t *testing.T) {
	tests := []struct {
		mode  string
		route string
	}{
		{"host", "usb-a"},
		{"device", "usb-a"},
		{"host", "bmc"},
		{"device", "bmc"},
	}

	for _, tt := range tests {
		t.Run(tt.mode+"_"+tt.route, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.String(), "opt=set") {
					w.WriteHeader(http.StatusOK)
					return
				}
				response := map[string]interface{}{
					"response": [][]interface{}{
						{"mode", tt.mode},
						{"node", float64(0)},
						{"route", tt.route},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			r := resourceUSB()
			d := r.TestResourceData()

			_ = d.Set("node", 1)
			_ = d.Set("mode", tt.mode)
			_ = d.Set("route", tt.route)

			config := &ProviderConfig{
				Token:    "test-token",
				Endpoint: server.URL,
			}

			diags := resourceUSBCreate(context.Background(), d, config)
			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}
		})
	}
}

func TestResourceUSBRead_Success(t *testing.T) {
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

	r := resourceUSB()
	d := r.TestResourceData()
	d.SetId("usb-node-2")

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := resourceUSBRead(context.Background(), d, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Get("current_mode").(string) != "device" {
		t.Errorf("expected current_mode 'device', got '%s'", d.Get("current_mode").(string))
	}
	if d.Get("current_node").(int) != 2 {
		t.Errorf("expected current_node 2, got %d", d.Get("current_node").(int))
	}
	if d.Get("current_route").(string) != "bmc" {
		t.Errorf("expected current_route 'bmc', got '%s'", d.Get("current_route").(string))
	}
}

func TestResourceUSBDelete_ClearsId(t *testing.T) {
	r := resourceUSB()
	d := r.TestResourceData()
	d.SetId("usb-node-1")

	diags := resourceUSBDelete(context.Background(), d, nil)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected empty ID after delete, got '%s'", d.Id())
	}
}

func TestResourceUSBUpdate_ChangesMode(t *testing.T) {
	var capturedMode string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "opt=set") {
			// Capture the mode from the URL
			if strings.Contains(r.URL.String(), "mode=1") {
				capturedMode = "device"
			} else if strings.Contains(r.URL.String(), "mode=0") {
				capturedMode = "host"
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		response := map[string]interface{}{
			"response": [][]interface{}{
				{"mode", "Device"},
				{"node", float64(0)},
				{"route", "USB-A"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	r := resourceUSB()
	d := r.TestResourceData()
	d.SetId("usb-node-1")

	_ = d.Set("node", 1)
	_ = d.Set("mode", "device")
	_ = d.Set("route", "usb-a")

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	diags := resourceUSBUpdate(context.Background(), d, config)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if capturedMode != "device" {
		t.Errorf("expected mode 'device' to be set, got '%s'", capturedMode)
	}
}
