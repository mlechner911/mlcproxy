package main

import (
	"flag"
	"fmt"
	"log"
	"mlc_goproxy/internal/config"
	"mlc_goproxy/internal/proxy"
)

func main() {
	// Command line flags
	proxyPort := flag.Int("port", 0, "Port für den Proxy-Server (überschreibt config.ini)")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting MLCProxy...")

	// Lade Konfiguration
	if err := config.LoadConfig(); err != nil {
		log.Printf("Warnung: Konnte Konfiguration nicht laden: %v", err)
		log.Println("Verwende Standard-Einstellungen")
	}

	// Command line flags überschreiben Konfiguration
	port := config.Cfg.Server.Port
	if *proxyPort != 0 {
		port = *proxyPort
	}

	// Start proxy server
	proxyAddr := fmt.Sprintf(":%d", port)
	if err := proxy.Start(proxyAddr); err != nil {
		log.Fatalf("Proxy server error: %v", err)
	}
}
