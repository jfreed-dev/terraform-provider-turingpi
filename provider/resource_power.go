package provider

import (
	"fmt"

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

func resourcePowerSet(d *schema.ResourceData, meta interface{}) error {
	node := d.Get("node").(int)
	state := d.Get("state").(bool)

	stateStr := "0"
	if state {
		stateStr = "1"
	}

	// Example logic for setting power state
	fmt.Printf("Setting power state for node %d to %s\n", node, stateStr)

	// Set a unique ID for the resource
	d.SetId(fmt.Sprintf("node-%d", node))
	return nil
}

func resourcePowerRead(d *schema.ResourceData, meta interface{}) error {
	// Example logic for reading power status
	fmt.Printf("Reading power state for node %s\n", d.Id())
	return nil
}

func resourcePowerDelete(d *schema.ResourceData, meta interface{}) error {
	// Example logic for cleanup if needed
	fmt.Printf("Deleting power resource for node %s\n", d.Id())
	return nil
}
