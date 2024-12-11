# Terraform AWS SSM Parameters Module

The main goal of the module is to provide a consistent way to manage service SSM parameters. Suitable for use with [External Secrets](https://external-secrets.io/latest/).


This module manages parameters in AWS SSM Parameter Store.

This module is capable of taking strings as a values. If you need to store something in a different format, please convert it to a string. 
Strings are stored as SecureString (Standard Tier) with maximum size `4 KB`. 
See limitations on tiers in [Managing parameter Tiers](https://docs.aws.amazon.com/systems-manager/latest/userguide/sysman-paramstore-su-create.html)

For proper usage, refer to the example in this guide and the [Examples](./examples) folder.

## Usage example:

```terraform
module "krabby" {
  source = "hazelops/terraform-aws-ssm-parameters/aws"
  name   = "krabby"
  env    = "dev"

  parameters = {
    "API_KEY" : "api-XXXXXXXXXXXXXXXXXXXXX",
    "S3_BUCKET_ARN" : "arn:aws:s3:::dev-krabby",
    "S3_BUCKET_NAME": "dev-krabby"
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