# IDEAS - Envirou Code Review & Improvement Notes

Review conducted by approaching the project as a new user: reading all documentation,
building from source, installing, and testing every command against the docs and help output.

---

## Errors Fixed (on branch `fix/docs-and-code-review`)

These have been fixed directly:

1. **`--config` flag help text showed wrong default path** - Said `$HOME/.envirou/config.ini`
   but the actual default is `$HOME/.config/envirou/config.ini` (`cmd/root.go:143`).

2. **README regex example uses wrong syntax** - The find example `ev find -r 'PATH\|HOME'`
   doesn't work because Go regex uses `|` not `\|`. The backslash-escaped pipe produces zero
   results. Fixed the markdown to render correctly with `&#124;`.

3. **Deprecated `ioutil` usage** - `ioutil.ReadFile`, `ioutil.WriteFile`, and `ioutil.TempFile`
   are deprecated since Go 1.16. Replaced with `os.ReadFile`, `os.WriteFile`, and `os.CreateTemp`
   in `pkg/ini/ini.go`, `pkg/config/default.go`, and `pkg/ini/ini_test.go`.

4. **Wrong godoc comment on `GetBool`** - `pkg/ini/ini.go:139` had `// GetString` as the
   comment for the `GetBool` function (copy-paste error).

5. **Typo in `profile.go`** - "explitly" -> "explicitly" in `GetNil` comment.

6. **`path` command missing from README** - The `ev path` command exists and works but was
   not listed in any of the README command tables.

---

## Documentation Inconsistencies

### README vs actual behavior

- **Help output shows `envirou` not `ev`**: When a user runs `ev help`, the usage section
  says `envirou [flags]` and `envirou [command]`. This is technically correct (ev is a shell
  function that calls envirou) but could confuse new users who were told to use `ev`.
  Consider aliasing the binary name in help output when invoked via the ev wrapper.

- **Undocumented command aliases**: Several aliases exist but are not mentioned in README:
  - `set` -> `.` (dot)
  - `snapshot` -> `snap`
  - `profiles` -> `profile`, `p`
  - `groups` -> `group`, `g`
  - `dotenv` -> `.env`
  Only `search` (alias for `find`) is documented. Consider adding a note about aliases.

- **`path` command has no dedicated documentation**: Unlike `dotenv`, `profiles`, and
  `snapshots`, which each have a guide in `docs/`, the `path` command has no documentation
  page. It's a useful command (`--check` for finding duplicates/missing dirs is great) and
  deserves a `docs/path.md`.

### Snapshots documentation

- The snapshots doc says the snapshot is stored at `~/.config/envirou/snapshot.ini` which
  is correct, but this differs from what the `--config` help used to say (now fixed).

### dotenv documentation

- The dotenv guide is well-written. One minor note: it doesn't mention what happens when
  `ev dotenv` is run without the `ev` wrapper (i.e., as `envirou dotenv`). Running
  `envirou dotenv file` just prints the export commands to stdout without executing them,
  which could confuse users.

---

## Code Quality Observations

### Things that work well
- Clean cobra CLI structure with well-organized subcommands
- The `--check` flag on `path` is genuinely useful (finds duplicates and missing dirs)
- Prepend/append operators (`^=`, `+=`) for PATH-like variables with deduplication is clever
- Snapshot/diff workflow is intuitive and well-implemented
- Profile "already active" detection works correctly even with prepend/append modes
- All 30+ tests pass, covering bootstrap, set, profiles, dotenv, find, path, snapshot, diff
- The default config file has excellent group categorization (shell, locale, network, etc.)

### Potential improvements

1. **`ev config` without EDITOR exits with code 3** - Could be friendlier. Maybe suggest
   common editors or offer to use `vi`/`nano` as fallback? Or at least exit with code 1
   (the conventional "general error") rather than 3.

2. **Password masking is limited** - The default config masks `AWS_SECRET_ACCESS_KEY` and
   `AWS_SESSION_TOKEN`, but in testing, `CLAUDE_CODE_OAUTH_TOKEN` was displayed in plain
   text. Consider:
   - Adding common token/secret patterns to default config (e.g., `*TOKEN*`, `*SECRET*`)
   - Or a "smart" default that masks any variable containing TOKEN, SECRET, PASSWORD, KEY
     in its name

3. **`diff` and `snapshot` are in "Configuration commands" group** - In the help output these
   appear under "Configuration commands" alongside `bootstrap`, `config`, and `install`. The
   README organizes them under "Tracking changes" which feels more natural. Consider moving
   them to a "Tracking" command group or into "Profile commands".

4. **`set` command doesn't validate profile names** - `ev set nonexistent` prints a warning
   but exits 0. Some users might prefer a non-zero exit code when no profiles were
   successfully applied. At minimum, `ev set nonexistent` (all profiles missing) could exit
   non-zero.

5. **No way to list what a profile contains** - `ev profiles` lists profile names but there's
   no `ev profiles show <name>` or `ev profile <name>` to see what variables a profile would
   set. Users have to open the config file. A `ev profiles --show <name>` flag or similar
   would be useful.

6. **Completion not integrated into install** - The `completion` command exists (bash, zsh,
   fish, powershell) but `envirou install` doesn't set up completions. Users have to
   manually run `envirou completion bash > /etc/bash_completion.d/envirou` or similar.
   Consider adding completion setup to the install command, or at least mentioning it.

7. **`ev dotenv` without `.env` file exits with code 1** - This is correct behavior, but
   could be friendlier. A message like "No .env file found in current directory. Specify
   a file: ev dotenv <file>" would help new users.

8. **No `ev unset` command** - To deactivate a profile, users have to create a "reset"
   profile that unsets everything, or manually unset variables. A complement to `ev set`
   would be useful (this is partially covered in TODO.md's "chained profiles" idea).

9. **Variables can appear in multiple groups** - In `-a` output, variables like `HOME`, `PWD`,
   and `SHLVL` appear under multiple groups (`.powershell`, `.shell`, `.system`, `..ignore`).
   This might be intentional for completeness but could be confusing. Consider noting this
   behavior or adding a setting to show each variable only once.

---

## Build & Release

- **Version shows "dev"** when built without ldflags. The `go install` path in the README
  would produce a binary that shows `Version dev`. Consider documenting the ldflags needed
  or using a build system that sets the version automatically.

- **`go.mod` specifies `go 1.21`** but the code now uses `os.ReadFile`/`os.WriteFile` which
  require Go 1.16+. The min version could be bumped if desired, but 1.21 is fine (it's
  higher than needed).

- **Vendor directory is committed** - This is a valid Go pattern but adds significant bulk
  to the repo. Modern Go projects typically rely on `go.sum` and the module proxy. Worth
  considering whether vendor/ is still needed.

---

## Shell Integration Observations

- **bash and zsh use the same bootstrap** - Both get `function ev() { eval "$(envirou "$@")"; }`.
  The zsh bootstrap doesn't use any zsh-specific features. This is fine but means there's
  no special zsh completion integration via the bootstrap.

- **PowerShell bootstrap is collapsed to one line** - The `collapseToOneLine` function in
  `bootstrap.go` joins the multiline PowerShell script with `; `. This works but makes
  debugging harder if something goes wrong. Consider keeping multiline for PowerShell.

- **Oh-My-Zsh theme calls `envirou profiles --active`** - This runs the full binary on
  every prompt render. For most users this is fast enough, but on slow systems or with
  large configs it could add latency. The PowerShell prompt has the same pattern.

---

## Testing Gaps

- **No integration tests for the `install` command** - The install command modifies real
  files (.bashrc, .zshrc, $PROFILE). Consider adding tests that use temp files.

- **No tests for the `config` command without EDITOR** - The exit-code-3 case isn't tested.

- **No test for `dotenv` with missing default `.env`** - The error path when no `.env`
  exists in the current directory isn't tested.

- **`find --regex` not tested with complex patterns** - The README example regex pattern
  wasn't tested to verify it actually works (and it didn't, with the escaped pipe).

---

## Minor Nits

- `shell.go:101` has a comment saying "seperator" (should be "separator")
- `default.go` uses `full_path` and `current_user` (snake_case) instead of Go-idiomatic
  `fullPath` and `currentUser`
- `ini.go:58` uses `first_char` (snake_case) instead of `firstChar`
- `config.go:13` and `bootstrap.go:12` both have `// setCmd represents the set command`
  as copy-paste godoc comments that don't match the actual command
