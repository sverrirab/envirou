#!/usr/bin/env bash

ev() {
  local output
  output="$(envirou "$@")"
  if [ -n "${output}" ]; then
    eval "${output}";
  fi
}
