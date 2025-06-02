# Build-Skript für MLCProxy

function Start-MLCProxyBuild {
    try {
        # Definiere Pfade
        $distPath = "dist"
        $staticPath = "$distPath\static"
        
        # Sicherstellen, dass keine alte Instanz läuft
        Write-Host "Beende laufende Instanzen..." -ForegroundColor Yellow
        Get-Process "mlcproxy" -ErrorAction SilentlyContinue | 
            Stop-Process -Force -ErrorAction SilentlyContinue
        
        # Warten bis Prozesse beendet sind
        Start-Sleep -Seconds 2
        
        # Alte Dateien entfernen
        if (Test-Path $distPath) {
            Write-Host "Entferne altes Build-Verzeichnis..." -ForegroundColor Yellow
            Remove-Item -Path $distPath -Recurse -Force
        }

        # Erstelle dist-Ordner und Unterordner
        Write-Host "Erstelle Build-Verzeichnisse..." -ForegroundColor Yellow
        New-Item -ItemType Directory -Force -Path $distPath | Out-Null
        New-Item -ItemType Directory -Force -Path $staticPath | Out-Null

        # Baue das Programm
        Write-Host "Kompiliere MLCProxy..." -ForegroundColor Yellow
        & go build -o "$distPath\mlcproxy.exe" cmd\proxy\main.go
        if (-not $?) { throw "Build fehlgeschlagen" }

        # Kopiere statische Dateien
        Write-Host "Kopiere statische Dateien..." -ForegroundColor Yellow
        Copy-Item "internal\stats\static\*" -Destination "$staticPath" -Recurse -Force
        
        # Kopiere Dokumentation und Konfiguration
        Write-Host "Kopiere Dokumentation und Konfiguration..." -ForegroundColor Yellow
        $filesToCopy = @("LICENSE", "README.md", "README.de.md")
        foreach ($file in $filesToCopy) {
            Copy-Item $file -Destination "$distPath" -Force
        }
        
        # Kopiere oder erstelle Konfigurationsdatei
        if (Test-Path "config.ini") {
            Copy-Item "config.ini" -Destination "$distPath" -Force
        } else {
            Write-Host "Erstelle initiale config.ini..." -ForegroundColor Yellow
            Copy-Item "config.ini.example" -Destination "$distPath\config.ini" -Force
        }

        Write-Host "`nBuild erfolgreich abgeschlossen!" -ForegroundColor Green
        Write-Host "MLCProxy kann jetzt aus dem 'dist'-Verzeichnis gestartet werden."
        return $true
    }
    catch {
        Write-Host "`nBuild fehlgeschlagen: $_" -ForegroundColor Red
        return $false
    }
}

# Build ausführen
Start-MLCProxyBuild
