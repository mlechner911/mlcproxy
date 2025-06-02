package proxy

import (
	"io"
	"log"
	"mlc_goproxy/internal/stats"
	"net"
	"net/http"
)

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func Start(addr string) error {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	proxy := &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Incoming request: %s %s", r.Method, r.URL.String())

			// Log request for statistics
			stats.LogRequest(r)

			if r.Method == http.MethodConnect {
				log.Printf("Handling HTTPS request to: %s", r.URL.Host)
				// HTTPS proxy handling
				handleHTTPS(w, r)
			} else {
				log.Printf("Handling HTTP request to: %s", r.URL.String())
				// HTTP proxy handling
				handleHTTP(w, r)
			}
		}),
	}

	log.Printf("Starting proxy server on %s", addr)
	return proxy.ListenAndServe()
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	// Create a new client for each request
	client := &http.Client{}

	// Create a new request
	req, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy headers
	copyHeader(req.Header, r.Header)

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error copying response: %v", err)
	}
}

func handleHTTPS(w http.ResponseWriter, r *http.Request) {
	// Try to connect to the destination
	destConn, err := net.Dial("tcp", r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer destConn.Close()

	// Hijack the connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	// Get the client connection
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	// Send 200 OK to client
	clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))

	// Start proxying data
	go func() {
		_, err := io.Copy(destConn, clientConn)
		if err != nil {
			log.Printf("Error copying to dest: %v", err)
		}
	}()

	_, err = io.Copy(clientConn, destConn)
	if err != nil {
		log.Printf("Error copying to client: %v", err)
	}
}
