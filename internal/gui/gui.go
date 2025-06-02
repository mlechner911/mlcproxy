package gui

import (
	"fmt"
	"html/template"
	"log"
	"mlc_goproxy/internal/stats"
	"net/http"
	"sync"
)

type guiData struct {
	Stats string
	Port  int
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
            font-family: monospace;
            margin-top: 20px;
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
        <div class="stats">
            {{.Stats}}
        </div>
    </div>
    <script>
        setInterval(() => {
            fetch('/stats')
                .then(response => response.text())
                .then(stats => {
                    document.querySelector('.stats').textContent = stats;
                });
        }, 1000);
    </script>
</body>
</html>
`

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
				Stats: stats.GetCurrentStats().String(),
				Port:  port,
			}

			tmpl.Execute(w, data)
		})

		// Handle stats updates
		mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, stats.GetCurrentStats().String())
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
