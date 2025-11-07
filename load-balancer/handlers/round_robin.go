package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type RoundRobinHandler struct {
	Handler
}

type Counter struct {
	index  int
	length int
}

func (c *Counter) Next() {
	c.index++
	if c.index >= c.length {
		c.index = 0
	}
}

func NewRoundRobinHandler(servers []Server) *RoundRobinHandler {
	return &RoundRobinHandler{
		Handler: Handler{
			Servers: servers,
		},
	}
}

func (h *RoundRobinHandler) GetUrl() (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	serverCounter := len(h.Servers)

	for range serverCounter {
		counter := &h.Counter
		server := &h.Servers[counter.index]

		if server.IsAlive {
			url := server.Url
			h.Counter.Next()
			return url, nil
		}

		h.Counter.Next()
	}

	return "", fmt.Errorf("no active destinations")
}

func (handler *RoundRobinHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	destination, err := handler.GetUrl()

	if err != nil {
		w.WriteHeader(int(http.StatusServiceUnavailable))
		fmt.Fprintf(w, "All servers failed on health check")
		return
	}

	targetUrl, err := url.Parse(destination)
	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetUrl)
	proxy.ServeHTTP(w, r)
}

func healthCheck(h *RoundRobinHandler, duration int) {
	sleepTime := time.Duration(duration) * time.Second

	for {
		i := 0
		count := len(h.Servers)
		for i < count {
			resp, err := http.Get(h.Servers[i].GetHealthUrl())
			if err != nil {
				fmt.Printf("health check error: %v", err)
			}
			isAlive := (err == nil && resp.Status == "200 OK")
			h.mu.Lock()
			h.Servers[i].IsAlive = isAlive
			h.mu.Unlock()
			if !isAlive {
				fmt.Printf("server %d is down\n", i)
			}
			i++
		}

		time.Sleep(sleepTime)
	}
}
