# Output
output "ssm_parameter_paths" {
  description = "A list of paths to created parameters"
  value = [
    for param in aws_ssm_parameter.this : param.name
  ]
}
