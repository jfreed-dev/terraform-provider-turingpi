package provider

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNode() *schema.Resource {
	return &schema.Resource{
		Create: resourceNodeProvision,
		Read:   resourceNodeStatus,
		Update: resourceNodeProvision,
		Delete: resourceNodeDelete,
		Schema: map[string]*schema.Schema{
			"node": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Node ID to manage",
			},
			"firmware_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Path to the firmware file (required for flashing)",
			},
			"power_state": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "on",
				Description: "Desired power state of the node (on/off)",
			},
			"boot_check": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Check if the node successfully boots by monitoring UART output",
			},
			"login_prompt_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     60,
				Description: "Timeout in seconds to wait for login prompt via UART",
			},
		},
	}
}

func resourceNodeProvision(d *schema.ResourceData, meta interface{}) error {
	node := d.Get("node").(int)
	firmware := d.Get("firmware_file").(string)
	powerState := d.Get("power_state").(string)
	bootCheck := d.Get("boot_check").(bool)
	timeout := d.Get("login_prompt_timeout").(int)

	// Step 1: Check and adjust power state
	currentPower := checkPowerStatus(node)
	if currentPower == "on" && powerState == "off" {
		turnOffNode(node)
	} else if currentPower == "off" && powerState == "on" {
		turnOnNode(node)
	}

	// Step 2: Flash firmware if needed
	if firmware != "" {
		fmt.Printf("Flashing node %d with firmware %s\n", node, firmware)
		flashNode(node, firmware)
	}

	// Step 3: Boot the node
	if powerState == "on" {
		turnOnNode(node)
	}

	// Step 4: Optional boot check via UART
	if bootCheck {
		fmt.Printf("Checking boot status for node %d\n", node)
		if !waitForLoginPrompt(node, timeout) {
			return fmt.Errorf("node %d failed to reach login prompt within %d seconds", node, timeout)
		}
	}

	// Set resource ID to uniquely identify the node
	d.SetId(fmt.Sprintf("node-%d", node))
	return nil
}

func resourceNodeStatus(d *schema.ResourceData, meta interface{}) error {
	node := d.Get("node").(int)
	currentPower := checkPowerStatus(node)

	d.Set("power_state", currentPower)
	return nil
}

func resourceNodeDelete(d *schema.ResourceData, meta interface{}) error {
	node := d.Get("node").(int)

	// Turn off the node during resource deletion
	turnOffNode(node)
	return nil
}

// Helper functions
func checkPowerStatus(node int) string {
	// Simulate checking power status
	fmt.Printf("Checking power status for node %d\n", node)
	// Replace this with API call
	return "off"
}

func turnOffNode(node int) {
	fmt.Printf("Turning off node %d\n", node)
	// Replace this with API call to turn off the node
}

func turnOnNode(node int) {
	fmt.Printf("Turning on node %d\n", node)
	// Replace this with API call to turn on the node
}

func flashNode(node int, firmware string) {
	fmt.Printf("Flashing node %d with firmware %s\n", node, firmware)
	// Replace this with API call to flash the firmware
}

func waitForLoginPrompt(node int, timeout int) bool {
	fmt.Printf("Waiting for login prompt on node %d for up to %d seconds\n", node, timeout)
	// Simulate waiting with a sleep (replace with UART API call)
	time.Sleep(5 * time.Second)
	// Simulate successful boot
	return true
}
