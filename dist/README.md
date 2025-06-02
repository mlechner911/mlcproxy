# MLCProxy

A robust HTTP(S) proxy server with integrated statistics display and traffic monitoring.

## Features

- Complete HTTP and HTTPS proxy support with CONNECT handling
- Integrated real-time statistics display via stats.local
- Live traffic monitoring and analysis
- Configurable proxy port (default: 3128)
- Advanced error handling and status messages
- User-friendly web interface with error feedback
- Detailed client statistics and byte tracking
- Automatic display updates with reconnection attempts
- Chrome DevTools compatibility
- Improved host detection and routing logic

## Installation

```powershell
# PowerShell
go build -o mlcproxy.exe cmd/proxy/main.go
```

```bash
# Bash/CMD
go build -o mlcproxy.exe cmd/proxy/main.go
```

## Usage

Start the proxy with the default port (3128):
```powershell
.\mlcproxy.exe
```

Or specify a custom port:
```powershell
.\mlcproxy.exe -port 8080
```

The statistics page can be accessed in two ways:
1. http://stats.local (requires proxy configuration)
2. http://localhost:3128/stat (direct)

## Proxy Configuration

Configure your browser or client with the following settings:
- Host: localhost or 127.0.0.1
- Port: 3128 (or your custom port)

## Curl Examples

HTTP test:
```powershell
# PowerShell
curl.exe -v --proxy http://localhost:3128 http://httpbin.org/get

# Alternative using Invoke-WebRequest
Invoke-WebRequest -Proxy "http://localhost:3128" -Uri "http://httpbin.org/get" -Verbose
```

HTTPS test:
```powershell
# PowerShell
curl.exe -v --proxy http://localhost:3128 https://httpbin.org/get

# Alternative using Invoke-WebRequest
Invoke-WebRequest -Proxy "http://localhost:3128" -Uri "https://httpbin.org/get" -Verbose
```

Get statistics:
```powershell
# PowerShell - Direct
curl.exe http://localhost:3128/stat
# or
Invoke-WebRequest -Uri "http://localhost:3128/stat"

# PowerShell - Via Proxy (stats.local)
curl.exe --proxy http://localhost:3128 http://stats.local
# or
Invoke-WebRequest -Proxy "http://localhost:3128" -Uri "http://stats.local"
```

### Note for PowerShell Users
In PowerShell, commands are chained using `;` instead of `&&`. Example:
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

## Languages

- [Deutsche Version (German Version)](README.de.md)
