$ErrorActionPreference = "Stop"

Write-Host "=========================================================="
Write-Host " Offline Queue & Peer ID Stability Test"
Write-Host "=========================================================="

# Cleanup any previous runs
Write-Host "`n[1] Cleaning up old state..."
$ErrorActionPreference = "Continue"
docker-compose down -v --remove-orphans 2>&1 | Out-Null
$ErrorActionPreference = "Stop"
$ErrorActionPreference = "Continue"
cmd /c rmdir /s /q "C:\tmp\offline-test" 2>$null
cmd /c rmdir /s /q "C:\tmp\offline-verify" 2>$null
cmd /c rmdir /s /q "C:\tmp\p2p-with-commits" 2>$null
cmd /c rmdir /s /q "C:\tmp\smoke-test-p2p" 2>$null
Remove-Item -Recurse -Force "data" -ErrorAction SilentlyContinue
$ErrorActionPreference = "Stop"

Write-Host "`n[2] Starting Host Node (Peer A)..."
$ErrorActionPreference = "Continue"
docker-compose up -d --build
$ErrorActionPreference = "Stop"

Write-Host "Waiting for Peer A to initialize..."
Start-Sleep -Seconds 15

# Get Peer ID on first run
$ErrorActionPreference = "Continue"
$peerA_ID_1 = docker logs p2p-git-libp2p 2>&1 | Select-String "libp2p host started:" | ForEach-Object { $_.Line.Split()[-1] }
$ErrorActionPreference = "Stop"
Write-Host "Peer A ID (Run 1): $peerA_ID_1"

if (-not $peerA_ID_1) {
    Write-Host "FAILED: Could not extract Peer A ID."
    exit 1
}

# Verify Peer ID Stability
Write-Host "`n[3] Verifying Peer ID Stability..."
$ErrorActionPreference = "Continue"
docker-compose restart libp2p-node
Start-Sleep -Seconds 5
$peerA_ID_2 = docker logs p2p-git-libp2p 2>&1 | Select-String "libp2p host started:" | Select-Object -Last 1 | ForEach-Object { $_.Line.Split()[-1] }
$ErrorActionPreference = "Stop"
Write-Host "Peer A ID (Run 2): $peerA_ID_2"

if ($peerA_ID_1 -ne $peerA_ID_2) {
    Write-Host "FAILED: Peer ID changed across restarts! Identity persistence is broken."
    exit 1
} else {
    Write-Host "SUCCESS: Peer ID is stable."
}

# Create a repo on Peer A
Write-Host "`n[4] Creating 'smoke-test-p2p' repo on Peer A..."
$headers = @{ "Authorization" = "ssh-ed25519 test-peer" }
$createResponse = Invoke-RestMethod -Uri "http://localhost:3000/api/v1/repos" -Method Post -Headers $headers -ContentType "application/json" -Body '{"name": "smoke-test-p2p", "description": "Offline Queue Test"}'
Write-Host "Repository created."

# Push a real commit to Peer A directly
Write-Host "`n[5] Pushing a real commit to Peer A locally..."
New-Item -ItemType Directory -Path "/tmp/smoke-test-p2p" -Force | Out-Null
Set-Location "/tmp/smoke-test-p2p"
git init
"A test commit" | Out-File "README.md"
git add .
git config user.email "test@local"
git config user.name "Test"
git commit -m "test commit"
git push http://localhost:3000/smoke-test-p2p master
Set-Location -Path $PSScriptRoot

# Start Peer B (Client)
Write-Host "`n[6] Starting Client Node (Peer B)..."
$env:P2P_LISTEN_PORT="4002"
$env:PROXY_PORT="4003"
$peerBContainer = "p2p-client-test"
$ErrorActionPreference = "Continue"
docker rm -f $peerBContainer 2>&1 | Out-Null
$ErrorActionPreference = "Stop"

$peerBProcess = Start-Process -FilePath "docker" -ArgumentList "run --name $peerBContainer -p 4003:4003 -p 4002:4002 -e P2P_LISTEN_PORT=$env:P2P_LISTEN_PORT -e PROXY_PORT=$env:PROXY_PORT -v gitproject_git-socket:/run/git -v gitproject_p2p-socket:/run/p2p gitproject-libp2p-node" -WindowStyle Hidden -PassThru

Write-Host "Waiting for Peer B to initialize..."
Start-Sleep -Seconds 5

# Clone with commits
Write-Host "`n[7] Cloning from Peer A via Peer B's local HTTP proxy (with commits)..."
$ErrorActionPreference = "Continue"
cmd /c rmdir /s /q "C:\tmp\p2p-with-commits" 2>$null
$ErrorActionPreference = "Stop"
git clone http://127.0.0.1:4003/p2p/$peerA_ID_1/smoke-test-p2p /tmp/p2p-with-commits

if (-not (Test-Path "/tmp/p2p-with-commits/README.md")) {
    Write-Host "FAILED: Cloned repo did not contain commits!"
    exit 1
} else {
    Write-Host "SUCCESS: Repository cloned with commits!"
}

# Offline Queue Test
Write-Host "`n[8] Testing Offline Queue..."
Write-Host "Stopping Peer A..."
$ErrorActionPreference = "Continue"
docker-compose stop
$ErrorActionPreference = "Stop"

Write-Host "Making an offline commit on Peer B's clone..."
Set-Location "/tmp/p2p-with-commits"
"offline commit" | Out-File "offline.txt"
git add .
git commit -m "made while host was offline"

Write-Host "Attempting push while Peer A is offline..."
git push origin master
# The proxy should catch this and queue it.
Set-Location -Path $PSScriptRoot

Write-Host "Bringing Peer A back online..."
$ErrorActionPreference = "Continue"
docker-compose start
$ErrorActionPreference = "Stop"
Start-Sleep -Seconds 15

Write-Host "Waiting 30 seconds for auto-flush queue..."
Start-Sleep -Seconds 30

Write-Host "Verifying offline commit synced to Peer A..."
$ErrorActionPreference = "Continue"
cmd /c rmdir /s /q "C:\tmp\offline-verify" 2>$null
$ErrorActionPreference = "Stop"
git clone http://localhost:3000/smoke-test-p2p /tmp/offline-verify

if (-not (Test-Path "/tmp/offline-verify/offline.txt")) {
    Write-Host "FAILED: Offline commit was not synced!"
    exit 1
} else {
    Write-Host "SUCCESS: Offline queue synced successfully!"
}

# Cleanup
Write-Host "`n[9] Cleaning up..."
$ErrorActionPreference = "Continue"
docker rm -f $peerBContainer 2>&1 | Out-Null
docker-compose down -v 2>&1 | Out-Null
$ErrorActionPreference = "Stop"

Write-Host "`nSUCCESS: All P2P Verification Tests Passed!"
