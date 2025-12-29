package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// USB mode constants for the BMC API
// Mode values encode both the operation mode and routing:
// | Mode | Operation | Route |
// |------|-----------|-------|
// | 0    | Host      | USB-A |
// | 1    | Device    | USB-A |
// | 2    | Flash     | USB-A | (Host)
// | 3    | Flash     | USB-A | (Device)
// | 4    | Host      | BMC   |
// | 5    | Device    | BMC   |
// | 6    | Flash     | BMC   | (Host)
// | 7    | Flash     | BMC   | (Device)

const (
	usbModeHostUSBA   = 0
	usbModeDeviceUSBA = 1
	usbModeHostBMC    = 4
	usbModeDeviceBMC  = 5
)

// usbStatusResponse represents the response from GET /api/bmc?opt=get&type=usb
type usbStatusResponse struct {
	Response [][]interface{} `json:"response"`
}

func resourceUSB() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages USB routing configuration on the Turing Pi BMC. The USB bus can only be routed to one node at a time.",
		CreateContext: resourceUSBCreate,
		ReadContext:   resourceUSBRead,
		UpdateContext: resourceUSBUpdate,
		DeleteContext: resourceUSBDelete,
		Schema: map[string]*schema.Schema{
			"node": {
				Type:             schema.TypeInt,
				Required:         true,
				Description:      "Node ID to route USB to (1-4)",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 4)),
			},
			"mode": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "USB mode: 'host' (node acts as USB host) or 'device' (node acts as USB device)",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"host", "device"}, false)),
			},
			"route": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "usb-a",
				Description:      "USB routing destination: 'usb-a' (external USB-A connector) or 'bmc' (route to BMC chip)",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"usb-a", "bmc"}, false)),
			},
			// Computed attributes from reading current state
			"current_mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current USB mode as reported by BMC",
			},
			"current_node": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Current node that USB is routed to",
			},
			"current_route": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current USB routing destination",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceUSBCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*ProviderConfig)

	node := d.Get("node").(int)
	mode := d.Get("mode").(string)
	route := d.Get("route").(string)

	// Convert to API mode integer
	apiMode := getUSBAPIMode(mode, route)

	// Set USB configuration
	if err := setUSBMode(config.Endpoint, config.Token, node, apiMode); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set USB mode: %w", err))
	}

	d.SetId(fmt.Sprintf("usb-node-%d", node))

	// Read back the state
	return resourceUSBRead(ctx, d, meta)
}

func resourceUSBRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*ProviderConfig)
	var diags diag.Diagnostics

	// Fetch current USB status
	status, err := getUSBStatus(config.Endpoint, config.Token)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to read USB status: %w", err))
	}

	// Parse the response
	currentMode, currentNode, currentRoute := parseUSBStatus(status)

	if err := d.Set("current_mode", currentMode); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set current_mode: %w", err))
	}
	if err := d.Set("current_node", currentNode); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set current_node: %w", err))
	}
	if err := d.Set("current_route", currentRoute); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set current_route: %w", err))
	}

	return diags
}

func resourceUSBUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*ProviderConfig)

	node := d.Get("node").(int)
	mode := d.Get("mode").(string)
	route := d.Get("route").(string)

	// Convert to API mode integer
	apiMode := getUSBAPIMode(mode, route)

	// Set USB configuration
	if err := setUSBMode(config.Endpoint, config.Token, node, apiMode); err != nil {
		return diag.FromErr(fmt.Errorf("failed to update USB mode: %w", err))
	}

	// Update the ID if node changed
	d.SetId(fmt.Sprintf("usb-node-%d", node))

	// Read back the state
	return resourceUSBRead(ctx, d, meta)
}

func resourceUSBDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// USB routing cannot be truly "deleted" - it's always routed somewhere
	// On delete, we just remove from state. The USB configuration remains on the BMC.
	d.SetId("")
	return nil
}

// getUSBAPIMode converts human-readable mode and route to API mode integer
func getUSBAPIMode(mode, route string) int {
	switch {
	case mode == "host" && route == "usb-a":
		return usbModeHostUSBA
	case mode == "device" && route == "usb-a":
		return usbModeDeviceUSBA
	case mode == "host" && route == "bmc":
		return usbModeHostBMC
	case mode == "device" && route == "bmc":
		return usbModeDeviceBMC
	default:
		return usbModeHostUSBA
	}
}

// setUSBMode calls the BMC API to set USB configuration
func setUSBMode(endpoint, token string, node, mode int) error {
	// API uses 0-indexed nodes
	apiNode := node - 1
	url := fmt.Sprintf("%s/api/bmc?opt=set&type=usb&mode=%d&node=%d", endpoint, mode, apiNode)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// getUSBStatus fetches current USB configuration from BMC
func getUSBStatus(endpoint, token string) (*usbStatusResponse, error) {
	url := fmt.Sprintf("%s/api/bmc?opt=get&type=usb", endpoint)

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

	var result usbStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// parseUSBStatus extracts mode, node, and route from USB status response
func parseUSBStatus(status *usbStatusResponse) (mode string, node int, route string) {
	// Default values
	mode = "host"
	node = 1
	route = "usb-a"

	// Response format: [[key, value], [key, value], ...]
	statusMap := make(map[string]interface{})
	for _, item := range status.Response {
		if len(item) >= 2 {
			key, keyOk := item[0].(string)
			if keyOk {
				statusMap[key] = item[1]
			}
		}
	}

	// Parse mode
	if m, ok := statusMap["mode"].(string); ok {
		switch m {
		case "Host", "host":
			mode = "host"
		case "Device", "device":
			mode = "device"
		default:
			mode = m
		}
	}

	// Parse node (API returns 0-indexed, we use 1-indexed)
	if n, ok := statusMap["node"].(float64); ok {
		node = int(n) + 1
	} else if n, ok := statusMap["node"].(int); ok {
		node = n + 1
	}

	// Parse route
	if r, ok := statusMap["route"].(string); ok {
		switch r {
		case "BMC", "bmc":
			route = "bmc"
		case "USB-A", "usb-a", "USB-2.0":
			route = "usb-a"
		default:
			route = r
		}
	}

	return mode, node, route
}
