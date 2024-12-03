module "s3_bucket" {
  source  = "terraform-aws-modules/s3-bucket/aws//examples/object"
  version = "~>4"
  name    = "krabby_bucket"
}

module "krabby" {
  source = "hazelops/terraform-aws-ssm-app/aws"
  env    = "dev"
  name   = "krabby"

  parameters = {
    "API_KEY" : "api-XXXXXXX"
    "s3_bucket_arn" : module.s3_bucket.s3_bucket_arn
    "s3_bucket_id" : module.s3_bucket.s3_bucket_id
  }
}
