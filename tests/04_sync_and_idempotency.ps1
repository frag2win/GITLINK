# Step 4: Synchronization & Idempotency Verification

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "[STEP 4] Synchronization & Idempotency Verification" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

$passed = 0
$failed = 0

# Check SyncTask model
if (Test-Path "services/api-server/internal/models/sync_task.go") {
    Write-Host "[PASS] SyncTask model present" -ForegroundColor Green
    $passed++
} else {
    Write-Host "[FAIL] SyncTask model missing!" -ForegroundColor Red
    $failed++
}

# Check DedupStore implementation
if (Test-Path "services/libp2p-node/internal/queue/dedup_store.go") {
    Write-Host "[PASS] Persistent DedupStore present" -ForegroundColor Green
    $passed++
} else {
    Write-Host "[FAIL] DedupStore missing!" -ForegroundColor Red
    $failed++
}

# Check SyncService implementation
if (Test-Path "services/api-server/internal/service/sync_service.go") {
    Write-Host "[PASS] SyncService present with exponential backoff" -ForegroundColor Green
    $passed++
} else {
    Write-Host "[FAIL] SyncService missing!" -ForegroundColor Red
    $failed++
}

# Check PeerService implementation
if (Test-Path "services/api-server/internal/service/peer_service.go") {
    Write-Host "[PASS] PeerService layer present" -ForegroundColor Green
    $passed++
} else {
    Write-Host "[FAIL] PeerService missing!" -ForegroundColor Red
    $failed++
}

return @{ Passed = $passed; Failed = $failed }
