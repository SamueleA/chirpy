package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

var apiCfg = apiConfig{
	fileserverHits: atomic.Int32{},
}

func main() {
	serveMux := http.NewServeMux()

	assetsHandler := http.StripPrefix("/app/assets", http.FileServer(http.Dir("assets")))
	serveMux.Handle("/app/assets/", apiCfg.middlewareMetricsInc(assetsHandler))

	rootHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	serveMux.Handle("/app", apiCfg.middlewareMetricsInc(rootHandler))

	healthHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})
	serveMux.Handle("GET /api/healthz", apiCfg.middlewareMetricsInc(healthHandler))
	
	metricsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8" )
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf("Hits: %v", apiCfg.fileserverHits.Load())))
	})
	serveMux.Handle("GET /api/metrics", metricsHandler)

	resetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCfg.fileserverHits.Store(0)
		w.WriteHeader(200)
	})
	serveMux.Handle("POST /api/reset", resetHandler)

	server := &http.Server{
		Addr: ":8080",
		Handler: serveMux,
	}

	server.ListenAndServe()
}