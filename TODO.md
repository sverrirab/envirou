# TODO

## Features

### Chained profiles
Support profile dependencies — if activating `profileX`, automatically import `profileY`
first. This enables an `init` profile that always runs as a baseline, and layered profiles
that build on each other (e.g., `ev set aws-prod` could automatically chain a base `aws` profile).

### Improve config editing
Make the config file easier to read and modify. Ideas:
- Validate config on save
- Show config diff after editing
- Better formatting/comments in generated default config

### Diff improvements
Support reset to snapshot? This might be a footgun so potentially a bad idea.


## Bugs / Cleanup
* Test and document command line completion scripts
