package handlers

import (
	"cmp"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"slices"
)

type LeastConnectionsHandler struct {
	Handler
	LCM uint
}

func NewLeastConnectionsHandler(servers []Server) *LeastConnectionsHandler {

	var leastCommonMultiple uint = 1
	serversPtrs := make([]*Server, len(servers))

	for i := range servers {
		serversPtrs[i] = &servers[i]
	}

	for _, server := range serversPtrs {
		if server.Weight == 0 {
			server.Weight = 1
		}
		leastCommonMultiple = leastCommonMultiple * server.Weight
	}

	for _, server := range serversPtrs {
		server.LoadScore = 0
		server.LoadCost = leastCommonMultiple / server.Weight
	}

	return &LeastConnectionsHandler{
		Handler: Handler{
			Servers: serversPtrs,
		},
		LCM: leastCommonMultiple,
	}
}

func (h *Handler) DecrementScore(server *Server) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if server.LoadScore >= server.LoadCost {
		server.LoadScore -= server.LoadCost
	}
}

func (h *LeastConnectionsHandler) GetServer() (*Server, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	slices.SortFunc(h.Servers, func(a, b *Server) int {
		return cmp.Compare(a.LoadScore, b.LoadScore)
	})

	for i := 0; i < len(h.Servers); i++ {
		server := h.Servers[i]
		if server.IsAlive {
			server.LoadScore += server.LoadCost
			return server, nil
		}
	}

	return &Server{}, fmt.Errorf("no active destinations")
}

type responseBodyWrapper struct {
	Body    io.ReadCloser // Embed the original body (so Read() works automatically)
	server  *Server
	handler *LeastConnectionsHandler
}

func (r *responseBodyWrapper) Close() error {
	r.handler.DecrementScore(r.server)
	return r.Body.Close()
}

func (r *responseBodyWrapper) Read(p []byte) (n int, err error) {
	return r.Body.Read(p)
}

func (handler *LeastConnectionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	server, err := handler.GetServer()

	if err != nil {
		w.WriteHeader(int(http.StatusServiceUnavailable))
		fmt.Fprintf(w, "All servers failed on health check")
		return
	}

	targetUrl, err := url.Parse(server.Url)
	if err != nil {
		handler.DecrementScore(server)
		log.Printf("ERROR: Could not parse server URL %s: %v", server.Url, err)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(targetUrl)

	proxy.ModifyResponse = func(res *http.Response) error {
		res.Body = &responseBodyWrapper{
			Body:    res.Body,
			server:  server,
			handler: handler,
		}
		return nil
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		log.Printf("Proxy error to %s: %v", server.Url, e)
		handler.DecrementScore(server)
		w.WriteHeader(http.StatusBadGateway)
	}

	proxy.ServeHTTP(w, r)
}
