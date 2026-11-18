[CmdletBinding()]
param(
    [ValidatePattern('^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$')][string]$Repository,
    [ValidatePattern('^\d+\.\d+\.\d+$')][string]$Version,
    [ValidatePattern('^[0-9a-fA-F]{40}$')][string]$Revision
)
. (Join-Path $PSScriptRoot 'common.ps1')
if ($env:GITHUB_ACTIONS -ne 'true' -or $env:GITHUB_EVENT_NAME -ne 'push' -or $env:GITHUB_REF -ne 'refs/heads/main') { throw 'Publicación permitida solo en GitHub Actions por push a main.' }
if ($Repository -ne $env:GITHUB_REPOSITORY -or $Revision -ne $env:GITHUB_SHA) { throw 'Repositorio o revisión no corresponden a la ejecución.' }
$root = Split-Path $PSScriptRoot -Parent
$notes = Get-ReleaseNotesPath $root
$tag = "v$Version"
$zip = Join-Path $root "artifacts\release\Trove-$Version-windows-x64.zip"
$sum = Join-Path $root 'artifacts\release\SHA256SUMS.txt'
if (-not (Test-Path -LiteralPath $zip) -or -not (Test-Path -LiteralPath $sum)) { throw 'Faltan los archivos de publicación.' }

function Read-GHJson {
    param([string[]]$Arguments)
    $data = & gh @Arguments
    if ($LASTEXITCODE -ne 0) { throw 'Falló una consulta de GitHub.' }
    return (($data -join "`n") | ConvertFrom-Json)
}
function Read-OptionalGHJson {
    param([string[]]$Arguments)
    $start = New-Object System.Diagnostics.ProcessStartInfo
    $start.FileName = (Get-Command gh).Source
    $start.Arguments = $Arguments -join ' '
    $start.UseShellExecute = $false
    $start.CreateNoWindow = $true
    $start.RedirectStandardOutput = $true
    $start.RedirectStandardError = $true
    $process = New-Object System.Diagnostics.Process
    $process.StartInfo = $start
    try {
        [void]$process.Start()
        $stdout = $process.StandardOutput.ReadToEndAsync()
        $stderr = $process.StandardError.ReadToEndAsync()
        if (-not $process.WaitForExit(30000)) { $process.Kill(); throw 'GitHub no respondió dentro del plazo.' }
        $output = $stdout.Result
        $failure = $stderr.Result
        if ($process.ExitCode -eq 0) { return ($output | ConvertFrom-Json) }
        if ($failure.Contains('HTTP 404') -or $failure.Contains('release not found')) { return $null }
        throw "Falló la consulta de GitHub: $failure"
    }
    finally { $process.Dispose() }
}
function Assert-TagRevision {
    $reference = Read-GHJson @('api', "repos/$Repository/git/ref/tags/$tag")
    $object = $reference.object
    for ($i = 0; $object.type -eq 'tag' -and $i -lt 5; $i++) {
        $annotated = Read-GHJson @('api', "repos/$Repository/git/tags/$($object.sha)")
        $object = $annotated.object
    }
    if ($object.type -ne 'commit' -or $object.sha -ne $Revision) { throw 'El tag ya identifica otra revisión; no se modificará.' }
}

$existing = Read-OptionalGHJson @('release', 'view', $tag, '--repo', $Repository, '--json', 'tagName,isDraft,targetCommitish,assets,url')
if ($null -ne $existing -and -not $existing.isDraft) {
    Assert-TagRevision
    $names = @($existing.assets | ForEach-Object { $_.name })
    if ((Split-Path $zip -Leaf) -notin $names -or 'SHA256SUMS.txt' -notin $names) { throw 'La release publicada está incompleta; requiere revisión, no se sobrescribe.' }
    $existing.url
    return
}
if ($null -eq $existing) {
    # A pre-existing tag must not be silently reused for another commit.
    $reference = Read-OptionalGHJson @('api', "repos/$Repository/git/ref/tags/$tag")
    if ($null -eq $reference) { Invoke-Checked 'gh' @('api', '--method', 'POST', "repos/$Repository/git/refs", '-f', "ref=refs/tags/$tag", '-f', "sha=$Revision") }
    Assert-TagRevision
    Invoke-Checked 'gh' @('release', 'create', $tag, $zip, $sum, '--repo', $Repository, '--target', $Revision, '--verify-tag', '--title', "Trove $tag", '--notes-file', $notes, '--draft')
}
else {
    Assert-TagRevision
    Invoke-Checked 'gh' @('release', 'upload', $tag, $zip, $sum, '--repo', $Repository, '--clobber')
}
Assert-TagRevision
Invoke-Checked 'gh' @('release', 'edit', $tag, '--repo', $Repository, '--notes-file', $notes, '--draft=false', '--latest')
Assert-TagRevision
Invoke-Checked 'gh' @('release', 'view', $tag, '--repo', $Repository, '--json', 'url', '--jq', '.url')
