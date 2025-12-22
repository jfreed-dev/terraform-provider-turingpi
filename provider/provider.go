package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const defaultEndpoint = "https://turingpi.local"

// ProviderConfig holds the configuration for the provider
type ProviderConfig struct {
	Token    string
	Endpoint string
}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("TURINGPI_USERNAME", nil),
				Description: "The username for BMC authentication",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("TURINGPI_PASSWORD", nil),
				Description: "The password for BMC authentication",
			},
			"endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("TURINGPI_ENDPOINT", defaultEndpoint),
				Description: "The BMC API endpoint URL (e.g., https://turingpi.local or https://192.168.1.100)",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"turingpi_power": resourcePower(),
			"turingpi_flash": resourceFlash(),
			"turingpi_node":  resourceNode(),
		},
		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	endpoint := d.Get("endpoint").(string)

	token, err := authenticate(endpoint, username, password)
	if err != nil {
		return nil, err
	}

	return &ProviderConfig{
		Token:    token,
		Endpoint: endpoint,
	}, nil
}
