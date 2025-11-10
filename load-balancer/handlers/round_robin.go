package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type RoundRobinHandler struct {
	Handler
}

func NewRoundRobinHandler(servers []Server) *RoundRobinHandler {
	serversPtrs := make([]*Server, len(servers))

	for i := range servers {
		serversPtrs[i] = &servers[i]
	}

	return &RoundRobinHandler{
		Handler: Handler{
			Servers: serversPtrs,
		},
	}
}

func (h *RoundRobinHandler) GetUrl() (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	serverCounter := len(h.Servers)
	counter := h.GetCounter()
	for range serverCounter {
		server := h.Servers[counter.index]
		if server.IsAlive {
			url := server.Url
			if server.Counter.NextAndWrap() {
				counter.Next()
			}
			return url, nil
		}
		counter.Next()
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
