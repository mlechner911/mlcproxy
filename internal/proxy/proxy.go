package proxy

import (
	"io"
	"log"
	"mlc_goproxy/internal/stats"
	"net"
	"net/http"
	"strings"
)

type TrackingReader struct {
	r         io.Reader
	bytesRead uint64
}

func (t *TrackingReader) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	t.bytesRead += uint64(n)
	return
}

func getClientIP(r *http.Request) string {
	return strings.Split(r.RemoteAddr, ":")[0]
}

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
	clientIP := getClientIP(r)

	// Create tracking reader for request body
	var requestReader TrackingReader
	if r.Body != nil {
		requestReader = TrackingReader{r: r.Body}
		r.Body = io.NopCloser(&requestReader)
	}

	client := &http.Client{}
	req, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	copyHeader(req.Header, r.Header)
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	// Create tracking reader for response body
	responseReader := &TrackingReader{r: resp.Body}
	_, err = io.Copy(w, responseReader)
	if err != nil {
		log.Printf("Error copying response: %v", err)
	}

	// Log transfer statistics
	stats.LogTransfer(clientIP, requestReader.bytesRead, responseReader.bytesRead)
}

func handleHTTPS(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)

	destConn, err := net.Dial("tcp", r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer destConn.Close()

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))

	// Create tracking readers for both directions
	clientReader := &TrackingReader{r: clientConn}
	destReader := &TrackingReader{r: destConn}

	// Proxy data in both directions
	go func() {
		io.Copy(destConn, clientReader)
	}()
	io.Copy(clientConn, destReader)

	// Log transfer statistics
	stats.LogTransfer(clientIP, clientReader.bytesRead, destReader.bytesRead)
}
