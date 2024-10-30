variable "environment" {
  description = "Environment name"
  type        = string
}

variable "project_name" {
  description = "Project name"
  type        = string
}

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
}

variable "instance_count" {
  description = "Number of EC2 instances"
  type        = number
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
