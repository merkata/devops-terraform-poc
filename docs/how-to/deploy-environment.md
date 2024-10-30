# How to Deploy an Environment

This guide explains how to deploy a new environment using this infrastructure.

## Prerequisites

- AWS credentials configured
- Terraform installed
- Required domain and SSL certificate

## Steps

1. Create Environment Configuration

Create a new directory for your environment:

```bash
mkdir environments/[env-name]
cd environments/[env-name]
```

2. Create terraform.tfvars:

```hcl
environment     = "[env-name]"
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

3. Initialize Terraform:

```bash
terraform init -backend-config="key=[env-name]/terraform.tfstate"
```

4. Plan and Apply:

```bash
terraform plan -out=tfplan
terraform apply tfplan
```

## Environment-Specific Considerations

### Development
- Use smaller instance types
- Single NAT Gateway
- Minimal instance count

### Production
- Use production-grade instance types
- Multiple NAT Gateways
- Higher instance count
- Enable deletion protection

## Monitoring the Deployment

1. Check AWS Console for resources
2. Monitor CloudWatch logs
3. Verify ALB health checks
4. Test application endpoints

## Cleanup

To destroy the environment:

```bash
terraform destroy
```

Note: In production, enable deletion protection and use appropriate IAM policies.
