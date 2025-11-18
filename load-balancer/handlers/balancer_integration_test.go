//go:build integration

package handlers

import (
	"emaiorov/load-balancer/config"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestHealthCheckIntegration(t *testing.T) {

	// Mock test http servers
	aliveServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	}))
	defer aliveServer.Close()

	deadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer deadServer.Close()

	//Is completely down (Connection refused)
	downServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This will never be called
	}))
	downServerURL := downServer.URL
	downServer.Close()

	h := &Handler{
		Servers: []*Server{
			{
				ServerConfig: config.ServerConfig{Url: aliveServer.URL},
				IsAlive:      false, // Start as FALSE, expect to become TRUE
			},
			{
				ServerConfig: config.ServerConfig{Url: deadServer.URL},
				IsAlive:      true, // Start as TRUE, expect to become FALSE
			},
			{
				ServerConfig: config.ServerConfig{Url: downServerURL},
				IsAlive:      true, // Start as TRUE, expect to become FALSE
			},
		},
		mu: sync.Mutex{},
	}

	// Run the HealthCheck in the background.
	go HealthCheck(h, 1)

	// We must wait for the goroutine to run its first loop.
	time.Sleep(2500 * time.Millisecond)

	//Assert the Results

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.Servers[0].IsAlive != true {
		t.Errorf("Expected aliveServer (%s) to be IsAlive=true, but got false", h.Servers[0].Url)
	}
	if h.Servers[1].IsAlive != false {
		t.Errorf("Expected deadServer (%s) to be IsAlive=false, but got true", h.Servers[1].Url)
	}
	if h.Servers[2].IsAlive != false {
		t.Errorf("Expected downServer (%s) to be IsAlive=false, but got true", h.Servers[2].Url)
	}
}

func TestRoundRobinServeHttpServersUp(t *testing.T) {

	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/data" {
			t.Errorf("Backend 1 got wrong path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Response from Backend 1"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Response from Backend 2"))
	}))
	defer backend2.Close()

	servers := []Server{
		{ServerConfig: config.ServerConfig{Url: backend1.URL}, IsAlive: true},
		{ServerConfig: config.ServerConfig{Url: backend2.URL}, IsAlive: true},
	}

	rrHandler := NewRoundRobinHandler(servers)

	req1 := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	// Create a recorder to capture the response "to the client"
	w1 := httptest.NewRecorder()

	// Execute the handler directly
	rrHandler.ServeHTTP(w1, req1)

	resp1 := w1.Result()
	body1, _ := io.ReadAll(resp1.Body)

	if resp1.StatusCode != http.StatusOK {
		t.Errorf("Request 1: Expected 200 OK, got %d", resp1.StatusCode)
	}
	if string(body1) != "Response from Backend 1" {
		t.Errorf("Request 1: Expected 'Response from Backend 1', got '%s'", string(body1))
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	w2 := httptest.NewRecorder()

	rrHandler.ServeHTTP(w2, req2)

	resp2 := w2.Result()
	body2, _ := io.ReadAll(resp2.Body)

	if string(body2) != "Response from Backend 2" {
		t.Errorf("Request 2: Expected 'Response from Backend 2', got '%s'", string(body2))
	}
}

func TestRoundRobinHandlerServeHTTPAllServersDown(t *testing.T) {

	servers := []Server{
		{ServerConfig: config.ServerConfig{Url: "http://dead1"}, IsAlive: false},
		{ServerConfig: config.ServerConfig{Url: "http://dead2"}, IsAlive: false},
	}
	rrHandler := NewRoundRobinHandler(servers)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	rrHandler.ServeHTTP(w, req)

	//Verify 503 Error
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected 503 Service Unavailable, got %d", resp.StatusCode)
	}

	expectedBody := "All servers failed on health check"
	if string(body) != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, string(body))
	}
}

func TestLeastConnectionsServeHttpServersUp(t *testing.T) {

	blocker := make(chan bool)
	//This one will WAITS until we tell it to finish.
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-blocker
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Response from Backend 1"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Response from Backend 2"))
	}))
	defer backend2.Close()

	servers := []Server{
		{ServerConfig: config.ServerConfig{Url: backend1.URL}, IsAlive: true},
		{ServerConfig: config.ServerConfig{Url: backend2.URL}, IsAlive: true},
	}

	lcHandler := NewLeastConnectionsHandler(servers)

	req1Finished := make(chan bool)
	// Create a recorder to capture the response "to the client"
	w1 := httptest.NewRecorder()

	go func() {
		req1 := httptest.NewRequest(http.MethodGet, "/api/data", nil)
		lcHandler.ServeHTTP(w1, req1)
		req1Finished <- true
	}()

	// Give the goroutine a tiny moment to start and register the load on Server 1
	time.Sleep(50 * time.Millisecond)

	req2 := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	w2 := httptest.NewRecorder()

	lcHandler.ServeHTTP(w2, req2)

	resp2 := w2.Result()
	body2, _ := io.ReadAll(resp2.Body)

	if string(body2) != "Response from Backend 2" {
		t.Errorf("Request 2 failed load balancing. Expected 'Response from Backend 2', got '%s'. (Did Req 1 finish too early?)", string(body2))
	}

	// Now that we verified Req 2 went to Server 2, we can release Server 1
	blocker <- true

	// Wait for Req 1 to actually return
	<-req1Finished

	resp1 := w1.Result()
	body1, _ := io.ReadAll(resp1.Body)
	if string(body1) != "Response from Backend 1" {
		t.Errorf("Request 1: Expected 'Response from Backend 1', got '%s'", string(body1))
	}
}

func TestLeastConnectionsHandlerServeHTTPAllServersDown(t *testing.T) {

	servers := []Server{
		{ServerConfig: config.ServerConfig{Url: "http://dead1"}, IsAlive: false},
		{ServerConfig: config.ServerConfig{Url: "http://dead2"}, IsAlive: false},
	}
	lcHandler := NewLeastConnectionsHandler(servers)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	lcHandler.ServeHTTP(w, req)

	//Verify 503 Error
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected 503 Service Unavailable, got %d", resp.StatusCode)
	}

	expectedBody := "All servers failed on health check"
	if string(body) != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, string(body))
	}
}
