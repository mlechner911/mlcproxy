package gui

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"mlc_goproxy/internal/stats"
	"net/http"
	"sync"
)

type guiData struct {
	Stats    template.HTML
	Port     int
	TopStats []*stats.ClientStats
}

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>MLCProxy Statistics</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 20px;
            background-color: #f0f0f0;
        }
        .container {
            background-color: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .status {
            color: #2c3e50;
            font-weight: bold;
            margin-bottom: 20px;
        }
        .instructions {
            background-color: #f8f9fa;
            padding: 15px;
            border-radius: 4px;
            margin: 15px 0;
        }
        .stats {
            margin-top: 20px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 20px;
            background: white;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background-color: #f8f9fa;
            font-weight: bold;
            color: #2c3e50;
        }
        tr:hover {
            background-color: #f5f5f5;
        }
        .summary {
            margin: 20px 0;
            padding: 15px;
            background-color: #e8f4f8;
            border-radius: 4px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h2>MLCProxy Status</h2>
        <div class="status">Proxy läuft auf Port {{.Port}}</div>
        <div class="instructions">
            <h3>Proxy-Konfiguration:</h3>
            <p>Host: localhost<br>Port: {{.Port}}</p>
        </div>
        <div class="summary">
            {{.Stats}}
        </div>
        <div class="stats">
            <h3>Top 10 Clients nach Traffic</h3>
            <table>
                <thead>
                    <tr>
                        <th>IP</th>
                        <th>Traffic In</th>
                        <th>Traffic Out</th>
                        <th>Gesamt Traffic</th>
                        <th>Anfragen</th>
                        <th>Letzter Zugriff</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .TopStats}}
                    <tr>
                        <td>{{.IP}}</td>
                        <td>{{.BytesInFormatted}}</td>
                        <td>{{.BytesOutFormatted}}</td>
                        <td>{{.TotalBytesFormatted}}</td>
                        <td>{{.RequestCount}}</td>
                        <td>{{.LastAccessFormatted}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>
    </div>
    <script>
        setInterval(() => {
            fetch('/stats')
                .then(response => response.text())
                .then(stats => {
                    document.querySelector('.summary').innerHTML = stats;
                });
            fetch('/topstats')
                .then(response => response.json())
                .then(stats => {
                    const tbody = document.querySelector('tbody');
                    tbody.innerHTML = stats.map(client => ` + "`" + `
                        <tr>
                            <td>${client.IP}</td>
                            <td>${client.BytesInFormatted}</td>
                            <td>${client.BytesOutFormatted}</td>
                            <td>${client.TotalBytesFormatted}</td>
                            <td>${client.RequestCount}</td>
                            <td>${client.LastAccessFormatted}</td>
                        </tr>
                    ` + "`" + `).join('');
                });
        }, 1000);
    </script>
</body>
</html>`

var (
	srv  *http.Server
	once sync.Once
	port int = 3128 // Default port
)

// SetPort sets the proxy port for the GUI
func SetPort(p int) {
	port = p
}

func Start() {
	once.Do(func() {
		mux := http.NewServeMux()

		// Handle main page
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			tmpl, err := template.New("index").Parse(htmlTemplate)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			data := guiData{
				Stats:    template.HTML(stats.GetCurrentStats().String()),
				Port:     port,
				TopStats: stats.GetTopClients(10),
			}

			tmpl.Execute(w, data)
		})

		// Handle stats updates
		mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, stats.GetCurrentStats().String())
		})

		// Handle top stats
		mux.HandleFunc("/topstats", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(stats.GetTopClients(10))
		})

		srv = &http.Server{
			Addr:    "127.0.0.1:9090",
			Handler: mux,
		}

		log.Printf("Starting web interface on http://127.0.0.1:9090")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("Web interface error: %v", err)
		}
	})
}
