terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}

provider "aws" {
  region = "us-west-2"
}

resource "aws_s3_bucket" "example" {
  bucket = "example-bucket-123456"
  tags = {
    Name        = "ExampleBucket"
    Environment = "Dev"
  }
}
