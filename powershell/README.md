# PowerShell and envirou

## Install
The easiest way to install:
```powershell
envirou install powershell
```

To also customize your prompt with active profile display and exit code:
```powershell
envirou install powershell --prompt
```

Use `envirou install --dry-run` to preview what will be added and where.

### Manual install
If you prefer to edit your profile yourself, add this to your PowerShell profile (`$PROFILE`):
```powershell
Invoke-Expression (& envirou bootstrap powershell)
```

To also customize your prompt with active profile display and exit code:
```powershell
Invoke-Expression (& envirou bootstrap powershell --prompt)
```

## Uninstall
```powershell
envirou install --uninstall
```

Or manually:
1. Remove the `Invoke-Expression` line from your `$PROFILE`
2. Remove the binary:
```powershell
Remove-Item (Get-Command envirou).Source
```
