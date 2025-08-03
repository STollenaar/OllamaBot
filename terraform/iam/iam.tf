resource "aws_iam_role" "ollamabot_role" {
  name               = "OllamabotRole"
  description        = "Role for the ollamabot"
  assume_role_policy = data.aws_iam_policy_document.assume_policy_document.json
}

resource "aws_iam_role_policy" "ollamabot_role_policy" {
  role   = aws_iam_role.ollamabot_role.id
  name   = "inline-role"
  policy = data.aws_iam_policy_document.ssm_access_role_policy_document.json
}
