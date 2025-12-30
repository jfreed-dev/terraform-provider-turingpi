# TFLint configuration for terraform-provider-turingpi examples
# https://github.com/terraform-linters/tflint

config {
  format = "compact"
  plugin_dir = "~/.tflint.d/plugins"

  call_module_type = "local"
  force = false
  disabled_by_default = false
}

# Enable the Terraform ruleset
plugin "terraform" {
  enabled = true
  preset  = "recommended"
}

# Naming conventions
rule "terraform_naming_convention" {
  enabled = true
  format  = "snake_case"
}

# Documentation rules
rule "terraform_documented_outputs" {
  enabled = true
}

rule "terraform_documented_variables" {
  enabled = true
}

# Standard file naming
rule "terraform_standard_module_structure" {
  enabled = true
}

# Require version constraints (disabled for examples - users choose their TF version)
rule "terraform_required_version" {
  enabled = false
}

rule "terraform_required_providers" {
  enabled = true
}

# Unused declarations
rule "terraform_unused_declarations" {
  enabled = true
}

# Workspace naming
rule "terraform_workspace_remote" {
  enabled = true
}
