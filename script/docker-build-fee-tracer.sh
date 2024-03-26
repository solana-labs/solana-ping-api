#!/usr/bin/env bash

here=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

docker build \
  -f "$here/../docker/Dockerfile.default" \
  --build-arg COMPONENT=fee-tracer \
  -t anzaxyz/fee-tracer \
  .