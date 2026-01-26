package provider

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

	//nolint:staticcheck // SA1019: intentionally testing deprecated Create field
	if r.Create == nil {
		t.Error("resource should have Create function")
	}

	//nolint:staticcheck // SA1019: intentionally testing deprecated Read field
	if r.Read == nil {
		t.Error("resource should have Read function")
	}

	// Flash resource should NOT have Update (uses ForceNew instead)
	//nolint:staticcheck // SA1019: intentionally testing deprecated Update field
	if r.Update != nil {
		t.Error("resource should NOT have Update function (uses ForceNew)")
	}

	//nolint:staticcheck // SA1019: intentionally testing deprecated Delete field
	if r.Delete == nil {
		t.Error("resource should have Delete function")
	}
}

func TestResourceFlashCreate_FileNotFound(t *testing.T) {
	r := resourceFlash()
	d := r.TestResourceData()

	_ = d.Set("node", 1)
	_ = d.Set("firmware_file", "/nonexistent/firmware.img")

	config := &ProviderConfig{
		Endpoint: "https://example.com",
		Token:    "test-token",
	}

	err := resourceFlashCreate(d, config)
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
	if !strings.Contains(err.Error(), "failed to open firmware file") {
		t.Errorf("expected file open error, got: %s", err)
	}
}

func TestGetFlashStatus_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `{"Done": [{"secs": 100, "nanos": 123456}, 12345678]}`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	status, err := getFlashStatus(server.URL, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Done == nil {
		t.Error("expected Done status")
	}
}

func TestGetFlashStatus_Flashing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `{"Flashing": {"bytes_written": 5000, "total_bytes": 10000}}`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	status, err := getFlashStatus(server.URL, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Flashing == nil {
		t.Fatal("expected Flashing status")
	}
	if status.Flashing.BytesWritten != 5000 {
		t.Errorf("expected 5000 bytes written, got %d", status.Flashing.BytesWritten)
	}
}

func TestGetFlashStatus_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer server.Close()

	_, err := getFlashStatus(server.URL, "test-token")
	if err == nil {
		t.Error("expected error for API failure")
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

func TestFlashStatusResponse_isTransferring_NewFormat(t *testing.T) {
	// BMC firmware 2.0.5+ returns an object format
	status := &flashStatusResponse{
		Transferring: []byte(`{"id":395533509,"process_name":"Node 1 upgrade service","size":1103253504,"cancelled":false,"bytes_written":550000000}`),
	}

	inProgress, bytesWritten, totalBytes := status.isTransferring()
	if !inProgress {
		t.Error("expected transfer to be in progress")
	}
	if bytesWritten != 550000000 {
		t.Errorf("expected bytesWritten=550000000, got %d", bytesWritten)
	}
	if totalBytes != 1103253504 {
		t.Errorf("expected totalBytes=1103253504, got %d", totalBytes)
	}
}

func TestFlashStatusResponse_isTransferring_OldFormat(t *testing.T) {
	// Legacy BMC firmware returns an array format [bytes_written, total_bytes]
	status := &flashStatusResponse{
		Transferring: []byte(`[550000000, 1103253504]`),
	}

	inProgress, bytesWritten, totalBytes := status.isTransferring()
	if !inProgress {
		t.Error("expected transfer to be in progress")
	}
	if bytesWritten != 550000000 {
		t.Errorf("expected bytesWritten=550000000, got %d", bytesWritten)
	}
	if totalBytes != 1103253504 {
		t.Errorf("expected totalBytes=1103253504, got %d", totalBytes)
	}
}

func TestFlashStatusResponse_isTransferring_Nil(t *testing.T) {
	status := &flashStatusResponse{
		Transferring: nil,
	}

	inProgress, bytesWritten, totalBytes := status.isTransferring()
	if inProgress {
		t.Error("expected no transfer in progress")
	}
	if bytesWritten != 0 || totalBytes != 0 {
		t.Error("expected zero values for nil Transferring")
	}
}

func TestFlashStatusResponse_isTransferring_Empty(t *testing.T) {
	status := &flashStatusResponse{
		Transferring: []byte{},
	}

	inProgress, bytesWritten, totalBytes := status.isTransferring()
	if inProgress {
		t.Error("expected no transfer in progress for empty data")
	}
	if bytesWritten != 0 || totalBytes != 0 {
		t.Error("expected zero values for empty Transferring")
	}
}

func TestGetFlashStatus_TransferringNewFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// BMC firmware 2.0.5+ response format
		response := `{"Transferring":{"id":395533509,"process_name":"Node 1 upgrade service","size":1103253504,"cancelled":false,"bytes_written":550000000}}`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	status, err := getFlashStatus(server.URL, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	inProgress, bytesWritten, totalBytes := status.isTransferring()
	if !inProgress {
		t.Error("expected transfer in progress")
	}
	if bytesWritten != 550000000 {
		t.Errorf("expected 550000000 bytes written, got %d", bytesWritten)
	}
	if totalBytes != 1103253504 {
		t.Errorf("expected 1103253504 total bytes, got %d", totalBytes)
	}
}

func TestGetFlashStatus_TransferringOldFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Legacy BMC firmware response format
		response := `{"Transferring":[550000000, 1103253504]}`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	status, err := getFlashStatus(server.URL, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	inProgress, bytesWritten, totalBytes := status.isTransferring()
	if !inProgress {
		t.Error("expected transfer in progress")
	}
	if bytesWritten != 550000000 {
		t.Errorf("expected 550000000 bytes written, got %d", bytesWritten)
	}
	if totalBytes != 1103253504 {
		t.Errorf("expected 1103253504 total bytes, got %d", totalBytes)
	}
}
