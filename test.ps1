$ErrorActionPreference = 'Continue'
Write-Host "=== STAGE 2: Health and Basic API ==="
curl.exe -s http://localhost:3000/health
Write-Host ""
curl.exe -s http://localhost:3000/api/v1/repos
Write-Host ""

Write-Host "=== STAGE 3: Create repo ==="
$out = curl.exe -s -X POST http://localhost:3000/api/v1/contributors -H "Content-Type: application/json" -d '{\"name\":\"test-user\",\"ssh_key\":\"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIObW2kKzB+m5o5dD7qQh1D3U7tM9y2D8fT0rBfN\"}'
Write-Host "Contributor created: $out"
$key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIObW2kKzB+m5o5dD7qQh1D3U7tM9y2D8fT0rBfN"
$repoOut = curl.exe -s -X POST http://localhost:3000/api/v1/repos -H "Authorization: $key" -H "Content-Type: application/json" -d '{\"name\":\"smoke-test\"}'
Write-Host "Repo created: $repoOut"

$ErrorActionPreference = 'Stop'
Write-Host "=== STAGE 4: Real git clone and push ==="
Remove-Item -Recurse -Force C:\tmp\smoke-test -ErrorAction SilentlyContinue
Remove-Item -Recurse -Force C:\tmp\smoke-test-verify -ErrorAction SilentlyContinue
if (!(Test-Path -Path C:\tmp)) { New-Item -ItemType Directory -Force -Path C:\tmp | Out-Null }

git.exe clone http://localhost:3000/smoke-test C:\tmp\smoke-test
Set-Location C:\tmp\smoke-test
git.exe config user.email "test@local"
git.exe config user.name "Test"
Set-Content -Path README.md -Value "# Smoke Test"
git.exe add README.md
git.exe commit -m "initial commit"
git.exe branch -M main
git.exe push origin main

Write-Host "Verify push..."
Set-Location C:\tmp
git.exe clone -b main http://localhost:3000/smoke-test C:\tmp\smoke-test-verify
Get-Content C:\tmp\smoke-test-verify\README.md

Write-Host "=== STAGE 5: Container isolation holds ==="
docker exec p2p-git-server ping -c 1 8.8.8.8 2>&1 | Out-String
docker exec p2p-git-server ls /repos | Out-String
docker exec p2p-git-server ls /repos/smoke-test | Out-String
curl.exe -s --connect-timeout 2 http://0.0.0.0:3000/health 2>&1

