variable "region" {
  description = "AWS region for all resources"
  type        = string
  default     = "eu-central-1"
}

variable "service_name" {
  description = "App Runner service name and resource prefix"
  type        = string
  default     = "pingpong"
}

variable "image_tag" {
  description = "Immutable image tag to deploy"
  type        = string
}

variable "container_port" {
  description = "Port the container listens on"
  type        = number
  default     = 8080
}

variable "cpu" {
  description = "App Runner CPU"
  type        = string
  default     = "0.25 vCPU"
}

variable "memory" {
  description = "App Runner memory"
  type        = string
  default     = "0.5 GB"
}

variable "min_instances" {
  description = "Autoscaling minimum instances"
  type        = number
  default     = 1
}

variable "max_instances" {
  description = "Autoscaling maximum instances"
  type        = number
  default     = 2
}

variable "max_concurrency" {
  description = "Max concurrent requests per instance"
  type        = number
  default     = 100
}

variable "custom_domain" {
  description = "Custom domain; empty uses the free *.awsapprunner.com domain"
  type        = string
  default     = ""
}

variable "hosted_zone_id" {
  description = "Route53 hosted zone id for the custom domain; empty skips DNS records"
  type        = string
  default     = ""
}

variable "enable_waf" {
  description = "Attach an AWS WAF web ACL to the service"
  type        = bool
  default     = false
}

variable "waf_rate_limit" {
  description = "WAF rate-based limit: max requests per IP per 5-minute window"
  type        = number
  default     = 1000
}

#runtime config
variable "max_body_bytes" {
  description = "Max request body bytes (PINGPONG_MAX_BODY_BYTES)"
  type        = number
  default     = 4096
}

variable "read_header_timeout" {
  description = "Header read timeout (PINGPONG_READ_HEADER_TIMEOUT)"
  type        = string
  default     = "5s"
}

variable "read_timeout" {
  description = "Full request read timeout (PINGPONG_READ_TIMEOUT)"
  type        = string
  default     = "10s"
}

variable "write_timeout" {
  description = "Response write timeout (PINGPONG_WRITE_TIMEOUT)"
  type        = string
  default     = "10s"
}

variable "idle_timeout" {
  description = "Keep-alive idle timeout (PINGPONG_IDLE_TIMEOUT)"
  type        = string
  default     = "60s"
}

variable "shutdown_timeout" {
  description = "Graceful shutdown timeout (PINGPONG_SHUTDOWN_TIMEOUT)"
  type        = string
  default     = "10s"
}

variable "max_inflight" {
  description = "Max concurrent in-flight requests (PINGPONG_MAX_INFLIGHT)"
  type        = number
  default     = 256
}

variable "log_level" {
  description = "Log level (PINGPONG_LOG_LEVEL): debug|info|warn|error"
  type        = string
  default     = "info"
}
