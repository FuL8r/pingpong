terraform {
  # Partial config — real values are passed at init:
  #   terraform init \
  #     -backend-config="bucket=<state_bucket>" \
  #     -backend-config="key=pingpong/app-runner.tfstate" \
  #     -backend-config="region=eu-central-1" \
  #     -backend-config="dynamodb_table=<lock_table>"
  backend "s3" {}
}
