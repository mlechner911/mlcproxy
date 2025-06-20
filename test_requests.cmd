@echo off
setlocal enabledelayedexpansion

REM Set proxy URL
set PROXY=http://localhost:3128

REM Set target URL (using a reliable test site)
set TARGET=http://example.com

REM Counter for progress display
set /a count=0
set /a total=500

echo Testing proxy with %total% requests to %TARGET%
echo Using proxy: %PROXY%
echo.

REM Perform requests
for /l %%i in (1,1,%total%) do (
    set /a count+=1
    set /a percent=count*100/total
    echo !percent!%% complete - Request !count! of %total%
    curl --proxy %PROXY% -s -o nul %TARGET%
)

echo.
echo Test completed: %total% requests sent
pause
