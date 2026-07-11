#!/usr/bin/env bash
function ev() {
  if [[ "${1:-}" == "__complete" || "${1:-}" == "__completeNoDesc" ]]; then
    envirou "$@"
    return
  fi
  eval "$(envirou "$@")"
}
