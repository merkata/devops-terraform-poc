data "aws_availability_zones" "available" {
  state = "available"
}

locals {
  # Get the first 3 AZs from the region
  azs = slice(data.aws_availability_zones.available.names, 0, 3)

  common_tags = {
    Environment = var.environment
    Project     = var.project_name
    ManagedBy   = "terraform"
  }
}

module "vpc" {
  source = "git::https://github.com/terraform-aws-modules/terraform-aws-vpc.git?ref=12caf8056992a6e327c584475f1522d05e77284d" # v5.14.0

  name = "${var.project_name}-${var.environment}"
  cidr = var.vpc_cidr

  azs             = local.azs
  private_subnets = [for i in range(3) : cidrsubnet(var.vpc_cidr, 4, i)]
  public_subnets  = [for i in range(3) : cidrsubnet(var.vpc_cidr, 4, i + 3)]

  # Enable public IP mapping for public subnets
  map_public_ip_on_launch = true

  enable_nat_gateway   = true
  single_nat_gateway   = var.environment != "prod"
  enable_dns_hostnames = true
  enable_dns_support   = true

  # VPC Flow Logs
  enable_flow_log                      = true
  create_flow_log_cloudwatch_log_group = true
  create_flow_log_cloudwatch_iam_role  = true

  # Tags
  tags = local.common_tags

  # Subnet specific tags
  private_subnet_tags = local.common_tags
  public_subnet_tags  = local.common_tags

  # Resource tags
  vpc_tags                 = local.common_tags
  igw_tags                 = local.common_tags
  nat_gateway_tags         = local.common_tags
  nat_eip_tags             = local.common_tags
  private_route_table_tags = local.common_tags
  public_route_table_tags  = local.common_tags
}
