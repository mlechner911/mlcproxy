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

```bash
go build -o mlcproxy.exe cmd/proxy/main.go
```

## Verwendung

Starten Sie den Proxy mit dem Standardport (3128):
```bash
./mlcproxy.exe
```

Oder geben Sie einen benutzerdefinierten Port an:
```bash
./mlcproxy.exe -port 8080
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
```bash
curl -v --proxy http://localhost:3128 http://httpbin.org/get
```

HTTPS-Test:
```bash
curl -v --proxy http://localhost:3128 https://httpbin.org/get
```

Statistik abrufen:
```bash
# Direkt
curl http://localhost:3128/stat

# Über Proxy (stats.local)
curl --proxy http://localhost:3128 http://stats.local
```
