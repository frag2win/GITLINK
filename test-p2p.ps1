$ErrorActionPreference = "Stop"

Write-Host "=========================================================="
Write-Host "   End-to-End P2P Smoke Test: Git Clone Over libp2p"
Write-Host "=========================================================="

Write-Host "`n[1] Starting Host Node (Peer A) via docker-compose..."
docker-compose down -v 
if (Test-Path "data") { Remove-Item -Recurse -Force "data" } 
docker-compose up -d --build

Write-Host "Waiting for Peer A to initialize..."
Start-Sleep -Seconds 20

# Create a repository on Peer A using its API Server
Write-Host "`n[2] Creating 'smoke-test-p2p' repository on Peer A..."
$headers = @{ "Authorization" = "ssh-ed25519 test-peer" }

try {
    Invoke-RestMethod -Method Post -Uri "http://localhost:3000/api/v1/repos" -Headers $headers -Body '{"name":"smoke-test-p2p", "description":"P2P Test Repository"}' -ContentType "application/json" | Out-Null
    Write-Host "Repository created."
} catch {
    Write-Host "Failed to create repository: $_"
    exit 1
}

# We need the Peer ID of Peer A. We can grep docker logs.
Write-Host "`n[3] Extracting Peer A's Peer ID..."
$ErrorActionPreference = "Continue"
$logs = docker logs p2p-git-libp2p 2>&1
$ErrorActionPreference = "Stop"
$peerIdLine = $logs | Select-String -Pattern "libp2p host started:" | Select-Object -First 1
if (-not $peerIdLine) {
    Write-Host "Failed to find Peer ID in logs!"
    exit 1
}
$peerId = ($peerIdLine -split "started: ")[1].Trim()
Write-Host "Peer A ID: $peerId"

Write-Host "`n[4] Starting Client Node (Peer B) locally..."
# Peer B needs a different libp2p port and proxy port
$env:P2P_LISTEN_PORT="4002"
$env:PROXY_PORT="4003"
$env:PEER_KEY_PATH="peerB.key"
$env:API_SOCKET_PATH="C:\Temp\peerB_dummy.sock"

# Start peer B using Docker instead of compiling locally
Write-Host "Starting Peer B using docker..."
$peerBContainer = "p2p-client-test"
$ErrorActionPreference = "Continue"
docker rm -f $peerBContainer 2>&1 | Out-Null
$ErrorActionPreference = "Stop"
$peerBProcess = Start-Process -FilePath "docker" -ArgumentList "run --name $peerBContainer -p 4003:4003 -p 4002:4002 -e P2P_LISTEN_PORT=$env:P2P_LISTEN_PORT -e PROXY_PORT=$env:PROXY_PORT -v gitproject_git-socket:/run/git -v gitproject_p2p-socket:/run/p2p gitproject-libp2p-node" -WindowStyle Hidden -PassThru

Write-Host "Waiting for Peer B to initialize..."
Start-Sleep -Seconds 10

Write-Host "`n[5] Cloning from Peer A via Peer B's local HTTP proxy..."
$tempDir = Join-Path $env:TEMP "p2p-test-clone"
if (Test-Path $tempDir) { Remove-Item -Recurse -Force $tempDir }
New-Item -ItemType Directory -Force -Path $tempDir | Out-Null

Set-Location $tempDir
# git clone http://localhost:4003/p2p/<peer-A-id>/smoke-test-p2p
git clone "http://127.0.0.1:4003/p2p/$peerId/smoke-test-p2p"

$exitCode = $LASTEXITCODE

Set-Location $PSScriptRoot

Write-Host "`n[6] Cleaning up..."
docker-compose down -v
$ErrorActionPreference = "Continue"
docker rm -f $peerBContainer 2>&1 | Out-Null
$ErrorActionPreference = "Stop"

if ($exitCode -eq 0) {
    Write-Host "`nSUCCESS: Successfully cloned a git repository peer-to-peer over libp2p streams!" -ForegroundColor Green
} else {
    Write-Host "`nFAILED: P2P Clone failed." -ForegroundColor Red
}
