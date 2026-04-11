# IDEAS - Envirou Improvement Notes

Review conducted by approaching the project as a new user: reading all documentation,
building from source, installing, and testing every command against the docs and help output.

---

## Documentation

- **Help output shows `envirou` not `ev`**: When a user runs `ev help`, the usage section
  says `envirou [flags]` and `envirou [command]`. This is technically correct (ev is a shell
  function that calls envirou) but could confuse new users who were told to use `ev`.
  Consider aliasing the binary name in help output when invoked via the ev wrapper.

- **dotenv without wrapper**: The dotenv guide doesn't mention what happens when `ev dotenv`
  is run without the `ev` wrapper (i.e., as `envirou dotenv`). Running `envirou dotenv file`
  just prints the export commands to stdout without executing them, which could confuse users.

---

## Potential Improvements

1. **`ev config` without EDITOR exits with code 3** - Could be friendlier. Maybe suggest
   common editors or offer to use `vi`/`nano` as fallback? Or at least exit with code 1
   (the conventional "general error") rather than 3.

2. **`diff` and `snapshot` are in "Configuration commands" group** - In the help output these
   appear under "Configuration commands" alongside `bootstrap`, `config`, and `install`. The
   README organizes them under "Tracking changes" which feels more natural. Consider moving
   them to a "Tracking" command group or into "Profile commands".

3. **`set` command doesn't validate profile names** - `ev set nonexistent` prints a warning
   but exits 0. Some users might prefer a non-zero exit code when no profiles were
   successfully applied. At minimum, `ev set nonexistent` (all profiles missing) could exit
   non-zero.

4. **No way to list what a profile contains** - `ev profiles` lists profile names but there's
   no `ev profiles show <name>` or `ev profile <name>` to see what variables a profile would
   set. Users have to open the config file. A `ev profiles --show <name>` flag or similar
   would be useful.

5. **Completion not integrated into install** - The `completion` command exists (bash, zsh,
   fish, powershell) but `envirou install` doesn't set up completions. Users have to
   manually run `envirou completion bash > /etc/bash_completion.d/envirou` or similar.
   Consider adding completion setup to the install command, or at least mentioning it.

6. **`ev dotenv` without `.env` file exits with code 1** - This is correct behavior, but
   could be friendlier. A message like "No .env file found in current directory. Specify
   a file: ev dotenv <file>" would help new users.

7. **No `ev unset` command** - To deactivate a profile, users have to create a "reset"
   profile that unsets everything, or manually unset variables. A complement to `ev set`
   would be useful (this is partially covered in TODO.md's "chained profiles" idea).

8. **Variables can appear in multiple groups** - In `-a` output, variables like `HOME`, `PWD`,
   and `SHLVL` appear under multiple groups (`.powershell`, `.shell`, `.system`, `..ignore`).
   This might be intentional for completeness but could be confusing. Consider noting this
   behavior or adding a setting to show each variable only once.

---

## Build & Release

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
