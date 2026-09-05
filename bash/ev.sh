#!/usr/bin/env bash
# ENVIROU_KEY is a non-exported shell variable (set by "ev unlock") so other
# programs never inherit it; pass it to the envirou child process only.
function ev() {
  if [[ "${1:-}" == "__complete" || "${1:-}" == "__completeNoDesc" ]]; then
    ENVIROU_KEY="${ENVIROU_KEY:-}" envirou "$@"
    return
  fi
  eval "$(ENVIROU_KEY="${ENVIROU_KEY:-}" envirou "$@")"
}
