package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func authenticate(endpoint, username, password string) (string, error) {
	url := fmt.Sprintf("%s/api/bmc/authenticate", endpoint)
	data := map[string]string{"username": username, "password": password}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("authentication failed with status: %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode authentication response: %v", err)
	}
	return result["id"], nil
}
