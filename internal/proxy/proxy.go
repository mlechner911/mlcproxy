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

// TrackingReader wraps an io.Reader to track the number of bytes read
type TrackingReader struct {
	r         io.Reader
	bytesRead uint64
}

func (t *TrackingReader) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	t.bytesRead += uint64(n)
	return
}

// getClientIP extracts the client's IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		// Take first IP from list
		ips := strings.Split(forwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	// Fallback to RemoteAddr
	remoteAddr := r.RemoteAddr

	// Handle IPv4 with port
	if strings.Count(remoteAddr, ":") == 1 {
		host, _, err := net.SplitHostPort(remoteAddr)
		if err == nil {
			return host
		}
	}

	// Handle IPv6 with port
	if strings.HasPrefix(remoteAddr, "[") {
		host, _, err := net.SplitHostPort(remoteAddr)
		if err == nil {
			return strings.Trim(host, "[]")
		}
		// IPv6 without port
		return strings.Trim(remoteAddr, "[]")
	}

	return remoteAddr
}

// copyHeader copies HTTP headers from src to dst
func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// Start initializes and starts the proxy server
func Start(addr string) error {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	handler := &ProxyHandler{
		statsPath:   config.Cfg.Paths.StatsPath,
		apiPath:     config.Cfg.Paths.APIPath,
		statsHost:   config.Cfg.Features.StatsHost,
		authManager: &AuthManager{},
	}

	// Log security settings
	log.Printf("Security settings:")
	if config.Cfg.Auth.EnableAuth {
		log.Printf("- Basic Auth enabled with %d users", len(config.Cfg.Auth.Credentials))
	} else {
		log.Printf("- Basic Auth disabled")
	}

	log.Printf("- Allowed networks: %v", config.Cfg.Security.AllowedNetworks)
	log.Printf("- Stats host %s is always allowed", handler.statsHost)

	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	log.Printf("Starting proxy server on %s", addr)
	log.Printf("Statistics available at http://%s%s", handler.statsHost, handler.statsPath)
	log.Printf("Configure your browser to use http://%s as proxy", addr)
	return server.ListenAndServe()
}

// ProxyHandler handles proxy requests and implements http.Handler
type ProxyHandler struct {
	statsPath   string
	apiPath     string
	statsHost   string
	authManager *AuthManager
}

// ServeHTTP handles all incoming HTTP requests
func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) { // Direct request to stats path
	if strings.HasPrefix(r.URL.Path, h.statsPath) {
		h.handleStats(w, r)
		return
	}

	// Extract actual host without port
	host := r.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	// Check for stats host (always allowed) - legacy support
	if host == h.statsHost {
		h.handleStats(w, r)
		return
	}

	// Extract client IP
	clientIP := getClientIP(r)

	// Verify client IP
	if !h.authManager.IsIPAllowed(clientIP) {
		log.Printf("Access denied for IP %s - not in allowed networks (%v)",
			clientIP, config.Cfg.Security.AllowedNetworks)
		http.Error(w, fmt.Sprintf("Access denied - IP %s not in allowed networks (%s)", clientIP, strings.Join(config.Cfg.Security.AllowedNetworks, ", ")), http.StatusForbidden)
		stats.LogRequest(r, http.StatusForbidden, 0, 0)
		return
	}

	// Check auth if enabled
	if config.Cfg.Auth.EnableAuth && !h.authManager.CheckAuth(r) {
		log.Printf("Auth failed for IP %s", clientIP)
		h.authManager.RequireAuth(w)
		stats.LogRequest(r, http.StatusProxyAuthRequired, 0, 0)
		return
	}

	// Log all other requests
	log.Printf("Proxy request: %s %s %s from IP %s", r.Method, r.Host, r.URL.String(), clientIP)

	// Handle HTTPS CONNECT requests
	if r.Method == http.MethodConnect {
		h.handleHTTPS(w, r)
		return
	}

	// Handle standard HTTP proxy requests
	h.handleHTTP(w, r)
}

// getPreferredLanguage extracts the preferred language from Accept-Language header
func (h *ProxyHandler) getPreferredLanguage(r *http.Request) string {
	acceptLang := r.Header.Get("Accept-Language")
	if acceptLang == "" {
		return "de" // Default to German
	}

	// Extract language from header (e.g., "en-US,en;q=0.9" -> "en")
	parts := strings.Split(acceptLang, ",")
	if len(parts) > 0 {
		langParts := strings.Split(parts[0], "-")
		if len(langParts) > 0 {
			return strings.ToLower(langParts[0])
		}
	}

	return "de" // Default to German
}

// handleStats serves the statistics interface and API
func (h *ProxyHandler) handleStats(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case path == h.apiPath+"/stats":
		// API endpoint for statistics
		handleStatsAPI(w, r)
	case path == h.statsPath || path == "/" || path == "":
		// Main page with language selection
		lang := h.getPreferredLanguage(r)

		// Try language-specific file first
		htmlFile := filepath.Join(config.Cfg.Paths.StaticDir, fmt.Sprintf("index.%s.html", lang))
		log.Printf("Trying to serve HTML file: %s", htmlFile)

		if _, err := os.Stat(htmlFile); os.IsNotExist(err) {
			// Fall back to default file
			htmlFile = filepath.Join(config.Cfg.Paths.StaticDir, "index.html")
			log.Printf("File not found, falling back to: %s", htmlFile)
		}

		// Check if file exists
		if _, err := os.Stat(htmlFile); os.IsNotExist(err) {
			log.Printf("HTML file not found: %s", htmlFile)
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, htmlFile)

	case strings.HasPrefix(path, "/styles.css"):
		// Serve CSS file
		cssFile := filepath.Join(config.Cfg.Paths.StaticDir, "styles.css")
		log.Printf("Trying to serve CSS file: %s", cssFile)
		if _, err := os.Stat(cssFile); os.IsNotExist(err) {
			log.Printf("CSS file not found: %s", cssFile)
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, cssFile)

	case strings.HasPrefix(path, "/script.js"):
		// Serve JavaScript file
		jsFile := filepath.Join(config.Cfg.Paths.StaticDir, "script.js")
		log.Printf("Trying to serve JS file: %s", jsFile)
		if _, err := os.Stat(jsFile); os.IsNotExist(err) {
			log.Printf("JavaScript file not found: %s", jsFile)
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, jsFile)

	default:
		// All other paths are not allowed
		http.NotFound(w, r)
	}
}

// handleHTTP handles standard HTTP proxy requests
func (h *ProxyHandler) handleHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle Chrome DevTools requests
	if strings.Contains(r.URL.Path, "/.well-known/appspecific/com.chrome.devtools") {
		handleDevToolsRequest(w)
		stats.LogRequest(r, http.StatusOK, 2, 2)
		return
	}

	// Track request body size
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

	// Track response body size
	responseReader := &TrackingReader{r: resp.Body}
	_, err = io.Copy(w, responseReader)
	if err != nil {
		log.Printf("Error copying response: %v", err)
	}

	stats.LogRequest(r, resp.StatusCode, int64(requestReader.bytesRead), int64(responseReader.bytesRead))
}

// handleHTTPS handles HTTPS CONNECT tunnel requests
func (h *ProxyHandler) handleHTTPS(w http.ResponseWriter, r *http.Request) {
	log.Printf("HTTPS CONNECT request to: %s", r.URL.Host)

	// Ensure we have a host with port
	host := r.URL.Host
	if !strings.Contains(host, ":") {
		host += ":443"
	}

	// Hijack the connection
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

	// Connect to target
	targetConn, err := net.DialTimeout("tcp", host, 10*time.Second)
	if err != nil {
		log.Printf("Failed to connect to %s: %v", host, err)
		clientConn.Write([]byte(fmt.Sprintf("HTTP/1.1 504 Gateway Timeout\r\n\r\n")))
		return
	}
	defer targetConn.Close()

	// Send connection established response
	_, err = clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	if err != nil {
		log.Printf("Failed to send 200 response: %v", err)
		return
	}

	// Set up traffic tracking
	clientReader := &TrackingReader{r: clientConn}
	targetReader := &TrackingReader{r: targetConn}

	// Create bidirectional tunnel
	done := make(chan bool, 2)

	// Client -> Target tunnel
	go func() {
		io.Copy(targetConn, clientReader)
		targetConn.(*net.TCPConn).CloseWrite()
		done <- true
	}()

	// Target -> Client tunnel
	go func() {
		io.Copy(clientConn, targetReader)
		clientConn.(*net.TCPConn).CloseWrite()
		done <- true
	}()

	// Wait for either direction to finish
	<-done

	// Log transfer statistics
	stats.LogRequest(r, http.StatusOK, int64(clientReader.bytesRead), int64(targetReader.bytesRead))
}

// handleDevToolsRequest handles Chrome DevTools protocol requests
func handleDevToolsRequest(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

// handleStatsAPI serves the statistics API endpoint
func handleStatsAPI(w http.ResponseWriter, r *http.Request) {
	stats.GetStats().ServeHTTP(w, r)
}
