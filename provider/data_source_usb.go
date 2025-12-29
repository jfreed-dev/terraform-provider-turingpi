package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUSB() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves the current USB routing configuration from the Turing Pi BMC.",
		ReadContext: dataSourceUSBRead,
		Schema: map[string]*schema.Schema{
			"mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current USB mode: 'host' (node acts as USB host) or 'device' (node acts as USB device)",
			},
			"node": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Node ID that USB is currently routed to (1-4)",
			},
			"route": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current USB routing destination: 'usb-a' (external connector) or 'bmc' (BMC chip)",
			},
		},
	}
}

func dataSourceUSBRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*ProviderConfig)
	var diags diag.Diagnostics

	// Fetch current USB status using the function from resource_usb.go
	status, err := getUSBStatus(config.Endpoint, config.Token)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to read USB status: %w", err))
	}

	// Parse the response using the function from resource_usb.go
	mode, node, route := parseUSBStatus(status)

	if err := d.Set("mode", mode); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set mode: %w", err))
	}
	if err := d.Set("node", node); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set node: %w", err))
	}
	if err := d.Set("route", route); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set route: %w", err))
	}

	// Set a stable ID for the data source
	d.SetId("turingpi-usb-status")

	return diags
}
