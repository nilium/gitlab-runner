﻿$goVersion = "1.14.0"
$gitVersion = "2.23.0"
$srcFolder = "C:\GitLab-Runner"

[environment]::SetEnvironmentVariable("RUNNER_SRC", $srcFolder, "Machine")

Write-Host "Installing Chocolatey"
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))

Write-Host "Installing Go"
choco install golang -y --version $goVersion

Write-Host "Installing Git"
choco install git -y --version $gitVersion
