resource "aws_security_group" "alb" {
  name        = substr(local.alb_name, 0, 32)
  description = "ALB Security Group"
  vpc_id      = var.vpc_id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }


  tags = merge(local.common_tags, {
    Name = local.alb_name
  })
}

resource "aws_lb" "main" {
  name               = substr(local.alb_name, 0, 32)
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets           = var.public_subnets

  enable_deletion_protection = var.environment == "prod"

  tags = local.common_tags
}

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.main.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = var.certificate_arn

  default_action {
    type = "fixed-response"

    fixed_response {
      content_type = "text/plain"
      message_body = "No routes matched"
      status_code  = "404"
    }
  }
}

resource "aws_lb_target_group" "apps" {
  for_each = var.apps

  name        = "${var.project_name}-${var.environment}-${each.key}"
  port        = each.value.port
  protocol    = "HTTP"
  vpc_id      = var.vpc_id
  target_type = "instance"

  health_check {
    enabled             = true
    path                = each.value.health_check_url
    healthy_threshold   = 3
    unhealthy_threshold = 3
    timeout             = 5
    interval            = 30
  }

  tags = {
    Environment = var.environment
    Project     = var.project_name
  }
}

resource "aws_lb_listener_rule" "apps" {
  for_each = var.apps

  listener_arn = aws_lb_listener.https.arn
  priority     = each.value.priority

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.apps[each.key].arn
  }

  condition {
    path_pattern {
      values = [each.value.path]
    }
  }

  condition {
    host_header {
      values = each.value.domain
    }
  }
}
