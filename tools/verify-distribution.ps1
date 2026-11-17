[CmdletBinding()]
param([ValidatePattern('^[0-9a-fA-F]{40}$')][string]$Revision)
. (Join-Path $PSScriptRoot 'common.ps1')
$root = Split-Path $PSScriptRoot -Parent
Set-ProjectTemp $root
$version = '1.2.3'
$relative = 'artifacts\distribution-contract'
& (Join-Path $PSScriptRoot 'build.ps1') -Version $version -Revision $Revision -Profile habitual -BuildID contract -OutputDirectory $relative -SkipFrontend
$source = Get-ArtifactDirectory $root $relative
& (Join-Path $PSScriptRoot 'package.ps1') -PortableDirectory $relative -OutputDirectory 'artifacts\distribution-zip' -Version $version -Revision $Revision
$zip = Join-Path $root 'artifacts\distribution-zip\Trove-1.2.3-windows-x64.zip'
$expanded = Join-Path $root '.local\distribution-expanded'
if (Test-Path -LiteralPath $expanded) { Remove-Item -LiteralPath $expanded -Recurse -Force }
Expand-Archive -LiteralPath $zip -DestinationPath $expanded
Test-ExecutableIdentity (Join-Path $expanded 'Trove.exe') (Join-Path $root '.local\distribution-cwd') $version $Revision habitual

function Assert-Rejected {
    param([scriptblock]$Action, [string]$Scenario)
    $rejected = $false
    try { & $Action } catch { $rejected = $true }
    if (-not $rejected) { throw "No se rechazó: $Scenario" }
}
Assert-Rejected { & (Join-Path $PSScriptRoot 'package.ps1') -PortableDirectory $relative -OutputDirectory 'artifacts\distribution-negative' -Version '9.9.9' -Revision $Revision } 'versión no correspondiente'
$extra = Join-Path $source 'unexpected-personal-data.txt'
try {
    'synthetic' | Set-Content -LiteralPath $extra -Encoding ASCII
    Assert-Rejected { & (Join-Path $PSScriptRoot 'package.ps1') -PortableDirectory $relative -OutputDirectory 'artifacts\distribution-negative' -Version $version -Revision $Revision } 'archivo no permitido'
}
finally { Remove-Item -LiteralPath $extra -Force }
$sumFile = Join-Path $source 'SHA256SUMS.txt'
$original = Get-Content -LiteralPath $sumFile -Raw
try {
    '' | Set-Content -LiteralPath $sumFile -Encoding ASCII
    Assert-Rejected { & (Join-Path $PSScriptRoot 'package.ps1') -PortableDirectory $relative -OutputDirectory 'artifacts\distribution-negative' -Version $version -Revision $Revision } 'manifiesto incompleto'
}
finally { $original | Set-Content -LiteralPath $sumFile -Encoding ASCII -NoNewline }

$oldActions = $env:GITHUB_ACTIONS
$oldEvent = $env:GITHUB_EVENT_NAME
$oldRef = $env:GITHUB_REF
try {
    $env:GITHUB_ACTIONS = 'true'
    $env:GITHUB_EVENT_NAME = 'pull_request'
    $env:GITHUB_REF = 'refs/heads/main'
    Assert-Rejected { & (Join-Path $PSScriptRoot 'publish.ps1') -Repository 'synthetic/Trove' -Version $version -Revision $Revision } 'publicación desde PR'
    $env:GITHUB_EVENT_NAME = 'push'
    $env:GITHUB_REF = 'refs/heads/dev'
    Assert-Rejected { & (Join-Path $PSScriptRoot 'publish.ps1') -Repository 'synthetic/Trove' -Version $version -Revision $Revision } 'publicación desde dev'
}
finally { $env:GITHUB_ACTIONS = $oldActions; $env:GITHUB_EVENT_NAME = $oldEvent; $env:GITHUB_REF = $oldRef }
