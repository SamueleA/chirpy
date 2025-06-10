package main

import (
	"net/http"
)

func main() {
	serveMux := http.NewServeMux()
	
	server := http.Server{
		Addr: ":8080",
		Handler: serveMux,
	}

	server.ListenAndServe()
}