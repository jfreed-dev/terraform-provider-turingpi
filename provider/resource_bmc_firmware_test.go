package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResourceBMCFirmwareSchema(t *testing.T) {
	resource := resourceBMCFirmware()

	// Check that required fields exist
	if resource.Schema["firmware_file"] == nil {
		t.Error("expected firmware_file field in schema")
	}
	if resource.Schema["bmc_local"] == nil {
		t.Error("expected bmc_local field in schema")
	}
	if resource.Schema["triggers"] == nil {
		t.Error("expected triggers field in schema")
	}
	if resource.Schema["timeout"] == nil {
		t.Error("expected timeout field in schema")
	}
	if resource.Schema["last_upgrade"] == nil {
		t.Error("expected last_upgrade field in schema")
	}
	if resource.Schema["previous_version"] == nil {
		t.Error("expected previous_version field in schema")
	}

	// Check firmware_file properties
	firmwareFile := resource.Schema["firmware_file"]
	if !firmwareFile.Required {
		t.Error("firmware_file should be required")
	}

	// Check bmc_local properties
	bmcLocal := resource.Schema["bmc_local"]
	if !bmcLocal.Optional {
		t.Error("bmc_local should be optional")
	}
	if bmcLocal.Default != false {
		t.Error("bmc_local should default to false")
	}

	// Check timeout properties
	timeout := resource.Schema["timeout"]
	if !timeout.Optional {
		t.Error("timeout should be optional")
	}
	if timeout.Default != 300 {
		t.Error("timeout should default to 300")
	}

	// Check computed fields
	if !resource.Schema["last_upgrade"].Computed {
		t.Error("last_upgrade should be computed")
	}
	if !resource.Schema["previous_version"].Computed {
		t.Error("previous_version should be computed")
	}
}

func TestExtractHandle(t *testing.T) {
	tests := []struct {
		name     string
		response firmwareInitResponse
		want     string
	}{
		{
			name: "valid handle",
			response: firmwareInitResponse{
				Response: [][]interface{}{
					{"handle", "abc123"},
				},
			},
			want: "abc123",
		},
		{
			name: "empty response",
			response: firmwareInitResponse{
				Response: [][]interface{}{},
			},
			want: "",
		},
		{
			name: "no handle key",
			response: firmwareInitResponse{
				Response: [][]interface{}{
					{"status", "ok"},
				},
			},
			want: "",
		},
		{
			name: "handle with other fields",
			response: firmwareInitResponse{
				Response: [][]interface{}{
					{"status", "ok"},
					{"handle", "xyz789"},
					{"length", float64(1024)},
				},
			},
			want: "xyz789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractHandle(tt.response)
			if got != tt.want {
				t.Errorf("extractHandle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractFlashStatus(t *testing.T) {
	tests := []struct {
		name     string
		response *flashProgressResponse
		want     string
	}{
		{
			name: "done status",
			response: &flashProgressResponse{
				Response: [][]interface{}{
					{"status", "done"},
				},
			},
			want: "done",
		},
		{
			name: "in progress",
			response: &flashProgressResponse{
				Response: [][]interface{}{
					{"status", "flashing"},
					{"progress", float64(50)},
				},
			},
			want: "flashing",
		},
		{
			name: "error status",
			response: &flashProgressResponse{
				Response: [][]interface{}{
					{"status", "error"},
					{"message", "failed to write"},
				},
			},
			want: "error",
		},
		{
			name: "empty response",
			response: &flashProgressResponse{
				Response: [][]interface{}{},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractFlashStatus(tt.response)
			if got != tt.want {
				t.Errorf("extractFlashStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractFirmwareVersion(t *testing.T) {
	tests := []struct {
		name     string
		response *bmcAboutResponse
		want     string
	}{
		{
			name: "valid firmware version",
			response: &bmcAboutResponse{
				Response: [][]interface{}{
					{"api", "1.0"},
					{"firmware", "2.0.5"},
					{"buildroot", "2023.02"},
				},
			},
			want: "2.0.5",
		},
		{
			name: "no firmware field",
			response: &bmcAboutResponse{
				Response: [][]interface{}{
					{"api", "1.0"},
					{"buildroot", "2023.02"},
				},
			},
			want: "",
		},
		{
			name: "empty response",
			response: &bmcAboutResponse{
				Response: [][]interface{}{},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractFirmwareVersion(tt.response)
			if got != tt.want {
				t.Errorf("extractFirmwareVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitBMCLocalFirmwareUpgrade(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if !strings.Contains(r.URL.String(), "opt=set") {
			t.Errorf("expected opt=set in URL: %s", r.URL.String())
		}
		if !strings.Contains(r.URL.String(), "type=firmware") {
			t.Errorf("expected type=firmware in URL: %s", r.URL.String())
		}
		if !strings.Contains(r.URL.String(), "local") {
			t.Errorf("expected local in URL: %s", r.URL.String())
		}
		if !strings.Contains(r.URL.String(), "file=/tmp/firmware.bin") {
			t.Errorf("expected file path in URL: %s", r.URL.String())
		}

		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %s", auth)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"response":[["handle","abc123"]]}`))
	}))
	defer server.Close()

	originalClient := HTTPClient
	HTTPClient = server.Client()
	defer func() { HTTPClient = originalClient }()

	handle, err := initBMCLocalFirmwareUpgrade(server.URL, "test-token", "/tmp/firmware.bin")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if handle != "abc123" {
		t.Errorf("expected handle abc123, got %s", handle)
	}
}

func TestInitBMCLocalFirmwareUpgrade_NoHandle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"response":[["status","ok"]]}`))
	}))
	defer server.Close()

	originalClient := HTTPClient
	HTTPClient = server.Client()
	defer func() { HTTPClient = originalClient }()

	_, err := initBMCLocalFirmwareUpgrade(server.URL, "test-token", "/tmp/firmware.bin")
	if err == nil {
		t.Error("expected error when no handle returned")
	}
	if !strings.Contains(err.Error(), "no handle") {
		t.Errorf("expected 'no handle' error, got: %v", err)
	}
}

func TestGetFlashProgress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.String(), "opt=get") {
			t.Errorf("expected opt=get in URL: %s", r.URL.String())
		}
		if !strings.Contains(r.URL.String(), "type=flash") {
			t.Errorf("expected type=flash in URL: %s", r.URL.String())
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"response":[["status","done"],["progress",100]]}`))
	}))
	defer server.Close()

	originalClient := HTTPClient
	HTTPClient = server.Client()
	defer func() { HTTPClient = originalClient }()

	progress, err := getFlashProgress(server.URL, "test-token")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	status := extractFlashStatus(progress)
	if status != "done" {
		t.Errorf("expected status done, got %s", status)
	}
}

func TestCancelFirmwareUpload(t *testing.T) {
	cancelCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/upload/test-handle/cancel") {
			cancelCalled = true
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	originalClient := HTTPClient
	HTTPClient = server.Client()
	defer func() { HTTPClient = originalClient }()

	err := cancelFirmwareUpload(server.URL, "test-token", "test-handle")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !cancelCalled {
		t.Error("cancel endpoint was not called")
	}
}

func TestUploadFirmwareData(t *testing.T) {
	uploadCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/upload/test-handle") && r.Method == "POST" {
			uploadCalled = true

			// Verify content type is multipart
			contentType := r.Header.Get("Content-Type")
			if !strings.Contains(contentType, "multipart/form-data") {
				t.Errorf("expected multipart/form-data, got %s", contentType)
			}

			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	originalClient := HTTPClient
	HTTPClient = server.Client()
	defer func() { HTTPClient = originalClient }()

	// Create a temporary test file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-firmware.bin")
	if err := os.WriteFile(tmpFile, []byte("test firmware content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	file, err := os.Open(tmpFile)
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer func() { _ = file.Close() }()

	err = uploadFirmwareData(server.URL, "test-token", "test-handle", file, tmpFile)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !uploadCalled {
		t.Error("upload endpoint was not called")
	}
}

func TestResourceBMCFirmwareCRUD(t *testing.T) {
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		// About endpoint
		if strings.Contains(r.URL.String(), "type=about") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"response":[["api","1.0"],["firmware","2.0.5"]]}`))
			return
		}

		// Firmware init endpoint
		if strings.Contains(r.URL.String(), "type=firmware") && strings.Contains(r.URL.String(), "opt=set") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"response":[["handle","test-handle"]]}`))
			return
		}

		// Flash progress endpoint
		if strings.Contains(r.URL.String(), "type=flash") && strings.Contains(r.URL.String(), "opt=get") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"response":[["status","done"]]}`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	originalClient := HTTPClient
	HTTPClient = server.Client()
	defer func() { HTTPClient = originalClient }()

	config := &ProviderConfig{
		Endpoint: server.URL,
		Token:    "test-token",
	}

	resource := resourceBMCFirmware()
	d := resource.TestResourceData()

	// Set required values - use bmc_local=true to avoid file upload
	if err := d.Set("firmware_file", "/tmp/test-firmware.bin"); err != nil {
		t.Fatalf("failed to set firmware_file: %v", err)
	}
	if err := d.Set("bmc_local", true); err != nil {
		t.Fatalf("failed to set bmc_local: %v", err)
	}
	if err := d.Set("timeout", 10); err != nil {
		t.Fatalf("failed to set timeout: %v", err)
	}

	// Test Create
	diags := resourceBMCFirmwareCreate(context.TODO(), d, config)
	if diags.HasError() {
		t.Errorf("Create returned error: %v", diags)
	}
	if d.Id() != "bmc-firmware" {
		t.Errorf("expected ID 'bmc-firmware', got '%s'", d.Id())
	}

	previousVersion := d.Get("previous_version").(string)
	if previousVersion != "2.0.5" {
		t.Errorf("expected previous_version '2.0.5', got '%s'", previousVersion)
	}

	lastUpgrade := d.Get("last_upgrade").(string)
	if lastUpgrade == "" {
		t.Error("expected last_upgrade to be set")
	}

	// Test Read (should be no-op)
	diags = resourceBMCFirmwareRead(context.TODO(), d, config)
	if diags.HasError() {
		t.Errorf("Read returned error: %v", diags)
	}

	// Test Delete
	diags = resourceBMCFirmwareDelete(context.TODO(), d, config)
	if diags.HasError() {
		t.Errorf("Delete returned error: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected empty ID after delete, got '%s'", d.Id())
	}
}
