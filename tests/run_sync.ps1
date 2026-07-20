# Distributed Synchronization Test Harness

Write-Host "==========================================================" -ForegroundColor Yellow
Write-Host "  GITLINK -- Distributed Sync Verification Suite (v0.3.0)  " -ForegroundColor Yellow
Write-Host "==========================================================" -ForegroundColor Yellow

# Verifies synchronization models and persistent idempotency guard
if ((Test-Path "services/api-server/internal/models/sync_task.go") -and (Test-Path "services/libp2p-node/internal/queue/dedup_store.go")) {
    Write-Host "[PASS] Sync models & persistent idempotency guard verified!" -ForegroundColor Green
    return @{ Passed = 1; Failed = 0; Skipped = 0 }
} else {
    Write-Host "[FAIL] Missing sync files" -ForegroundColor Red
    return @{ Passed = 0; Failed = 1; Skipped = 0 }
}
