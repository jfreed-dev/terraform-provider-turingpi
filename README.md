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
3. **Git**: Install Git for version control.

---

## Building the Plugin

To build the Terraform provider:

1. Clone the repository:

   ```bash
   git clone https://github.com/jfreed-dev/turingpi-terraform-provider.git
   cd terraform-provider-turingpi

2. Initialize the Go module:

   ```bash
   go mod tidy

3. Build the binary:

   ```bash
   go build -o terraform-provider-turingpi

---

## Installation

### For Linux/macOS:

1. Create the plugin directory:

   ```bash
   mkdir -p ~/.terraform.d/plugins/local/turingpi/turingpi/1.0.0/linux_amd64/

2. Move the binary to the plugin directory:

   ```bash
   mv terraform-provider-turingpi ~/.terraform.d/plugins/local/turingpi/turingpi/1.0.0/linux_amd64/

### For Windows:

1. Create the plugin directory:

   ```cmd
   mkdir -p %APPDATA%\terraform.d\plugins\local\turingpi\turingpi\1.0.0\windows_amd64\

2. Move the binary to the plugin directory:

   ```cmd
   move terraform-provider-turingpi.exe %APPDATA%\terraform.d\plugins\local\turingpi\turingpi\1.0.0\windows_amd64\

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

2. Initialize Terraform:

   ```bash
   terraform init

3. Apply Your Configuration:

   ```bash
   terraform apply

---
## Debugging During Development

1. Use `go run` to quickly test changes:

   ```bash
   go run main.go

2. Enable verbose logs during Terraform runs:

   ```bash
   export TF_LOG=DEBUG
   terraform apply

---

## Contributing

Contributions are welcome! Feel free to fork this repository, make changes, and submit a pull request.

---

## License

This project is licensed under the MIT License. [MIT License](https://mit-license.org/).
