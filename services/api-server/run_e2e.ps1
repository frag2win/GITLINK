$ErrorActionPreference = "Stop"

# Setup environment for Rust
$env:PATH += ";C:\ProgramData\mingw64\mingw64\bin"
$env:CARGO_TARGET_DIR = "C:\Users\sunanda.AMFIIND\cargo-target"
$env:RUSTFLAGS = "-l advapi32 -l crypt32 -l user32 -l ole32 -l ws2_32 -l bcrypt"

Write-Host "Building git-server..."
Push-Location "..\git-server"
cargo build
$gitServerExe = "C:\Users\sunanda.AMFIIND\cargo-target\debug\git-server.exe"
Pop-Location

Write-Host "Building api-server..."
go build -o api-server.exe ./cmd/server

Write-Host "Starting git-server..."
$env:GIT_SERVER_IPC_NETWORK = "tcp"
$env:GIT_SERVER_IPC_ADDRESS = "127.0.0.1:9099"
$env:GIT_SERVER_REPOS_PATH = "C:\temp\git-repos-test"
$gitServerProcess = Start-Process -FilePath $gitServerExe -PassThru -WindowStyle Hidden

Write-Host "Starting api-server..."
$env:API_PORT = "8080"
$env:API_GIT_IPC_NETWORK = "tcp"
$env:API_GIT_IPC_ADDRESS = "127.0.0.1:9099"
$env:API_DB_URL = "postgres://gitadmin:gitpassword@localhost:5432/gitplatform?sslmode=disable"
$env:JWT_SECRET = "supersecret"
$env:API_DEV_MODE = "true"
$apiServerProcess = Start-Process -FilePath ".\api-server.exe" -RedirectStandardOutput "api-server.log" -RedirectStandardError "api-server.err" -PassThru -WindowStyle Hidden

Start-Sleep -Seconds 5

Write-Host "Running tests..."
try {
    go test -v -count=1 ./test/e2e/...
} finally {
    Write-Host "Cleaning up servers..."
    Stop-Process -Id $gitServerProcess.Id -Force -ErrorAction SilentlyContinue
    Stop-Process -Id $apiServerProcess.Id -Force -ErrorAction SilentlyContinue
    Remove-Item -Recurse -Force "C:\temp\git-repos-test" -ErrorAction SilentlyContinue
}
