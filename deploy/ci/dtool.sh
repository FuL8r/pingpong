#!/usr/bin/env sh
# Run a CLI tool inside its official Docker image so the Jenkins agent needs only docker
set -eu

IMAGE="$1"
shift
WS="${WORKSPACE:-$(pwd)}"

# When Jenkins itself runs in a container (docker.sock mounted), the workspace
# path only exists inside that container
if [ -f /.dockerenv ]; then
  exec docker run --rm --volumes-from "$(hostname)" -w "$WS" \
    -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_DEFAULT_REGION \
    "$IMAGE" "$@"
else
  exec docker run --rm -v "$WS:$WS" -w "$WS" \
    -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_DEFAULT_REGION \
    "$IMAGE" "$@"
fi
