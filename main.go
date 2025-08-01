package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/samuelea/chirpy/internal/auth"
	"github.com/samuelea/chirpy/internal/database"
	"github.com/samuelea/chirpy/internal/utils"
)

var genericErrorMessage string =  "Something went wrong"

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
	godotenv.Load()
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	dbQueries := database.New(db)

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
	
	createChirp := http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		type successResponse struct {
			Id				uuid.UUID	`json:"id"`
			UserId		uuid.UUID	`json:"user_id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Body			string		`json:"body"`
		}

		type reqBody struct {
			Body 		string `json:"body"`
			UserId	string `json:"user_id"`
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

		parsedUUID, err := uuid.Parse(decodedRedBody.UserId)

		if err != nil {
			utils.RespondWithError(w, 400, "invalid user id")
			return
		}

		chirp, err := dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
			UserID: parsedUUID,
			Body: cleanMsg.CorrectedMsg,
		})

		if err != nil {
			utils.RespondWithError(w, 500, genericErrorMessage)
		}

		utils.RespondWithJSon(w, 201, successResponse{
			Id: chirp.ID,
			UserId: chirp.UserID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
		})
	})

	serveMux.Handle("POST /api/chirps", apiCfg.middlewareMetricsInc(createChirp))

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
		environment := os.Getenv("PLATFORM")
		
		if environment == "dev" {
			err := dbQueries.ClearUsers(r.Context())
			if err != nil {
				utils.RespondWithError(w, 500, "Failed to reset users")
			}
		}

		apiCfg.fileserverHits.Store(0)
		w.WriteHeader(200)
	})


	createUser := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type Input struct {
			Email 		string `json:"email"`
			Password 	string `json:"password"`
		}	

		decoder := json.NewDecoder(r.Body)

		var decodedInput Input

		err := decoder.Decode(&decodedInput)

		if err != nil {
			utils.RespondWithError(w, 400, "Wrong input data")
			return
		}

		hashedPassword, err := auth.HashPassword(decodedInput.Password)
		
		if err != nil {
			utils.RespondWithError(w, 500, genericErrorMessage)
			return
		}

		user, err := dbQueries.CreateUser(r.Context(), database.CreateUserParams{
			Email: decodedInput.Email,
			HashedPassword: hashedPassword,
		})

		if err != nil {
			utils.RespondWithError(w, 500, "a user with that email already exists")
			return
		}

		type CreateUserResponse struct {
			ID 				uuid.UUID `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt	time.Time	`json:"updated_at"`
			Email			string		`json:"email"`
		}

		utils.RespondWithJSon(w, 201, CreateUserResponse{
			ID: user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email: user.Email,
		})
	})

	serveMux.Handle("POST /api/users", apiCfg.middlewareMetricsInc(createUser))
	
	login := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type body struct {
			Email			string	`json:"email"`
			Password	string	`json:"password"`
		}

		type sucessResponse struct {
			Id				uuid.UUID	`json:"id"`
			UpdatedAt	time.Time	`json:"updated_at"`
			CreatedAt	time.Time	`json:"created_at"`
			Email			string		`json:"email"`
		}

		decoder := json.NewDecoder(r.Body)
		
		var decodedBody body

		err := decoder.Decode(&decodedBody)

		if err != nil {
			utils.RespondWithError(w, 400, genericErrorMessage)
			return
		}

		user, err := dbQueries.GetUser(r.Context(), decodedBody.Email)	

		if err != nil {
			utils.RespondWithError(w, 401, "Invalid email or password")
		}

		err = auth.CheckPasswordHash(decodedBody.Password, user.HashedPassword)

		if err != nil {
			utils.RespondWithError(w, 401, "Invalid email or password")
		}

		response := sucessResponse{
			Id: user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email: user.Email,
		}

		utils.RespondWithJSon(w, 200, response)
	})

	serveMux.Handle("POST /api/login", apiCfg.middlewareMetricsInc(login))

	getChirps := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type successResponse struct {
			Id				uuid.UUID	`json:"id"`
			UserId		uuid.UUID	`json:"user_id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Body			string		`json:"body"`
		}

		type response []successResponse

		chirps, err := dbQueries.GetChirps(r.Context())

		if err != nil {
			utils.RespondWithError(w, 500, genericErrorMessage)
		}

		var chirpList response
		for i := range(len(chirps)) {
			chirp := chirps[i]
			chirpList = append(chirpList, successResponse{
				Id: chirp.ID,
				UserId: chirp.UserID,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
				Body: chirp.Body,
			})
		}

		utils.RespondWithJSon(w, 200, &chirpList)
	})

	serveMux.Handle("GET /api/chirps", apiCfg.middlewareMetricsInc(getChirps))

	getChirp := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chirpID := r.PathValue("chirpID")

		uChirpID, err := uuid.Parse(chirpID)

		if err != nil {
			utils.RespondWithError(w, 400, "invalid id")
			return
		}
		
		chirp, err := dbQueries.GetChirp(r.Context(), uChirpID)

		if err != nil {
			utils.RespondWithError(w, 404, "user not found")
			return
		}

		type successResponse struct {
			Id				uuid.UUID	`json:"id"`
			CreatedAt	time.Time `json:"created_at"`
			UpdatedAt time.Time	`json:"updated_at"`
			Body 			string 		`json:"body"`	
			UserId 		uuid.UUID `json:"user_id"`
		}

		res := successResponse{
			Id: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserId: chirp.UserID,
		}

		utils.RespondWithJSon(w, 200, res)
	})
	
	serveMux.Handle("GET /api/chirps/{chirpID}", apiCfg.middlewareMetricsInc(getChirp))

	serveMux.Handle("POST /admin/reset", resetHandler)
	
	server := &http.Server{
		Addr: ":8080",
		Handler: serveMux,
	}

	server.ListenAndServe()
}


