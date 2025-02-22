#!/bin/sh

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
