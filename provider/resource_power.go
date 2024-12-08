package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourcePower() *schema.Resource {
	return &schema.Resource{
		Create: resourcePowerSet,
		Read:   resourcePowerRead,
		Update: resourcePowerSet,
		Delete: resourcePowerDelete,
		Schema: map[string]*schema.Schema{
			"node": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Node ID to control power",
			},
			"state": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Power state (true for on, false for off)",
			},
		},
	}
}

func resourcePowerSet(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	node := d.Get("node").(int)
	state := d.Get("state").(bool)
	stateStr := "0"
	if state {
		stateStr = "1"
	}

	url := fmt.Sprintf("https://turingpi.local/api/bmc?opt=power&type=set&node%d=%s", node, stateStr)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+meta.(string))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to set power state for node %d", node)
	}

	d.SetId(fmt.Sprintf("node-%d", node))
	return nil
}

func resourcePowerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	// Implementation of state reading
	return nil
}

func resourcePowerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	// No action required for deletion
	return nil
}
