@echo off

powershell.exe -File .\install-ca-cert.ps1 -CAFile C:\etc\gitlab-runner\certs\ca.crt

%*
