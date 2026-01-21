# Pre-commit hook for FitWatch (Windows PowerShell version)
# Ensures CI will pass before allowing commit

$ErrorActionPreference = "Stop"

Write-Host "🔍 Running pre-commit checks..." -ForegroundColor Cyan

# Change to fitwatch directory
try {
    $repoRoot = & git rev-parse --show-toplevel 2>&1
    if ($LASTEXITCODE -eq 0 -and (Test-Path "$repoRoot/fitwatch")) {
        Set-Location "$repoRoot/fitwatch"
    } elseif (Test-Path "./fitwatch") {
        Set-Location "./fitwatch"
    } else {
        # Already in fitwatch or running from scripts
        $scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
        if ($scriptDir -match "scripts$") {
            Set-Location (Split-Path -Parent $scriptDir)
        }
    }
} catch {
    # Not in a git repo, try to find fitwatch directory relative to script
    $scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
    $fitDir = Split-Path -Parent $scriptDir
    Set-Location $fitDir
}

# 1. Format check
Write-Host "📝 Checking formatting..." -ForegroundColor Yellow
$unformatted = & gofmt -l . 2>$null
if ($unformatted) {
    Write-Host "❌ The following files need formatting:" -ForegroundColor Red
    Write-Host $unformatted
    Write-Host ""
    Write-Host "Run: gofmt -w ." -ForegroundColor Yellow
    exit 1
}
Write-Host "✓ Formatting OK" -ForegroundColor Green

# 2. Vet
Write-Host "🔬 Running go vet..." -ForegroundColor Yellow
go vet ./...
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ go vet failed" -ForegroundColor Red
    exit 1
}
Write-Host "✓ Vet OK" -ForegroundColor Green

# 3. Build
Write-Host "🔨 Building..." -ForegroundColor Yellow
go build ./...
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed" -ForegroundColor Red
    exit 1
}
Write-Host "✓ Build OK" -ForegroundColor Green

# 4. Tests (skip integration tests that need credentials)
Write-Host "🧪 Running tests..." -ForegroundColor Yellow
go test -short ./...
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Tests failed" -ForegroundColor Red
    exit 1
}
Write-Host "✓ Tests OK" -ForegroundColor Green

# 5. Check for secrets in staged files (CRITICAL)
Write-Host "🔐 Scanning for secrets..." -ForegroundColor Yellow

$secretsFound = $false
$stagedFiles = git diff --cached --name-only 2>$null
if (-not $stagedFiles) {
    # No staged files, check all tracked files that differ from HEAD
    $stagedFiles = git diff --name-only HEAD 2>$null
}

# Secret patterns to detect
$secretPatterns = @(
    @{ Pattern = 'eyJ[A-Za-z0-9_-]{20,}\.eyJ[A-Za-z0-9_-]{20,}'; Name = 'JWT Token' },
    @{ Pattern = '(api[_-]?key|apikey)\s*[:=]\s*[''"][A-Za-z0-9_-]{16,}'; Name = 'API Key' },
    @{ Pattern = '(password|passwd|pwd)\s*[:=]\s*[''"][^''"]{8,}'; Name = 'Password' },
    @{ Pattern = '(secret|token)\s*[:=]\s*[''"][A-Za-z0-9_-]{16,}'; Name = 'Secret/Token' },
    @{ Pattern = 'Bearer\s+[A-Za-z0-9_-]{20,}'; Name = 'Bearer Token' },
    @{ Pattern = 'AKIA[0-9A-Z]{16}'; Name = 'AWS Access Key' },
    @{ Pattern = 'sk-[A-Za-z0-9]{32,}'; Name = 'OpenAI API Key' },
    @{ Pattern = 'ghp_[A-Za-z0-9]{36}'; Name = 'GitHub Token' },
    @{ Pattern = '-----BEGIN\s+(RSA\s+)?PRIVATE KEY-----'; Name = 'Private Key' }
)

foreach ($file in $stagedFiles) {
    if (-not $file) { continue }
    
    # Skip test files, examples, and binary files
    if ($file -match '_test\.(go|js|ts)$') { continue }
    if ($file -match '\.(example|sample|template)$') { continue }
    if ($file -match '\.(exe|dll|bin|fit|db|png|jpg|gif)$') { continue }
    if ($file -match 'testdata[/\\]') { continue }
    
    if (Test-Path $file -ErrorAction SilentlyContinue) {
        $content = Get-Content $file -Raw -ErrorAction SilentlyContinue
        if (-not $content) { continue }
        
        foreach ($pattern in $secretPatterns) {
            if ($content -match $pattern.Pattern) {
                Write-Host "❌ SECRETS DETECTED: $($pattern.Name) in $file" -ForegroundColor Red
                $secretsFound = $true
            }
        }
    }
}

if ($secretsFound) {
    Write-Host ""
    Write-Host "🚨 Commit blocked: Secrets detected in staged files!" -ForegroundColor Red
    Write-Host "   Remove the secrets and try again." -ForegroundColor Red
    Write-Host "   If this is a false positive, add the file to .gitignore or update the pattern." -ForegroundColor Yellow
    exit 1
}
Write-Host "✓ No secrets detected" -ForegroundColor Green

# 6. Check for common issues
Write-Host "🔎 Checking for common issues..." -ForegroundColor Yellow

# Check for large files
foreach ($file in $stagedFiles) {
    if ($file -and (Test-Path $file -ErrorAction SilentlyContinue)) {
        $size = (Get-Item $file).Length
        if ($size -gt 1048576) {
            $sizeKB = [math]::Round($size / 1024)
            Write-Host "⚠️  Warning: Large file $file (${sizeKB}KB)" -ForegroundColor Yellow
        }
    }
}

Write-Host ""
Write-Host "✅ All pre-commit checks passed!" -ForegroundColor Green
exit 0
