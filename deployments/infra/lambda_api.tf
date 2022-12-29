data "archive_file" "api_lambda_zip" {
  type        = "zip"
  source_file = "../../cmd/lambda/api_get/main"
  output_path = "api_main.zip"
}

resource "aws_lambda_function" "api" {
  function_name = "${local.service_name}_api"

  role = aws_iam_role.lambda_exec.arn

  filename          = data.archive_file.api_lambda_zip.output_path
  handler           = "main"
  source_code_hash  = filebase64sha256(data.archive_file.api_lambda_zip.output_path)
  runtime           = "go1.x"

  environment {
    variables = {
      aws_region = var.aws_region
      codeowners_ttl_minutes = var.codeowners_ttl_minutes
      codeowners_host_table = aws_dynamodb_table.hosts.name
      codeowners_repositoryowner_table = aws_dynamodb_table.repository_owners.name
    }
  }
  
}

resource "aws_cloudwatch_log_group" "api" {
  name = "/aws/lambda/${aws_lambda_function.api.function_name}"

  retention_in_days = 30
}

resource "aws_apigatewayv2_integration" "api" {
  api_id = aws_apigatewayv2_api.lambda_gateway.id

  integration_uri    = aws_lambda_function.api.invoke_arn
  integration_type   = "AWS_PROXY"
  integration_method = "POST"
  
}

resource "aws_apigatewayv2_route" "api" {
  api_id = aws_apigatewayv2_api.lambda_gateway.id

  route_key = "GET /repository/owner"
  target    = "integrations/${aws_apigatewayv2_integration.api.id}"
  
}

resource "aws_lambda_permission" "api" {

  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.api.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.lambda_gateway.execution_arn}/*/*"
}