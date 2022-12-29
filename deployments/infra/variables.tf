variable "aws_region" {
  description = "AWS region for all resources."

  type    = string
  default = "us-west-2"
}

variable "environment" {
  description = "Environment the infrasructure is for."

  type    = string
  default = "prd"
}

variable "codeowners_ttl_minutes" {
    description = "How long code owners data lasts before expiring"

    type = string
    default = "180"
}