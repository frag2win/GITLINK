# Orchestration Master Test Suite Runner for GITLINK

Write-Host "==========================================================" -ForegroundColor Magenta
Write-Host "  GITLINK -- Automated Master Verification Suite v1.0   " -ForegroundColor Magenta
Write-Host "==========================================================" -ForegroundColor Magenta

$startTime = Get-Date

# 1. Run Static & Unit Suite
$staticRes = & ".\tests\run_static.ps1"

# 2. Run Sync Suite
$syncRes = & ".\tests\run_sync.ps1"

# 3. Run Security Suite
$secRes = & ".\tests\run_security.ps1"

# 4. Run Integration Suite (Container check)
$integRes = & ".\tests\run_integration.ps1"

$totalPassed = $staticRes.Passed + $syncRes.Passed + $secRes.Passed + $integRes.Passed
$totalFailed = $staticRes.Failed + $syncRes.Failed + $secRes.Failed + $integRes.Failed
$totalSkipped = $integRes.Skipped
$totalTests = $totalPassed + $totalFailed + $totalSkipped

$duration = (Get-Date) - $startTime

Write-Host ""
Write-Host "==========================================================" -ForegroundColor Magenta
Write-Host "              MASTER VERIFICATION SUMMARY REPORT          " -ForegroundColor Magenta
Write-Host "==========================================================" -ForegroundColor Magenta
Write-Host "Total Verification Checks Executed : $totalTests"
Write-Host "Passed                             : $totalPassed" -ForegroundColor Green
Write-Host "Failed                             : $totalFailed" -ForegroundColor $(if($totalFailed -eq 0){"Green"}else{"Red"})
Write-Host "Skipped (Live Container Tests)    : $totalSkipped" -ForegroundColor Yellow
Write-Host "Duration                           : $($duration.TotalSeconds) seconds"
Write-Host "==========================================================" -ForegroundColor Magenta

if ($totalFailed -eq 0) {
    Write-Host ""
    Write-Host "VERDICT: [PASS] PRODUCTION CANDIDATE -- Static & Unit Verification Passed (v0.3.0)" -ForegroundColor Green
    Exit 0
} else {
    Write-Host ""
    Write-Host "VERDICT: [FAIL] Issues Detected!" -ForegroundColor Red
    Exit 1
}
