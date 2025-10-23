# Cloud Storage Bucket
resource "google_storage_bucket" "main" {
  name          = var.bucket_name
  location      = var.location
  project       = var.project_id
  force_destroy = var.force_destroy

  # Uniform bucket-level access (recommended)
  uniform_bucket_level_access = true

  # Versioning
  versioning {
    enabled = var.versioning_enabled
  }

  # CORS configuration
  dynamic "cors" {
    for_each = var.cors_rules
    content {
      origin          = cors.value.allowed_origins
      method          = cors.value.allowed_methods
      response_header = lookup(cors.value, "expose_headers", [])
      max_age_seconds = lookup(cors.value, "max_age_seconds", 3600)
    }
  }

  # Lifecycle rules
  dynamic "lifecycle_rule" {
    for_each = var.lifecycle_rules
    content {
      action {
        type          = lifecycle_rule.value.action_type
        storage_class = lookup(lifecycle_rule.value, "storage_class", null)
      }

      condition {
        age                = lookup(lifecycle_rule.value, "age_days", null)
        num_newer_versions = lookup(lifecycle_rule.value, "num_newer_versions", null)
        with_state         = lookup(lifecycle_rule.value, "with_state", null)
      }
    }
  }

  labels = merge(
    var.labels,
    {
      managed_by = "infrar"
    }
  )
}
