# --- Proxy Settings ---
$port = "7890" 
$env:HTTP_PROXY="http://127.0.0.1:$port"
$env:HTTPS_PROXY="http://127.0.0.1:$port"

Write-Host "-------------------------------------------" -ForegroundColor Cyan
Write-Host "[1/3] Proxy set to 127.0.0.1:$port" -ForegroundColor Green

# --- Smart Node Switch ---
$currentNodeV = node -v
if ($currentNodeV -like "v20*") {
    Write-Host "[2/3] Node is already $currentNodeV, skipping switch." -ForegroundColor Gray
} else {
    Write-Host "[2/3] Current Node is $currentNodeV. Switching to v20..." -ForegroundColor Cyan
    nvm use 20
}

# --- Network Test ---
Write-Host "[3/3] Checking connection to Google..." -ForegroundColor Cyan
try {
    $response = curl.exe -I -s --connect-timeout 5 https://www.google.com
    if ($response -match "200 OK") {
        Write-Host "✅ Network OK! Launching Gemini CLI..." -ForegroundColor Green
        Write-Host "-------------------------------------------" -ForegroundColor Cyan
        gemini
    } else {
        throw "Connection failed"
    }
} catch {
    Write-Host "❌ ERROR: Cannot reach Google. Check your Clash settings." -ForegroundColor Red
    Write-Host "Press any key to exit..."
    $null = [System.Console]::ReadKey()
}