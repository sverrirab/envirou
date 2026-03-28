function ev {
    $output = (& envirou --output-powershell $args) -join "`n"
    if ($output.Length -ne 0) {
        Invoke-Expression $output
    }
}
