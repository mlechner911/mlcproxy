/*
Copyright (c) 2025 Michael Lechner
This software is released under the MIT License.
See the LICENSE file for further details.
*/

package proxy

import (
	"encoding/base64"
	"log"
	"mlc_goproxy/internal/config"
	"net"
	"net/http"
	"strings"
)

// AuthManager verwaltet die Authentifizierung und IP-Berechtigungen
type AuthManager struct{}

// CheckAuth prüft die Basic Authentication
func (am *AuthManager) CheckAuth(r *http.Request) bool {
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

// RequireAuth sendet den Auth-Header
func (am *AuthManager) RequireAuth(w http.ResponseWriter) {
	w.Header().Set("Proxy-Authenticate", `Basic realm="MLCProxy Access"`)
	http.Error(w, "Proxy authentication required", http.StatusProxyAuthRequired)
}

// IsIPAllowed prüft ob die IP-Adresse in den erlaubten Netzwerken liegt
func (am *AuthManager) IsIPAllowed(ipStr string) bool {
	if len(config.Cfg.Security.AllowedNetworks) == 0 {
		return true
	}

	// Extrahiere IP-Adresse aus Host:Port Format
	host := ipStr
	if strings.Count(ipStr, ":") == 1 {
		// IPv4 mit Port
		host, _, _ = net.SplitHostPort(ipStr)
	} else if strings.HasPrefix(ipStr, "[") && strings.Contains(ipStr, "]:") {
		// IPv6 mit Port: [2001:db8::1]:8080
		host, _, _ = net.SplitHostPort(ipStr)
		host = strings.Trim(host, "[]")
	}

	clientIP := net.ParseIP(host)
	if clientIP == nil {
		log.Printf("Warnung: Konnte IP-Adresse nicht parsen: %s", host)
		return false
	}

	// Konvertiere zu 4-byte IPv4 wenn möglich
	if ip4 := clientIP.To4(); ip4 != nil {
		clientIP = ip4
	}

	for _, network := range config.Cfg.Security.AllowedNetworks {
		_, ipNet, err := net.ParseCIDR(network)
		if err != nil {
			log.Printf("Warnung: Ungültiges Netzwerk in Konfiguration: %s", network)
			continue
		}
		if ipNet.Contains(clientIP) {
			return true
		}
	}

	log.Printf("Zugriff verweigert für IP %s - nicht in erlaubten Netzwerken: %v",
		host, config.Cfg.Security.AllowedNetworks)
	return false
}
