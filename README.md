# Terraform AWS SSM APP Module

##  

## Usage example:
```terraform
module "test_ssm_app" {
  source = "../terraform/terraform-aws-ssm-app"
  env    = var.env
  parameters = {
    "vpc_id" = {
      name        = "vpc_id"
      value       = module.vpc.vpc_id
      description = "VPC ID"
    }
    "vpc_cidr" = {
      name        = "vpc_cidr"
      value       = module.vpc.vpc_cidr_block
      description = "VPC CIDR block"
    }
  }

  app_name = "krabby"
}
```

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >=1.2.0 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | >=4.30.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | >=4.30.0 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [aws_ssm_parameter.this](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ssm_parameter) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_app_name"></a> [app\_name](#input\_app\_name) | Name of the application | `string` | n/a | yes |
| <a name="input_env"></a> [env](#input\_env) | Environment name | `string` | n/a | yes |
| <a name="input_parameters"></a> [parameters](#input\_parameters) | List of values to put to AWS SSM Parameter Store | <pre>map(object({<br>    name        = string<br>    value       = string<br>    description = optional(string)<br>  }))</pre> | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_ssm_parameter_paths"></a> [ssm\_parameter\_paths](#output\_ssm\_parameter\_paths) | A list of paths to created parameters |
<!-- END_TF_DOCS -->