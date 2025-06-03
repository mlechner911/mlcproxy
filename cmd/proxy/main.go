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
	proxyPort := flag.Int("port", 0, "Port for the proxy server (overrides config.ini)")
	showVersion := flag.Bool("version", false, "Show version information and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("MLCProxy %s\n", version.GetVersionInfo())
		fmt.Println(version.Copyright)
		return
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Starting MLCProxy %s...", version.GetVersionInfo())

	// Load configuration
	if err := config.LoadConfig(); err != nil {
		log.Printf("Warning: Could not load configuration: %v", err)
		log.Println("Using default settings")
	}

	// Command line flags override configuration
	port := config.Cfg.Server.Port
	if *proxyPort != 0 {
		port = *proxyPort
	}
	// Start proxy server
	proxyAddr := fmt.Sprintf(":%d", port)
	log.Printf("Starting proxy server on port %d...", port)

	if err := proxy.Start(proxyAddr); err != nil {
		log.Printf("Error starting proxy server: %v", err)
		fmt.Println("\nPress any key to exit...")
		fmt.Scanln()
		log.Fatal("Terminating program due to error")
	}
}
