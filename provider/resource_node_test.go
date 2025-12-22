package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceNode(t *testing.T) {
	r := resourceNode()
	if err := r.InternalValidate(nil, true); err != nil {
		t.Fatalf("resource internal validation failed: %s", err)
	}
}

func TestResourceNode_Schema(t *testing.T) {
	r := resourceNode()

	expectedFields := []string{
		"node",
		"firmware_file",
		"power_state",
		"boot_check",
		"login_prompt_timeout",
	}

	for _, field := range expectedFields {
		if _, ok := r.Schema[field]; !ok {
			t.Errorf("schema missing '%s' field", field)
		}
	}
}

func TestResourceNode_SchemaTypes(t *testing.T) {
	r := resourceNode()

	tests := []struct {
		field    string
		expected schema.ValueType
	}{
		{"node", schema.TypeInt},
		{"firmware_file", schema.TypeString},
		{"power_state", schema.TypeString},
		{"boot_check", schema.TypeBool},
		{"login_prompt_timeout", schema.TypeInt},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			if r.Schema[tt.field].Type != tt.expected {
				t.Errorf("expected %s to be type %v, got %v", tt.field, tt.expected, r.Schema[tt.field].Type)
			}
		})
	}
}

func TestResourceNode_RequiredFields(t *testing.T) {
	r := resourceNode()

	if !r.Schema["node"].Required {
		t.Error("node should be required")
	}
}

func TestResourceNode_OptionalFields(t *testing.T) {
	r := resourceNode()

	optionalFields := []string{
		"firmware_file",
		"power_state",
		"boot_check",
		"login_prompt_timeout",
	}

	for _, field := range optionalFields {
		if !r.Schema[field].Optional {
			t.Errorf("%s should be optional", field)
		}
	}
}

func TestResourceNode_DefaultValues(t *testing.T) {
	r := resourceNode()

	// power_state defaults to "on"
	if r.Schema["power_state"].Default != "on" {
		t.Errorf("power_state should default to 'on', got %v", r.Schema["power_state"].Default)
	}

	// boot_check defaults to false
	if r.Schema["boot_check"].Default != false {
		t.Errorf("boot_check should default to false, got %v", r.Schema["boot_check"].Default)
	}

	// login_prompt_timeout defaults to 60
	if r.Schema["login_prompt_timeout"].Default != 60 {
		t.Errorf("login_prompt_timeout should default to 60, got %v", r.Schema["login_prompt_timeout"].Default)
	}
}

func TestResourceNode_HasCRUDFunctions(t *testing.T) {
	r := resourceNode()

	if r.Create == nil {
		t.Error("resource should have Create function")
	}

	if r.Read == nil {
		t.Error("resource should have Read function")
	}

	if r.Update == nil {
		t.Error("resource should have Update function")
	}

	if r.Delete == nil {
		t.Error("resource should have Delete function")
	}
}

func TestResourceNodeProvision_SetsId(t *testing.T) {
	r := resourceNode()
	d := r.TestResourceData()

	d.Set("node", 1)
	d.Set("power_state", "on")
	d.Set("boot_check", false)

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: "https://test.local",
	}

	err := resourceNodeProvision(d, config)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	expectedId := "node-1"
	if d.Id() != expectedId {
		t.Errorf("expected ID %s, got %s", expectedId, d.Id())
	}
}

func TestResourceNodeProvision_DifferentNodes(t *testing.T) {
	r := resourceNode()

	testCases := []struct {
		node       int
		expectedId string
	}{
		{1, "node-1"},
		{2, "node-2"},
		{3, "node-3"},
		{4, "node-4"},
	}

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: "https://test.local",
	}

	for _, tc := range testCases {
		t.Run(tc.expectedId, func(t *testing.T) {
			d := r.TestResourceData()
			d.Set("node", tc.node)
			d.Set("power_state", "on")
			d.Set("boot_check", false)

			err := resourceNodeProvision(d, config)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if d.Id() != tc.expectedId {
				t.Errorf("expected ID %s, got %s", tc.expectedId, d.Id())
			}
		})
	}
}

func TestResourceNodeProvision_PowerStateOn(t *testing.T) {
	r := resourceNode()
	d := r.TestResourceData()

	d.Set("node", 1)
	d.Set("power_state", "on")
	d.Set("boot_check", false)

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: "https://test.local",
	}

	err := resourceNodeProvision(d, config)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResourceNodeProvision_PowerStateOff(t *testing.T) {
	r := resourceNode()
	d := r.TestResourceData()

	d.Set("node", 1)
	d.Set("power_state", "off")
	d.Set("boot_check", false)

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: "https://test.local",
	}

	err := resourceNodeProvision(d, config)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResourceNodeProvision_WithFirmware(t *testing.T) {
	r := resourceNode()
	d := r.TestResourceData()

	d.Set("node", 1)
	d.Set("firmware_file", "/path/to/firmware.img")
	d.Set("power_state", "on")
	d.Set("boot_check", false)

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: "https://test.local",
	}

	err := resourceNodeProvision(d, config)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResourceNodeProvision_WithBootCheck(t *testing.T) {
	// Create mock server that returns login prompt
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Boot complete\nlogin:"))
	}))
	defer server.Close()

	r := resourceNode()
	d := r.TestResourceData()

	d.Set("node", 1)
	d.Set("power_state", "on")
	d.Set("boot_check", true)
	d.Set("login_prompt_timeout", 1)

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	err := resourceNodeProvision(d, config)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResourceNodeProvision_BootCheckTimeout(t *testing.T) {
	// Create mock server that never returns login prompt
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Still booting..."))
	}))
	defer server.Close()

	r := resourceNode()
	d := r.TestResourceData()

	d.Set("node", 1)
	d.Set("power_state", "on")
	d.Set("boot_check", true)
	d.Set("login_prompt_timeout", 1)

	config := &ProviderConfig{
		Token:    "test-token",
		Endpoint: server.URL,
	}

	err := resourceNodeProvision(d, config)
	if err == nil {
		t.Fatal("expected error for boot check timeout, got nil")
	}
}

func TestResourceNodeStatus_SetsPowerState(t *testing.T) {
	r := resourceNode()
	d := r.TestResourceData()

	d.Set("node", 1)
	d.SetId("node-1")

	err := resourceNodeStatus(d, nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// checkPowerStatus currently returns "off"
	powerState := d.Get("power_state").(string)
	if powerState != "off" {
		t.Errorf("expected power_state 'off', got %s", powerState)
	}
}

func TestResourceNodeDelete_TurnsOffNode(t *testing.T) {
	r := resourceNode()
	d := r.TestResourceData()

	d.Set("node", 1)
	d.SetId("node-1")

	err := resourceNodeDelete(d, nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResourceNodeDelete_DifferentNodes(t *testing.T) {
	r := resourceNode()

	nodes := []int{1, 2, 3, 4}

	for _, node := range nodes {
		t.Run("node_"+string(rune('0'+node)), func(t *testing.T) {
			d := r.TestResourceData()
			d.Set("node", node)
			d.SetId("node-" + string(rune('0'+node)))

			err := resourceNodeDelete(d, nil)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}
