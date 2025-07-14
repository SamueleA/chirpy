package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/samuelea/chirpy/internal/utils"
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

var prohibitedWords = []string{"kerfuffle", "sharbert", "fornax"}

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
			Valid bool 		`json:"valid,omitempty"`
			CleanedBody string `json:"cleaned_body,omitempty"`
		}

		type reqBody struct {
			Body string `json:"body"`
		} 

		var decodedRedBody reqBody

		decoder := json.NewDecoder(r.Body)
		
		err := decoder.Decode(&decodedRedBody)

		if err != nil {
			utils.RespondWithError(w, 500, genericErrorMessage)
			return
		}

		isValid := len(decodedRedBody.Body) <= 140


		if !isValid {
			utils.RespondWithError(w, 400, "Chirp is too long")
			return
		}		

		cleanMsg := utils.GetCorrectedString(decodedRedBody.Body, prohibitedWords)

		if cleanMsg.WasCensored {
			err = utils.RespondWithJSon(w, 200, successResponse{
				CleanedBody: cleanMsg.CorrectedMsg,
			})
		} else {
			err = utils.RespondWithJSon(w, 200, successResponse{
				CleanedBody: cleanMsg.CorrectedMsg,
			})
		}

		if err != nil {
			utils.RespondWithError(w, 500, genericErrorMessage)
		}
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
