# Complete Web Application Deployment on GCP
# This example shows the full deployment flow using Infrar plugins

terraform {
  required_version = ">= 1.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# 1. Storage - Cloud Storage bucket for application data
module "storage" {
  source = "../../packages/storage/gcp/terraform"

  bucket_name        = "${var.app_name}-data"
  project_id         = var.project_id
  location           = var.location
  versioning_enabled = var.environment == "prod" ? true : false

  lifecycle_rules = [{
    action_type = "SetStorageClass"
    storage_class = "NEARLINE"
    age_days    = 90
  }]

  labels = {
    environment = var.environment
    application = var.app_name
  }
}

# 2. Secrets - Database password
module "database_password" {
  source = "../../packages/secrets/gcp/terraform"

  secret_name  = "database-password"
  project_id   = var.project_id
  secret_value = var.db_password  # In production, use Secret Manager rotation

  labels = {
    environment = var.environment
    application = var.app_name
  }
}

# 3. Web Application - Cloud Run
module "web_app" {
  source = "../../packages/compute/gcp/terraform"

  app_name        = var.app_name
  project_id      = var.project_id
  container_image = var.container_image
  container_port  = var.container_port
  region          = var.region

  cpu           = var.cpu
  memory        = var.memory
  min_instances = var.min_instances
  max_instances = var.max_instances

  # Pass bucket name and secret name to the application
  environment_variables = {
    BUCKET_NAME     = module.storage.bucket_name
    DATABASE_SECRET = module.database_password.secret_name
    GCP_PROJECT     = var.project_id
    ENVIRONMENT     = var.environment
  }

  allow_public_access    = true
  create_service_account = true

  labels = {
    environment = var.environment
    application = var.app_name
  }
}

# 4. Grant application access to Cloud Storage bucket
resource "google_storage_bucket_iam_member" "app_access" {
  bucket = module.storage.bucket_name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${module.web_app.service_account_email}"
}

# 5. Grant application access to database secret
resource "google_secret_manager_secret_iam_member" "app_access" {
  secret_id = module.database_password.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${module.web_app.service_account_email}"
}

# Outputs
output "application_url" {
  description = "URL of the deployed application"
  value       = module.web_app.service_url
}

output "bucket_name" {
  description = "Name of the Cloud Storage bucket"
  value       = module.storage.bucket_name
}

output "secret_name" {
  description = "Name of the database password secret"
  value       = module.database_password.secret_name
}

output "service_account_email" {
  description = "Email of the service account"
  value       = module.web_app.service_account_email
}
