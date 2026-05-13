# Go Load Balancer

A custom load balancer implementation in Go using reverse proxies, round-robin scheduling, retry handling, failover routing, and active backend health monitoring.

## Features

- Reverse proxy based request forwarding
- Round-robin load balancing
- Atomic peer rotation
- Backend health checks
- Automatic unhealthy backend detection
- Retry mechanism for failed requests
- Failover routing to alternate backends
- Context-based retry and attempt tracking
- Concurrency-safe backend state management

---

## Backend Structure

```go
type Server struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}
```

### Fields

- `URL` → Backend server URL
- `Alive` → Backend health status
- `mux` → Read-write mutex for concurrent access safety
- `ReverseProxy` → Reverse proxy instance for request forwarding

---

## Concurrency-Safe Backend State

### Read Backend Status

```go
func (s *Server) isAlive() bool
```

Uses `RLock()` for safe concurrent reads.

### Update Backend Status

```go
func (s *Server) setAlive(alive bool)
```

Uses `Lock()` for safe concurrent writes.

---

## Server Pool

```go
type ServerPool struct {
	servers []*Server
	current uint64
}
```

Maintains backend servers and the current round-robin index.

### Register Backend

```go
func (s *ServerPool) AddServer(server *Server)
```

Adds backend servers into the pool.

---

## Round Robin Load Balancing

### Next Backend Index

```go
func (s *ServerPool) NextIndex() int
```

Uses atomic operations to safely rotate backend selection.

### Get Next Alive Backend

```go
func (s *ServerPool) GetNextPeer() *Server
```

Selects the next available backend while skipping unhealthy servers.

Features:
- Cyclic traversal
- Dead backend skipping
- Atomic index synchronization

---

## Backend Health Monitoring

### TCP Health Probe

```go
func isServerAlive(u *url.URL) bool
```

Uses `net.DialTimeout()` to verify backend availability.

### Health Check Runner

```go
func (s *ServerPool) HealthCheck()
```

Checks all registered backends and updates their health status.

### Periodic Health Monitoring

```go
func HealthCheckHandler()
```

Runs backend health checks continuously using `time.Ticker`.

---

## Request Retry & Failover

### Retry Tracking

```go
func GetRetryFromContext(r *http.Request) int
```

Tracks retries for the same backend using request context.

### Attempt Tracking

```go
func GetAttemptsFromContext(r *http.Request) int
```

Tracks failover attempts across different backends.

---

## Load Balancer Handler

```go
func loadBalancer(w http.ResponseWriter, r *http.Request)
```

Responsibilities:
- Select backend server
- Forward requests through reverse proxy
- Handle unavailable backend scenarios
- Return `503 Service Unavailable` when retries exceed limit

---

## Reverse Proxy Error Handling

Custom proxy error handler implementation:

```go
proxy.ErrorHandler = func(...)
```

Features:
- Retry failed backend requests
- Mark failed backends as unhealthy
- Trigger failover routing
- Context-based retry propagation

---

## Concurrency Primitives Used

- `sync.RWMutex`
- `sync/atomic`

---

## Tech Stack

- Go
- `net/http`
- `net/http/httputil`
- `context`
- `sync`
- `sync/atomic`
- `net`

---

## Running

```bash
go run main.go --port=3030
```

The load balancer starts on the provided port and distributes requests across configured backend servers.

---

## Planned Improvements

- Graceful shutdown handling
- Config-based backend registration
- Weighted load balancing
- Structured logging
- Metrics and observability
- Circuit breaker implementation