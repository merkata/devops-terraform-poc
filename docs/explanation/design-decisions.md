# Design Decisions

This document explains key architectural and implementation decisions made in this infrastructure.

## Infrastructure as Code

### Why Terraform?
- Declarative approach to infrastructure
- Strong AWS provider support
- State management capabilities
- Module system for reusability

### Module Structure
We chose to separate the infrastructure into three main modules:
- VPC: Network isolation and routing
- ALB: Load balancing and SSL termination
- Compute: EC2 instance management

This separation allows:
- Independent scaling of components
- Clear responsibility boundaries
- Reuse across different environments

## Security Decisions

### HTTPS Only
- All HTTP traffic is redirected to HTTPS
- TLS 1.2+ required for all connections
- Security headers are stripped

TODO:
- Access logs enabled for audit trail

### Network Security
- Private subnets for compute instances
- Public subnets only for ALB
- Security groups with minimal required access
- IMDSv2 required on EC2 instances

### IAM Security
- Least privilege principle
- Instance profiles for EC2 access
- Service-linked roles where possible

## Scalability Decisions

### Auto Scaling
- Instance count based on environment
- Double max capacity for burst loads
- Health checks for automatic replacement

### Multi-AZ
- Resources spread across 3 AZs
- Single NAT Gateway in non-prod
- Multiple NAT Gateways in prod

## Operational Decisions

### State Management
- Remote state in S3
- State locking with DynamoDB
- Separate state files per environment

### Testing Strategy
- Terratest for infrastructure testing
- End-to-end deployment testing
- Security scanning with checkov

### CI/CD Integration
- Automated formatting checks
- Security scanning
- Documentation generation
- Test execution
