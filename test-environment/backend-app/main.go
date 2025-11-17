package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func main() {
	http.HandleFunc("/", requestHandler)
	log.Println("Backend server starting on port 8000...")

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Just write a 200 OK status
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
		log.Printf("[%s] Health check request: %s %s", os.Getenv("SERVER_ID"), r.Method, r.URL.Path)
	})

	http.ListenAndServe(":8000", nil)
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	serverID := os.Getenv("SERVER_ID")
	if serverID == "" {
		serverID = "Unknown server"
	}

	log.Printf("[%s] Received request: %s %s", serverID, r.Method, r.URL.Path)

	sleepTime := time.Duration(rand.Intn(1000)+200) * time.Millisecond
	time.Sleep(sleepTime)

	w.Header().Set("Content-Type", "application/json")

	// Using fmt.Fprintf to write directly to the ResponseWriter
	fmt.Fprintf(w, `{"message": "Hello from %s", "duration_ms": %d}`,
		serverID, sleepTime.Milliseconds())

	//log.Printf("[%s] Finished request (took %v)", serverID, sleepTime)
}
