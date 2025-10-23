# Cloud Run Service
resource "google_cloud_run_service" "main" {
  name     = var.service_name
  location = var.region
  project  = var.project_id

  template {
    spec {
      containers {
        image = var.container_image

        # Environment variables
        dynamic "env" {
          for_each = var.environment_variables
          content {
            name  = env.key
            value = env.value
          }
        }

        # Resource limits
        resources {
          limits = {
            cpu    = var.cpu_limit
            memory = var.memory_limit
          }
        }
      }

      # Service account for Cloud Run
      service_account_name = var.service_account_email

      # Timeout
      timeout_seconds = var.timeout
    }

    metadata {
      annotations = {
        "autoscaling.knative.dev/maxScale" = tostring(var.max_instances)
        "autoscaling.knative.dev/minScale" = tostring(var.min_instances)
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }

  metadata {
    labels = var.labels
  }
}

# IAM policy to allow public access (if enabled)
resource "google_cloud_run_service_iam_member" "public_access" {
  count = var.allow_public_access ? 1 : 0

  service  = google_cloud_run_service.main.name
  location = google_cloud_run_service.main.location
  project  = google_cloud_run_service.main.project
  role     = "roles/run.invoker"
  member   = "allUsers"
}
