/*
Copyright (c) 2025 Michael Lechner

This software is released under the MIT License.
See the LICENSE file for further details.
*/

package proxy

import (
	"fmt"
	"io"
	"log"
	"mlc_goproxy/internal/config"
	"mlc_goproxy/internal/stats"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type TrackingReader struct {
	r         io.Reader
	bytesRead uint64
}

func (t *TrackingReader) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	t.bytesRead += uint64(n)
	return
}

func getClientIP(r *http.Request) string {
	return strings.Split(r.RemoteAddr, ":")[0]
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func Start(addr string) error {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	handler := &ProxyHandler{
		statsPath:   config.Cfg.Paths.StatsPath,
		apiPath:     config.Cfg.Paths.APIPath,
		statsHost:   config.Cfg.Features.StatsHost,
		authManager: &AuthManager{},
	}

	// Logge Sicherheitseinstellungen
	log.Printf("Sicherheitseinstellungen:")
	if config.Cfg.Auth.EnableAuth {
		log.Printf("- Basic Auth aktiviert mit %d Benutzern", len(config.Cfg.Auth.Credentials))
	} else {
		log.Printf("- Basic Auth deaktiviert")
	}

	log.Printf("- Erlaubte Netzwerke: %v", config.Cfg.Security.AllowedNetworks)
	log.Printf("- Statistik-Host %s ist immer zugelassen", handler.statsHost)

	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	log.Printf("Starte Proxy-Server auf %s", addr)
	log.Printf("Statistiken verfügbar unter http://%s%s", handler.statsHost, handler.statsPath)
	log.Printf("Konfigurieren Sie Ihren Browser für http://%s als Proxy", addr)
	return server.ListenAndServe()
}

// ProxyHandler verarbeitet die Proxy-Anfragen
type ProxyHandler struct {
	statsPath   string
	apiPath     string
	statsHost   string
	authManager *AuthManager
}

func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extrahiere den tatsächlichen Host ohne Port
	host := r.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	// Prüfe auf Statistik-Host (immer erlaubt)
	if host == h.statsHost {
		h.handleStats(w, r)
		return
	}

	// Extrahiere Client-IP
	clientIP := getClientIP(r)

	// Überprüfe Client-IP
	if !h.authManager.IsIPAllowed(clientIP) {
		log.Printf("Zugriff verweigert für IP %s - nicht im erlaubten Netzwerk (%v)",
			clientIP, config.Cfg.Security.AllowedNetworks)
		http.Error(w, "Access denied - IP not in allowed networks", http.StatusForbidden)
		stats.LogRequest(r, http.StatusForbidden, 0, 0)
		return
	}

	// Überprüfe Auth wenn aktiviert
	if config.Cfg.Auth.EnableAuth && !h.authManager.CheckAuth(r) {
		log.Printf("Auth fehlgeschlagen für IP %s", clientIP)
		h.authManager.RequireAuth(w)
		stats.LogRequest(r, http.StatusProxyAuthRequired, 0, 0)
		return
	}

	// Logge alle anderen Anfragen
	log.Printf("Proxy request: %s %s %s von IP %s", r.Method, r.Host, r.URL.String(), clientIP)

	// HTTPS CONNECT requests
	if r.Method == http.MethodConnect {
		h.handleHTTPS(w, r)
		return
	}

	// Standard HTTP Proxy Requests
	h.handleHTTP(w, r)
}

func (h *ProxyHandler) getPreferredLanguage(r *http.Request) string {
	// Accept-Language Header auslesen
	acceptLang := r.Header.Get("Accept-Language")
	if acceptLang == "" {
		return "de" // Fallback auf Deutsch
	}

	// Sprache aus dem Header extrahieren (z.B. "en-US,en;q=0.9" -> "en")
	parts := strings.Split(acceptLang, ",")
	if len(parts) > 0 {
		langParts := strings.Split(parts[0], "-")
		if len(langParts) > 0 {
			return strings.ToLower(langParts[0])
		}
	}

	return "de" // Fallback auf Deutsch
}

func (h *ProxyHandler) handleStats(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	log.Printf("Handling stats request: %s", path)
	log.Printf("Static dir configured as: %s", config.Cfg.Paths.StaticDir)

	switch {
	case path == h.apiPath+"/stats":
		// API endpoint für Statistiken
		handleStatsAPI(w, r)
	case path == "/" || path == "":
		// Hauptseite mit Sprachauswahl
		lang := h.getPreferredLanguage(r)

		// Versuche zuerst die sprachspezifische Datei
		htmlFile := filepath.Join(config.Cfg.Paths.StaticDir, fmt.Sprintf("index.%s.html", lang))
		log.Printf("Trying to serve HTML file: %s", htmlFile)

		if _, err := os.Stat(htmlFile); os.IsNotExist(err) {
			// Fallback auf Standard-Datei
			htmlFile = filepath.Join(config.Cfg.Paths.StaticDir, "index.html")
			log.Printf("File not found, falling back to: %s", htmlFile)
		}

		// Prüfe ob die Datei existiert
		if _, err := os.Stat(htmlFile); os.IsNotExist(err) {
			log.Printf("HTML file not found: %s", htmlFile)
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, htmlFile)

	case strings.HasPrefix(path, "/styles.css"):
		// CSS-Datei
		cssFile := filepath.Join(config.Cfg.Paths.StaticDir, "styles.css")
		log.Printf("Trying to serve CSS file: %s", cssFile)
		if _, err := os.Stat(cssFile); os.IsNotExist(err) {
			log.Printf("CSS file not found: %s", cssFile)
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, cssFile)

	case strings.HasPrefix(path, "/script.js"):
		// JavaScript-Datei
		jsFile := filepath.Join(config.Cfg.Paths.StaticDir, "script.js")
		log.Printf("Trying to serve JS file: %s", jsFile)
		if _, err := os.Stat(jsFile); os.IsNotExist(err) {
			log.Printf("JavaScript file not found: %s", jsFile)
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, jsFile)

	default:
		// Alle anderen Pfade sind nicht erlaubt
		http.NotFound(w, r)
	}
}

func (h *ProxyHandler) handleHTTP(w http.ResponseWriter, r *http.Request) {
	// Chrome DevTools check
	if strings.Contains(r.URL.Path, "/.well-known/appspecific/com.chrome.devtools") {
		handleDevToolsRequest(w)
		stats.LogRequest(r, http.StatusOK, 2, 2)
		return
	}

	// Standard proxy logic
	var requestReader TrackingReader
	if r.Body != nil {
		requestReader = TrackingReader{r: r.Body}
		r.Body = io.NopCloser(&requestReader)
	}

	// Ensure complete URL
	targetURL := r.URL.String()
	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		targetURL = "http://" + r.Host + targetURL
	}

	// Create and send request
	client := &http.Client{}
	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		stats.LogRequest(r, http.StatusInternalServerError, 0, 0)
		return
	}

	copyHeader(req.Header, r.Header)
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		stats.LogRequest(r, http.StatusInternalServerError, 0, 0)
		return
	}
	defer resp.Body.Close()

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	responseReader := &TrackingReader{r: resp.Body}
	_, err = io.Copy(w, responseReader)
	if err != nil {
		log.Printf("Error copying response: %v", err)
	}

	stats.LogRequest(r, resp.StatusCode, int64(requestReader.bytesRead), int64(responseReader.bytesRead))
}

func (h *ProxyHandler) handleHTTPS(w http.ResponseWriter, r *http.Request) {
	log.Printf("HTTPS CONNECT request to: %s", r.URL.Host)

	// Ensure we have a host with port
	host := r.URL.Host
	if !strings.Contains(host, ":") {
		host += ":443"
	}

	// First hijack the connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		msg := "Proxy server doesn't support hijacking"
		log.Print(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		msg := fmt.Sprintf("Hijacking failed: %v", err)
		log.Print(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	// Then connect to target
	targetConn, err := net.DialTimeout("tcp", host, 10*time.Second)
	if err != nil {
		log.Printf("Failed to connect to %s: %v", host, err)
		clientConn.Write([]byte(fmt.Sprintf("HTTP/1.1 504 Gateway Timeout\r\n\r\n")))
		return
	}
	defer targetConn.Close()

	// Send connection established
	_, err = clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	if err != nil {
		log.Printf("Failed to send 200 response: %v", err)
		return
	}

	// Set up tracking
	clientReader := &TrackingReader{r: clientConn}
	targetReader := &TrackingReader{r: targetConn}

	// Create tunnels
	done := make(chan bool, 2)

	// Client -> Target
	go func() {
		io.Copy(targetConn, clientReader)
		targetConn.(*net.TCPConn).CloseWrite()
		done <- true
	}()

	// Target -> Client
	go func() {
		io.Copy(clientConn, targetReader)
		clientConn.(*net.TCPConn).CloseWrite()
		done <- true
	}()

	// Wait for either direction to finish
	<-done

	// Log statistics
	stats.LogRequest(r, http.StatusOK, int64(clientReader.bytesRead), int64(targetReader.bytesRead))
}

func handleDevToolsRequest(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func handleStatsAPI(w http.ResponseWriter, r *http.Request) {
	stats.GetStats().ServeHTTP(w, r)
}

func handleStatsPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := stats.WriteHTMLStats(w, r); err != nil {
		log.Printf("Error writing HTML: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Diese Funktionen wurden in den AuthManager verschoben
