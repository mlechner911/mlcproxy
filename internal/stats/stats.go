package stats

import (
	"fmt"
	"net"
	"net/http"
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

type ClientStats struct {
	IP                  string    `json:"IP"`
	BytesIn             uint64    `json:"BytesIn"`
	BytesOut            uint64    `json:"BytesOut"`
	BytesInFormatted    string    `json:"BytesInFormatted"`
	BytesOutFormatted   string    `json:"BytesOutFormatted"`
	TotalBytesFormatted string    `json:"TotalBytesFormatted"`
	LastAccess          time.Time `json:"-"`
	LastAccessFormatted string    `json:"LastAccessFormatted"`
	RequestCount        uint64    `json:"RequestCount"`
}

type Statistics struct {
	TotalRequests uint64
	StartTime     time.Time
	Clients       map[string]*ClientStats
	mu            sync.RWMutex
}

var stats = &Statistics{
	StartTime: time.Now(),
	Clients:   make(map[string]*ClientStats),
}

func LogRequest(req *http.Request) {
	stats.mu.Lock()
	defer stats.mu.Unlock()

	stats.TotalRequests++
	// Get client IP
	ip := getClientIP(req)

	// Update or create client stats
	if _, exists := stats.Clients[ip]; !exists {
		stats.Clients[ip] = &ClientStats{
			IP: ip,
		}
	}

	client := stats.Clients[ip]
	client.LastAccess = time.Now()
	client.RequestCount++
}

func LogTransfer(ip string, bytesIn, bytesOut uint64) {
	stats.mu.Lock()
	defer stats.mu.Unlock()

	if client, exists := stats.Clients[ip]; exists {
		client.BytesIn += bytesIn
		client.BytesOut += bytesOut
	}
}

func GetCurrentStats() *Statistics {
	stats.mu.RLock()
	defer stats.mu.RUnlock()

	// Create a deep copy
	statsCopy := &Statistics{
		TotalRequests: stats.TotalRequests,
		StartTime:     stats.StartTime,
		Clients:       make(map[string]*ClientStats),
	}

	for ip, client := range stats.Clients {
		statsCopy.Clients[ip] = &ClientStats{
			IP:           client.IP,
			BytesIn:      client.BytesIn,
			BytesOut:     client.BytesOut,
			LastAccess:   client.LastAccess,
			RequestCount: client.RequestCount,
		}
	}

	return statsCopy
}

func GetTopClients(n int) []*ClientStats {
	stats.mu.RLock()
	defer stats.mu.RUnlock()
	// Convert map to slice for sorting
	clients := make([]*ClientStats, 0, len(stats.Clients))
	for _, client := range stats.Clients {
		total := client.BytesIn + client.BytesOut
		clients = append(clients, &ClientStats{
			IP:                  client.IP,
			BytesIn:             client.BytesIn,
			BytesOut:            client.BytesOut,
			BytesInFormatted:    formatBytes(client.BytesIn),
			BytesOutFormatted:   formatBytes(client.BytesOut),
			TotalBytesFormatted: formatBytes(total),
			LastAccess:          client.LastAccess,
			LastAccessFormatted: client.LastAccess.Format("15:04:05"),
			RequestCount:        client.RequestCount,
		})
	}

	// Sort by total traffic (BytesIn + BytesOut)
	sort.Slice(clients, func(i, j int) bool {
		totalI := clients[i].BytesIn + clients[i].BytesOut
		totalJ := clients[j].BytesIn + clients[j].BytesOut
		return totalI > totalJ
	})

	// Return top N clients
	if n > len(clients) {
		n = len(clients)
	}
	return clients[:n]
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

func (s *Statistics) String() string {
	uptime := time.Since(s.StartTime).Round(time.Second)
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Uptime: %v\nTotal Requests: %d\n\n", uptime, s.TotalRequests))
	sb.WriteString("Top 10 Clients by Traffic:\n")
	sb.WriteString("------------------------\n")

	for _, client := range GetTopClients(10) {
		sb.WriteString(fmt.Sprintf("\nIP: %s\n", client.IP))
		sb.WriteString(fmt.Sprintf("Bytes In: %s\n", formatBytes(client.BytesIn)))
		sb.WriteString(fmt.Sprintf("Bytes Out: %s\n", formatBytes(client.BytesOut)))
		sb.WriteString(fmt.Sprintf("Requests: %d\n", client.RequestCount))
		sb.WriteString(fmt.Sprintf("Last Access: %s\n", client.LastAccess.Format("15:04:05")))
	}

	return sb.String()
}
