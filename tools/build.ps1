[CmdletBinding()]
param(
    [ValidatePattern('^[0-9A-Za-z][0-9A-Za-z._+-]*$')][string]$Version = 'development',
    [ValidatePattern('^([0-9a-fA-F]{40}|unknown|local)$')][string]$Revision = 'unknown',
    [ValidateSet('test', 'habitual')][string]$Profile = 'test',
    [ValidatePattern('^[0-9A-Za-z][0-9A-Za-z._+-]*$')][string]$BuildID = 'local',
    [string]$OutputDirectory = 'artifacts\portable',
    [switch]$SkipFrontend
)
. (Join-Path $PSScriptRoot 'common.ps1')
$root = Split-Path $PSScriptRoot -Parent
Set-ProjectTemp $root
$out = Get-ArtifactDirectory $root $OutputDirectory
Push-Location $root
try {
    if (-not $SkipFrontend) {
        Invoke-Checked 'wails3' @('generate', 'bindings', '-clean=true', '-ts', '-d', 'frontend/bindings')
        Invoke-Checked 'npm' @('--prefix', 'frontend', 'run', 'build')
    }
    if (Test-Path -LiteralPath $out) { Remove-Item -LiteralPath $out -Recurse -Force }
    [System.IO.Directory]::CreateDirectory($out) | Out-Null
    $exe = Join-Path $out 'Trove.exe'
    $env:CGO_ENABLED = '0'
    $flags = "-H windowsgui -X main.defaultProfile=$Profile -X main.version=$Version -X main.revision=$Revision -X main.buildID=$BuildID"
    Invoke-Checked 'go' @('build', '-tags', 'production', '-trimpath', '-buildvcs=false', '-ldflags', $flags, '-o', $exe, '.')
    Copy-Item -LiteralPath (Join-Path $root 'LICENSE') -Destination (Join-Path $out 'LICENSE')
    $identity = [ordered]@{ version = $Version; revision = $Revision; profile = $Profile; platform = 'windows-x64'; buildID = $BuildID }
    $identity | ConvertTo-Json | Set-Content -LiteralPath (Join-Path $out 'build-info.json') -Encoding UTF8
    $instructions = @(
        "Trove $Version - Windows x64 portable",
        "Revision: $Revision",
        "Perfil predeterminado: $Profile",
        '',
        'Extraer toda esta carpeta en una ubicación con permiso de escritura y abrir Trove.exe.',
        'Requisito: Microsoft WebView2 Runtime. No requiere Go, Node ni un servidor SQLite.',
        "Datos: data\$Profile junto al ejecutable. No contiene bibliotecas personales.",
        'La configuración de respaldo se realiza desde Respaldo y restauración.',
        'Cerrar otras instancias del mismo perfil antes de restaurar o cambiar ejecutable.',
        'El paquete no está firmado digitalmente.',
        'Documentación y código: https://github.com/Santiparra/Trove'
    )
    $instructions | Set-Content -LiteralPath (Join-Path $out 'LEEME.txt') -Encoding UTF8
    $checksums = foreach ($name in @('Trove.exe', 'LICENSE', 'LEEME.txt', 'build-info.json')) {
        $hash = (Get-FileHash -LiteralPath (Join-Path $out $name) -Algorithm SHA256).Hash.ToLowerInvariant()
        "$hash  $name"
    }
    $checksums | Set-Content -LiteralPath (Join-Path $out 'SHA256SUMS.txt') -Encoding ASCII
    Test-ExecutableIdentity $exe (Join-Path $root '.local\portable-identity') $Version $Revision $Profile
}
finally { Pop-Location }
