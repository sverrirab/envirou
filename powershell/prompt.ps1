function prompt {
    $suffix = "> "
    if ($LASTEXITCODE -ne 0) {
        $suffix = " $([char]27)[31m[$LASTEXITCODE]$([char]27)[0m " + $suffix
    }
    $hasLocalKey = ($null -ne $global:ENVIROU_KEY)
    if ($hasLocalKey) { $env:ENVIROU_KEY = $global:ENVIROU_KEY }
    $envirouProfiles = envirou profiles --active 2>&1
    if ($hasLocalKey) { Remove-Item Env:ENVIROU_KEY -ErrorAction SilentlyContinue }
    "$envirouProfiles${pwd}`r`n" + $suffix
}
