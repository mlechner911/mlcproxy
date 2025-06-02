/*
Copyright (c) 2025 Michael Lechner
This software is released under the MIT License.
See the LICENSE file for further details.
*/

package proxy

import (
	"encoding/base64"
	"mlc_goproxy/internal/config"
	"net"
	"net/http"
	"strings"
)

// checkAuth prüft die Basic Authentication
func checkAuth(r *http.Request) bool {
	if !config.Cfg.Auth.EnableAuth {
		return true
	}

	auth := r.Header.Get("Proxy-Authorization")
	if auth == "" {
		return false
	}

	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return false
	}

	decoded, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return false
	}

	credentials := strings.SplitN(string(decoded), ":", 2)
	if len(credentials) != 2 {
		return false
	}

	username, password := credentials[0], credentials[1]
	if storedPass, ok := config.Cfg.Auth.Credentials[username]; ok {
		return password == storedPass
	}

	return false
}

// requireAuth sendet den Auth-Header
func requireAuth(w http.ResponseWriter) {
	w.Header().Set("Proxy-Authenticate", `Basic realm="MLCProxy Access"`)
	http.Error(w, "Proxy authentication required", http.StatusProxyAuthRequired)
}

// isIPAllowed prüft ob die IP-Adresse in den erlaubten Netzwerken liegt
func isIPAllowed(ipStr string) bool {
	if len(config.Cfg.Security.AllowedNetworks) == 0 {
		return true
	}

	// Entferne Port-Nummer wenn vorhanden
	if strings.Contains(ipStr, ":") {
		ipStr = strings.Split(ipStr, ":")[0]
	}

	clientIP := net.ParseIP(ipStr)
	if clientIP == nil {
		return false
	}

	for _, network := range config.Cfg.Security.AllowedNetworks {
		_, ipNet, err := net.ParseCIDR(network)
		if err != nil {
			continue
		}
		if ipNet.Contains(clientIP) {
			return true
		}
	}

	return false
}
