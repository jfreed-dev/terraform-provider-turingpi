package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func authenticate(username, password string) (string, error) {
	url := "https://turingpi.local/api/bmc/authenticate"
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
	json.NewDecoder(resp.Body).Decode(&result)
	return result["id"], nil
}
