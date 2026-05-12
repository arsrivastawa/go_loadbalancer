package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

const (
	Attempts int = 0
	Retry    int = 0
)

type Server struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func (s *Server) isAlive() bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.Alive
}

func (s *Server) setAlive(alive bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.Alive = alive
}

type ServerPool struct {
	servers []*Server
	current uint64
}

func (s *ServerPool) AddServer(server *Server) {
	s.servers = append(s.servers, server)
}

func (s *ServerPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range s.servers {
		if b.URL.Host == backendUrl.Host {
			b.setAlive(alive)
			break
		}
	}
}

func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.servers)))
}

func (s *ServerPool) GetNextPeer() *Server {
	next := s.NextIndex()
	l := len(s.servers) + next

	for i := next; i < l; i++ {
		idx := i % len(s.servers)
		if s.servers[idx].isAlive() {
			if idx != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return s.servers[idx]
		}
	}
	return nil
}

func isServerAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)

	if err != nil {
		return false
	}

	defer conn.Close()
	return true
}

func (s *ServerPool) HealthCheck() {
	for _, s := range s.servers {
		alive := isServerAlive(s.URL)
		s.setAlive(alive)

		status := "down"
		if alive {
			status = "up"
		}

		fmt.Printf("server:%s status:%s", s.URL, status)
	}
}

func requestHandler(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get("id")

	response := map[string]string{
		"message": "success",
		"name":    "Aditya",
	}

	switch req.Method {
	case "GET":
		json.NewEncoder(w).Encode(response)
	case "POST":
		json.NewEncoder(w).Encode(response)
	default:
		fmt.Fprintf(w, "Hello for different type of request your user id is %s", id)
	}
}

func main() {

	http.HandleFunc("/hello", requestHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
