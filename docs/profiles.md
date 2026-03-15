# Working with profiles

Profiles are the core of envirou. A profile is a named set of environment variables that you can activate with a single command.

## Creating profiles

Open the config file:

```bash
ev config
```

Add a profile section:

```ini
[profile:dev]
AWS_PROFILE=dev
AWS_DEFAULT_REGION=us-west-2
```

Variables can have three states in a profile:
- `KEY=value` — set the variable to this value
- `KEY=` — set the variable to an empty string
- `KEY` — unset (remove) the variable

### Path operators (prepend and append)

For PATH-like variables, you can use `^=` to prepend or `+=` to append instead of replacing the entire value:

| Operator | Syntax | Description |
|----------|--------|-------------|
| Replace | `VAR=value` | Set the variable to an exact value (default) |
| Prepend | `VAR^=value` | Prepend to a path-like variable |
| Append | `VAR+=value` | Append to a path-like variable |

These operators split on the platform path separator (`:` on Unix, `;` on Windows) and **deduplicate**: if a component already exists anywhere in the current value, it is skipped. This means applying the same profile twice is safe — envirou will report "already active" and make no changes.

Example:
```ini
[profile:py3]
PATH^=/Users/you/envs/py3/bin
VIRTUAL_ENV=/Users/you/envs/py3
```

If your PATH is `/usr/local/bin:/usr/bin:/bin`, running `ev set py3` produces:
```
PATH=/Users/you/envs/py3/bin:/usr/local/bin:/usr/bin:/bin
```

You can prepend multiple components at once:
```ini
PATH^=/a:/b
```

**Note:** `ev diff --save` always writes profiles using `=` (full replacement), since it captures the exact environment state rather than the delta. You can manually edit saved profiles to use `^=` or `+=` if desired.

## Activating profiles

```bash
ev set dev
```

Activate multiple profiles at once — they're applied in order:

```bash
ev set dev eu-region
```

## Viewing profiles

List all profiles (active ones are highlighted):

```bash
ev profiles
```

Show only active or inactive:

```bash
ev profiles --active
ev profiles --inactive
```

## Creating profiles from your current environment

If you've configured your environment manually and want to capture it, use the snapshot and diff workflow:

```bash
ev snapshot          # save current state
# ... make changes ...
ev diff --save dev   # save changes as a profile
ev snapshot --reset  # clean up
```

See the [snapshot and diff guide](./snapshots.md) for details.

## Example: Multi-environment setup

```ini
[profile:basic]
PATH=/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin
VIRTUAL_ENV
AWS_PROFILE

[profile:awsdev]
AWS_PROFILE=dev
AWS_DEFAULT_REGION=us-west-2

[profile:awsprod]
AWS_PROFILE=prod
AWS_DEFAULT_REGION=eu-west-1

[profile:py3]
PATH^=/Users/you/envs/py3/bin
VIRTUAL_ENV=/Users/you/envs/py3
```

Switch between environments:

```bash
ev set basic              # reset to clean baseline
ev set py3 awsdev         # Python 3 + AWS dev
ev set py3 awsprod        # Python 3 + AWS prod
```

The `basic` profile unsets `VIRTUAL_ENV` and `AWS_PROFILE` (note the bare variable names without `=`), giving you a clean slate. The `py3` profile uses `^=` to prepend to PATH rather than replacing it entirely — no need to hardcode the full path.
