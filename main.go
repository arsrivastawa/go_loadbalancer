package main

import (
	"context"
	"flag"
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

func HealthCheckHandler() {
	t := time.NewTicker(2 * time.Second)
	for range t.C {
		serverPool.HealthCheck()
	}
}

func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 1
}

func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(Retry).(int); ok {
		return retry
	}
	return 0
}

func loadBalancer(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)

	if attempts > 3 {
		fmt.Printf("Max attempts reached for request, addr: %s, url: %s", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
	}

	server := serverPool.GetNextPeer()
	if server != nil {
		server.ReverseProxy.ServeHTTP(w, r)
		return
	}

	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

var serverPool ServerPool

func main() {
	var port int
	flag.IntVar(&port, "port", 3030, "Port on which loadbalncer will serve")
	flag.Parse()

	serverList := []string{}

	serverList = append(serverList, "http://localhost:8079")
	serverList = append(serverList, "http://localhost:8080")
	serverList = append(serverList, "http://localhost:8081")

	if len(serverList) == 0 {
		fmt.Println("Please provide one or more servers to loadbalancer")
	}

	for _, server := range serverList {
		serverUrl, err := url.Parse(server)

		if err != nil {
			println("Unable to parse serve URL, please provide valid URL")
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(serverUrl)

		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
			fmt.Printf("server [%s] failed, error: %s", serverUrl.Host, e.Error())

			retries := GetRetryFromContext(r)
			if retries < 3 {
				time.Sleep(10 * time.Millisecond)
				ctx := context.WithValue(r.Context(), Retry, retries+1)
				r = r.WithContext(ctx)
				proxy.ServeHTTP(w, r)

				return
			}

			serverPool.MarkBackendStatus(serverUrl, false)

			attempts := GetAttemptsFromContext(r)

			fmt.Println("Attempting retry for the request")

			ctx := context.WithValue(r.Context(), Attempts, attempts+1)

			r = r.WithContext(ctx)

			loadBalancer(w, r)
		}

		serverPool.AddServer(&Server{
			URL:          serverUrl,
			Alive:        true,
			ReverseProxy: proxy,
		})
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(loadBalancer),
	}

	HealthCheckHandler()

	fmt.Println(server.Addr)
	fmt.Printf("Starting loadbalancer at :%d\n", port)

	if err := server.ListenAndServe(); err != nil {
		fmt.Println("failed to start loadbalancer")
		log.Fatal(err)
	}
}
