# Service-specific variables for Cloud Run
# Common variables (project_id, region, labels, etc.) are defined at provider level

variable "service_name" {
  description = "Name of the Cloud Run service"
  type        = string
}

variable "container_image" {
  description = "Container image URL (e.g., gcr.io/project/image:tag)"
  type        = string
}

variable "service_account_email" {
  description = "Service account email for Cloud Run"
  type        = string
  default     = ""
}

variable "cpu_limit" {
  description = "CPU limit for the container"
  type        = string
  default     = "1000m"
}

variable "memory_limit" {
  description = "Memory limit for the container"
  type        = string
  default     = "512Mi"
}

variable "timeout" {
  description = "Request timeout in seconds"
  type        = number
  default     = 300
}

variable "min_instances" {
  description = "Minimum number of instances"
  type        = number
  default     = 0
}

variable "max_instances" {
  description = "Maximum number of instances"
  type        = number
  default     = 10
}

variable "environment_variables" {
  description = "Environment variables for the container"
  type        = map(string)
  default     = {}
}

variable "allow_public_access" {
  description = "Allow unauthenticated public access"
  type        = bool
  default     = false
}
