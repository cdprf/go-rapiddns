$ErrorActionPreference = 'Stop'
$buildDir = "build"
if (-not (Test-Path $buildDir)) { New-Item -ItemType Directory -Path $buildDir | Out-Null }
$targets = @(
    @{ OS = 'windows'; ARCH = 'amd64' },
    @{ OS = 'windows'; ARCH = '386' },
    @{ OS = 'linux'; ARCH = 'amd64' },
    @{ OS = 'linux'; ARCH = '386' },
    @{ OS = 'linux'; ARCH = 'arm64' },
    @{ OS = 'linux'; ARCH = 'arm' },
    @{ OS = 'darwin'; ARCH = 'amd64' },
    @{ OS = 'darwin'; ARCH = 'arm64' }
)
foreach ($t in $targets) {
    $out = "$buildDir/rapiddnsquery-$($t.OS)-$($t.ARCH)"
    if ($t.OS -eq 'windows') { $out += '.exe' }
    Write-Host "Building $out ..."
    $env:GOOS = $t.OS
    $env:GOARCH = $t.ARCH
    go build -o $out main.go
    Write-Host "OK: $out"
}
