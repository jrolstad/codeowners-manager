data "archive_file" "cron_lambda_zip" {
  type        = "zip"
  source_file = "../../cmd/lambda/cron_loader/main"
  output_path = "loader_main.zip"
}

resource "aws_lambda_function" "cron_loader" {
  function_name = "${local.service_name}_cron_loader"

  role = aws_iam_role.lambda_exec.arn

  filename          = data.archive_file.cron_lambda_zip.output_path
  handler           = "main"
  source_code_hash  = filebase64sha256(data.archive_file.cron_lambda_zip.output_path)
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

resource "aws_cloudwatch_log_group" "cron_loader" {
  name = "/aws/lambda/${aws_lambda_function.cron_loader.function_name}"

  retention_in_days = 30
}

resource "aws_cloudwatch_event_rule" "every_twelve_hours" {
  name                = "every-twelve-hours"
  description         = "Fires every 12 hours"
  schedule_expression = "rate(12 hours)"
}

resource "aws_cloudwatch_event_target" "load_owners_every_twelve_hours" {
  rule      = "${aws_cloudwatch_event_rule.every_twelve_hours.name}"
  target_id = "lambda"
  arn       = "${aws_lambda_function.cron_loader.arn}"
}

resource "aws_lambda_permission" "allow_cloudwatch_to_call_cron_loader" {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.cron_loader.function_name}"
  principal     = "events.amazonaws.com"
  source_arn    = "${aws_cloudwatch_event_rule.every_twelve_hours.arn}"
}