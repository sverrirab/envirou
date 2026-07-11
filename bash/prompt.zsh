# Optional additive Zsh prompt integration.
autoload -Uz add-zsh-hook

__envirou_prompt_update() {
  local previous_status=$?
  local profiles
  profiles="$(envirou profiles --prompt-output 2>/dev/null)"
  while [[ "$profiles" == *[[:space:]] ]]; do
    profiles="${profiles%?}"
  done
  profiles="${profiles//\%/%%}"
  if [[ -n "$profiles" ]]; then
    ENVIROU_PROMPT_SEGMENT="[$profiles] "
  else
    ENVIROU_PROMPT_SEGMENT=""
  fi
  return "$previous_status"
}

setopt prompt_subst
if [[ "${PROMPT:-}" != *'${ENVIROU_PROMPT_SEGMENT}'* ]]; then
  PROMPT='${ENVIROU_PROMPT_SEGMENT}'"${PROMPT:-}"
fi

add-zsh-hook -d precmd __envirou_prompt_update 2>/dev/null
add-zsh-hook precmd __envirou_prompt_update
