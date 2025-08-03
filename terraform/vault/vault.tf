
resource "vault_aws_secret_backend_role" "ollamabot_role" {
  backend         = data.terraform_remote_state.vault_setup.outputs.vault_aws_client
  name            = data.terraform_remote_state.iam_role.outputs.iam.ollamabot_role.name
  credential_type = "assumed_role"
  role_arns       = [data.terraform_remote_state.iam_role.outputs.iam.ollamabot_role.arn]
}
