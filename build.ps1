$ErrorActionPreference = "Stop"

Write-Host "Testing..."
go test ./...

Write-Host "Preparing icon resource..."
$rsrc = Get-Command rsrc -ErrorAction SilentlyContinue
if (-not $rsrc) {
    go install github.com/akavel/rsrc@latest
    $goBin = Join-Path (go env GOPATH) "bin"
    $env:PATH = "$goBin;$env:PATH"
}

rsrc -ico assets/app-icon.ico -o rsrc_windows_amd64.syso

New-Item -ItemType Directory -Force -Path dist | Out-Null
Write-Host "Building Windows x64 GUI..."
go build -trimpath -ldflags="-s -w -H=windowsgui" -o dist/JSON-Google-Docs-Converter.exe .

Remove-Item rsrc_windows_amd64.syso -ErrorAction SilentlyContinue
Write-Host "Built dist/JSON-Google-Docs-Converter.exe"
