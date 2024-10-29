output "alb_name" {
  description = "Name of ALB"
  value       = aws_lb.main.name
}

output "alb_dns_name" {
  description = "DNS name of ALB"
  value       = aws_lb.main.dns_name
}

output "target_group_arns" {
  description = "Map of target group ARNs"
  value       = { for k, v in aws_lb_target_group.apps : k => v.arn }
}

output "alb_security_group_id" {
  description = "Security group ID of the ALB"
  value       = aws_security_group.alb.id
}
