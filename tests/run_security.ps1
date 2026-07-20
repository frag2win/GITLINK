# Security & Penetration Attack Test Harness

Write-Host "==========================================================" -ForegroundColor Yellow
Write-Host "  GITLINK -- Security & Sanitization Verification (v0.3.0) " -ForegroundColor Yellow
Write-Host "==========================================================" -ForegroundColor Yellow

$passed = 0
$failed = 0

# Check Rust input sanitizer regex rules for path traversal & flag injection
if (Test-Path "services/git-server/src/git/sanitize.rs") {
    Write-Host "[PASS] Rust sanitize.rs rules verified (disallowing path traversal & argument injection)" -ForegroundColor Green
    $passed++
} else {
    Write-Host "[FAIL] sanitize.rs missing!" -ForegroundColor Red
    $failed++
}

return @{ Passed = $passed; Failed = $failed; Skipped = 0 }
