# Set window title
$host.UI.RawUI.WindowTitle = "MLCProxy"
Write-Host "Starting MLCProxy..." -ForegroundColor Green

# Kill any running instances of mlcproxy
Get-Process "mlcproxy" -ErrorAction SilentlyContinue | Stop-Process -Force

# Set proxy settings for the current session
$env:http_proxy = "http://localhost:3128"
$env:https_proxy = "http://localhost:3128"

# Function to add stats.local to hosts file if not present
function Ensure-StatsLocal {
    $hostsPath = "$env:SystemRoot\System32\drivers\etc\hosts"
    $statsEntry = "127.0.0.1 stats.local"
    
    # Check if entry exists
    $content = Get-Content $hostsPath
    if ($content -notcontains $statsEntry) {
        try {
            # Add entry to hosts file (requires admin rights)
            Add-Content -Path $hostsPath -Value "`n$statsEntry" -ErrorAction Stop
            Write-Host "Added stats.local to hosts file" -ForegroundColor Green
        }
        catch {
            Write-Host "Warning: Could not add stats.local to hosts file. Run as administrator if needed." -ForegroundColor Yellow
        }
    }
}

# Ensure stats.local is in hosts file
Ensure-StatsLocal

# Start MLCProxy in a new window
Start-Process "mlcproxy.exe" -WindowStyle Normal

# Wait for the proxy to start
Start-Sleep -Seconds 2

# Open the stats page in the default browser
Start-Process "http://stats.local"

Write-Host "`nMLCProxy is running. Press Ctrl+C to stop." -ForegroundColor Green
Write-Host "Statistics page should open in your default browser." -ForegroundColor Green

# Keep the window open
pause
