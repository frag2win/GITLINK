# test-phase3.ps1
# End-to-End Test for Phase 3: PostgreSQL, Branch Protection, and Pull Requests

$ErrorActionPreference = "Stop"

Write-Host "=============================================" -ForegroundColor Cyan
Write-Host "Phase 3: Self-Hostable Platform Verification" -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host ""

$API_URL = "http://localhost:3000"

# Wait for API server to be healthy
Write-Host "Waiting for API server at $API_URL..." -ForegroundColor Yellow
$retries = 0
while ($retries -lt 30) {
    try {
        $response = Invoke-WebRequest "$API_URL/" -UseBasicParsing -ErrorAction SilentlyContinue
        if ($response.StatusCode -eq 200) {
            Write-Host "API server is up." -ForegroundColor Green
            break
        }
    } catch {
        # ignore
    }
    Start-Sleep -Seconds 1
    $retries++
}

if ($retries -eq 30) {
    Write-Host "API server did not start in time." -ForegroundColor Red
    exit 1
}

# 1. Create a repository
Write-Host "Creating repository 'phase3-test'..." -ForegroundColor Yellow
$repoJson = @{ name = "phase3-test" } | ConvertTo-Json
$repoResponse = Invoke-RestMethod -Uri "$API_URL/api/v1/repos" -Method Post -Body $repoJson -ContentType "application/json"
$repoId = $repoResponse.id
Write-Host "Repository created with ID $repoId." -ForegroundColor Green

# 2. Protect the main branch
Write-Host "Protecting the 'main' branch..." -ForegroundColor Yellow
$protectJson = @{ requirePR = $true } | ConvertTo-Json
Invoke-RestMethod -Uri "$API_URL/api/v1/repos/$repoId/branches/main/protect" -Method Post -Body $protectJson -ContentType "application/json"
Write-Host "Branch 'main' is now protected." -ForegroundColor Green

# 3. Clone and try to push directly to main (should fail)
Write-Host "Testing branch protection..." -ForegroundColor Yellow
if (Test-Path "phase3-test") { Remove-Item -Recurse -Force "phase3-test" }
git clone "$API_URL/phase3-test.git" phase3-test
Set-Location phase3-test
New-Item -ItemType File -Name "test.txt" -Value "Hello Phase 3"
git add test.txt
git commit -m "Direct push to main"

try {
    # This should fail because of branch protection
    git push origin main 2>&1
    Write-Host "ERROR: Push to main succeeded, but it should have failed!" -ForegroundColor Red
    exit 1
} catch {
    Write-Host "Push to main rejected as expected (Branch Protection working)." -ForegroundColor Green
}

# 4. Push to a feature branch instead (should succeed)
Write-Host "Pushing to a feature branch..." -ForegroundColor Yellow
git checkout -b feature/test
git push origin feature/test
Write-Host "Push to feature branch succeeded." -ForegroundColor Green
Set-Location ..

# 5. Create a Pull Request
Write-Host "Creating Pull Request..." -ForegroundColor Yellow
$prJson = @{
    title = "Add test.txt"
    description = "Testing PR creation"
    baseBranch = "main"
    headBranch = "feature/test"
} | ConvertTo-Json

$prResponse = Invoke-RestMethod -Uri "$API_URL/api/v1/repos/$repoId/pulls" -Method Post -Body $prJson -ContentType "application/json"
$prId = $prResponse.id
Write-Host "Pull Request created with ID $prId." -ForegroundColor Green

# 6. Merge the Pull Request
Write-Host "Merging Pull Request..." -ForegroundColor Yellow
$mergeResponse = Invoke-RestMethod -Uri "$API_URL/api/v1/repos/$repoId/pulls/$prId/merge" -Method Post
Write-Host "Merge Response: $($mergeResponse.status), Hash: $($mergeResponse.hash)" -ForegroundColor Green
Write-Host "Pull Request merged successfully." -ForegroundColor Green

# Cleanup
if (Test-Path "phase3-test") { Remove-Item -Recurse -Force "phase3-test" }

Write-Host "=============================================" -ForegroundColor Cyan
Write-Host "Phase 3 verification complete! All tests passed." -ForegroundColor Green
Write-Host "=============================================" -ForegroundColor Cyan
