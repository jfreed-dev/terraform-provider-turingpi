package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFlash() *schema.Resource {
	return &schema.Resource{
		Create: resourceFlashCreate,
		Read:   resourceFlashRead,
		Delete: resourceFlashDelete,
		Schema: map[string]*schema.Schema{
			"node": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Node ID to flash firmware",
			},
			"firmware_file": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Path to the firmware file",
			},
		},
	}
}

func resourceFlashCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	// Flashing logic implementation
	node := d.Get("node").(int)
	firmware := d.Get("firmware_file").(string)
	fmt.Printf("Flashing node %d with firmware %s", node, firmware)
	d.SetId(fmt.Sprintf("node-%d", node))
	return nil
}

func resourceFlashRead(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	// Implement read logic for flash progress
	return nil
}

func resourceFlashDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	// No action needed for deletion
	return nil
}
