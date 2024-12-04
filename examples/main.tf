resource "aws_s3_bucket" "this" {
  bucket = "dev-krabby-bucket"
  tags = {
    Name        = "Bucket for storing Krabby logs"
    Environment = "dev"
  }
}
module "krabby" {
  source = "hazelops/terraform-aws-ssm-app/aws"
  env    = "dev"
  name   = "krabby"

  parameters = {
    "API_KEY" : "api-XXXXXXXXXXXXXXXXXXXXX"
    "S3_BUCKET_ARN"  : aws_s3_bucket.this.arn
    "S3_BUCKET_NAME" : aws_s3_bucket.this.id
  }
}
