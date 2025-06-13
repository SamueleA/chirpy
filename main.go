package main

import (
	"net/http"
)

func main() {
	serveMux := http.NewServeMux()

	serveMux.Handle("/", http.FileServer(http.Dir(".")))
	
	server := &http.Server{
		Addr: ":8080",
		Handler: serveMux,
	}

	server.ListenAndServe()
}