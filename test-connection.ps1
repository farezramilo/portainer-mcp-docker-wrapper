# Test Portainer MCP Wrapper Connection
# Run this script to verify the wrapper is accessible

Write-Host "Testing Portainer MCP Wrapper..." -ForegroundColor Cyan
Write-Host ""

# Test 1: Basic connectivity
Write-Host "Test 1: Testing basic connectivity to port 8081..." -ForegroundColor Yellow
try {
    $result = Test-NetConnection -ComputerName 192.168.0.242 -Port 8081 -WarningAction SilentlyContinue
    if ($result.TcpTestSucceeded) {
        Write-Host "  [OK] Port 8081 is accessible" -ForegroundColor Green
    } else {
        Write-Host "  [FAIL] Port 8081 is NOT accessible" -ForegroundColor Red
        Write-Host "    Possible firewall issue on Docker host" -ForegroundColor Red
    }
} catch {
    Write-Host "  [FAIL] Connection test failed: $_" -ForegroundColor Red
}

Write-Host ""

# Test 2: Health endpoint
Write-Host "Test 2: Testing health endpoint..." -ForegroundColor Yellow
try {
    $headers = @{
        "Authorization" = "Bearer 57rGjKsvCB1xPC56gg3Ty1vFmSQOSQWrXyUyuydJEh4="
    }
    $response = Invoke-WebRequest -Uri "http://192.168.0.242:8081/health" -Headers $headers -TimeoutSec 5 -ErrorAction Stop
    Write-Host "  [OK] Health endpoint responded: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "  Response: $($response.Content)" -ForegroundColor Gray
} catch {
    Write-Host "  [FAIL] Health endpoint failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# Test 3: MCP SSE endpoint
Write-Host "Test 3: Testing MCP SSE endpoint..." -ForegroundColor Yellow
try {
    $headers = @{
        "Authorization" = "Bearer 57rGjKsvCB1xPC56gg3Ty1vFmSQOSQWrXyUyuydJEh4="
    }
    $response = Invoke-WebRequest -Uri "http://192.168.0.242:8081/sse" -Headers $headers -TimeoutSec 5 -ErrorAction Stop
    Write-Host "  [OK] SSE endpoint accessible: $($response.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "  [FAIL] SSE endpoint failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "Testing complete!" -ForegroundColor Cyan
