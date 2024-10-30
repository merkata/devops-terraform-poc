# Components

This document describes the core components of the infrastructure.

## VPC Module

The VPC module creates a Virtual Private Cloud with the following resources:

- Public and private subnets across multiple availability zones
- Internet Gateway for public subnets
- NAT Gateway for private subnets
- Route tables and associations

### Variables

- `environment` - Environment name (e.g., dev, staging, prod)
- `project_name` - Name of the project
- `vpc_cidr` - CIDR block for the VPC

## ALB Module

The Application Load Balancer module creates:

- HTTPS load balancer with TLS 1.2+
- Target groups for applications
- Security groups
- Listener rules based on path patterns
- Access logging to S3

### Variables

- `environment` - Environment name
- `project_name` - Project name
- `vpc_id` - VPC ID where ALB will be created
- `public_subnets` - List of public subnet IDs
- `certificate_arn` - ARN of ACM certificate
- `access_logs_bucket` - S3 bucket for ALB logs
- `apps` - Map of application configurations

## Compute Module

The Compute module manages EC2 instances with:

- Auto Scaling Group
- Launch Template
- IAM roles and instance profile
- Security groups
- Target group attachments

### Variables

- `environment` - Environment name
- `project_name` - Project name
- `vpc_id` - VPC ID
- `private_subnets` - List of private subnet IDs
- `instance_type` - EC2 instance type
- `instance_count` - Number of EC2 instances
- `target_group_arns` - List of target group ARNs
