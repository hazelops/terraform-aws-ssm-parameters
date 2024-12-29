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

  validation {
    condition     = alltrue([for v in var.parameters : (v != null && v != "" && can(v) || can(int(v)))])
    error_message = "All values in the 'parameters' map must be non-null, non-empty strings."
  }
}

locals {
  parameters = merge(
    { ENV = var.env },
    var.parameters
  )
}
