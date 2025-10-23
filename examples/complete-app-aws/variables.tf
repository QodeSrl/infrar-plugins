variable "app_name" {
  description = "Name of the application"
  type        = string
  default     = "my-web-app"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "container_image" {
  description = "Docker container image URL"
  type        = string
  # Example: "123456789.dkr.ecr.us-east-1.amazonaws.com/my-app:latest"
}

variable "container_port" {
  description = "Port the container listens on"
  type        = number
  default     = 8080
}

variable "cpu" {
  description = "CPU units (256 = 0.25 vCPU, 512 = 0.5 vCPU, 1024 = 1 vCPU)"
  type        = number
  default     = 512
}

variable "memory" {
  description = "Memory in MB"
  type        = number
  default     = 1024
}

variable "desired_count" {
  description = "Number of instances to run"
  type        = number
  default     = 2
}

variable "db_password" {
  description = "Database password (sensitive)"
  type        = string
  sensitive   = true
  default     = null
}
