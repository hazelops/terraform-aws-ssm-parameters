# Simple Terraform Module Example doesn't have provider configuration
plugin "aws" {
  enabled = true
}

# Ignore missing providers warning
rule "missing-provider" {
  enabled = false
}
