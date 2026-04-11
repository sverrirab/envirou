# Inspecting path variables

Envirou can display path-like variables with one entry per line, making them much easier to read than the raw colon-separated (or semicolon-separated on Windows) value.

## Basic usage

Show all variables that envirou recognises as paths (configured via the `path` setting):

```bash
ev path
```

Show a specific variable:

```bash
ev path PATH
```

## Checking for problems

Use `--check` to flag missing directories and duplicates:

```bash
ev path --check
```

Each entry is annotated:
- **not found** — the directory does not exist on disk
- **duplicate** — the same entry appears more than once
- **empty** — a blank entry (e.g. from a trailing `:`)

A summary line is printed at the end of each variable showing the total count of entries and any issues found.

## Which variables are shown?

The `path` setting in your config file controls which variables are treated as path-like:

```ini
[settings]
path=HOME, PATH, GOPATH, JAVA_HOME, KUBECONFIG, VIRTUAL_ENV
```

Only these variables appear in `ev path` output (without an explicit argument) and receive the alternating-colour formatting in normal `ev` output.

You can always pass any variable name explicitly, even if it is not in the path list:

```bash
ev path MANPATH
```

## Example

```
$ ev path --check PATH
# PATH
/home/you/.local/bin
/usr/local/bin
/usr/bin
/bin
/usr/local/games  [not found]
/usr/games  [not found]
/snap/bin
/home/you/.local/bin  [duplicate]
# 8 entries -- 2 missing, 1 duplicate
```
