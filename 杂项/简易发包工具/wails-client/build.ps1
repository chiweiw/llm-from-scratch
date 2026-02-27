                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            $ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot
go version | Out-Null
if (-not $?) { Write-Host "Go not installed" ; exit 2 }
go mod tidy
if (-not (Test-Path "bin")) { New-Item -ItemType Directory -Path "bin" | Out-Null }
go build -tags "desktop,production" -ldflags '-H=windowsgui -s -w' -o "bin/wails-client.exe"
$out = Resolve-Path "bin/wails-client.exe"
Write-Host ("Build OK: " + $out)
