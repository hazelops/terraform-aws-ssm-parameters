# SSM Parameter
resource "aws_ssm_parameter" "this" {
  for_each = var.parameters

  name        = "/${var.env}/${var.app_name}/${each.value.name}"
  value       = each.value.value
  type        = "SecureString"
  description = each.value.description
  lifecycle {
    ignore_changes = [value]
  }
  tags = {
    Application = each.value.name
  }
}
