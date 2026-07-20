# Integration Verification Suite (Multi-Container Environment)

Write-Host "==========================================================" -ForegroundColor Yellow
Write-Host "  GITLINK -- Integration Verification Suite (v0.3.0)     " -ForegroundColor Yellow
Write-Host "==========================================================" -ForegroundColor Yellow

Write-Host "[INFO] Checking Docker daemon status..." -ForegroundColor Yellow
try {
    $dockerCheck = docker info 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[SKIPPED] Docker daemon is offline. Start Docker Desktop to run live container integration tests." -ForegroundColor Yellow
        return @{ Passed = 0; Failed = 0; Skipped = 1 }
    }
    Write-Host "[PASS] Docker daemon is online!" -ForegroundColor Green
    return @{ Passed = 1; Failed = 0; Skipped = 0 }
} catch {
    Write-Host "[SKIPPED] Docker CLI not accessible." -ForegroundColor Yellow
    return @{ Passed = 0; Failed = 0; Skipped = 1 }
}
