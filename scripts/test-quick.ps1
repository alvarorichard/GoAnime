$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if (-not $env:GOCACHE) {
    $env:GOCACHE = Join-Path $root ".gocache"
}
New-Item -ItemType Directory -Force -Path $env:GOCACHE | Out-Null

function Invoke-Step {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Label,
        [Parameter(Mandatory = $true)]
        [string]$Command
    )

    Write-Host ""
    Write-Host "==> $Label"
    Invoke-Expression $Command
}

Invoke-Step "API compile smoke" "go test ./internal/api -run TestDoesNotExist -count=1"
Invoke-Step "Providers + source/util/streaming tests" "go test ./internal/api/providers ./internal/api/source ./internal/util ./internal/streaming -count=1"
Invoke-Step "Download stack compile smoke" "go test ./internal/download ./internal/downloader -run TestDoesNotExist -count=1"
Invoke-Step "Handlers compile smoke" "go test ./internal/handlers -run TestDoesNotExist -count=1"
Invoke-Step "Player compile smoke" "go test ./internal/player -run TestDoesNotExist -count=1"
Invoke-Step "Scraper compile smoke" "go test ./internal/scraper -run TestDoesNotExist -count=1"
Invoke-Step "Appflow compile artifact" "go test -c ./internal/appflow -o $env:TEMP\appflow.test.exe"
Invoke-Step "Playback compile artifact" "go test -c ./internal/playback -o $env:TEMP\playback.test.exe"

if ($env:GOANIME_LIVE_SMOKE -eq "1") {
    Invoke-Step "Live Goyabu smoke" "go test ./internal/player -run 'TestSmokeGoyabu(QuickDownload|DownloadResolver)' -count=1 -v"
}
