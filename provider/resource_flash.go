package provider

import (
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
				ForceNew:    true,
			},
			"firmware_file": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Path to the firmware file",
				ForceNew:    true,
			},
		},
	}
}

func resourceFlashCreate(d *schema.ResourceData, meta interface{}) error {
	node := d.Get("node").(int)
	firmware := d.Get("firmware_file").(string)

	fmt.Printf("Flashing node %d with firmware %s\n", node, firmware)
	d.SetId(fmt.Sprintf("node-%d", node))
	return nil
}

func resourceFlashRead(d *schema.ResourceData, meta interface{}) error {
	fmt.Printf("Reading flash status for node %s\n", d.Id())
	return nil
}

func resourceFlashDelete(d *schema.ResourceData, meta interface{}) error {
	fmt.Printf("Deleting flash resource for node %s\n", d.Id())
	return nil
}
