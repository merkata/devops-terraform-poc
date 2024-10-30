# Add New Application

This guide explains how to add a new application to the existing infrastructure.

## Prerequisites

Before adding a new application, ensure you have:
- Application health check endpoint
- Port number for the application
- Domain name (if using custom domain)
- SSL certificate (if using custom domain)

## Steps

1. Update terraform.tfvars with new application configuration:

```hcl
apps = {
  existing_app = {
    # Existing app config...
  }
  new_app = {
    port             = 8087
    path             = "/new-app/*"
    health_check_url = "/new-app/health"
    domain           = ["your-domain.com"]
    priority         = 300
  }
}
```

2. Update security group rules if needed:

If your application requires additional ports or protocols, modify the security group in the compute module.

3. Apply the changes:

```bash
terraform plan -out=tfplan
terraform apply tfplan
```

## Validation

1. Check ALB target group health:
```bash
aws elbv2 describe-target-health \
  --target-group-arn $(terraform output -raw target_group_arns | jq -r '.new_app')
```

2. Test the application endpoint:
```bash
curl -k https://<alb-dns-name>/new-app/
```

3. Monitor CloudWatch logs for any issues

## Troubleshooting

### Common Issues

1. Health Check Failures
- Verify health check URL is correct
- Check security group allows health check port
- Verify application is running and responding

2. Routing Issues
- Check path pattern matches application URLs
- Verify priority doesn't conflict with other rules
- Confirm domain name configuration

3. Security Group Issues
- Verify ALB security group allows traffic to application
- Check EC2 security group allows traffic from ALB
