locals {
  # Ensure the name length is within AWS limits (32 characters)
  name_max_length = 32

  # Take first 16 chars of project name to leave room for environment and suffix
  project_prefix = substr(lower(replace(var.project_name, "/[^a-zA-Z0-9-]/", "")), 0, 16)
  env_suffix     = substr(var.environment, 0, 8)

  # Final ALB name: "<project_prefix>-<env>-alb"
  alb_name = format(
    "%s-%s-alb",
    local.project_prefix,
    local.env_suffix
  )

  # Common tags
  common_tags = {
    Environment = var.environment
    Project     = var.project_name
    ManagedBy   = "terraform"
  }
}

# Add validation
resource "null_resource" "name_validation" {
  lifecycle {
    precondition {
      condition     = length(local.alb_name) <= local.name_max_length
      error_message = "ALB name '${local.alb_name}' exceeds the maximum length of ${local.name_max_length} characters"
    }
  }
}
