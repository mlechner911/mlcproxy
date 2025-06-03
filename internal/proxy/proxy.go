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
	"strings"
	"time"
)

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
func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract actual host without port
	host := r.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	// Check if this is a stats request (either via stats.local or /stats or /stat path)
	if host == h.statsHost || strings.HasPrefix(r.URL.Path, "/stats") || strings.HasPrefix(r.URL.Path, "/stat/") || r.URL.Path == "/stat" {
		// Check for recursion
		if r.Header.Get("X-MLCProxy-Internal") == "true" {
			http.Error(w, "Loop detected", http.StatusInternalServerError)
			return
		}
		r.Header.Set("X-MLCProxy-Internal", "true")

		// Normalize path: /stat -> /stats
		r.URL.Path = strings.ReplaceAll(r.URL.Path, "/stat/", "/stats/")
		if r.URL.Path == "/stat" {
			r.URL.Path = "/stats"
		}

		// Remove /stats prefix if present
		originalPath := r.URL.Path
		if strings.HasPrefix(originalPath, "/stats") {
			r.URL.Path = strings.TrimPrefix(originalPath, "/stats")
		}
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

// handleHTTP handles standard HTTP proxy requests
func (h *ProxyHandler) handleHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle Chrome DevTools requests
	if strings.Contains(r.URL.Path, "/.well-known/appspecific/com.chrome.devtools") {
		handleDevToolsRequest(w)
		stats.LogRequest(r, http.StatusOK, 2, 2)
		return
	}
	// Track request body size
	var requestReader *TrackingReader
	if r.Body != nil {
		requestReader = NewTrackingReader(r.Body)
		r.Body = io.NopCloser(requestReader)
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
	responseReader := NewTrackingReader(resp.Body)
	_, err = io.Copy(w, responseReader)
	if err != nil {
		log.Printf("Error copying response: %v", err)
	}

	var requestBytes int64
	if requestReader != nil {
		requestBytes = int64(requestReader.BytesRead())
	}
	stats.LogRequest(r, resp.StatusCode, requestBytes, int64(responseReader.BytesRead()))
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
	clientReader := NewTrackingReader(clientConn)
	targetReader := NewTrackingReader(targetConn)

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
	stats.LogRequest(r, http.StatusOK, int64(clientReader.BytesRead()), int64(targetReader.BytesRead()))
}
