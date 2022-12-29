resource "aws_dynamodb_table" "hosts" {
  name           = "${local.service_name}_hosts"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "Id"

  attribute {
    name = "Id"
    type = "S"
  }

}

resource "aws_dynamodb_table" "repository_owners" {
  name           = "${local.service_name}_repository_owners"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "Id"

  attribute {
    name = "Id"
    type = "S"
  }

  ttl {
    attribute_name = "ExpiresAt"
    enabled        = true
  }

}