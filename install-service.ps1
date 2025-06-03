# MLCProxy Service Installer
# Erfordert Administrative Rechte
#Requires -RunAsAdministrator

param(
    [string]$InstallDir = "$env:ProgramFiles\MLCProxy",
    [switch]$Uninstall
)

$serviceName = "MLCProxy"

if ($Uninstall) {
    Write-Host "Deinstalliere MLCProxy-Dienst..." -ForegroundColor Yellow
    if (Get-Service $serviceName -ErrorAction SilentlyContinue) {
        Stop-Service $serviceName
        sc.exe delete $serviceName
        Write-Host "MLCProxy-Dienst wurde deinstalliert." -ForegroundColor Green
    } else {
        Write-Host "MLCProxy-Dienst ist nicht installiert." -ForegroundColor Yellow
    }
    exit
}

# Stoppe existierenden Dienst
if (Get-Service $serviceName -ErrorAction SilentlyContinue) {
    Write-Host "Stoppe existierenden MLCProxy-Dienst..." -ForegroundColor Yellow
    Stop-Service $serviceName
    sc.exe delete $serviceName
    Start-Sleep -Seconds 2
}

# Erstelle Installationsverzeichnis und Unterverzeichnisse
Write-Host "Erstelle Installationsverzeichnisse..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
New-Item -ItemType Directory -Force -Path "$InstallDir\static" | Out-Null
New-Item -ItemType Directory -Force -Path "$InstallDir\logs" | Out-Null

# Kopiere und aktualisiere Konfigurationsdatei
Write-Host "Kopiere und konfiguriere Dateien..." -ForegroundColor Yellow
Copy-Item "mlcproxy.exe" -Destination $InstallDir -Force
Copy-Item "static\*" -Destination "$InstallDir\static" -Recurse -Force

# Erstelle angepasste config.ini
$configContent = @"
# MLCProxy Konfiguration

[server]
# Port auf dem der Proxy läuft
port = 3128

[paths]
# Basis-Pfad für statische Dateien
static_dir = static
# URL-Pfad für die Statistik-API
stats_path = /stat
# URL-Pfad für die API-Endpunkte
api_path = /api

[features]
# Hostname für die Statistik-Seite
stats_host = stats.local

[auth]
# Aktiviere Basic Auth (true/false)
enable_auth = false
# Benutzername:Passwort Paare (mehrere möglich)
credentials = admin:secret,user1:pass1

[security]
# Erlaube nur bestimmte Netzwerke (CIDR-Notation, mehrere mit Komma getrennt)
# IPv4 Beispiele: 
# - Einzelne IP: 192.168.1.1/32
# - Ganzes Subnetz: 192.168.1.0/24
# - Mehrere Netze: 192.168.1.0/24,10.0.0.0/8
# - Alles erlauben (IPv4): 0.0.0.0/0
# IPv6 Beispiele:
# - Localhost: ::1/128
# - Link-local: fe80::/10
# - Alles erlauben (IPv6): ::/0
allowed_networks = 127.0.0.1/32,172.16.0.0/12,172.23.0.0/16,192.168.0.0/16,::1/128,fe80::/10
"@

$configContent | Set-Content -Path "$InstallDir\config.ini" -Force

# Lade NSSM herunter falls nicht vorhanden
$nssm = "$InstallDir\nssm.exe"
if (-not (Test-Path $nssm)) {
    Write-Host "Lade NSSM herunter..." -ForegroundColor Yellow
    $nssmUrl = "https://nssm.cc/release/nssm-2.24.zip"
    $nssmZip = "$env:TEMP\nssm.zip"
    Invoke-WebRequest -Uri $nssmUrl -OutFile $nssmZip
    Expand-Archive -Path $nssmZip -DestinationPath "$env:TEMP\nssm"
    Copy-Item "$env:TEMP\nssm\nssm-2.24\win64\nssm.exe" -Destination $nssm
    Remove-Item $nssmZip
    Remove-Item "$env:TEMP\nssm" -Recurse
}

Write-Host "Installiere MLCProxy-Dienst..." -ForegroundColor Yellow

# Installiere und konfiguriere Dienst
& $nssm install $serviceName "$InstallDir\mlcproxy.exe"
& $nssm set $serviceName DisplayName "MLCProxy - HTTP(S) Proxy Server"
& $nssm set $serviceName Description "Ein robuster HTTP(S) Proxy-Server mit integrierter Statistik-Anzeige"
& $nssm set $serviceName AppDirectory $InstallDir
& $nssm set $serviceName AppEnvironmentExtra "MLCPROXY_HOME=$InstallDir"
& $nssm set $serviceName Start SERVICE_AUTO_START
& $nssm set $serviceName ObjectName LocalSystem
& $nssm set $serviceName AppStdout "$InstallDir\logs\service.log"
& $nssm set $serviceName AppStderr "$InstallDir\logs\error.log"
& $nssm set $serviceName AppRotateFiles 1
& $nssm set $serviceName AppRotateOnline 1
& $nssm set $serviceName AppRotateBytes 10485760

# Konfiguriere hosts-Datei
$hostsPath = "$env:SystemRoot\System32\drivers\etc\hosts"
$statsEntry = "127.0.0.1 stats.local"
$hostsContent = Get-Content $hostsPath
if ($hostsContent -notcontains $statsEntry) {
    Write-Host "Füge stats.local zu hosts-Datei hinzu..." -ForegroundColor Yellow
    Add-Content -Path $hostsPath -Value "`n$statsEntry"
}

# Konfiguriere Windows-Firewall
Write-Host "Konfiguriere Firewall..." -ForegroundColor Yellow
$firewallRule = Get-NetFirewallRule -DisplayName "MLCProxy" -ErrorAction SilentlyContinue
if ($firewallRule) {
    Remove-NetFirewallRule -DisplayName "MLCProxy"
}
New-NetFirewallRule -DisplayName "MLCProxy" -Direction Inbound -Action Allow -Program "$InstallDir\mlcproxy.exe" -Protocol TCP

# Starte den Dienst
Write-Host "Starte MLCProxy-Dienst..." -ForegroundColor Yellow
Start-Service $serviceName

Write-Host "`nMLCProxy wurde erfolgreich als Dienst installiert!" -ForegroundColor Green
Write-Host "Dienst-Name: $serviceName"
Write-Host "Installations-Verzeichnis: $InstallDir"
Write-Host "Proxy-URL: http://localhost:3128"
Write-Host "Statistik-URL: http://stats.local"
Write-Host "`nDienst-Verwaltung:"
Write-Host "- Start: Start-Service MLCProxy"
Write-Host "- Stop: Stop-Service MLCProxy"
Write-Host "- Status: Get-Service MLCProxy"
Write-Host "- Deinstallation: $PSCommandPath -Uninstall"

# Zeige Dienst-Status
Get-Service $serviceName | Format-List Name, DisplayName, Status, StartType
