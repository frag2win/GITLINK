# Step 3: Security & Sanitization Verification

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "[STEP 3] Security & Sanitization Verification" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

$passed = 0
$failed = 0

# Check Rust sanitize module presence
if (Test-Path "services/git-server/src/git/sanitize.rs") {
    Write-Host "[PASS] Rust input sanitizer module found (sanitize.rs)" -ForegroundColor Green
    $passed++
} else {
    Write-Host "[FAIL] Rust input sanitizer module missing!" -ForegroundColor Red
    $failed++
}

# Check Auth middleware presence
if (Test-Path "services/api-server/internal/middleware/auth.go") {
    Write-Host "[PASS] JWT Auth middleware found (auth.go)" -ForegroundColor Green
    $passed++
} else {
    Write-Host "[FAIL] JWT Auth middleware missing!" -ForegroundColor Red
    $failed++
}

# Check AuthorizationService presence
if (Test-Path "services/api-server/internal/service/authorization_service.go") {
    Write-Host "[PASS] AuthorizationService found (authorization_service.go)" -ForegroundColor Green
    $passed++
} else {
    Write-Host "[FAIL] AuthorizationService missing!" -ForegroundColor Red
    $failed++
}

return @{ Passed = $passed; Failed = $failed }
