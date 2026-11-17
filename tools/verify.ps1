[CmdletBinding()]
param([switch]$CheckBindings)
. (Join-Path $PSScriptRoot 'common.ps1')
$root = Split-Path $PSScriptRoot -Parent
Set-ProjectTemp $root
$env:CGO_ENABLED = '0'
Push-Location $root
try {
    foreach ($script in Get-ChildItem -LiteralPath (Join-Path $root 'tools') -Filter '*.ps1') {
        $tokens = $null
        $errors = $null
        [void][System.Management.Automation.Language.Parser]::ParseFile($script.FullName, [ref]$tokens, [ref]$errors)
        if ($errors.Count -ne 0) { throw "Sintaxis inválida en $($script.Name): $($errors -join '; ')" }
    }
    $documents = @((Get-Item -LiteralPath (Join-Path $root 'README.md'))) + @(Get-ChildItem -LiteralPath (Join-Path $root 'docs') -Recurse -Filter '*.md')
    foreach ($document in $documents) {
        $text = Get-Content -LiteralPath $document.FullName -Raw
        foreach ($link in [regex]::Matches($text, '\]\(([^)#]+\.md)(?:#[^)]*)?\)')) {
            $target = $link.Groups[1].Value
            if ([Uri]::IsWellFormedUriString($target, [UriKind]::Absolute)) { continue }
            if ($target -match '(^|/)\.local-docs/|(^|/)AGENTS\.md$') { throw "Enlace personal en documentación pública: $target" }
            if (-not (Test-Path -LiteralPath (Join-Path $document.DirectoryName $target))) { throw "Enlace inexistente en $($document.Name): $target" }
        }
    }
    Invoke-Checked 'go' @('test', './...')
    Invoke-Checked 'npm' @('--prefix', 'frontend', 'test')
    if ($CheckBindings) {
        Invoke-Checked 'wails3' @('generate', 'bindings', '-clean=true', '-ts', '-d', 'frontend/bindings')
        $changes = & git status --porcelain -- frontend/bindings
        if ($LASTEXITCODE -ne 0) { throw 'No se pudo comprobar el contrato generado.' }
        if ($changes) {
            $changes
            & git diff -- frontend/bindings
            & git ls-files --eol frontend/bindings
            throw 'Los bindings no corresponden al backend: regenerar y versionar frontend/bindings.'
        }
    }
}
finally { Pop-Location }
