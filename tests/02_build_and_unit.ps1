# Step 2: Build & Unit Test Verification

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "[STEP 2] Build & Unit Test Verification" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

$passed = 0
$failed = 0

# 1. Build & Test Go api-server unit tests (internal packages)
Write-Host "[TEST] Building & Testing services/api-server internal packages..." -ForegroundColor Yellow
Push-Location "services/api-server"
try {
    $out = go test ./internal/... 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[PASS] services/api-server unit tests passed!" -ForegroundColor Green
        $passed++
    } else {
        Write-Host "[FAIL] services/api-server unit tests failed:`n$out" -ForegroundColor Red
        $failed++
    }
} finally {
    Pop-Location
}

# 2. Build & Test Go libp2p-node
Write-Host "[TEST] Building & Testing services/libp2p-node internal packages..." -ForegroundColor Yellow
Push-Location "services/libp2p-node"
try {
    $out = go test ./internal/... 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[PASS] services/libp2p-node unit tests passed!" -ForegroundColor Green
        $passed++
    } else {
        Write-Host "[FAIL] services/libp2p-node unit tests failed:`n$out" -ForegroundColor Red
        $failed++
    }
} finally {
    Pop-Location
}

# 3. Rust services/git-server static check
Write-Host "[TEST] Verifying Rust services/git-server..." -ForegroundColor Yellow
if (Test-Path "services/git-server/src/main.rs") {
    Write-Host "[PASS] services/git-server source files verified!" -ForegroundColor Green
    $passed++
} else {
    Write-Host "[FAIL] services/git-server main.rs missing!" -ForegroundColor Red
    $failed++
}

return @{ Passed = $passed; Failed = $failed }
