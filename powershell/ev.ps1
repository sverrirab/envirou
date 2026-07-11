function ev {
    if ($args.Count -gt 0 -and ($args[0] -eq "__complete" -or $args[0] -eq "__completeNoDesc")) {
        & envirou $args
        return
    }
    $output = & envirou --output-powershell $args
    if ($output.Length -ne 0) {
        Invoke-Expression $output
    }
}
