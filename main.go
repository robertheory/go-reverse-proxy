package main

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type ReverseProxy struct {
	routes map[string][]string
}

func NewReverseProxy() *ReverseProxy {
	return &ReverseProxy{
		routes: map[string][]string{
			"/todos/1": {
				"http://jsonplaceholder.typicode.com",
				"http://jsonplaceholder.typicode.com",
			},
		},
	}
}

func (rp *ReverseProxy) selectBackend(path string) (string, bool) {
	backend, exists := rp.routes[path]

	if !exists || len(backend) == 0 {
		return "", false
	}

	return backend[rand.Intn(len(backend))], true
}

func (rp *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend, exists := rp.selectBackend(r.URL.Path)

	if !exists {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	remote, err := url.Parse(backend)

	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	proxyReq, err := http.NewRequest(r.Method, remote.String()+r.URL.String(), r.Body)

	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	proxyReq.Header = r.Header

	resp, err := http.DefaultClient.Do(proxyReq)

	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	defer resp.Body.Close()

	responseBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	for k, v := range resp.Header {
		w.Header()[k] = v
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(responseBytes)

	log.Printf("Request: %s, Backend: %s", r.URL.Path, backend)

}

func main() {
	rand.Seed(time.Now().UnixNano())

	rp := NewReverseProxy()
	http.Handle("/", rp)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
