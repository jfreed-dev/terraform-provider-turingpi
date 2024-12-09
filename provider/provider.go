package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The username for BMC authentication",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The password for BMC authentication",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"turingpi_power": resourcePower(),
			"turingpi_flash": resourceFlash(),
			"turingpi_node":  resourceNode(), // Add this resource
		},
	}
}
