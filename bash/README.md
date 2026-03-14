# Bash (or zsh) with envirou

## Install
Add this to your bash configuration file `.bashrc` (or `.zshrc` for zsh):
```bash
eval "$(envirou bootstrap bash)"
```
For zsh:
```bash
eval "$(envirou bootstrap zsh)"
```
Then restart your shell (or run the command directly in your current shell).

## Oh-My-Zsh
Link the theme folder in this repository into your local theme folder and add `ZSH_THEME="envirou"` to your startup.

## Uninstall

1. Remove the `eval` line from your `.bashrc` / `.zshrc`
2. Remove the binary `rm $(which envirou)`
3. If you don't want to restart your current shell run `unset ev` (or `unset -f ev` for zsh)
