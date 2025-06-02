# MLCProxy

Ein robuster HTTP(S) Proxy-Server mit integrierter Statistik-Anzeige und Traffic-Monitoring.

## Features

- Vollständige HTTP und HTTPS Proxy-Unterstützung mit CONNECT-Handling
- Integrierte Echtzeit-Statistik-Anzeige über stats.local
- Live Traffic-Monitoring und -Analyse
- Konfigurierbarer Proxy-Port (Standard: 3128)
- Erweiterte Fehlerbehandlung und Statusmeldungen
- Benutzerfreundliche Web-Oberfläche mit Fehler-Feedback
- Detaillierte Client-Statistiken und Byte-Tracking
- Automatische Aktualisierung der Anzeige mit Wiederverbindungsversuchen
- Chrome DevTools-Kompatibilität
- Verbesserte Host-Erkennung und Routing-Logik

## Installation

```powershell
# PowerShell
go build -o mlcproxy.exe cmd/proxy/main.go
```

```bash
# Bash/CMD
go build -o mlcproxy.exe cmd/proxy/main.go
```

## Verwendung

Starten Sie den Proxy mit dem Standardport (3128):
```powershell
.\mlcproxy.exe
```

Oder geben Sie einen benutzerdefinierten Port an:
```powershell
.\mlcproxy.exe -port 8080
```

Die Statistik-Seite ist auf zwei Arten erreichbar:
1. http://stats.local (erfordert Proxy-Konfiguration)
2. http://localhost:3128/stat (direkt)

## Proxy-Konfiguration

Konfigurieren Sie Ihren Browser oder Client mit folgenden Einstellungen:
- Host: localhost oder 127.0.0.1
- Port: 3128 (oder Ihr benutzerdefinierter Port)

## Curl-Beispiele

HTTP-Test:
```powershell
# PowerShell
curl.exe -v --proxy http://localhost:3128 http://httpbin.org/get

# Alternativ mit Invoke-WebRequest
Invoke-WebRequest -Proxy "http://localhost:3128" -Uri "http://httpbin.org/get" -Verbose
```

HTTPS-Test:
```powershell
# PowerShell
curl.exe -v --proxy http://localhost:3128 https://httpbin.org/get

# Alternativ mit Invoke-WebRequest
Invoke-WebRequest -Proxy "http://localhost:3128" -Uri "https://httpbin.org/get" -Verbose
```

Statistik abrufen:
```powershell
# PowerShell - Direkt
curl.exe http://localhost:3128/stat
# oder
Invoke-WebRequest -Uri "http://localhost:3128/stat"

# PowerShell - Über Proxy (stats.local)
curl.exe --proxy http://localhost:3128 http://stats.local
# oder
Invoke-WebRequest -Proxy "http://localhost:3128" -Uri "http://stats.local"
```

### Hinweis für PowerShell-Benutzer
In PowerShell werden Befehle mit `;` statt `&&` verkettet. Beispiel:
```powershell
go build -o mlcproxy.exe cmd/proxy/main.go; .\mlcproxy.exe
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Thanks to all contributors
- Icons from [Material Design Icons](https://material.io/icons/)
- Built with Go and modern web technologies

## Author

- **Michael Lechner** - *Initial work* - [MLCProxy](https://github.com/yourusername/mlcproxy)
