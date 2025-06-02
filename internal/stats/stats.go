package stats

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Statistics struct {
	TotalRequests uint64
	StartTime     time.Time
	mu            sync.RWMutex
}

var stats = &Statistics{
	StartTime: time.Now(),
}

func LogRequest(req *http.Request) {
	stats.mu.Lock()
	defer stats.mu.Unlock()
	stats.TotalRequests++
}

func GetCurrentStats() *Statistics {
	stats.mu.RLock()
	defer stats.mu.RUnlock()
	return &Statistics{
		TotalRequests: stats.TotalRequests,
		StartTime:     stats.StartTime,
	}
}

func (s *Statistics) String() string {
	uptime := time.Since(s.StartTime).Round(time.Second)
	return fmt.Sprintf("Uptime: %v\nTotal Requests: %d", uptime, s.TotalRequests)
}
