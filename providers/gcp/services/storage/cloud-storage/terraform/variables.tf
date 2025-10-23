# Service-specific variables for Cloud Storage
# Common variables (project_id, region, labels, etc.) are defined at provider level

variable "bucket_name" {
  description = "Name of the Cloud Storage bucket"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9][a-z0-9-._]*[a-z0-9]$", var.bucket_name))
    error_message = "Bucket name must be lowercase alphanumeric with hyphens, dots, or underscores"
  }
}

variable "location" {
  description = "GCP location or multi-region for the bucket"
  type        = string
  default     = "US"
}

variable "versioning_enabled" {
  description = "Enable object versioning"
  type        = bool
  default     = false
}

variable "force_destroy" {
  description = "Allow bucket deletion even when not empty"
  type        = bool
  default     = false
}

variable "cors_rules" {
  description = "CORS rules for the bucket"
  type = list(object({
    allowed_methods = list(string)
    allowed_origins = list(string)
    expose_headers  = optional(list(string))
    max_age_seconds = optional(number)
  }))
  default = []
}

variable "lifecycle_rules" {
  description = "Lifecycle rules for the bucket"
  type = list(object({
    action_type        = string
    storage_class      = optional(string)
    age_days           = optional(number)
    num_newer_versions = optional(number)
    with_state         = optional(string)
  }))
  default = []

  validation {
    condition = alltrue([
      for rule in var.lifecycle_rules :
      contains(["Delete", "SetStorageClass"], rule.action_type)
    ])
    error_message = "Action type must be Delete or SetStorageClass"
  }
}
