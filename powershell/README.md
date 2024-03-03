# PowerShell and envirou

To enable ev in your current session run this script: [ev.ps1](./ev.ps1) or (it's built into envirou):
```powershell
Invoke-Expression -Command $(envirou bootstrap --powershell)
```

## Uninstall
```powershell
Remove-Item (Get-Command envirou).Source
```

## Update prompt with envirou profile
Run this script [prompt.ps1](./prompt.ps1) 
