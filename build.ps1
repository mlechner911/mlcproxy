# Build-Skript für MLCProxy

function Start-MLCProxyBuild {
    try {
        # Definiere Pfade
        $distPath = "dist"
        $staticPath = "$distPath\static"        # Sicherstellen, dass keine alte Instanz läuft
        Write-Host "Searching and terminating running MLCProxy instances..." -ForegroundColor Yellow
        $processes = Get-Process | Where-Object {
            $_.Path -like "*mlcproxy*" -or 
            $_.ProcessName -like "*mlcproxy*" -or
            ($_.CommandLine -like "*mlcproxy*" -and $_.CommandLine -notlike "*build.ps1*")
        } -ErrorAction SilentlyContinue

        if ($processes) {
            Write-Host "Found processes:" -ForegroundColor Yellow
            $processes | ForEach-Object {
                Write-Host "  - $($_.ProcessName) (PID: $($_.Id))"
                try {
                    $_ | Stop-Process -Force
                    Write-Host "    Terminated." -ForegroundColor Green
                } catch {
                    Write-Host "    Error terminating process: $_" -ForegroundColor Red
                }
            }
            # Warten bis Prozesse wirklich beendet sind
            Start-Sleep -Seconds 2
            # Nochmal prüfen
            $remainingProcesses = Get-Process | Where-Object {
                $_.Path -like "*mlcproxy*" -or $_.ProcessName -like "*mlcproxy*"
            } -ErrorAction SilentlyContinue
            if ($remainingProcesses) {
                throw "Could not terminate all MLCProxy processes"
            }
        } else {
            Write-Host "No running MLCProxy instances found." -ForegroundColor Green
        }
        
        # Alte Dateien entfernen
        if (Test-Path $distPath) {
            Write-Host "Removing old build directory..." -ForegroundColor Yellow
            Remove-Item -Path $distPath -Recurse -Force
        }

        # Erstelle dist-Ordner und Unterordner
        Write-Host "Creating build directories..." -ForegroundColor Yellow
        New-Item -ItemType Directory -Force -Path $distPath | Out-Null
        New-Item -ItemType Directory -Force -Path $staticPath | Out-Null        # Baue das Programm mit Optimierungen und Metadaten
        Write-Host "Compiling MLCProxy (optimized)..." -ForegroundColor Yellow
        $env:GOOS = "windows"
        $env:GOARCH = "amd64"
        $env:CGO_ENABLED = "0"
        $companyName = "Michael Lechner"
        $productName = "MLCProxy"
        $description = "HTTP/HTTPS Proxy with Statistics Interface"
        $version = "1.0.1"
          $ldflags = @(
            "-s -w",  # Strip debugging info
            "-X 'main.CompanyName=$companyName'",
            "-X 'main.ProductName=$productName'",
            "-X 'main.FileDescription=$description'",
            "-X 'main.ProductVersion=$version'"
            # Entfernt: -H=windowsgui, damit Konsolenausgaben sichtbar sind
        )
        
        & go build -o "$distPath\mlcproxy.exe" -ldflags="$($ldflags -join ' ')" -trimpath cmd\proxy\main.go
        if (-not $?) { throw "Build failed" }

        # Kopiere statische Dateien
        Write-Host "Copying static files..." -ForegroundColor Yellow
        Copy-Item "internal\stats\static\*" -Destination "$staticPath" -Recurse -Force
          # Füge Windows-Manifest zur Binary hinzu
        Write-Host "Adding Windows manifest..." -ForegroundColor Yellow
        if (Get-Command "mt.exe" -ErrorAction SilentlyContinue) {
            & mt.exe -manifest mlcproxy.manifest -outputresource:"$distPath\mlcproxy.exe;1"
        } else {
            Write-Host "mt.exe not found - manifest will not be embedded" -ForegroundColor Yellow
        }

        # Kopiere Dokumentation und Konfiguration
        Write-Host "Copying documentation and configuration..." -ForegroundColor Yellow
        $filesToCopy = @("LICENSE", "README.md", "README.de.md")
        foreach ($file in $filesToCopy) {
            Copy-Item $file -Destination "$distPath" -Force
        }
        
        # Kopiere oder erstelle Konfigurationsdatei
        if (Test-Path "config.ini") {
            Copy-Item "config.ini" -Destination "$distPath" -Force
        } else {
            Write-Host "Creating initial config.ini..." -ForegroundColor Yellow
            Copy-Item "config.ini.example" -Destination "$distPath\config.ini" -Force
        }

        Write-Host "`nBuild completed successfully!" -ForegroundColor Green
        Write-Host "MLCProxy can now be started from the 'dist' directory."
        return $true
    }
    catch {
        Write-Host "`nBuild failed: $_" -ForegroundColor Red
        return $false
    }
}

# Build ausführen
Start-MLCProxyBuild
