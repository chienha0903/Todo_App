# bench.ps1 — So sanh GOMAXPROCS=1 vs GOMAXPROCS=2 tren todo-bff
#
# Usage:
#   .\scripts\benchmark\bench.ps1                    # chay ca 2 case
#   .\scripts\benchmark\bench.ps1 -Case 1            # chi chay GOMAXPROCS=1
#   .\scripts\benchmark\bench.ps1 -Case 2            # chi chay GOMAXPROCS=2
#   .\scripts\benchmark\bench.ps1 -Requests 1000 -Concurrency 100

param(
    [int]$Requests    = 500,
    [int]$Concurrency = 50,
    [int]$Case        = 0,     # 0 = ca 2, 1 = chi case 1, 2 = chi case 2
    [string]$BaseURL  = "http://localhost:8080"
)

$ErrorActionPreference = "Stop"

# ── Helpers ──────────────────────────────────────────────────────────────────

function Write-Header($msg) {
    Write-Host ""
    Write-Host "═══════════════════════════════════════════" -ForegroundColor Cyan
    Write-Host "  $msg" -ForegroundColor Cyan
    Write-Host "═══════════════════════════════════════════" -ForegroundColor Cyan
}

function Ensure-Hey {
    if (-not (Get-Command hey -ErrorAction SilentlyContinue)) {
        Write-Host "► Installing hey..." -ForegroundColor Yellow
        go install github.com/rakyll/hey@latest
        # Ensure GOPATH/bin is in PATH
        $gopath = go env GOPATH
        $env:PATH = "$gopath\bin;$env:PATH"
        if (-not (Get-Command hey -ErrorAction SilentlyContinue)) {
            Write-Error "hey not found after install. Add $(go env GOPATH)\bin to PATH."
        }
    }
    Write-Host "✓ hey found: $(Get-Command hey | Select-Object -ExpandProperty Source)" -ForegroundColor Green
}

function Wait-BFFReady($url) {
    Write-Host "► Waiting for BFF to be ready..." -ForegroundColor Yellow
    $deadline = (Get-Date).AddSeconds(60)
    while ((Get-Date) -lt $deadline) {
        try {
            $r = Invoke-WebRequest -Uri "$url/health" -Method GET -TimeoutSec 2 -ErrorAction Stop
            if ($r.StatusCode -eq 200) {
                Write-Host "✓ BFF is ready" -ForegroundColor Green
                return
            }
        } catch {}
        Start-Sleep -Seconds 2
    }
    Write-Error "BFF did not become ready within 60 seconds."
}

function Get-JWT($url) {
    $body = '{"query":"mutation { login(input: { email: \"test@example.com\", password: \"TestPassword123!\" }) { accessToken } }"}'
    try {
        $resp = Invoke-RestMethod -Uri "$url/graphql" `
            -Method POST `
            -ContentType "application/json" `
            -Body $body `
            -ErrorAction Stop
        $token = $resp.data.login.accessToken
        if (-not $token) { throw "empty token" }
        Write-Host "✓ JWT obtained" -ForegroundColor Green
        return $token
    } catch {
        Write-Host "⚠  Could not get JWT (login failed). Benchmarking /health instead." -ForegroundColor Yellow
        return $null
    }
}

function Restart-BFF($procs) {
    Write-Host "► Restarting todo-bff with GOMAXPROCS=$procs..." -ForegroundColor Yellow
    $env:GOMAXPROCS = "$procs"
    docker compose up -d --build --no-deps todo-bff 2>&1 | Out-Null
    # Khi container restart, can wait them healthcheck
    Start-Sleep -Seconds 5
}

function Confirm-GOMAXPROCS($expectedProcs) {
    # Doc log tu container de xac nhan GOMAXPROCS thuc te
    $logs = docker compose logs --tail=30 todo-bff 2>&1
    $line = $logs | Select-String "GOMAXPROCS" | Select-Object -Last 1
    if ($line) {
        Write-Host "✓ Log xac nhan: $($line.Line.Trim())" -ForegroundColor Green
    } else {
        Write-Host "⚠  Khong tim thay GOMAXPROCS trong log — kiem tra thu cong:" -ForegroundColor Yellow
        Write-Host "   docker compose logs todo-bff | findstr GOMAXPROCS"
    }
}

function Run-Hey($label, $url, $token, $reqs, $conc) {
    Write-Host ""
    Write-Host "► Running: $label  (n=$reqs, c=$conc)" -ForegroundColor Yellow

    $outFile = "bench_${label}.txt"

    if ($token) {
        $gqlBody = '{"query":"{ todos(userId: 1, page: 1, pageSize: 10) { todos { id title } totalCount } }"}'
        hey -n $reqs -c $conc `
            -m POST `
            -H "Authorization: Bearer $token" `
            -H "Content-Type: application/json" `
            -d $gqlBody `
            "$url/graphql" | Tee-Object -FilePath $outFile
    } else {
        # Fallback: /health endpoint (khong can auth)
        hey -n $reqs -c $conc "$url/health" | Tee-Object -FilePath $outFile
    }

    Write-Host "► Ket qua da luu vao $outFile" -ForegroundColor Cyan
    return $outFile
}

function Parse-HeyOutput($file) {
    if (-not (Test-Path $file)) { return $null }
    $content = Get-Content $file -Raw

    $rps      = if ($content -match 'Requests/sec:\s+([\d.]+)') { $matches[1] } else { "N/A" }
    $avg      = if ($content -match 'Average:\s+([\d.]+) secs')  { [math]::Round([double]$matches[1] * 1000, 2) } else { "N/A" }
    $p95      = if ($content -match '95% in\s+([\d.]+) secs')    { [math]::Round([double]$matches[1] * 1000, 2) } else { "N/A" }
    $p99      = if ($content -match '99% in\s+([\d.]+) secs')    { [math]::Round([double]$matches[1] * 1000, 2) } else { "N/A" }
    $errRate  = if ($content -match '\[non-2xx\]\s+(\d+)')        { $matches[1] } else { "0" }

    return [PSCustomObject]@{
        RPS      = $rps
        AvgMs    = $avg
        P95Ms    = $p95
        P99Ms    = $p99
        ErrCount = $errRate
    }
}

# ── Main ─────────────────────────────────────────────────────────────────────

Write-Header "GOMAXPROCS Benchmark — todo-bff"
Write-Host "Target  : $BaseURL"
Write-Host "Requests: $Requests  Concurrency: $Concurrency"

Ensure-Hey
Wait-BFFReady $BaseURL
$jwt = Get-JWT $BaseURL

$results = @{}

# ── Case 1: GOMAXPROCS=1 ─────────────────────────────────────────────────────
if ($Case -eq 0 -or $Case -eq 1) {
    Write-Header "Case A: GOMAXPROCS=1"
    Restart-BFF 1
    Wait-BFFReady $BaseURL
    Confirm-GOMAXPROCS 1
    $f1 = Run-Hey "GOMAXPROCS1" $BaseURL $jwt $Requests $Concurrency
    $results["GOMAXPROCS=1"] = Parse-HeyOutput $f1
}

# ── Case 2: GOMAXPROCS=2 ─────────────────────────────────────────────────────
if ($Case -eq 0 -or $Case -eq 2) {
    Write-Header "Case B: GOMAXPROCS=2"
    Restart-BFF 2
    Wait-BFFReady $BaseURL
    Confirm-GOMAXPROCS 2
    $f2 = Run-Hey "GOMAXPROCS2" $BaseURL $jwt $Requests $Concurrency
    $results["GOMAXPROCS=2"] = Parse-HeyOutput $f2
}

# ── Summary ──────────────────────────────────────────────────────────────────
Write-Header "Ket qua tong hop"

Write-Host ""
Write-Host ("{0,-15} {1,10} {2,12} {3,12} {4,12} {5,10}" -f "Case","RPS","Avg(ms)","p95(ms)","p99(ms)","Errors")
Write-Host ("{0,-15} {1,10} {2,12} {3,12} {4,12} {5,10}" -f "----","---","-------","-------","-------","------")

foreach ($key in $results.Keys | Sort-Object) {
    $r = $results[$key]
    Write-Host ("{0,-15} {1,10} {2,12} {3,12} {4,12} {5,10}" -f $key, $r.RPS, $r.AvgMs, $r.P95Ms, $r.P99Ms, $r.ErrCount)
}

Write-Host ""
Write-Host "Chi tiet: xem bench_GOMAXPROCS1.txt va bench_GOMAXPROCS2.txt" -ForegroundColor Cyan

# Reset GOMAXPROCS ve default sau khi benchmark xong
Remove-Item Env:GOMAXPROCS -ErrorAction SilentlyContinue
