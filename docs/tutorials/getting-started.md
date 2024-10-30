# Getting Started

This tutorial will guide you through setting up and deploying your first environment using this infrastructure.

## Prerequisites

Before you begin, ensure you have:

1. AWS CLI installed and configured
2. Terraform >= 1.0
3. Go >= 1.23
4. Python 3.x

## Initial Setup

1. Clone the repository and run the bootstrap script:

```bash
./bootstrap.sh
```

This script will:
- Verify all required tools are installed
- Set up Python virtual environment
- Check AWS credentials
- Create and configure the Terraform backend (S3 bucket and DynamoDB table)

2. Create your terraform.tfvars file:

```hcl
environment     = "dev"
project_name    = "myproject"
vpc_cidr        = "10.0.0.0/16"
instance_type   = "t3.micro"
instance_count  = 2
certificate_arn = "arn:aws:acm:region:account:certificate/id"

apps = {
  app1 = {
    port             = 8085
    path             = "/app1/*"
    health_check_url = "/app1/status"
    domain           = ["your-domain.com"]
    priority         = 100
  }
}
```

3. Initialize and apply Terraform:

```bash
terraform init
terraform plan
terraform apply
```

## Verifying the Deployment

1. Check the ALB DNS name in the outputs
2. Verify the EC2 instances are running
3. Test application endpoints through the ALB

## Next Steps

- Add more applications using the [Add New Application](../how-to/add-new-application.md) guide
- Configure monitoring and alerting
- Set up CI/CD pipelines
