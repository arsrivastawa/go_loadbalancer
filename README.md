# Go Load Balancer

A custom load balancer implementation in Go using reverse proxies, backend pooling, and concurrency-safe backend state management.

## Current Features

- Backend server abstraction
- Reverse proxy integration
- Server pool management
- Thread-safe backend health state handling
- Backend status update mechanism
- Foundation for retry and failover handling

---

## Backend Structure

Each backend server is represented using:

```go
type Server struct {
	URL          string
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

Maintains the collection of registered backend servers.

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

Updates the health status of a backend by matching its URL.

---

## Tech Stack

- Go
- `net/http`
- `net/http/httputil`
- `sync.RWMutex`

---

## Planned Features

- Round-robin load balancing
- Retry mechanism
- Health checks
- Automatic failover
- Request routing
- Context-based retry tracking
- Active backend monitoring