# TODO

## Features

### Chained profiles
Support profile dependencies - if activating `profileX`, automatically import `profileY`
first. This enables an `init` profile that always runs as a baseline, and layered profiles
that build on each other (e.g., `ev set aws-prod` could automatically chain a base `aws` profile).

### Profile inspection
`ev profiles` lists profile names but there is no way to see which variables a profile
would set without opening the config file. Add something like `ev profiles --show NAME`.

### Profile deactivation
There is no complement to `ev set`; deactivating a profile requires a "reset" profile or
manually unsetting variables. Add an `ev unset` (or similar) command. Partially overlaps
with chained profiles above.

### Improve config editing
Make the config file easier to read and modify. Ideas:
- Validate config on save
- Show config diff after editing
- Better formatting/comments in generated default config

### Diff improvements
Support reset to snapshot? This might be a footgun so potentially a bad idea.
