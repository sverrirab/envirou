# PowerShell and envirou

## Install
Add this to your PowerShell profile (`$PROFILE`):
```powershell
Invoke-Expression (& envirou bootstrap powershell)
```

To also customize your prompt with active profile display and exit code:
```powershell
Invoke-Expression (& envirou bootstrap powershell --prompt)
```

## Uninstall
1. Remove the `Invoke-Expression` line from your `$PROFILE`
2. Remove the binary:
```powershell
Remove-Item (Get-Command envirou).Source
```
