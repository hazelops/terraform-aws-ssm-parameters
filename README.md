# Terraform AWS SSM APP Module

The main goal of the module is to provide a consistent way to manage service SSM parameters. Suitable for use with [External Secrets](https://external-secrets.io/latest/).


This module uploads strings to AWS SSM Parameter Store.

If you need to store something in a different format, please convert it to a string. Strings are stored as SecureString.

For proper usage, refer to the example in this guide and the [Examples](./examples) folder.

## Usage example:

```terraform
module "krabby" {
  source = "hazelops/terraform-aws-ssm-app/aws"
  name   = "krabby"
  env    = "dev"

  parameters = {
    "API_KEY" : "api-XXXXXXX",
    "s3_bucket_arn" : "arn_XXXX",
    "s3_bucket_name": "dev-krabby"
  }
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
| <a name="input_env"></a> [env](#input\_env) | Environment name | `string` | n/a | yes |
| <a name="input_name"></a> [name](#input\_name) | Name of the service | `string` | n/a | yes |
| <a name="input_parameters"></a> [parameters](#input\_parameters) | Map of SSM ParameterStore parameters to store - by default, /$var.env/$var.name/* | `map(string)` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_ssm_parameter_paths"></a> [ssm\_parameter\_paths](#output\_ssm\_parameter\_paths) | A list of paths to created parameters |
<!-- END_TF_DOCS -->