# devops-terraform-poc
A DevOps POC for a mock assignment that showcases terraform

## Introduction

This project showcases how to create a sample AWS deployment via terraform in a DevOps fashion.

The project itself is about deploying (a dynamic set of) apps that are fronted by an ALB inside a VPC.

Notable components of the repo:

- modules for the project reside in a modules folder and are reusable
- a test folder with terratest tests for each module and an e2e (full deployment) test
- an examples folder that hosts an entire full deployment
- docs folder with documentation following Diataxis to get you started
- a bootstrap script that should aid in local testing and development
- CI that will run formatting, linting, security checks, tests and automated module documentation
- a backend configuration for using S3 (with DynamoDB) as a terraform state backend

## Getting Started

### Bootstrap Script

The `bootstrap.sh` script is the entry component that handles the initial setup of your development environment. It performs several important tasks:

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
