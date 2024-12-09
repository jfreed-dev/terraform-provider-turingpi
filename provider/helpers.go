package provider

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

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

func checkBootStatus(node int, timeout int, token string) (bool, error) {
	url := fmt.Sprintf("https://turingpi.local/api/bmc?opt=get&type=uart&node=%d", node)
	client := &http.Client{}

	deadline := time.Now().Add(time.Duration(timeout) * time.Second)

	for time.Now().Before(deadline) {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return false, fmt.Errorf("failed to create UART request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := client.Do(req)
		if err != nil {
			return false, fmt.Errorf("UART request failed: %v", err)
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return false, fmt.Errorf("failed to read UART response: %v", err)
		}

		// Simulate login prompt detection
		if strings.Contains(string(body), "login:") {
			fmt.Printf("Node %d booted successfully: login prompt detected.\n", node)
			return true, nil
		}

		time.Sleep(5 * time.Second)
	}

	return false, fmt.Errorf("timeout reached: node %d did not boot successfully", node)
}
