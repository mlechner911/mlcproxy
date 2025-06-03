/*
Copyright (c) 2025 Michael Lechner

This software is released under the MIT License.
See the LICENSE file for further details.
*/

package proxy

import (
	"fmt"
	"mlc_goproxy/internal/config"
	"mlc_goproxy/internal/stats"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// handleStats serves the statistics interface and API
func (h *ProxyHandler) handleStats(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" || path == "" {
		// Bei Root-URL direkt die Index-Seite ausliefern
		lang := h.getPreferredLanguage(r)
		htmlFile := filepath.Join(config.Cfg.Paths.StaticDir, fmt.Sprintf("index.%s.html", lang))

		// Check if language-specific file exists
		if _, err := os.Stat(htmlFile); os.IsNotExist(err) {
			htmlFile = filepath.Join(config.Cfg.Paths.StaticDir, "index.html")
		}
		http.ServeFile(w, r, htmlFile)
		return
	}

	// Remove leading slash and split path to get file extension
	path = strings.TrimPrefix(path, "/")
	ext := filepath.Ext(path)

	switch ext {
	case ".ico":
		// Favicon
		http.ServeFile(w, r, filepath.Join(config.Cfg.Paths.StaticDir, "favicon.svg"))
	case ".json":
		// Stats API endpoint
		handleStatsAPI(w, r)
	case ".css":
		// CSS file
		http.ServeFile(w, r, filepath.Join(config.Cfg.Paths.StaticDir, "styles.css"))
	case ".js":
		// JavaScript file
		http.ServeFile(w, r, filepath.Join(config.Cfg.Paths.StaticDir, "script.js"))
	case ".html":
		// Main page with language selection
		lang := h.getPreferredLanguage(r)
		htmlFile := filepath.Join(config.Cfg.Paths.StaticDir, fmt.Sprintf("index.%s.html", lang))

		// Check if language-specific file exists
		if _, err := os.Stat(htmlFile); os.IsNotExist(err) {
			htmlFile = filepath.Join(config.Cfg.Paths.StaticDir, "index.html")
		}
		http.ServeFile(w, r, htmlFile)
		return
	// http.NotFound(w, r)
	default:
		http.NotFound(w, r)
	}
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

// handleStatsAPI serves the statistics API endpoint
func handleStatsAPI(w http.ResponseWriter, r *http.Request) {
	stats.GetStats().ServeHTTP(w, r)
}

// handleDevToolsRequest handles Chrome DevTools protocol requests
func handleDevToolsRequest(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}
