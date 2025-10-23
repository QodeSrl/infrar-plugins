variable "app_name" {
  description = "Name of the application"
  type        = string
  default     = "my-web-app"
}

variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "us-central1"
}

variable "location" {
  description = "GCP location for storage (can be region or multi-region)"
  type        = string
  default     = "US"
}

variable "container_image" {
  description = "Docker container image URL"
  type        = string
  # Example: "gcr.io/my-project/my-app:latest"
}

variable "container_port" {
  description = "Port the container listens on"
  type        = number
  default     = 8080
}

variable "cpu" {
  description = "CPU in milliCPU (1000 = 1 vCPU)"
  type        = number
  default     = 1000
}

variable "memory" {
  description = "Memory in MB"
  type        = number
  default     = 512
}

variable "min_instances" {
  description = "Minimum number of instances (0 to scale to zero)"
  type        = number
  default     = 0
}

variable "max_instances" {
  description = "Maximum number of instances"
  type        = number
  default     = 10
}

variable "db_password" {
  description = "Database password (sensitive)"
  type        = string
  sensitive   = true
  default     = null
}
