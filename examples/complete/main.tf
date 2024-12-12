resource "aws_s3_bucket" "krabby_demo" {
  bucket = "dev-krabby-demo"
  tags = {
    Name        = "Demo Krabby bucket"
    Environment = "dev"
  }
}

module "krabby" {
  source = "../../"
  env    = "dev"
  name   = "krabby"

  parameters = {
    API_KEY        = "api-XXXXXXXXXXXXXXXXXXXXX"
    S3_BUCKET_ARN  = aws_s3_bucket.krabby_demo.arn
    S3_BUCKET_NAME = aws_s3_bucket.krabby_demo.id
  }
}
