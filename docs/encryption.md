# Encrypted values

Envirou can store individual environment variable values encrypted at rest,
in profiles and in .env files. Encrypted values look like
`enc:v1:...` and are decrypted only when a profile or .env file is applied
to your shell. Values are encrypted with AES-256-GCM using a key derived
from your passphrase (PBKDF2-SHA256).

If you never use this feature, nothing changes for you.

## Encrypting a value

```
envirou encrypt
```

The first use asks you to choose a passphrase of at least 12 characters and creates the key material
file `~/.config/envirou/crypt.ini`. **Back this file up** - encrypted values
cannot be recovered without it (it does not contain your passphrase, only
the salt used for key derivation).

You are then prompted (hidden input) for the value to encrypt, and the token
is printed:

```
Paste this into your profile or .env file:
enc:v1:aFb0...
```

Paste the token as a normal value:

```ini
[profile:prod]
AWS_PROFILE=prod
DB_PASSWORD=enc:v1:aFb0...
```

or in a .env file:

```bash
DB_PASSWORD=enc:v1:aFb0...
```

For scripting, `envirou encrypt --stdout VALUE` prints the bare token to
stdout (use with `envirou`, not the `ev` wrapper, and note that values
passed as arguments end up in shell history).

`envirou decrypt TOKEN` decrypts a token and shows the plaintext, for
verification.

## Using encrypted profiles

Two modes:

**Prompt per operation** - just use profiles normally:

```
ev set prod
Envirou passphrase: ...
Profile prod enabled
```

**Unlocked session** - enter the passphrase once:

```
ev unlock
Envirou passphrase: ...
Unlocked - encrypted values now decrypt automatically.

ev set prod      # no prompt
ev set staging   # no prompt
ev lock          # forget the key again
```

`ev unlock` stores the derived key (not the passphrase) in the `ENVIROU_KEY`
shell variable. While unlocked, profiles with encrypted values also show
their active/inactive state correctly; while locked they display as
inactive. Envirou masks `ENVIROU_KEY` in listings and excludes it from
snapshots and diffs.

## CI and non-interactive use

There is no terminal to prompt on in CI. Generate the key once on your
machine and store it in your CI secret store:

```
envirou unlock --print-key
```

Then expose it to the job as `ENVIROU_KEY`. Decryption works without
crypt.ini when `ENVIROU_KEY` is set.

## Security notes

- The passphrase never touches disk; crypt.ini stores a random salt, the
  PBKDF2 iteration count and a check token used to detect wrong passphrases.
- While unlocked, the derived key is in your shell's environment and is
  visible to processes running as your user. `ev lock` clears it.
- Tokens authenticate themselves (GCM): a tampered token or a wrong key
  fails loudly, it never applies garbage to your environment.
- `ev snapshot` and `ev diff --save` write the live environment - i.e. the
  decrypted values - to disk in plaintext. Re-encrypt values you move from
  a diff into a profile.
- Losing crypt.ini (specifically the salt in it) makes existing tokens
  unrecoverable unless you have the derived key (`unlock --print-key`)
  stored somewhere.
