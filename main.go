package main

import (
	"encoding/json"
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
	
	// validate_chirp
	validateChirpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		genericErrorMessage := "Something went wrong"

		type successResponse struct {
			Valid bool `json:"valid"`
		}

		type errorResponse struct {
			Error string `json:"error"`
		}

		type reqBody struct {
			Body string `json:"body"`
		} 

		var decodedRedBody reqBody

		decoder := json.NewDecoder(r.Body)
		
		err := decoder.Decode(&decodedRedBody)

		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			resp, _ := json.Marshal(errorResponse{
				Error: genericErrorMessage,
			})

			w.Write(resp)
			return
		}

		isValid := len(decodedRedBody.Body) <= 140


		if !isValid {
			bodyToSend, _ := json.Marshal(errorResponse{
				Error: "Chirp is too long",
			})
		
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)
			w.Write(bodyToSend)
			return
		}		

		bodyToSend, err := json.Marshal(successResponse{
			Valid: true,
		})

		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			genericErr, _ := json.Marshal(errorResponse{
				Error: genericErrorMessage,
			})
			
			w.Write(genericErr)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(bodyToSend)
	})
	
	serveMux.Handle("POST /api/validate_chirp", apiCfg.middlewareMetricsInc(validateChirpHandler))

	metricsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html" )
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf(
`
		<html>
  		<body>
    		<h1>Welcome, Chirpy Admin</h1>
   			<p>Chirpy has been visited %d times!</p>
 			</body>
		</html>
`, apiCfg.fileserverHits.Load())))
	})
	serveMux.Handle("GET /admin/metrics", metricsHandler)

	resetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCfg.fileserverHits.Store(0)
		w.WriteHeader(200)
	})

	serveMux.Handle("POST /admin/reset", resetHandler)

	server := &http.Server{
		Addr: ":8080",
		Handler: serveMux,
	}

	server.ListenAndServe()
}