# TODO

## Features

### search/find env variables
ev find -i path 
with simple regex support

### Chained profiles
Support profile dependencies — if activating `profileX`, automatically import `profileY`
first. This enables an `init` profile that always runs as a baseline, and layered profiles
that build on each other (e.g., `ev set aws-prod` could automatically chain a base `aws` profile).

### Base variables for paths
Store a base PATH (or other path-like variables) in the environment using a reserved prefix
(e.g., `_ENVIROU_BASE_PATH`). Profiles can then append/prepend to the base rather than
storing the full path. This makes switching Python virtualenvs or SDK versions cleaner —
profiles only specify the delta, not the entire PATH.

### Improve config editing
Make the config file easier to read and modify. Ideas:
- Validate config on save
- Show config diff after editing
- Better formatting/comments in generated default config

## Bugs / Cleanup
* Test and document command line completion scripts
