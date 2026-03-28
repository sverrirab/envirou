function prompt {
    $suffix = "> "
    if ($LASTEXITCODE -ne 0) {
        $suffix = " $([char]27)[31m[$LASTEXITCODE]$([char]27)[0m " + $suffix
    }
    "$(envirou profiles --active 2>&1)${pwd}`r`n" + $suffix
}
