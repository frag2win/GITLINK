# Step 1: Prerequisites Check

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "[STEP 1] Verifying System Prerequisites" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

$results = @{
    "Go" = $false
    "Cargo" = $false
    "Git" = $false
}

# Check Go
try {
    $goVer = go version
    Write-Host "[PASS] Go installed: $goVer" -ForegroundColor Green
    $results["Go"] = $true
} catch {
    Write-Host "[FAIL] Go is not installed or not in PATH" -ForegroundColor Red
}

# Check Cargo / Rust
try {
    $cargoVer = cargo --version
    Write-Host "[PASS] Cargo/Rust installed: $cargoVer" -ForegroundColor Green
    $results["Cargo"] = $true
} catch {
    Write-Host "[FAIL] Cargo/Rust is not installed or not in PATH" -ForegroundColor Red
}

# Check Git
try {
    $gitVer = git --version
    Write-Host "[PASS] Git installed: $gitVer" -ForegroundColor Green
    $results["Git"] = $true
} catch {
    Write-Host "[FAIL] Git is not installed or not in PATH" -ForegroundColor Red
}

return $results
