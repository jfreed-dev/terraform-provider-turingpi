package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// BMC API response structures
type bmcAboutResponse struct {
	Response [][]interface{} `json:"response"`
}

type bmcInfoResponse struct {
	Response struct {
		Network []networkInterface `json:"network"`
		Storage []storageDevice    `json:"storage"`
	} `json:"response"`
}

type networkInterface struct {
	Device string `json:"device"`
	IP     string `json:"ip"`
	MAC    string `json:"mac"`
}

type storageDevice struct {
	Name  string `json:"name"`
	Total int64  `json:"total"`
	Free  int64  `json:"free"`
	Used  int64  `json:"use"`
}

type bmcPowerResponse struct {
	Response [][]interface{} `json:"response"`
}

func dataSourceInfo() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves information about the Turing Pi BMC including version info, network configuration, storage metrics, and node power status.",
		ReadContext: dataSourceInfoRead,
		Schema: map[string]*schema.Schema{
			// Version information from /api/bmc?opt=get&type=about
			"api_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "BMC API version",
			},
			"daemon_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "BMC daemon version",
			},
			"buildroot_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Buildroot version",
			},
			"firmware_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "BMC firmware version",
			},
			"build_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "BMC build timestamp",
			},

			// Network information from /api/bmc?opt=get&type=info
			"network_interfaces": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of network interfaces on the BMC",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Network interface device name",
						},
						"ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IP address",
						},
						"mac": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "MAC address",
						},
					},
				},
			},

			// Storage information from /api/bmc?opt=get&type=info
			"storage_devices": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of storage devices",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Storage device name (e.g., 'bmc', 'microSD')",
						},
						"total_bytes": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Total storage capacity in bytes",
						},
						"used_bytes": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Used storage in bytes",
						},
						"free_bytes": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Free storage in bytes",
						},
					},
				},
			},

			// Power status from /api/bmc?opt=get&type=power
			"nodes": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Power status of each node (node1-node4)",
				Elem: &schema.Schema{
					Type: schema.TypeBool,
				},
			},
		},
	}
}

func dataSourceInfoRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*ProviderConfig)
	var diags diag.Diagnostics

	// Fetch version/about information
	aboutData, err := fetchBMCAbout(config.Endpoint, config.Token)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to fetch BMC about info: %w", err))
	}

	if err := setAboutData(d, aboutData); err != nil {
		return diag.FromErr(err)
	}

	// Fetch network and storage information
	infoData, err := fetchBMCInfo(config.Endpoint, config.Token)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to fetch BMC info: %w", err))
	}

	if err := setInfoData(d, infoData); err != nil {
		return diag.FromErr(err)
	}

	// Fetch power status
	powerData, err := fetchBMCPower(config.Endpoint, config.Token)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to fetch BMC power status: %w", err))
	}

	if err := setPowerData(d, powerData); err != nil {
		return diag.FromErr(err)
	}

	// Set a stable ID for the data source
	d.SetId("turingpi-bmc-info")

	return diags
}

func fetchBMCAbout(endpoint, token string) (*bmcAboutResponse, error) {
	url := fmt.Sprintf("%s/api/bmc?opt=get&type=about", endpoint)

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

	var result bmcAboutResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func fetchBMCInfo(endpoint, token string) (*bmcInfoResponse, error) {
	url := fmt.Sprintf("%s/api/bmc?opt=get&type=info", endpoint)

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

	var result bmcInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func fetchBMCPower(endpoint, token string) (*bmcPowerResponse, error) {
	url := fmt.Sprintf("%s/api/bmc?opt=get&type=power", endpoint)

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

	var result bmcPowerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func setAboutData(d *schema.ResourceData, data *bmcAboutResponse) error {
	// The about API returns data in a nested array format:
	// [[key, value], [key, value], ...]
	aboutMap := make(map[string]string)
	for _, item := range data.Response {
		if len(item) >= 2 {
			key, keyOk := item[0].(string)
			value, valueOk := item[1].(string)
			if keyOk && valueOk {
				aboutMap[key] = value
			}
		}
	}

	if v, ok := aboutMap["api"]; ok {
		if err := d.Set("api_version", v); err != nil {
			return fmt.Errorf("failed to set api_version: %w", err)
		}
	}
	if v, ok := aboutMap["version"]; ok {
		if err := d.Set("daemon_version", v); err != nil {
			return fmt.Errorf("failed to set daemon_version: %w", err)
		}
	}
	if v, ok := aboutMap["buildroot"]; ok {
		if err := d.Set("buildroot_version", v); err != nil {
			return fmt.Errorf("failed to set buildroot_version: %w", err)
		}
	}
	if v, ok := aboutMap["firmware"]; ok {
		if err := d.Set("firmware_version", v); err != nil {
			return fmt.Errorf("failed to set firmware_version: %w", err)
		}
	}
	if v, ok := aboutMap["buildtime"]; ok {
		if err := d.Set("build_time", v); err != nil {
			return fmt.Errorf("failed to set build_time: %w", err)
		}
	}

	return nil
}

func setInfoData(d *schema.ResourceData, data *bmcInfoResponse) error {
	// Set network interfaces
	networkInterfaces := make([]map[string]interface{}, 0, len(data.Response.Network))
	for _, iface := range data.Response.Network {
		networkInterfaces = append(networkInterfaces, map[string]interface{}{
			"device": iface.Device,
			"ip":     iface.IP,
			"mac":    iface.MAC,
		})
	}
	if err := d.Set("network_interfaces", networkInterfaces); err != nil {
		return fmt.Errorf("failed to set network_interfaces: %w", err)
	}

	// Set storage devices
	storageDevices := make([]map[string]interface{}, 0, len(data.Response.Storage))
	for _, storage := range data.Response.Storage {
		storageDevices = append(storageDevices, map[string]interface{}{
			"name":        storage.Name,
			"total_bytes": storage.Total,
			"used_bytes":  storage.Used,
			"free_bytes":  storage.Free,
		})
	}
	if err := d.Set("storage_devices", storageDevices); err != nil {
		return fmt.Errorf("failed to set storage_devices: %w", err)
	}

	return nil
}

func setPowerData(d *schema.ResourceData, data *bmcPowerResponse) error {
	// The power API returns data in format:
	// [[node1, status], [node2, status], ...]
	nodes := make(map[string]interface{})
	for _, item := range data.Response {
		if len(item) >= 2 {
			nodeName, nameOk := item[0].(string)
			// Status could be bool or int (0/1)
			if nameOk {
				switch v := item[1].(type) {
				case bool:
					nodes[nodeName] = v
				case float64:
					nodes[nodeName] = v == 1
				case int:
					nodes[nodeName] = v == 1
				}
			}
		}
	}

	if err := d.Set("nodes", nodes); err != nil {
		return fmt.Errorf("failed to set nodes: %w", err)
	}

	return nil
}
