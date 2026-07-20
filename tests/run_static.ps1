# Step 1: Static & Unit Verification Suite for GITLINK

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  GITLINK -- Static & Unit Verification Suite (v0.3.0)  " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

$startTime = Get-Date

$step1 = & ".\tests\01_prereqs.ps1"
$step2 = & ".\tests\02_build_and_unit.ps1"
$step3 = & ".\tests\03_security_and_sanitization.ps1"
$step4 = & ".\tests\04_sync_and_idempotency.ps1"

$totalPassed = $step2.Passed + $step3.Passed + $step4.Passed
$totalFailed = $step2.Failed + $step3.Failed + $step4.Failed
$totalTests = $totalPassed + $totalFailed
$duration = (Get-Date) - $startTime

Write-Host ""
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "            STATIC & UNIT VERIFICATION REPORT             " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "Total Verification Checks Executed : $totalTests"
Write-Host "Passed                             : $totalPassed" -ForegroundColor Green
Write-Host "Failed                             : $totalFailed" -ForegroundColor $(if($totalFailed -eq 0){"Green"}else{"Red"})
Write-Host "Duration                           : $($duration.TotalSeconds) seconds"
Write-Host "==========================================================" -ForegroundColor Cyan

return @{ Passed = $totalPassed; Failed = $totalFailed }
