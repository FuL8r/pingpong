#!/usr/bin/env sh
# Run a CLI tool inside its official Docker image so the Jenkins agent needs only
# Docker (no pre-installed terraform/aws/go). The current workspace and AWS creds
# are passed through, so tools see the checked-out repo and can reach AWS.
#
# Usage: deploy/ci/dtool.sh <image> <args...>
#   deploy/ci/dtool.sh hashicorp/terraform:1.9 -chdir=deploy/terraform init ...
#   deploy/ci/dtool.sh amazon/aws-cli:2 sts get-caller-identity
#   deploy/ci/dtool.sh golang:1.26 go test ./...
set -eu

IMAGE="$1"
shift
WS="${WORKSPACE:-$(pwd)}"

# When Jenkins itself runs in a container (docker.sock mounted), the workspace
# path only exists inside that container, so share its volumes instead of a bind
# mount whose host path would not resolve on the daemon.
if [ -f /.dockerenv ]; then
  exec docker run --rm --volumes-from "$(hostname)" -w "$WS" \
    -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_DEFAULT_REGION \
    "$IMAGE" "$@"
else
  exec docker run --rm -v "$WS:$WS" -w "$WS" \
    -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_DEFAULT_REGION \
    "$IMAGE" "$@"
fi
