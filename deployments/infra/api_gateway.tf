
resource "aws_cloudwatch_log_group" "lambda_gateway" {
  name = "/aws/api_gw/${aws_apigatewayv2_api.lambda_gateway.name}"

  retention_in_days = 30
}

resource "aws_apigatewayv2_api" "lambda_gateway" {
  name          = local.service_name
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_stage" "lambda_gateway" {
  name = local.service_name
  api_id = aws_apigatewayv2_api.lambda_gateway.id

  auto_deploy = true

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.lambda_gateway.arn

    format = jsonencode({
      requestId               = "$context.requestId"
      sourceIp                = "$context.identity.sourceIp"
      requestTime             = "$context.requestTime"
      protocol                = "$context.protocol"
      httpMethod              = "$context.httpMethod"
      resourcePath            = "$context.resourcePath"
      routeKey                = "$context.routeKey"
      status                  = "$context.status"
      responseLength          = "$context.responseLength"
      integrationErrorMessage = "$context.integrationErrorMessage"
      }
    )
  }
}