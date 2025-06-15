package main

import (
	"net/http"
)

func main() {
	serveMux := http.NewServeMux()

	serveMux.Handle("/app/assets/", http.StripPrefix("/app/assets", http.FileServer(http.Dir("assets"))))
	serveMux.Handle("/app", http.StripPrefix("/app", http.FileServer(http.Dir("."))))

	serveMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})
	
	server := &http.Server{
		Addr: ":8080",
		Handler: serveMux,
	}

	server.ListenAndServe()
}