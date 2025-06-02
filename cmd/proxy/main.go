/*
Copyright (c) 2025 Michael Lechner

This software is released under the MIT License.
See the LICENSE file for further details.
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"mlc_goproxy/internal/config"
	"mlc_goproxy/internal/proxy"
	"mlc_goproxy/internal/version"
)

func main() {
	// Command line flags
	proxyPort := flag.Int("port", 0, "Port für den Proxy-Server (überschreibt config.ini)")
	showVersion := flag.Bool("version", false, "Show version information and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("MLCProxy %s\n", version.GetVersionInfo())
		fmt.Println(version.Copyright)
		return
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Starting MLCProxy %s...", version.GetVersionInfo())

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
