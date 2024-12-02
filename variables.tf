# Variables
variable "env" {
  type        = string
  description = "Environment name"
}
variable "app_name" {
  type        = string
  description = "Name of the application"
}
variable "parameters" {
  description = "List of values to put to AWS SSM Parameter Store"
  type = map(object({
    name        = string
    value       = string
    description = optional(string)
  }))
}
