# Bash (or zsh) with envirou

## Install
The easiest way to install:
```bash
envirou install
```
This auto-detects your shell (bash or zsh) and adds the bootstrap line to your profile.

Use `envirou install --dry-run` to preview what will be added and where.

To prepend active profiles while preserving your existing Bash or Zsh prompt:
```bash
envirou install --prompt
```

### Manual install
If you prefer to edit your profile yourself, add this to `.bashrc` (or `.zshrc` for zsh):
```bash
eval "$(envirou bootstrap bash)"
```
For zsh:
```bash
eval "$(envirou bootstrap zsh)"
```
Then restart your shell (or run the command directly in your current shell).

For manual opt-in prompt integration, add `--prompt` to the bootstrap command:
```bash
eval "$(envirou bootstrap bash --prompt)"
# or, in zsh after Oh My Zsh has loaded:
eval "$(envirou bootstrap zsh --prompt)"
```

Prompt and completion integration can be enabled together:
```zsh
source <(envirou bootstrap zsh --prompt --completion)
```

## Oh-My-Zsh theme
To use the envirou zsh theme (shows active profiles in your prompt):

1. Copy (or symlink) the theme file to your Oh-My-Zsh custom themes folder:
   ```bash
   cp oh-my-zsh/envirou.zsh-theme ~/.oh-my-zsh/custom/themes/
   ```
2. Set the theme in your `.zshrc`:
   ```bash
   ZSH_THEME="envirou"
   ```
3. Restart your shell.

Note: the `ev` shell function must also be installed (see above).

## Uninstall
```bash
envirou install --uninstall
```

Or manually:
1. Remove the `eval` line from your `.bashrc` / `.zshrc`
2. Remove the binary `rm $(which envirou)`
3. If you don't want to restart your current shell run `unset ev` (or `unset -f ev` for zsh)
