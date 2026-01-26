package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceFlash() *schema.Resource {
	return &schema.Resource{
		Description: "Flashes firmware to a Turing Pi compute node. The node must be powered off before flashing.",
		Create:      resourceFlashCreate,
		Read:        resourceFlashRead,
		Delete:      resourceFlashDelete,
		Schema: map[string]*schema.Schema{
			"node": {
				Type:             schema.TypeInt,
				Required:         true,
				Description:      "Node ID to flash firmware (1-4)",
				ForceNew:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 4)),
			},
			"firmware_file": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Path to the firmware file to flash",
				ForceNew:    true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
		},
	}
}

// flashResponse represents the BMC flash initiation response
type flashResponse struct {
	Handle interface{} `json:"handle"` // Can be string or number
}

// flashStatusResponse represents the BMC flash status response
type flashStatusResponse struct {
	Transferring json.RawMessage `json:"Transferring,omitempty"`
	Flashing     *flashingStatus `json:"Flashing,omitempty"`
	Done         *[]interface{}  `json:"Done,omitempty"`
	Error        *string         `json:"Error,omitempty"`
}

type flashingStatus struct {
	BytesWritten int64 `json:"bytes_written"`
	TotalBytes   int64 `json:"total_bytes"`
}

// transferringStatus represents the new BMC firmware 2.0.5+ format
type transferringStatus struct {
	ID           int64  `json:"id"`
	ProcessName  string `json:"process_name"`
	Size         int64  `json:"size"`
	Cancelled    bool   `json:"cancelled"`
	BytesWritten int64  `json:"bytes_written"`
}

// isTransferring checks if the flash status indicates a transfer is in progress
// and returns progress info if available. Handles both old ([]int64) and new (object) formats.
func (f *flashStatusResponse) isTransferring() (inProgress bool, bytesWritten, totalBytes int64) {
	if len(f.Transferring) == 0 {
		return false, 0, 0
	}

	// Try new object format first (BMC firmware 2.0.5+)
	var newFormat transferringStatus
	if err := json.Unmarshal(f.Transferring, &newFormat); err == nil {
		return true, newFormat.BytesWritten, newFormat.Size
	}

	// Try old array format (legacy BMC firmware)
	var oldFormat []int64
	if err := json.Unmarshal(f.Transferring, &oldFormat); err == nil && len(oldFormat) >= 2 {
		return true, oldFormat[0], oldFormat[1]
	}

	// Unknown format but field is present - assume transferring
	return true, 0, 0
}

func resourceFlashCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)
	node := d.Get("node").(int)
	firmwarePath := d.Get("firmware_file").(string)

	// Open the firmware file
	file, err := os.Open(firmwarePath)
	if err != nil {
		return fmt.Errorf("failed to open firmware file: %w", err)
	}
	defer func() { _ = file.Close() }()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat firmware file: %w", err)
	}
	fileSize := fileInfo.Size()

	fmt.Printf("Flashing node %d with firmware %s (%d bytes)\n", node, firmwarePath, fileSize)

	// Step 1: Power off the node before flashing
	if err := setNodePower(config.Endpoint, config.Token, node, false); err != nil {
		return fmt.Errorf("failed to power off node before flash: %w", err)
	}
	time.Sleep(2 * time.Second) // Wait for node to power off

	// Step 2: Initiate flash operation
	// API uses 0-indexed nodes
	// file=stream indicates we'll upload via streaming, not from local SD card
	apiNode := node - 1
	url := fmt.Sprintf("%s/api/bmc?opt=set&type=flash&node=%d&file=stream&length=%d", config.Endpoint, apiNode, fileSize)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create flash request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+config.Token)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("flash initiation failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("flash initiation failed with status %d: %s", resp.StatusCode, string(body))
	}

	var flashResp flashResponse
	if err := json.NewDecoder(resp.Body).Decode(&flashResp); err != nil {
		return fmt.Errorf("failed to decode flash response: %w", err)
	}

	if flashResp.Handle == nil {
		return fmt.Errorf("no upload handle returned from BMC")
	}

	// Handle can be string or number
	var handleStr string
	switch h := flashResp.Handle.(type) {
	case string:
		handleStr = h
	case float64:
		handleStr = fmt.Sprintf("%.0f", h)
	default:
		handleStr = fmt.Sprintf("%v", h)
	}

	fmt.Printf("Got upload handle: %s\n", handleStr)

	// Step 3: Upload the firmware file using multipart form
	uploadURL := fmt.Sprintf("%s/api/bmc/upload/%s", config.Endpoint, handleStr)

	// Create a pipe for streaming the multipart form data
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	// Start a goroutine to write the multipart form data
	errChan := make(chan error, 1)
	go func() {
		defer func() { _ = pw.Close() }()
		defer func() { _ = writer.Close() }()

		part, err := writer.CreateFormFile("file", firmwarePath)
		if err != nil {
			errChan <- fmt.Errorf("failed to create form file: %w", err)
			return
		}

		// Re-open the file for reading (we need to read it again)
		uploadFile, err := os.Open(firmwarePath)
		if err != nil {
			errChan <- fmt.Errorf("failed to open firmware file for upload: %w", err)
			return
		}
		defer func() { _ = uploadFile.Close() }()

		if _, err := io.Copy(part, uploadFile); err != nil {
			errChan <- fmt.Errorf("failed to copy firmware data: %w", err)
			return
		}

		errChan <- nil
	}()

	uploadReq, err := http.NewRequest("POST", uploadURL, pr)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}
	uploadReq.Header.Set("Authorization", "Bearer "+config.Token)
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())

	fmt.Printf("Uploading firmware to BMC (%d bytes)...\n", fileSize)
	uploadResp, err := HTTPClient.Do(uploadReq)
	if err != nil {
		return fmt.Errorf("firmware upload failed: %w", err)
	}
	defer func() { _ = uploadResp.Body.Close() }()

	// Check for errors from the goroutine
	if uploadErr := <-errChan; uploadErr != nil {
		return uploadErr
	}

	if uploadResp.StatusCode != http.StatusOK && uploadResp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(uploadResp.Body)
		return fmt.Errorf("firmware upload failed with status %d: %s", uploadResp.StatusCode, string(body))
	}

	fmt.Printf("Upload complete, waiting for flash to finish...\n")

	// Step 4: Poll flash status until complete
	timeout := time.After(25 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("flash operation timed out")
		case <-ticker.C:
			status, err := getFlashStatus(config.Endpoint, config.Token)
			if err != nil {
				fmt.Printf("Warning: failed to get flash status: %v\n", err)
				continue
			}

			if status.Error != nil {
				return fmt.Errorf("flash failed: %s", *status.Error)
			}

			if status.Done != nil {
				fmt.Printf("Flash completed successfully\n")
				d.SetId(fmt.Sprintf("flash-node-%d", node))
				return nil
			}

			if status.Flashing != nil {
				pct := float64(status.Flashing.BytesWritten) / float64(status.Flashing.TotalBytes) * 100
				fmt.Printf("Flashing: %.1f%% (%d/%d bytes)\n", pct, status.Flashing.BytesWritten, status.Flashing.TotalBytes)
			}

			if inProgress, bytesWritten, totalBytes := status.isTransferring(); inProgress {
				if totalBytes > 0 {
					pct := float64(bytesWritten) / float64(totalBytes) * 100
					fmt.Printf("Transferring: %.1f%% (%d/%d bytes)\n", pct, bytesWritten, totalBytes)
				} else {
					fmt.Printf("Transferring...\n")
				}
			}
		}
	}
}

func getFlashStatus(endpoint, token string) (*flashStatusResponse, error) {
	url := fmt.Sprintf("%s/api/bmc?opt=get&type=flash", endpoint)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var status flashStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &status, nil
}

func resourceFlashRead(d *schema.ResourceData, meta interface{}) error {
	// Flash is a one-time operation - once completed, we just maintain state
	// The resource exists if it was successfully flashed
	id := d.Id()
	if id == "" || !strings.HasPrefix(id, "flash-node-") {
		d.SetId("")
	}
	return nil
}

func resourceFlashDelete(d *schema.ResourceData, meta interface{}) error {
	// Flash cannot be "undone" - we just remove from state
	// The node retains its flashed firmware
	fmt.Printf("Removing flash resource from state (firmware remains on node)\n")
	d.SetId("")
	return nil
}
