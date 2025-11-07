package handlers

import (
	"emaiorov/load-balancer/config"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	config.ServerConfig
	IsAlive bool
}

type Handler struct {
	mu      sync.Mutex
	Counter Counter
	Servers []Server
}

func (s *Server) GetHealthUrl() string {
	return s.Url + s.Health
}

type LoadBalancer interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

func HealthCheck(h *Handler, duration int) {
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
