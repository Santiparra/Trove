[CmdletBinding()]
param(
    [string]$PortableDirectory = 'artifacts\portable',
    [string]$OutputDirectory = 'artifacts\release',
    [ValidatePattern('^\d+\.\d+\.\d+$')][string]$Version,
    [ValidatePattern('^[0-9a-fA-F]{40}$')][string]$Revision
)
. (Join-Path $PSScriptRoot 'common.ps1')
$root = Split-Path $PSScriptRoot -Parent
$source = Get-ArtifactDirectory $root $PortableDirectory
$out = Get-ArtifactDirectory $root $OutputDirectory
$meta = Get-Content -LiteralPath (Join-Path $source 'build-info.json') -Raw | ConvertFrom-Json
if ($meta.version -ne $Version -or $meta.revision -ne $Revision -or $meta.profile -ne 'habitual' -or $meta.platform -ne 'windows-x64') { throw 'El artefacto no corresponde a la versión/commit/perfil de publicación.' }
$expected = @('Trove.exe', 'LICENSE', 'LEEME.txt', 'build-info.json', 'SHA256SUMS.txt')
$files = @(Get-ChildItem -LiteralPath $source -Force)
if ($files.Count -ne $expected.Count -or @($files | Where-Object { $_.PSIsContainer -or $_.Name -notin $expected }).Count -ne 0) { throw 'El artefacto contiene archivos inesperados.' }
$lines = @(Get-Content -LiteralPath (Join-Path $source 'SHA256SUMS.txt'))
if ($lines.Count -ne 4) { throw 'El manifiesto debe identificar los cuatro archivos del paquete.' }
$seen = @{}
foreach ($line in $lines) {
    if ($line -notmatch '^([a-f0-9]{64})  (Trove\.exe|LICENSE|LEEME\.txt|build-info\.json)$') { throw 'Manifiesto inválido.' }
    if ($seen.ContainsKey($Matches[2])) { throw 'Archivo duplicado en el manifiesto.' }
    $seen[$Matches[2]] = $true
    if ((Get-FileHash -LiteralPath (Join-Path $source $Matches[2]) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $Matches[1]) { throw "Checksum incorrecto: $($Matches[2])" }
}
if (Test-Path -LiteralPath $out) { Remove-Item -LiteralPath $out -Force -Recurse }
[System.IO.Directory]::CreateDirectory($out) | Out-Null
$zip = Join-Path $out "Trove-$Version-windows-x64.zip"
Compress-Archive -LiteralPath @($expected | ForEach-Object { Join-Path $source $_ }) -DestinationPath $zip -CompressionLevel Optimal
$sum = (Get-FileHash -LiteralPath $zip -Algorithm SHA256).Hash.ToLowerInvariant()
"$sum  $(Split-Path $zip -Leaf)" | Set-Content -LiteralPath (Join-Path $out 'SHA256SUMS.txt') -Encoding ASCII
