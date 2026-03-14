# Snapshots and Diff

Envirou can track changes to your environment over time. This is useful when installing tools, activating virtual environments, or debugging unexpected configuration changes.

## Basic workflow

### 1. Take a snapshot

Save your current environment as a baseline:

```bash
ev snapshot
```

This stores all current environment variables (excluding transient ones like `PWD` and `SHLVL`) to `~/.config/envirou/snapshot.ini`.

### 2. Make changes

Do whatever you normally do — activate a virtualenv, install a tool, source a script, switch Node versions, etc:

```bash
source venv/bin/activate
nvm use 18
```

### 3. See what changed

```bash
ev diff
```

This shows a summary of differences:
- `+` — variable was added (not present when you took the snapshot)
- `~` — variable existed but its value changed
- `-` — variable was removed

### 4. Highlighted display

When a snapshot exists, running `ev` (with no arguments) highlights modified variables in your normal environment display using the diff color (red by default). This gives you an at-a-glance view of what's different from your baseline.

## Creating profiles from changes

If you like the current changes and want to replay them later, save them as a profile:

```bash
ev diff --save myprofile
```

This appends a new `[profile:myprofile]` section to your config file with the added and changed variables. Removed variables are recorded as unset entries.

You can now activate this profile any time:

```bash
ev set myprofile
```

## Clearing the snapshot

When you're done tracking changes:

```bash
ev snapshot --reset
```

This removes the snapshot file. The diff highlights in the normal `ev` display disappear.

## Example: Capturing a Python virtualenv profile

```bash
# Start clean
ev snapshot

# Activate the virtualenv
source ~/envs/myproject/bin/activate

# See what changed
ev diff
# + VIRTUAL_ENV=/Users/you/envs/myproject
# ~ PATH=/Users/you/envs/myproject/bin:/usr/local/bin:...

# Save it as a reusable profile
ev diff --save myproject

# Clean up
ev snapshot --reset
```

Now you can switch to this environment any time with `ev set myproject` — without needing to remember the virtualenv path or source any activation scripts.

## Example: Debugging unexpected changes

Something changed your environment and you're not sure what:

```bash
# Take a snapshot at the start of your session
ev snapshot

# ... work for a while, run scripts, etc ...

# Check what's different
ev diff
```

This is especially useful for catching scripts that silently modify `PATH`, set proxy variables, or override tool configurations.

## Tips

- **Take snapshots early**: Add `ev snapshot` to your shell startup if you want a daily baseline to compare against.
- **Snapshot is global**: There's one snapshot at a time. Taking a new snapshot replaces the previous one.
- **Ignored variables**: Transient variables (`PWD`, `SHLVL`, `OLDPWD`, etc.) from the `..ignore` group in your config are automatically excluded from snapshots and diffs.
