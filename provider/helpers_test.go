package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCheckPowerStatus(t *testing.T) {
	// Current implementation is a stub that always returns "off"
	result := checkPowerStatus(1)
	if result != "off" {
		t.Errorf("expected 'off', got %s", result)
	}
}

func TestCheckPowerStatus_DifferentNodes(t *testing.T) {
	nodes := []int{1, 2, 3, 4}
	for _, node := range nodes {
		result := checkPowerStatus(node)
		if result != "off" {
			t.Errorf("node %d: expected 'off', got %s", node, result)
		}
	}
}

func TestTurnOnNode_DoesNotPanic(t *testing.T) {
	// Current implementation is a stub, just verify it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("turnOnNode panicked: %v", r)
		}
	}()

	turnOnNode(1)
	turnOnNode(2)
	turnOnNode(3)
	turnOnNode(4)
}

func TestTurnOffNode_DoesNotPanic(t *testing.T) {
	// Current implementation is a stub, just verify it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("turnOffNode panicked: %v", r)
		}
	}()

	turnOffNode(1)
	turnOffNode(2)
	turnOffNode(3)
	turnOffNode(4)
}

func TestFlashNode_DoesNotPanic(t *testing.T) {
	// Current implementation is a stub, just verify it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("flashNode panicked: %v", r)
		}
	}()

	flashNode(1, "/path/to/firmware.img")
	flashNode(2, "/another/firmware.bin")
}

func TestCheckBootStatus_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s", r.Method)
		}

		// Verify Authorization header
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			t.Errorf("expected Bearer token in Authorization header, got %s", auth)
		}

		// Verify query parameters
		if r.URL.Query().Get("opt") != "get" {
			t.Errorf("expected opt=get, got %s", r.URL.Query().Get("opt"))
		}
		if r.URL.Query().Get("type") != "uart" {
			t.Errorf("expected type=uart, got %s", r.URL.Query().Get("type"))
		}

		// Return response with login prompt
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Some boot output...\nlogin: "))
	}))
	defer server.Close()

	// Use short timeout since mock server returns immediately
	success, err := checkBootStatus(server.URL, 1, 1, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if !success {
		t.Error("expected success=true when login prompt is found")
	}
}

func TestCheckBootStatus_TokenInHeader(t *testing.T) {
	expectedToken := "my-secret-token"
	var capturedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("login:"))
	}))
	defer server.Close()

	checkBootStatus(server.URL, 1, 1, expectedToken)

	expectedHeader := "Bearer " + expectedToken
	if capturedAuth != expectedHeader {
		t.Errorf("expected Authorization header '%s', got '%s'", expectedHeader, capturedAuth)
	}
}

func TestCheckBootStatus_NodeInURL(t *testing.T) {
	testCases := []struct {
		node         int
		expectedNode string
	}{
		{1, "1"},
		{2, "2"},
		{3, "3"},
		{4, "4"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("node_%d", tc.node), func(t *testing.T) {
			var capturedNode string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedNode = r.URL.Query().Get("node")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("login:"))
			}))
			defer server.Close()

			checkBootStatus(server.URL, tc.node, 1, "token")

			if capturedNode != tc.expectedNode {
				t.Errorf("expected node=%s in URL, got node=%s", tc.expectedNode, capturedNode)
			}
		})
	}
}

func TestCheckBootStatus_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return response without login prompt
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Booting...\nStill booting..."))
	}))
	defer server.Close()

	// Use very short timeout to speed up test
	// Note: This test will take at least 1 second due to the timeout
	success, err := checkBootStatus(server.URL, 1, 1, "token")

	if success {
		t.Error("expected success=false on timeout")
	}

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "timeout reached") {
		t.Errorf("expected timeout error message, got: %s", err.Error())
	}
}

func TestCheckBootStatus_ConnectionError(t *testing.T) {
	// Use invalid URL to simulate connection error
	success, err := checkBootStatus("http://localhost:99999", 1, 1, "token")

	if success {
		t.Error("expected success=false on connection error")
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCheckBootStatus_URLConstruction(t *testing.T) {
	var capturedPath string
	var capturedQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("login:"))
	}))
	defer server.Close()

	checkBootStatus(server.URL, 2, 1, "token")

	if capturedPath != "/api/bmc" {
		t.Errorf("expected path /api/bmc, got %s", capturedPath)
	}

	if !strings.Contains(capturedQuery, "opt=get") {
		t.Errorf("expected query to contain opt=get, got %s", capturedQuery)
	}

	if !strings.Contains(capturedQuery, "type=uart") {
		t.Errorf("expected query to contain type=uart, got %s", capturedQuery)
	}

	if !strings.Contains(capturedQuery, "node=2") {
		t.Errorf("expected query to contain node=2, got %s", capturedQuery)
	}
}

func TestCheckBootStatus_LoginPromptVariations(t *testing.T) {
	testCases := []struct {
		name     string
		response string
		expected bool
	}{
		{"login prompt at end", "boot complete\nlogin:", true},
		{"login prompt with space", "boot complete\nlogin: ", true},
		{"login prompt in middle", "stuff\nlogin:\nmore stuff", true},
		{"no login prompt", "still booting...", false},
		{"empty response", "", false},
		{"partial match", "logging in...", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tc.response))
			}))
			defer server.Close()

			success, _ := checkBootStatus(server.URL, 1, 1, "token")

			if success != tc.expected {
				t.Errorf("expected success=%v for response '%s', got %v", tc.expected, tc.response, success)
			}
		})
	}
}
