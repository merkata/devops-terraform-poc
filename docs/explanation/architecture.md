# Architecture

This document explains the high-level architecture of the infrastructure.

## Overview

The infrastructure is designed as a modular, multi-tier application platform running on AWS. It follows AWS best practices for security, scalability, and high availability.

## Infrastructure Layers

### Networking Layer (VPC)

- Isolated network environment with public and private subnets
- High availability across multiple Availability Zones
- NAT Gateways for private subnet internet access
- Security groups for network access control

### Load Balancing Layer (ALB)

- HTTPS-only Application Load Balancer
- TLS 1.2+ for secure communication
- Path-based routing to different applications

TODO:
- Access logging for security and compliance
- Header cleaning for security

### Compute Layer (EC2)

- Auto Scaling Groups for high availability
- Launch Templates for consistent instance configuration
- Private subnet placement for security
- IAM roles for AWS service access
- Security groups for instance-level firewall

## Security Considerations

- All public traffic is HTTPS only
- Instances in private subnets
- Least privilege IAM roles
- Security groups limit network access
- Load balancer drops sensitive headers
- Access logging enabled

## Scalability

- Auto Scaling Groups adjust capacity
- Load balancer distributes traffic
- Multi-AZ deployment for availability
