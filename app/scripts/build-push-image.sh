#!/usr/bin/env bash

set -e

cd "$(dirname $0)/.." || exit 1

function assert_env() {
  missing=false
  for env_var_key in "$@"; do
    if [ -z "${!env_var_key}" ]; then
      echo "Error: ENV \"$env_var_key\" must be set and must not be empty" >&2
      missing=true
    fi
  done

  if [ "$missing" = true ]; then
    exit 1
  fi
}

assert_env SERVICE TAG

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin "cmd/${SERVICE}/main.go"
TARGET="cr.yandex/$(../terraform/tf output -json -no-color | jq -cMr .container_registry.value.repository.${SERVICE}.name):${TAG}"
docker build -f build/${SERVICE}/Dockerfile -t "${TARGET}" .
rm bin
yc iam create-token | docker login cr.yandex -u iam --password-stdin
docker push "${TARGET}"