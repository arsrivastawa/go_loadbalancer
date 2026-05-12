# Go Load Balancer

A custom load balancer implementation in Go using reverse proxies, round-robin server selection, health checks, and concurrency-safe backend state management.

## Current Features

- Backend server abstraction
- Reverse proxy integration
- Server pool management
- Thread-safe backend health state handling
- Round-robin backend selection
- Atomic counter based peer rotation
- TCP-based health checks
- Backend availability monitoring
- Backend status update mechanism

---

## Backend Structure

Each backend server is represented using:

```go
type Server struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}
```

### Fields

- `URL` → Backend server address
- `Alive` → Backend availability status
- `mux` → Read-write mutex for concurrency safety
- `ReverseProxy` → Reverse proxy instance for request forwarding

---

## Concurrency-Safe State Management

### Read Backend Status

```go
func (s *Server) isAlive() bool
```

Uses `RLock()` for concurrent-safe reads.

### Update Backend Status

```go
func (s *Server) setAlive(alive bool)
```

Uses `Lock()` for concurrent-safe writes.

---

## Server Pool

```go
type ServerPool struct {
	servers []*Server
	current uint64
}
```

Maintains the collection of backend servers and the current round-robin index.

### Add Backend

```go
func (s *ServerPool) AddServer(server *Server)
```

Registers a backend into the pool.

---

## Backend Status Management

```go
func (s *ServerPool) MarkBackendStatus(backendUrl *url.URL, alive bool)
```

Updates the health status of a backend by matching its host.

---

## Round Robin Load Balancing

### Next Backend Index

```go
func (s *ServerPool) NextIndex() int
```

Uses atomic operations to safely increment and retrieve the next backend index.

### Get Next Available Backend

```go
func (s *ServerPool) GetNextPeer() *Server
```

Selects the next alive backend server using round-robin scheduling.

Features:
- Cyclic backend traversal
- Dead backend skipping
- Atomic index updates

---

## Health Checks

### Backend Health Probe

```go
func isServerAlive(u *url.URL) bool
```

Uses `net.DialTimeout()` to verify backend TCP connectivity.

### Pool Health Monitoring

```go
func (s *ServerPool) HealthCheck()
```

Checks all registered backends and updates their availability status.

---

## Concurrency Primitives Used

- `sync.RWMutex`
- `sync/atomic`

---

## Tech Stack

- Go
- `net/http`
- `net/http/httputil`
- `sync`
- `sync/atomic`
- `net`

---

## Planned Features

- Request forwarding
- Retry mechanism
- Automatic failover
- Context-based retry tracking
- Active health monitoring scheduler
- Reverse proxy error handling