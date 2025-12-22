package provider

import (
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

	// Check node field exists
	if _, ok := r.Schema["node"]; !ok {
		t.Error("schema missing 'node' field")
	}

	// Check state field exists
	if _, ok := r.Schema["state"]; !ok {
		t.Error("schema missing 'state' field")
	}
}

func TestResourcePower_SchemaTypes(t *testing.T) {
	r := resourcePower()

	tests := []struct {
		field    string
		expected schema.ValueType
	}{
		{"node", schema.TypeInt},
		{"state", schema.TypeBool},
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

func TestResourcePower_HasCRUDFunctions(t *testing.T) {
	r := resourcePower()

	//nolint:staticcheck // SA1019: intentionally testing deprecated Create field
	if r.Create == nil {
		t.Error("resource should have Create function")
	}

	//nolint:staticcheck // SA1019: intentionally testing deprecated Read field
	if r.Read == nil {
		t.Error("resource should have Read function")
	}

	//nolint:staticcheck // SA1019: intentionally testing deprecated Update field
	if r.Update == nil {
		t.Error("resource should have Update function")
	}

	//nolint:staticcheck // SA1019: intentionally testing deprecated Delete field
	if r.Delete == nil {
		t.Error("resource should have Delete function")
	}
}

func TestResourcePowerSet_SetsId(t *testing.T) {
	r := resourcePower()
	d := r.TestResourceData()

	_ = d.Set("node", 1)
	_ = d.Set("state", true)

	err := resourcePowerSet(d, nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	expectedId := "node-1"
	if d.Id() != expectedId {
		t.Errorf("expected ID %s, got %s", expectedId, d.Id())
	}
}

func TestResourcePowerSet_DifferentNodes(t *testing.T) {
	r := resourcePower()

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
			_ = d.Set("node", tc.node)
			_ = d.Set("state", true)

			err := resourcePowerSet(d, nil)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if d.Id() != tc.expectedId {
				t.Errorf("expected ID %s, got %s", tc.expectedId, d.Id())
			}
		})
	}
}

func TestResourcePowerSet_PowerStates(t *testing.T) {
	r := resourcePower()

	testCases := []struct {
		name  string
		state bool
	}{
		{"power on", true},
		{"power off", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := r.TestResourceData()
			_ = d.Set("node", 1)
			_ = d.Set("state", tc.state)

			err := resourcePowerSet(d, nil)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			// Verify the function completes without error
			if d.Id() == "" {
				t.Error("expected ID to be set")
			}
		})
	}
}

func TestResourcePowerRead_DoesNotError(t *testing.T) {
	r := resourcePower()
	d := r.TestResourceData()
	d.SetId("node-1")

	err := resourcePowerRead(d, nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResourcePowerDelete_DoesNotError(t *testing.T) {
	r := resourcePower()
	d := r.TestResourceData()
	d.SetId("node-1")

	err := resourcePowerDelete(d, nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}
