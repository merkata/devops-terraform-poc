variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "public_subnets" {
  description = "List of public subnet IDs"
  type        = list(string)
}

variable "environment" {
  description = "Environment name"
  type        = string
}

variable "project_name" {
  description = "Project name"
  type        = string
}

variable "certificate_arn" {
  description = "ARN of ACM certificate"
  type        = string
}

variable "apps" {
  description = "Map of application configurations"
  type = map(object({
    port             = number
    path             = string
    health_check_url = string
    domain           = list(string)
    priority         = number
  }))
}
