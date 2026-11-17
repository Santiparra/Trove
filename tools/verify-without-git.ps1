[CmdletBinding()]
param([string]$PortableDirectory = 'artifacts\portable')
. (Join-Path $PSScriptRoot 'common.ps1')
$root = Split-Path $PSScriptRoot -Parent
Set-ProjectTemp $root
$portable = Get-ArtifactDirectory $root $PortableDirectory
$identity = Get-Content -LiteralPath (Join-Path $portable 'build-info.json') -Raw | ConvertFrom-Json
$work = Join-Path $root '.local\without-git'
if (Test-Path -LiteralPath $work) { Remove-Item -LiteralPath $work -Force -Recurse }
[System.IO.Directory]::CreateDirectory($work) | Out-Null
$source = Join-Path $work 'source'
[System.IO.Directory]::CreateDirectory((Join-Path $source 'frontend')) | Out-Null
foreach ($name in @('main.go', 'go.mod', 'go.sum', 'internal')) { Copy-Item -LiteralPath (Join-Path $root $name) -Destination $source -Recurse }
Copy-Item -LiteralPath (Join-Path $root 'frontend\dist') -Destination (Join-Path $source 'frontend') -Recurse
$oldCeiling = $env:GIT_CEILING_DIRECTORIES
$env:GIT_CEILING_DIRECTORIES = $work
Push-Location $source
try {
    $gitStart = New-Object System.Diagnostics.ProcessStartInfo
    $gitStart.FileName = (Get-Command git).Source
    $gitStart.Arguments = 'rev-parse --show-toplevel'
    $gitStart.WorkingDirectory = $source
    $gitStart.UseShellExecute = $false
    $gitStart.CreateNoWindow = $true
    $gitStart.RedirectStandardOutput = $true
    $gitStart.RedirectStandardError = $true
    $gitProbe = New-Object System.Diagnostics.Process
    $gitProbe.StartInfo = $gitStart
    try {
        [void]$gitProbe.Start()
        if (-not $gitProbe.WaitForExit(10000)) { $gitProbe.Kill(); throw 'No terminó la comprobación de Git.' }
        if ($gitProbe.ExitCode -eq 0) { throw 'La copia de admisión todavía descubre un repositorio Git.' }
        if ($gitProbe.ExitCode -ne 128) { throw 'La comprobación de Git falló por un motivo distinto de ausencia de repositorio.' }
    }
    finally { $gitProbe.Dispose() }
    $flags = "-H windowsgui -X main.defaultProfile=$($identity.profile) -X main.version=$($identity.version) -X main.revision=$($identity.revision) -X main.buildID=$($identity.buildID)"
    $env:CGO_ENABLED = '0'
    $exe = Join-Path $work 'Trove.exe'
    Invoke-Checked 'go' @('build', '-tags', 'production', '-trimpath', '-buildvcs=false', '-ldflags', $flags, '-o', $exe, '.')
    Test-ExecutableIdentity $exe (Join-Path $work 'empty-cwd') $identity.version $identity.revision $identity.profile
    Test-ExecutableIdentity (Join-Path $portable 'Trove.exe') (Join-Path $work 'portable-cwd') $identity.version $identity.revision $identity.profile
}
finally { Pop-Location; $env:GIT_CEILING_DIRECTORIES = $oldCeiling }
