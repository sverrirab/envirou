function ev {
    if ($args.Count -gt 0 -and ($args[0] -eq "__complete" -or $args[0] -eq "__completeNoDesc")) {
        & envirou $args
        return
    }
    $hasLocalKey = ($null -ne $global:ENVIROU_KEY)
    if ($hasLocalKey) { $env:ENVIROU_KEY = $global:ENVIROU_KEY }
    $output = & envirou --output-powershell $args
    if ($hasLocalKey) { Remove-Item Env:ENVIROU_KEY -ErrorAction SilentlyContinue }
    if ($output.Length -ne 0) {
        Invoke-Expression $output
    }
}
