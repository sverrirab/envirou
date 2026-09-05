# Optional additive Bash prompt integration.
__envirou_prompt_update() {
  local previous_status=$?
  local profiles
  profiles="$(ENVIROU_KEY="${ENVIROU_KEY:-}" envirou profiles --prompt-output 2>/dev/null)"
  while [[ "$profiles" == *[[:space:]] ]]; do
    profiles="${profiles%?}"
  done
  if [[ -n "$profiles" ]]; then
    ENVIROU_PROMPT_SEGMENT="[$profiles] "
  else
    ENVIROU_PROMPT_SEGMENT=""
  fi
  return "$previous_status"
}

if [[ "${PS1:-}" != *'${ENVIROU_PROMPT_SEGMENT}'* ]]; then
  PS1='${ENVIROU_PROMPT_SEGMENT}'"${PS1:-}"
fi

if [[ "$(declare -p PROMPT_COMMAND 2>/dev/null)" == "declare -a"* ]]; then
  __envirou_prompt_hook_found=0
  for __envirou_prompt_hook in "${PROMPT_COMMAND[@]}"; do
    if [[ "$__envirou_prompt_hook" == "__envirou_prompt_update" ]]; then
      __envirou_prompt_hook_found=1
      break
    fi
  done
  if [[ $__envirou_prompt_hook_found -eq 0 ]]; then
    PROMPT_COMMAND=(__envirou_prompt_update "${PROMPT_COMMAND[@]}")
  fi
  unset __envirou_prompt_hook __envirou_prompt_hook_found
elif [[ ";${PROMPT_COMMAND:-};" != *";__envirou_prompt_update;"* ]]; then
  PROMPT_COMMAND="__envirou_prompt_update${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
fi
