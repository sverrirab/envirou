# envirou - Manage your shell environment variables

Simple utility to manage your shell `env`.  
Provides simple customizable and colorized output in addition to profiles and reset support.

# Quickstart

```bash
git clone git@github.com:sverrirab/envirou.git
alias ev="source $PWD/envirou/envirou"
```

# Background

Everyone that works with complex infrastructure from the command line has gathered dozens and sometimes hundreds of files that manipulate your command line environment.  Classical examples are PATH's to tools/SDK versions, external service endpoints for your PROD and DEV environments etc etc.

There are two basic problems with this: firstly you are having to memorize a bunch of script names and secondly you are never 100% which environment is active at any point and in which shell window.
 
Tools such as zsh (oh-my-zsh) and shell prompt configuration but those risk adding too much clutter to your terminal session window.

The name Envirou is inspired by Spirou the comic book character.  
The alias `ev` is both short for *Envirou* and `env`. 
