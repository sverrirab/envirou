# Bash (or zsh) with envirou

## Install
The simplest way to that is to add this to your bash configuration file `.bashrc` (or `.zshrc` if you are using zsh):
```bash
eval $(envirou bootstrap --bash)
```
and restart your shell (or run this in your local shell)

## Oh-My-Zsh
Link the theme folder in this repository into your local theme folder and add `ZSH_THEME="envirou"` to your startup.

## Uninstall

1. Remove the shell function from your `.bashrc` / `.zshrc`
2. Remove the binary `rm $(which envirou)`
3. If you don't want to restart your current shell run `unset ev` (or `unset -f ev` if you are running zsh) 
