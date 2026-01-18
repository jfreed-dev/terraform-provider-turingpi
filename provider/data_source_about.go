package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAbout() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves detailed version and build information about the Turing Pi BMC daemon, including API version, firmware version, buildroot version, and build timestamp.",
		ReadContext: dataSourceAboutRead,
		Schema: map[string]*schema.Schema{
			"api_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "BMC API version (e.g., '1.0.0').",
			},
			"daemon_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "BMC daemon version.",
			},
			"buildroot_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Buildroot version used to build the BMC firmware.",
			},
			"firmware_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "BMC firmware version.",
			},
			"build_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the BMC firmware was built.",
			},
		},
	}
}

func dataSourceAboutRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*ProviderConfig)
	var diags diag.Diagnostics

	// Reuse the existing fetchBMCAbout function from data_source_info.go
	aboutData, err := fetchBMCAbout(config.Endpoint, config.Token)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to fetch BMC about info: %w", err))
	}

	// Parse the response using the shared function that handles both formats
	aboutMap := parseAboutResponse(aboutData)

	if v, ok := aboutMap["api"]; ok {
		if err := d.Set("api_version", v); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set api_version: %w", err))
		}
	}
	if v, ok := aboutMap["version"]; ok {
		if err := d.Set("daemon_version", v); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set daemon_version: %w", err))
		}
	}
	if v, ok := aboutMap["buildroot"]; ok {
		if err := d.Set("buildroot_version", v); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set buildroot_version: %w", err))
		}
	}
	if v, ok := aboutMap["firmware"]; ok {
		if err := d.Set("firmware_version", v); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set firmware_version: %w", err))
		}
	}
	// Handle both "buildtime" (legacy) and "build_version" (new) field names
	if v, ok := aboutMap["buildtime"]; ok {
		if err := d.Set("build_time", v); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set build_time: %w", err))
		}
	} else if v, ok := aboutMap["build_version"]; ok {
		if err := d.Set("build_time", v); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set build_time: %w", err))
		}
	}

	d.SetId("turingpi-about")

	return diags
}
