# Variables

This document details all the variables used across the infrastructure modules.

## Global Variables

These variables are commonly used across multiple modules:

| Variable | Description | Type | Required |
|----------|-------------|------|----------|
| environment | Environment name (dev/staging/prod) | string | yes |
| project_name | Name of the project | string | yes |
| region | AWS region | string | yes |

## Module-Specific Variables

### VPC Module

| Variable | Description | Type | Required |
|----------|-------------|------|----------|
| vpc_cidr | CIDR block for VPC | string | yes |

### ALB Module

| Variable | Description | Type | Required |
|----------|-------------|------|----------|
| vpc_id | VPC ID | string | yes |
| public_subnets | List of public subnet IDs | list(string) | yes |
| certificate_arn | ACM certificate ARN | string | yes |
| access_logs_bucket | S3 bucket for ALB logs | string | yes |
| apps | Application configurations | map(object) | yes |

### Compute Module

| Variable | Description | Type | Required |
|----------|-------------|------|----------|
| vpc_id | VPC ID | string | yes |
| private_subnets | List of private subnet IDs | list(string) | yes |
| instance_type | EC2 instance type | string | yes |
| instance_count | Number of EC2 instances | number | yes |
| target_group_arns | List of target group ARNs | list(string) | yes |
