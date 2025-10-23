# GCP Cloud Storage Bucket Module

Simple, production-ready Cloud Storage bucket module for Infrar deployments.

## Features

- ✅ Google-managed encryption by default
- ✅ Uniform bucket-level access
- ✅ Optional versioning
- ✅ Optional CORS configuration
- ✅ Optional lifecycle rules
- ✅ Secure defaults

## Usage

```hcl
module "storage" {
  source = "./packages/storage/gcp/terraform"

  bucket_name        = "my-app-data"
  project_id         = "my-gcp-project"
  location           = "US"
  versioning_enabled = true

  lifecycle_rules = [{
    action_type = "SetStorageClass"
    storage_class = "NEARLINE"
    age_days    = 90
  }]

  labels = {
    environment = "production"
    project     = "my-app"
  }
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| bucket_name | Name of the Cloud Storage bucket | `string` | n/a | yes |
| project_id | GCP project ID | `string` | n/a | yes |
| location | GCP location or multi-region | `string` | `"US"` | no |
| versioning_enabled | Enable object versioning | `bool` | `false` | no |
| force_destroy | Allow bucket deletion when not empty | `bool` | `false` | no |
| cors_rules | CORS configuration rules | `list(object)` | `[]` | no |
| lifecycle_rules | Lifecycle management rules | `list(object)` | `[]` | no |
| labels | Additional resource labels | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| bucket_name | Name of the Cloud Storage bucket |
| bucket_url | URL of the bucket |
| bucket_self_link | Self-link of the bucket |

## Example with CORS

```hcl
module "storage" {
  source = "./packages/storage/gcp/terraform"

  bucket_name = "my-app-uploads"
  project_id  = "my-gcp-project"

  cors_rules = [{
    allowed_methods = ["GET", "PUT", "POST"]
    allowed_origins = ["https://example.com"]
    max_age_seconds = 3600
  }]
}
```

## Storage Classes

Available storage classes for lifecycle rules:
- `STANDARD` - Best for frequently accessed data
- `NEARLINE` - Best for data accessed < 1/month
- `COLDLINE` - Best for data accessed < 1/quarter
- `ARCHIVE` - Best for data accessed < 1/year
