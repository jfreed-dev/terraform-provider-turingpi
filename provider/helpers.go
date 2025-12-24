package provider

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Note: Uses HTTPClient from provider.go for TLS configuration

func checkPowerStatus(node int) string {
	// Simulate checking power status
	fmt.Printf("Checking power status for node %d\n", node)
	// Replace this with an actual API call
	return "off"
}

func turnOffNode(node int) {
	fmt.Printf("Turning off node %d\n", node)
	// Replace this with an API call to turn off the node
}

func turnOnNode(node int) {
	fmt.Printf("Turning on node %d\n", node)
	// Replace this with an API call to turn on the node
}

func flashNode(node int, firmware string) {
	fmt.Printf("Flashing node %d with firmware %s\n", node, firmware)
	// Replace this with an API call to flash the firmware
}

func checkBootStatus(endpoint string, node int, timeout int, token string, pattern string) (bool, error) {
	url := fmt.Sprintf("%s/api/bmc?opt=get&type=uart&node=%d", endpoint, node)

	deadline := time.Now().Add(time.Duration(timeout) * time.Second)

	for time.Now().Before(deadline) {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return false, fmt.Errorf("failed to create UART request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := HTTPClient.Do(req)
		if err != nil {
			return false, fmt.Errorf("UART request failed: %v", err)
		}

		defer func() { _ = resp.Body.Close() }()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, fmt.Errorf("failed to read UART response: %v", err)
		}

		// Check for configured boot pattern in UART output
		if strings.Contains(string(body), pattern) {
			fmt.Printf("Node %d booted successfully: pattern %q detected.\n", node, pattern)
			return true, nil
		}

		time.Sleep(5 * time.Second)
	}

	return false, fmt.Errorf("timeout reached: node %d did not boot successfully (pattern %q not found)", node, pattern)
}
