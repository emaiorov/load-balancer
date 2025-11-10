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
	IsAlive   bool
	Counter   Counter
	LoadScore uint
	LoadCost  uint
}

type Handler struct {
	mu      sync.Mutex
	Counter Counter
	Servers []*Server
}

type Counter struct {
	index  int
	length int
}

func (c *Counter) SetLenth(length int) {
	c.length = length
}

func (h *Handler) GetCounter() *Counter {
	h.Counter.SetLenth(len(h.Servers))
	return &h.Counter
}

func (c *Counter) Next() {
	c.index++
	if c.index >= c.length {
		c.index = 0
	}
}

func (c *Counter) NextAndWrap() bool {
	c.index++
	if c.index >= c.length {
		c.index = 0
		return true
	}
	return false
}

func (s *Server) GetHealthUrl() string {
	return s.Url + s.Health
}

type LoadBalancer interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

/* Implement connection timeout */
func getClient() *http.Client {
	return &http.Client{
		Timeout: time.Duration(3) * time.Second,
	}
}

func HealthCheck(h *Handler, seconds int) {
	sleepTime := time.Duration(seconds) * time.Second

	for {
		i := 0
		count := len(h.Servers)
		for i < count {
			resp, err := getClient().Get(h.Servers[i].GetHealthUrl())
			if err != nil {
				fmt.Printf("health check error: %v", err)
				fmt.Println("")
			}
			isAlive := (err == nil && resp.Status == "200 OK")
			h.mu.Lock()
			h.Servers[i].IsAlive = isAlive
			h.mu.Unlock()
			if !isAlive {
				fmt.Printf("server %s is down\n", h.Servers[i].Url)
				fmt.Println("")
			}
			i++
		}

		time.Sleep(sleepTime)
	}
}
