environment     = "dev"
project_name    = "demo"
vpc_cidr        = "10.0.0.0/16"
instance_type   = "t3.micro"
instance_count  = 2
certificate_arn = "arn:aws:acm:us-east-1:683721267198:certificate/aa67a8ae-f2fe-4cef-95e6-a676fd11f5be"

apps = {
  app1 = {
    port             = 8085
    path             = "/app1/*"
    health_check_url = "/app1/status"
    domain           = ["merkata.cloudns.be"]
    priority         = 100
  }
  app2 = {
    port             = 8086
    path             = "/app2/*"
    health_check_url = "/app2/status"
    domain           = ["merkata.cloudns.be"]
    priority         = 200
  }
}
