# Complete Web Application Deployment on AWS
# This example shows the full deployment flow using Infrar plugins

terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.region
}

# 1. Storage - S3 bucket for application data
module "storage" {
  source = "../../packages/storage/aws/terraform"

  bucket_name        = "${var.app_name}-data"
  versioning_enabled = var.environment == "prod" ? true : false

  lifecycle_rules = [{
    id     = "archive-old-data"
    status = "Enabled"
    transitions = [{
      days          = 90
      storage_class = "GLACIER"
    }]
    expiration = {
      days = 365
    }
  }]

  tags = {
    Environment = var.environment
    Application = var.app_name
  }
}

# 2. Secrets - Database password
module "database_password" {
  source = "../../packages/secrets/aws/terraform"

  secret_name  = "${var.app_name}/database/password"
  description  = "Database password for ${var.app_name}"
  secret_value = var.db_password  # In production, use AWS Secrets Manager rotation

  tags = {
    Environment = var.environment
    Application = var.app_name
  }
}

# 3. Web Application - ECS Fargate
module "web_app" {
  source = "../../packages/compute/aws/terraform"

  app_name        = var.app_name
  cluster_name    = "${var.app_name}-cluster"
  container_image = var.container_image
  container_port  = var.container_port
  region          = var.region

  cpu            = var.cpu
  memory         = var.memory
  desired_count  = var.desired_count

  # Pass bucket name and secret ARN to the application
  environment_variables = {
    BUCKET_NAME      = module.storage.bucket_name
    DATABASE_SECRET  = module.database_password.secret_arn
    AWS_REGION       = var.region
    ENVIRONMENT      = var.environment
  }

  tags = {
    Environment = var.environment
    Application = var.app_name
  }
}

# 4. Grant application access to S3 bucket
resource "aws_iam_role_policy" "app_storage_access" {
  name = "${var.app_name}-storage-access"
  role = module.web_app.task_role_arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject",
        "s3:ListBucket"
      ]
      Resource = [
        module.storage.bucket_arn,
        "${module.storage.bucket_arn}/*"
      ]
    }]
  })
}

# 5. Grant application access to database secret
resource "aws_iam_role_policy" "app_secrets_access" {
  name = "${var.app_name}-secrets-access"
  role = module.web_app.task_role_arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "secretsmanager:GetSecretValue"
      ]
      Resource = module.database_password.secret_arn
    }]
  })
}

# Outputs
output "application_url" {
  description = "URL of the deployed application"
  value       = module.web_app.service_url
}

output "bucket_name" {
  description = "Name of the S3 bucket"
  value       = module.storage.bucket_name
}

output "secret_arn" {
  description = "ARN of the database password secret"
  value       = module.database_password.secret_arn
}

output "load_balancer_dns" {
  description = "DNS name of the load balancer"
  value       = module.web_app.load_balancer_dns
}

output "log_group_name" {
  description = "CloudWatch log group name"
  value       = module.web_app.log_group_name
}
