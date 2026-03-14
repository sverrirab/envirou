function ev {
    $output = & envirou --output-powershell $args
    if ($output.Length -ne 0) {
        Invoke-Expression $output
    }
}
