package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type Handler struct {
	mu      sync.Mutex
	counter int
	servers []Servers
}

type Servers struct {
	mu      sync.Mutex
	isAlive bool
	url     string
}

func (s *Servers) Ready(isAlive bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.isAlive = isAlive
}

func (handler *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.mu.Lock()

	for !handler.servers[handler.counter].isAlive {
		handler.counter++
		if handler.counter >= len(handler.servers) {
			handler.counter = 0
		}
	}

	targetUrl, err := url.Parse(handler.servers[handler.counter].url)
	if err != nil {
		log.Fatal(err)
	}

	handler.counter++
	if handler.counter >= len(handler.servers) {
		handler.counter = 0
	}

	handler.mu.Unlock()

	proxy := httputil.NewSingleHostReverseProxy(targetUrl)
	proxy.ServeHTTP(w, r)
}

func main() {

	var handler = &Handler{
		counter: 0,
		servers: []Servers{
			{url: "http://localhost:9001", isAlive: true},
			{url: "http://localhost:9002", isAlive: true},
			{url: "http://localhost:9003", isAlive: true},
		},
	}

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
