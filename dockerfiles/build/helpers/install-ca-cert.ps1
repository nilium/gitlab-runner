param(
    [string]$CAFile
)

if (Test-Path -LiteralPath $CAFile -PathType Leaf) {
    Write-Output "Adding CA certificate..."
    Import-Certificate -FilePath $CAFile -CertStoreLocation cert:\LocalMachine\CA
}
