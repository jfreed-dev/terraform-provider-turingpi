package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceFlash(t *testing.T) {
	r := resourceFlash()
	if err := r.InternalValidate(nil, true); err != nil {
		t.Fatalf("resource internal validation failed: %s", err)
	}
}

func TestResourceFlash_Schema(t *testing.T) {
	r := resourceFlash()

	// Check node field exists
	if _, ok := r.Schema["node"]; !ok {
		t.Error("schema missing 'node' field")
	}

	// Check firmware_file field exists
	if _, ok := r.Schema["firmware_file"]; !ok {
		t.Error("schema missing 'firmware_file' field")
	}
}

func TestResourceFlash_SchemaTypes(t *testing.T) {
	r := resourceFlash()

	tests := []struct {
		field    string
		expected schema.ValueType
	}{
		{"node", schema.TypeInt},
		{"firmware_file", schema.TypeString},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			if r.Schema[tt.field].Type != tt.expected {
				t.Errorf("expected %s to be type %v, got %v", tt.field, tt.expected, r.Schema[tt.field].Type)
			}
		})
	}
}

func TestResourceFlash_RequiredFields(t *testing.T) {
	r := resourceFlash()

	if !r.Schema["node"].Required {
		t.Error("node should be required")
	}

	if !r.Schema["firmware_file"].Required {
		t.Error("firmware_file should be required")
	}
}

func TestResourceFlash_ForceNewFields(t *testing.T) {
	r := resourceFlash()

	if !r.Schema["node"].ForceNew {
		t.Error("node should have ForceNew=true")
	}

	if !r.Schema["firmware_file"].ForceNew {
		t.Error("firmware_file should have ForceNew=true")
	}
}

func TestResourceFlash_HasCRUDFunctions(t *testing.T) {
	r := resourceFlash()

	if r.Create == nil {
		t.Error("resource should have Create function")
	}

	if r.Read == nil {
		t.Error("resource should have Read function")
	}

	// Flash resource should NOT have Update (uses ForceNew instead)
	if r.Update != nil {
		t.Error("resource should NOT have Update function (uses ForceNew)")
	}

	if r.Delete == nil {
		t.Error("resource should have Delete function")
	}
}

func TestResourceFlashCreate_SetsId(t *testing.T) {
	r := resourceFlash()
	d := r.TestResourceData()

	d.Set("node", 1)
	d.Set("firmware_file", "/path/to/firmware.img")

	err := resourceFlashCreate(d, nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	expectedId := "node-1"
	if d.Id() != expectedId {
		t.Errorf("expected ID %s, got %s", expectedId, d.Id())
	}
}

func TestResourceFlashCreate_DifferentNodes(t *testing.T) {
	r := resourceFlash()

	testCases := []struct {
		node       int
		expectedId string
	}{
		{1, "node-1"},
		{2, "node-2"},
		{3, "node-3"},
		{4, "node-4"},
	}

	for _, tc := range testCases {
		t.Run(tc.expectedId, func(t *testing.T) {
			d := r.TestResourceData()
			d.Set("node", tc.node)
			d.Set("firmware_file", "/path/to/firmware.img")

			err := resourceFlashCreate(d, nil)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if d.Id() != tc.expectedId {
				t.Errorf("expected ID %s, got %s", tc.expectedId, d.Id())
			}
		})
	}
}

func TestResourceFlashCreate_DifferentFirmwareFiles(t *testing.T) {
	r := resourceFlash()

	firmwareFiles := []string{
		"/path/to/firmware.img",
		"/another/path/firmware.bin",
		"./relative/path/image.img",
	}

	for _, firmware := range firmwareFiles {
		t.Run(firmware, func(t *testing.T) {
			d := r.TestResourceData()
			d.Set("node", 1)
			d.Set("firmware_file", firmware)

			err := resourceFlashCreate(d, nil)
			if err != nil {
				t.Fatalf("unexpected error for firmware %s: %s", firmware, err)
			}

			if d.Id() == "" {
				t.Error("expected ID to be set")
			}
		})
	}
}

func TestResourceFlashRead_DoesNotError(t *testing.T) {
	r := resourceFlash()
	d := r.TestResourceData()
	d.SetId("node-1")

	err := resourceFlashRead(d, nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResourceFlashDelete_DoesNotError(t *testing.T) {
	r := resourceFlash()
	d := r.TestResourceData()
	d.SetId("node-1")

	err := resourceFlashDelete(d, nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}
