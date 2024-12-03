# Variables
variable "env" {
  type        = string
  description = "Environment name"
}

variable "name" {
  type        = string
  description = "Name of the service"
}

variable "parameters" {
  type        = map(string)
  description = "Map of SSM ParameterStore parameters to store - by default, /$var.env/$var.name/*"
}

locals {
  parameters = merge(
    { ENV = var.env },
    var.parameters
  )
}
