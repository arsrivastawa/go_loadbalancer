package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

const (
	Attempts int = 0
	Retry    int = 0
)

type Server struct {
	URL          string
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
		if b.URL == backendUrl.String() {
			b.setAlive(alive)
			break
		}
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
