# Go HTTP Load Balancer

[![codecov](https://codecov.io/github/emaiorov/load-balancer/graph/badge.svg?token=6GLC4AILXX)](https://codecov.io/github/emaiorov/load-balancer)

A simple, concurrent HTTP load balancer written in Go, built as a learning project.

## Features
* **Round Robin Load Balancing:** Distributes requests evenly across multiple backend servers.
* **Concurrent & Fast:** Uses Go's concurrency primitives (`sync.Mutex`) to handle thousands of requests in parallel without race conditions.
* **(WIP) Health Checks:** (You can add this here once you build it)

---

## 🚀 How to Run

This project includes the load balancer and a complete Docker-based test environment.

**Prerequisites:**
* [Go](https://go.dev/doc/install) (1.23+)
* [Docker](https://www.docker.com/get-started/)

### 1. Start the Backend Servers

The `test-environment` folder contains three simple Go servers. `docker-compose` will build and run all three.

```bash
# From the project root
cd test-environment/
docker-compose up --build
```
This will start 3 servers on:

http://localhost:9001  
http://localhost:9002  
http://localhost:9003  

### 2. Start the Load Balancer

In a new terminal, run the load balancer Go app:
```bash
# From the project root
cd load-balancer/
go run .
```

The load balancer is now running on http://localhost:8080.

### 3. Test it!

You can now send traffic to the load balancer and watch it get distributed.

```bash
# Send a single request
curl http://localhost:8080
# {"message": "Hello from Server-1", ...}

# Send 100 requests at once
hey -n 100 http://localhost:8080
```