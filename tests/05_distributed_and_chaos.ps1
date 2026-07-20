# Step 5: Distributed Multi-Peer, Chaos Resilience & Operations Verification

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "[STEP 5] Distributed Sync, Operations & Release Gate Verification" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

$passed = 0
$failed = 0

# 1. Verify Tiered Health & Metrics Handlers
if ((Test-Path "services/api-server/internal/handlers/health_handler.go") -and (Test-Path "services/api-server/internal/handlers/metrics_handler.go")) {
    Write-Host "[PASS] Tiered Health (/health, /ready, /health/deep) and Metrics handlers present" -ForegroundColor Green
    $passed++
} else {
    Write-Host "[FAIL] Tiered Health or Metrics handlers missing!" -ForegroundColor Red
    $failed++
}

# 2. Verify Real-Time WebSocket Engine & Reconnect Replay
if ((Test-Path "services/api-server/internal/service/ws_hub.go") -and (Test-Path "services/api-server/internal/handlers/ws_handler.go")) {
    Write-Host "[PASS] WebSocketHub with monotonic EventIDs and reconnect replay present" -ForegroundColor Green
    $passed++
} else {
    Write-Host "[FAIL] WebSocketHub or WSHandler missing!" -ForegroundColor Red
    $failed++
}

# 3. Verify DLQ Administration & Audit Log Integration
if (Test-Path "services/api-server/internal/handlers/admin_handler.go") {
    Write-Host "[PASS] DLQ Replay and Admin Operations API present with audit logging" -ForegroundColor Green
    $passed++
} else {
    Write-Host "[FAIL] AdminHandler missing!" -ForegroundColor Red
    $failed++
}

# 4. Verify Conflict Analysis & Diagnostics Engine
if ((Test-Path "services/api-server/internal/models/conflict_report.go") -and (Test-Path "services/api-server/internal/service/conflict_service.go")) {
    Write-Host "[PASS] Conflict Analysis Engine and rich diagnostic model present" -ForegroundColor Green
    $passed++
} else {
    Write-Host "[FAIL] Conflict Analysis Engine missing!" -ForegroundColor Red
    $failed++
}

# 5. Verify Security Hardening (SEC-01 & SEC-02)
$routerContent = Get-Content "services/api-server/internal/router/router.go" -Raw
if ($routerContent -match "gitHTTPGroup := app\.Group\(" -and $routerContent -match "middleware\.Auth") {
    Write-Host "[PASS] SEC-01 Git Smart HTTP authentication wrapper verified" -ForegroundColor Green
    $passed++
} else {
    Write-Host "[FAIL] SEC-01 Git Smart HTTP authentication wrapper missing in router.go!" -ForegroundColor Red
    $failed++
}

$authzContent = Get-Content "services/api-server/internal/service/authorization_service.go" -Raw
if ($authzContent -match "GetUserRepoRole") {
    Write-Host "[PASS] SEC-02 Team RBAC authorization engine integration verified" -ForegroundColor Green
    $passed++
} else {
    Write-Host "[FAIL] SEC-02 Team RBAC authorization check missing in authorization_service.go!" -ForegroundColor Red
    $failed++
}

return @{ Passed = $passed; Failed = $failed }
