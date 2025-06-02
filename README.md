# MLCProxy

Ein HTTP(S) Proxy-Server mit Web-Interface für Statistiken.

## Features

- HTTP und HTTPS Proxy-Unterstützung
- Web-Interface für Statistiken (http://127.0.0.1:9090)
- Konfigurierbarer Proxy-Port (Standard: 3128)
- Echtzeit-Statistiken

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

Das Web-Interface ist unter http://127.0.0.1:9090 erreichbar.

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
