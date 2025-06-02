# MLCProxy Installer
# Erfordert Administrative Rechte
# Installiere MLCProxy auf Windows
#Requires -RunAsAdministrator

param(
    [string]$InstallDir = "$env:ProgramFiles\MLCProxy",
    [switch]$NoDesktopShortcut,
    [switch]$NoStartMenu,
    [switch]$NoAutostart
)

# Setze Ausführungsrichtlinie für das aktuelle Skript
Set-ExecutionPolicy Bypass -Scope Process -Force

# Funktion zum Erstellen von Verknüpfungen
function New-Shortcut {
    param(
        [string]$TargetPath,
        [string]$ShortcutPath,
        [string]$Arguments
    )
    $WScriptShell = New-Object -ComObject WScript.Shell
    $Shortcut = $WScriptShell.CreateShortcut($ShortcutPath)
    $Shortcut.TargetPath = $TargetPath
    $Shortcut.Arguments = $Arguments
    $Shortcut.Save()
}

# Stoppe laufende MLCProxy-Instanzen
Write-Host "Stoppe laufende MLCProxy-Instanzen..." -ForegroundColor Yellow
Get-Process "mlcproxy" -ErrorAction SilentlyContinue | Stop-Process -Force

# Erstelle Installationsverzeichnis
Write-Host "Erstelle Installationsverzeichnis..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
New-Item -ItemType Directory -Force -Path "$InstallDir\static" | Out-Null

# Kopiere Dateien
Write-Host "Kopiere Programmdateien..." -ForegroundColor Yellow
Copy-Item "dist\mlcproxy.exe" -Destination $InstallDir -Force
Copy-Item "dist\config.ini" -Destination $InstallDir -Force
Copy-Item "dist\static\*" -Destination "$InstallDir\static" -Recurse -Force
Copy-Item "dist\LICENSE" -Destination $InstallDir -Force
Copy-Item "dist\README.*" -Destination $InstallDir -Force

# Konfiguriere Windows-Firewall
Write-Host "Konfiguriere Firewall..." -ForegroundColor Yellow
$firewallRule = Get-NetFirewallRule -DisplayName "MLCProxy" -ErrorAction SilentlyContinue
if ($firewallRule) {
    Remove-NetFirewallRule -DisplayName "MLCProxy"
}
New-NetFirewallRule -DisplayName "MLCProxy" -Direction Inbound -Action Allow -Program "$InstallDir\mlcproxy.exe" -Protocol TCP

# Füge stats.local zu hosts-Datei hinzu
# Write-Host "Konfiguriere hosts-Datei..." -ForegroundColor Yellow
# $hostsPath = "$env:SystemRoot\System32\drivers\etc\hosts"
# $statsEntry = "127.0.0.1 stats.local"
# $hostsContent = Get-Content $hostsPath
# if ($hostsContent -notcontains $statsEntry) {
#     Add-Content -Path $hostsPath -Value "`n$statsEntry"
# }

# Erstelle Desktop-Verknüpfung
if (-not $NoDesktopShortcut) {
    Write-Host "Erstelle Desktop-Verknüpfung..." -ForegroundColor Yellow
    $desktopPath = [System.Environment]::GetFolderPath('Desktop')
    New-Shortcut -TargetPath "$InstallDir\mlcproxy.exe" -ShortcutPath "$desktopPath\MLCProxy.lnk"
}

# Erstelle Startmenü-Einträge
if (-not $NoStartMenu) {
    Write-Host "Erstelle Startmenü-Einträge..." -ForegroundColor Yellow
    $startMenuPath = "$env:ProgramData\Microsoft\Windows\Start Menu\Programs\MLCProxy"
    New-Item -ItemType Directory -Force -Path $startMenuPath | Out-Null
    New-Shortcut -TargetPath "$InstallDir\mlcproxy.exe" -ShortcutPath "$startMenuPath\MLCProxy.lnk"
}

# Erstelle und konfiguriere Windows-Dienst
if (-not $NoAutostart) {
    Write-Host "Konfiguriere Autostart-Dienst..." -ForegroundColor Yellow
    $serviceName = "MLCProxy"
    
    # Entferne existierenden Dienst
    if (Get-Service $serviceName -ErrorAction SilentlyContinue) {
        sc.exe delete $serviceName
        Start-Sleep -Seconds 2
    }
    
    # Erstelle neuen Dienst
    $nssm = "$InstallDir\nssm.exe"
    if (-not (Test-Path $nssm)) {
        # Download NSSM (Non-Sucking Service Manager)
        $nssmUrl = "https://nssm.cc/release/nssm-2.24.zip"
        $nssmZip = "$env:TEMP\nssm.zip"
        Invoke-WebRequest -Uri $nssmUrl -OutFile $nssmZip
        Expand-Archive -Path $nssmZip -DestinationPath "$env:TEMP\nssm"
        Copy-Item "$env:TEMP\nssm\nssm-2.24\win64\nssm.exe" -Destination $nssm
        Remove-Item $nssmZip
        Remove-Item "$env:TEMP\nssm" -Recurse
    }
    
    # Installiere und konfiguriere Dienst
    & $nssm install $serviceName "$InstallDir\mlcproxy.exe"
    & $nssm set $serviceName DisplayName "MLCProxy - HTTP(S) Proxy Server"
    & $nssm set $serviceName Description "Ein robuster HTTP(S) Proxy-Server mit integrierter Statistik-Anzeige"
    & $nssm set $serviceName Start SERVICE_AUTO_START
    Start-Service $serviceName
}

Write-Host "`nInstallation abgeschlossen!" -ForegroundColor Green
Write-Host "MLCProxy wurde installiert in: $InstallDir"
Write-Host "Proxy läuft auf: http://localhost:3128"
Write-Host "Statistik-Seite: http://stats.local"

if (-not $NoAutostart) {
    Write-Host "Der Dienst 'MLCProxy' wurde erstellt und gestartet."
}

Write-Host "`nDrücken Sie eine beliebige Taste zum Beenden..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
