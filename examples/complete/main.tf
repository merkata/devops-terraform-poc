module "vpc" {
  source = "../../modules/vpc"

  environment  = var.environment
  project_name = var.project_name
  vpc_cidr     = var.vpc_cidr
}

module "alb" {
  source = "../../modules/alb"

  environment     = var.environment
  project_name    = var.project_name
  vpc_id          = module.vpc.vpc_id
  public_subnets  = module.vpc.public_subnets
  certificate_arn = var.certificate_arn

  apps = var.apps
}

module "compute" {
  source = "../../modules/compute"

  environment           = var.environment
  project_name          = var.project_name
  vpc_id                = module.vpc.vpc_id
  private_subnets       = module.vpc.private_subnets
  instance_type         = var.instance_type
  instance_count        = var.instance_count
  alb_security_group_id = module.alb.alb_security_group_id
  target_group_arns     = [for app_name in keys(var.apps) : module.alb.target_group_arns[app_name]]
  apps                  = module.alb.apps
}
