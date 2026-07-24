resource "aws_apprunner_auto_scaling_configuration_version" "pingpong" {
  auto_scaling_configuration_name = var.service_name
  max_concurrency                 = var.max_concurrency
  min_size                        = var.min_instances
  max_size                        = var.max_instances

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_apprunner_service" "pingpong" {
  service_name = var.service_name

  source_configuration {
    auto_deployments_enabled = false

    authentication_configuration {
      access_role_arn = aws_iam_role.apprunner_ecr_access.arn
    }

    image_repository {
      image_identifier      = "${aws_ecr_repository.pingpong.repository_url}:${var.image_tag}"
      image_repository_type = "ECR"

      image_configuration {
        port = tostring(var.container_port)
        runtime_environment_variables = {
          PINGPONG_ADDR                = ":${var.container_port}"
          PINGPONG_MAX_BODY_BYTES      = tostring(var.max_body_bytes)
          PINGPONG_READ_HEADER_TIMEOUT = var.read_header_timeout
          PINGPONG_READ_TIMEOUT        = var.read_timeout
          PINGPONG_WRITE_TIMEOUT       = var.write_timeout
          PINGPONG_IDLE_TIMEOUT        = var.idle_timeout
          PINGPONG_SHUTDOWN_TIMEOUT    = var.shutdown_timeout
          PINGPONG_MAX_INFLIGHT        = tostring(var.max_inflight)
          PINGPONG_LOG_LEVEL           = var.log_level
        }
      }
    }
  }

  instance_configuration {
    cpu    = var.cpu
    memory = var.memory
  }

  auto_scaling_configuration_arn = aws_apprunner_auto_scaling_configuration_version.pingpong.arn

  health_check_configuration {
    protocol            = "TCP"
    interval            = 10
    timeout             = 5
    healthy_threshold   = 1
    unhealthy_threshold = 5
  }

  network_configuration {
    ingress_configuration {
      is_publicly_accessible = true
    }
  }
}
