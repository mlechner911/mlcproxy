package main

import (
	"flag"
	"fmt"
	"log"
	"mlc_goproxy/internal/gui"
	"mlc_goproxy/internal/proxy"
)

func main() {
	// Command line flags
	proxyPort := flag.Int("port", 3128, "Port für den Proxy-Server")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting MLCProxy...")

	// Create error channel for proxy
	errChan := make(chan error, 1) // Set the port in the GUI
	gui.SetPort(*proxyPort)

	// Start proxy server in a goroutine
	go func() {
		proxyAddr := fmt.Sprintf(":%d", *proxyPort)
		log.Printf("Starting proxy server on port %d...", *proxyPort)
		if err := proxy.Start(proxyAddr); err != nil {
			log.Printf("Proxy server error: %v", err)
			errChan <- err
		}
	}()

	// Start GUI in a goroutine
	go func() {
		log.Println("Starting GUI...")
		gui.Start()
		errChan <- nil
	}()

	// Wait for error or exit
	if err := <-errChan; err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
