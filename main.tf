# SSM Parameter
resource "aws_ssm_parameter" "this" {
  for_each = local.parameters

  name        = "/${var.env}/${var.name}/${each.key}"
  value       = each.value
  type        = "SecureString"
  description = "The parameter is set for the application: ${var.name}. Managed by Terraform."

  lifecycle {
    create_before_destroy = true
  }
}
