@echo off
title MLCProxy
echo Starting MLCProxy...

REM Kill any running instances of mlcproxy
taskkill /F /IM mlcproxy.exe 2>nul

REM Set proxy settings for the current session
set http_proxy=http://localhost:3128
set https_proxy=http://localhost:3128

REM Start MLCProxy in a new window
start "MLCProxy" cmd /c "mlcproxy.exe"

REM Wait for the proxy to start
timeout /t 2 /nobreak

REM Open the stats page in the default browser
start http://stats.local

REM Add the localhost exception to bypass proxy for stats.local
netsh winhttp set proxy proxy-server="localhost:3128" bypass-list="stats.local"

echo MLCProxy is running. Press Ctrl+C to stop.
echo Statistics page should open in your default browser.
pause
