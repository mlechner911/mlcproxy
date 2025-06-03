/*
Copyright (c) 2025 Michael Lechner

This software is released under the MIT License.
See the LICENSE file for further details.
*/

package stats

import (
	"encoding/json"
	"fmt"
	"mlc_goproxy/internal/config"
	"mlc_goproxy/internal/version"
	"net"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// getClientIP extrahiert die IP-Adresse aus der RemoteAddr oder X-Forwarded-For Header
func getClientIP(r *http.Request) string {
	// Zuerst prüfen wir X-Forwarded-For
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		// Nehmen die erste IP aus der Liste
		ips := strings.Split(forwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	// Fallback auf RemoteAddr
	remoteAddr := r.RemoteAddr

	// Entferne den Port für IPv4
	if strings.Count(remoteAddr, ":") == 1 {
		host, _, err := net.SplitHostPort(remoteAddr)
		if err == nil {
			return host
		}
	}

	// Behandle IPv6
	if strings.HasPrefix(remoteAddr, "[") {
		// IPv6 mit Port: [::1]:1234
		host, _, err := net.SplitHostPort(remoteAddr)
		if err == nil {
			return host
		}
		// IPv6 ohne Port: [::1]
		return strings.Trim(remoteAddr, "[]")
	}

	// Wenn alles andere fehlschlägt, gib die Adresse wie sie ist zurück
	return remoteAddr
}

type RequestInfo struct {
	Timestamp time.Time `json:"timestamp"`
	ClientIP  string    `json:"client_ip"`
	Method    string    `json:"method"`
	Host      string    `json:"host"`
	Path      string    `json:"path"`
	Status    int       `json:"status"`
	BytesIn   int64     `json:"bytes_in"`
	BytesOut  int64     `json:"bytes_out"`
}

type ClientStats struct {
	IP         string    `json:"ip"`
	BytesIn    int64     `json:"bytes_in"`
	BytesOut   int64     `json:"bytes_out"`
	BytesTotal int64     `json:"bytes_total"`
	Requests   int       `json:"requests"`
	LastSeen   time.Time `json:"last_seen"`
}

type Stats struct {
	mu             sync.RWMutex
	StartTime      time.Time               `json:"start_time"`
	TotalRequests  int64                   `json:"total_requests"`
	TotalBytesIn   int64                   `json:"total_bytes_in"`
	TotalBytesOut  int64                   `json:"total_bytes_out"`
	ActiveClients  int                     `json:"active_clients"`
	ClientStats    map[string]*ClientStats `json:"-"`
	RecentRequests []RequestInfo           `json:"-"`
}

var globalStats = New()

// GetStats gibt die globale Statistik-Instanz zurück
func GetStats() *Stats {
	return globalStats
}

func New() *Stats {
	return &Stats{
		StartTime:      time.Now(),
		ClientStats:    make(map[string]*ClientStats),
		RecentRequests: make([]RequestInfo, 0, 100),
	}
}

func LogRequest(req *http.Request, status int, bytesIn, bytesOut int64) {
	globalStats.mu.Lock()
	defer globalStats.mu.Unlock()

	globalStats.TotalRequests++

	// Get client IP
	ip := getClientIP(req)

	// Update or create client stats
	if _, exists := globalStats.ClientStats[ip]; !exists {
		globalStats.ClientStats[ip] = &ClientStats{
			IP: ip,
		}
	}

	client := globalStats.ClientStats[ip]
	client.LastSeen = time.Now()
	client.Requests++

	// Update bytes for client
	client.BytesIn += bytesIn
	client.BytesOut += bytesOut
	client.BytesTotal = client.BytesIn + client.BytesOut

	// Add to recent requests
	reqInfo := RequestInfo{
		Timestamp: time.Now(),
		ClientIP:  ip,
		Method:    req.Method,
		Host:      req.Host,
		Path:      req.URL.Path,
		Status:    status,
		BytesIn:   bytesIn,
		BytesOut:  bytesOut,
	}

	if len(globalStats.RecentRequests) >= 100 {
		globalStats.RecentRequests = append(globalStats.RecentRequests[1:], reqInfo)
	} else {
		globalStats.RecentRequests = append(globalStats.RecentRequests, reqInfo)
	}

	globalStats.updateActiveClients()

	// Update total bytes
	globalStats.TotalBytesIn += bytesIn
	globalStats.TotalBytesOut += bytesOut
}

func LogTransfer(ip string, bytesIn, bytesOut uint64) {
	globalStats.mu.Lock()
	defer globalStats.mu.Unlock()

	if client, exists := globalStats.ClientStats[ip]; exists {
		client.BytesIn += int64(bytesIn)
		client.BytesOut += int64(bytesOut)
		client.BytesTotal = client.BytesIn + client.BytesOut
		client.LastSeen = time.Now() // Aktualisiere auch den Zeitstempel
	} else {
		// Falls der Client noch nicht existiert, erstelle einen neuen Eintrag
		globalStats.ClientStats[ip] = &ClientStats{
			IP:         ip,
			BytesIn:    int64(bytesIn),
			BytesOut:   int64(bytesOut),
			BytesTotal: int64(bytesIn + bytesOut),
			LastSeen:   time.Now(),
		}
	}

	// Update global totals
	globalStats.TotalBytesIn += int64(bytesIn)
	globalStats.TotalBytesOut += int64(bytesOut)
}

func (s *Stats) updateActiveClients() {
	threshold := time.Now().Add(-5 * time.Minute)
	active := 0
	for _, stats := range s.ClientStats {
		if stats.LastSeen.After(threshold) {
			active++
		}
	}
	s.ActiveClients = active
}

func (s *Stats) getTopClients(n int) []ClientStats {
	// Convert map to slice for sorting
	clients := make([]ClientStats, 0, len(s.ClientStats))
	for _, c := range s.ClientStats {
		clients = append(clients, *c)
	}

	// Sort by total bytes (descending)
	sort.Slice(clients, func(i, j int) bool {
		return clients[i].BytesTotal > clients[j].BytesTotal
	})

	// Return top n clients
	if len(clients) > n {
		clients = clients[:n]
	}
	return clients
}

func (s *Stats) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	response := struct {
		*Stats
		Version        string        `json:"version"`
		BuildDate      string        `json:"build_date"`
		RecentRequests []RequestInfo `json:"recent_requests"`
		ClientStats    []ClientStats `json:"client_stats"`
	}{
		Stats:          s,
		Version:        version.Version,
		BuildDate:      version.BuildDate,
		RecentRequests: s.RecentRequests,
		ClientStats:    s.getTopClients(10),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// WriteHTMLStats schreibt die HTML-Statistikseite in den ResponseWriter
func WriteHTMLStats(w http.ResponseWriter, r *http.Request) error {
	http.ServeFile(w, r, filepath.Join(config.Cfg.Paths.StaticDir, "index.html"))
	return nil
}

// ServeStaticFiles registriert die Handler für statische Dateien
func ServeStaticFiles(mux *http.ServeMux) {
	fs := http.FileServer(http.Dir(config.Cfg.Paths.StaticDir))
	mux.Handle(config.Cfg.Paths.StatsPath+"/", http.StripPrefix(config.Cfg.Paths.StatsPath+"/", fs))
}

func formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
