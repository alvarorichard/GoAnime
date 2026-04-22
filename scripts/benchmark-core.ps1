$ErrorActionPreference = 'Stop'

$repoRoot = Split-Path -Parent $PSScriptRoot
$env:GOCACHE = Join-Path $repoRoot '.gocache'
New-Item -ItemType Directory -Force -Path $env:GOCACHE | Out-Null

$timestamp = Get-Date -Format 'yyyyMMdd_HHmmss'
$reportDir = Join-Path $repoRoot 'benchmarks'
$reportPath = Join-Path $reportDir ("core_{0}.txt" -f $timestamp)
New-Item -ItemType Directory -Force -Path $reportDir | Out-Null

$packages = @(
    './internal/api/source',
    './internal/api/providers',
    './internal/streaming',
    './internal/util',
    './internal/player'
)

"Go benchmark report - $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss zzz')" | Tee-Object -FilePath $reportPath
"go version: $(& go version)" | Tee-Object -FilePath $reportPath -Append
"GOOS/GOARCH: $(& go env GOOS)/$(& go env GOARCH)" | Tee-Object -FilePath $reportPath -Append
$gomaxprocs = if ($env:GOMAXPROCS) { $env:GOMAXPROCS } else { 'default' }
"GOMAXPROCS: $gomaxprocs" | Tee-Object -FilePath $reportPath -Append
"" | Tee-Object -FilePath $reportPath -Append

foreach ($pkg in $packages) {
    "### $pkg" | Tee-Object -FilePath $reportPath -Append
    & go test $pkg -run TestDoesNotExist -bench . -benchmem -count 3 -benchtime=200ms 2>&1 |
        Tee-Object -FilePath $reportPath -Append
    "" | Tee-Object -FilePath $reportPath -Append
}

Write-Output "Benchmark report -> $reportPath"
