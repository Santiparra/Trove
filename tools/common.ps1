Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Invoke-Checked {
    param([string]$Command, [string[]]$Arguments = @())
    & $Command @Arguments
    if ($LASTEXITCODE -ne 0) { throw "$Command terminó con código $LASTEXITCODE" }
}

function Set-ProjectTemp {
    param([string]$Root)
    $temp = Join-Path $Root '.local\automation-tmp'
    [System.IO.Directory]::CreateDirectory($temp) | Out-Null
    $env:TEMP = $temp
    $env:TMP = $temp
}

function Get-ReleaseNotesPath {
    param([string]$Root)
    $folder = Join-Path $Root 'docs/releases'
    $name = (Get-Content -LiteralPath (Join-Path $folder 'current.txt') -Raw -Encoding UTF8).Trim()
    if ($name -cnotmatch '^[a-z0-9][a-z0-9-]*\.md$' -or $name -ieq 'README.md') { throw 'La selección de novedades debe identificar un documento de corte dentro de docs/releases/.' }
    $path = Join-Path $folder $name
    if (-not (Test-Path -LiteralPath $path -PathType Leaf)) { throw 'No existe el documento de novedades seleccionado.' }
    $notes = Get-Content -LiteralPath $path -Raw -Encoding UTF8
    if ([string]::IsNullOrWhiteSpace($notes) -or $notes -notmatch '(?m)^#\s+\S') { throw 'El documento de novedades está vacío o no tiene título.' }
    return $path
}

function Get-ArtifactDirectory {
    param([string]$Root, [string]$RelativePath)
    $path = [System.IO.Path]::GetFullPath((Join-Path $Root $RelativePath))
    $allowed = [System.IO.Path]::GetFullPath((Join-Path $Root 'artifacts')) + [System.IO.Path]::DirectorySeparatorChar
    if (-not $path.StartsWith($allowed, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw 'La salida debe ser un subdirectorio de artifacts/ dentro del proyecto.'
    }
    return $path
}

function Test-ExecutableIdentity {
    param([string]$Executable, [string]$WorkingDirectory, [string]$Version, [string]$Revision, [string]$Profile)
    [System.IO.Directory]::CreateDirectory($WorkingDirectory) | Out-Null
    $start = New-Object System.Diagnostics.ProcessStartInfo
    $start.FileName = $Executable
    $start.Arguments = '--version'
    $start.WorkingDirectory = $WorkingDirectory
    $start.UseShellExecute = $false
    $start.CreateNoWindow = $true
    $start.RedirectStandardOutput = $true
    $start.RedirectStandardError = $true
    $process = New-Object System.Diagnostics.Process
    $process.StartInfo = $start
    try {
        if (-not $process.Start()) { throw 'No se pudo ejecutar la prueba de identidad.' }
        if (-not $process.WaitForExit(30000)) { $process.Kill(); throw 'La consulta de versión no terminó en 30 segundos.' }
        $output = $process.StandardOutput.ReadToEnd()
        $errors = $process.StandardError.ReadToEnd()
        if ($process.ExitCode -ne 0) { throw "El portable falló: $errors" }
        foreach ($expected in @("Trove $Version;", "revision=$Revision;", "profile=$Profile;", 'platform=windows/amd64;')) {
            if (-not $output.Contains($expected)) { throw "Identidad inesperada: $output (falta $expected)" }
        }
        if ((Test-Path -LiteralPath (Join-Path $WorkingDirectory 'data')) -or (Test-Path -LiteralPath (Join-Path (Split-Path $Executable -Parent) 'data'))) { throw '--version creó una biblioteca inesperadamente.' }
    }
    finally { $process.Dispose() }
}
