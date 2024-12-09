# Terraform Provider for Turing Pi 2.5

A Terraform provider for managing Turing Pi's Bare Metal Controller (BMC). This plugin enables you to interact with Turing Pi's BMC API, allowing you to:

- Authenticate with the BMC.
- Control power for specific nodes.
- Flash firmware to nodes.
- Retrieve system and node statuses.

## Features

- **Power Management**: Turn nodes on/off and check power status.
- **Firmware Management**: Flash firmware to individual nodes.
- **Node Interaction**: Communicate with nodes via the Turing Pi BMC API.

---

## Prerequisites

1. **Go Installation**: Install Go from the [official Go website](https://go.dev/).
2. **Terraform**: Install Terraform from the [official Terraform website](https://www.terraform.io/).

---

## Building the Plugin

### To build the Terraform provider:

1. Clone the repository:

   ```bash
   git clone https://github.com/jfreed-dev/turingpi-terraform-provider.git
   cd terraform-provider-turingpi
   ```

2. Initialize the Go module:

   ```bash
   go mod tidy
   ```

3. Build the binary:

   ```bash
   go build -o terraform-provider-turingpi
   ```

---

## Installation

### For Linux/macOS:

1. Create the plugin directory:

   ```bash
   mkdir -p ~/.terraform.d/plugins/local/turingpi/turingpi/1.0.0/linux_amd64/
   ```

2. Move the binary to the plugin directory:

   ```bash
   mv terraform-provider-turingpi ~/.terraform.d/plugins/local/turingpi/turingpi/1.0.0/linux_amd64/
   ```

### For Windows:

1. Create the plugin directory:

   ```cmd
   mkdir -p %APPDATA%\terraform.d\plugins\local\turingpi\turingpi\1.0.0\windows_amd64\
   ```

2. Move the binary to the plugin directory:

   ```cmd
   move terraform-provider-turingpi.exe %APPDATA%\terraform.d\plugins\local\turingpi\turingpi\1.0.0\windows_amd64\
   ```

---
## Using the Provider

1. Define the Provider in Terraform Configuration
   Create a `main.tf` file with the following content:

   ```hcl
   terraform {
     required_providers {
       turingpi = {
         source  = "local/turingpi/turingpi"
         version = "1.0.0"
     }
    }
   }

   provider "turingpi" {
    username = "your-username"
    password = "your-password"
   }
   ```

2. Initialize Terraform:

   ```bash
   terraform init
   ```

3. Apply Your Configuration:

   ```bash
   terraform apply
   ```

---
## Debugging During Development

1. Use `go run` to quickly test changes:

   ```bash
   go run main.go
   ```

2. Enable verbose logs during Terraform runs:

   ```bash
   export TF_LOG=DEBUG
   terraform apply
   ```

---

## Security Considerations

To avoid exposing sensitive credentials directly in your Terraform configuration files:

1. **Use Environment Variables:** Terraform supports environment variables for sensitive provider configurations. You can set `TURINGPI_USERNAME` and `TURINGPI_PASSWORD` in your shell environment:

   ```bash
   export TURINGPI_USERNAME=root
   export TURINGPI_PASSWORD=turing
   ```
   
   **Update** the provider block to use environment variables:

   ```hcl
   provider "turingpi" {}
   ```

2. Use a `.tfvars` File: Store credentials in a separate `.tfvars` file:

   ```plaintext
   username = "root"
   password = "turing"
   ```

   Reference the `.tfvars` file in your Terraform commands:

   ```bash
   terraform apply -var-file="credentials.tfvars"
   ```
---

## Terraform Example

Here’s a complete example of a Terraform configuration using the Turing Pi provider:

   ```hcl
   terraform {
     required_providers {
       turingpi = {
         source  = "local/turingpi/turingpi"
         version = "1.0.0"
       }
     }
   }

   provider "turingpi" {
     username = "root"      # Replace with your BMC username
     password = "turing"    # Replace with your BMC password
   }

   resource "turingpi_power" "node1" {
     node  = 1
     state = true           # Turn on power for node 1
   }

   resource "turingpi_flash" "node1" {
     node          = 1
     firmware_file = "/path/to/firmware.img"
   }
   ```

---

## Contributing

Contributions are welcome! Feel free to fork this repository, make changes, and submit a pull request.

---

## License

This project is licensed under the MIT License. [MIT License](https://mit-license.org/).
