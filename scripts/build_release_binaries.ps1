$ErrorActionPreference = "Stop"

$distDir = "B:\workgit\brokenlens\dist"
if (-not (Test-Path $distDir)) {
    New-Item -ItemType Directory -Path $distDir | Out-Null
}

$targets = @(
    @{ OS = "linux"; ARCH = "amd64"; Name = "brokenlens-linux-amd64" },
    @{ OS = "linux"; ARCH = "arm64"; Name = "brokenlens-linux-arm64" },
    @{ OS = "darwin"; ARCH = "amd64"; Name = "brokenlens-darwin-amd64" },
    @{ OS = "darwin"; ARCH = "arm64"; Name = "brokenlens-darwin-arm64" },
    @{ OS = "windows"; ARCH = "amd64"; Name = "brokenlens-windows-amd64.exe" },
    @{ OS = "windows"; ARCH = "arm64"; Name = "brokenlens-windows-arm64.exe" }
)

Write-Host "Building release binaries..."
foreach ($t in $targets) {
    Write-Host "Compiling $($t.OS)/$($t.ARCH) -> $($t.Name)..."
    $env:CGO_ENABLED = "0"
    $env:GOOS = $t.OS
    $env:GOARCH = $t.ARCH
    $outPath = Join-Path $distDir $t.Name
    go build -ldflags="-s -w" -o $outPath ./cmd/brokenlens
}

Write-Host "Generating checksums..."
$checksumFile = Join-Path $distDir "checksums.txt"
Set-Content -Path $checksumFile -Value ""

Get-ChildItem -Path $distDir -File | Where-Object { $_.Name -ne "checksums.txt" } | ForEach-Object {
    $hash = (Get-FileHash -Algorithm SHA256 $_.FullName).Hash.ToLower()
    $line = "$hash  $($_.Name)"
    Add-Content -Path $checksumFile -Value $line
}

Write-Host "All binaries built successfully!"
Get-Content $checksumFile
