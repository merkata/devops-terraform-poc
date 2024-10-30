# devops-terraform-poc
A DevOps POC for a mock assignment that showcases terraform

## Getting Started

### Bootstrap Script

The `bootstrap.sh` script is a crucial component that handles the initial setup of your development environment. It performs several important tasks:

1. **Environment Verification**
   - Checks for required tools (Terraform, Go, Python)
   - Verifies correct versions and configurations

2. **AWS Setup**
   - Validates AWS credentials
   - Ensures proper AWS CLI configuration

3. **Terraform Backend Configuration**
   - Creates and configures S3 bucket for state storage
   - Sets up DynamoDB table for state locking
   - Initializes backend infrastructure

Run the bootstrap script once when setting up a new development environment:
```bash
./bootstrap.sh
```

<!-- BEGIN_TF_DOCS -->
<!-- END_TF_DOCS -->
