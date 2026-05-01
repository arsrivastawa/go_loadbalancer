package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

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
