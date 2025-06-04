package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/janiv/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,
		`<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>`,
		cfg.fileserverHits.Load())
}
func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform == "dev" {
		cfg.fileserverHits.Store(0)
		cfg.db.Reset(r.Context())
		w.WriteHeader(200)
		return
	}
	respondWithError(w, 403, "get outta here")
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	w.Write([]byte(msg))
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling payload JSON: %s", err)
	}
	w.Write(dat)
}

func (cfg *apiConfig) handlerChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error decoding json: %s", err))
		return
	}
	if len(params.Body) > 140 {
		type returnVals struct {
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Chirp too long, calm down Shoresy",
		}
		respondWithJSON(w, 400, respBody)
		log.Printf("Chirp too long, chirp is %d characters", len(params.Body))

	} else {
		type returnVals struct {
			Cleaned_body string `json:"cleaned_body"`
		}
		respBody := returnVals{
			Cleaned_body: chirperCleaner(params.Body),
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshaling JSON: %s", err)
			respondWithError(w, 500, "something went wrong")
			return
		}
		respondWithJSON(w, 201, dat)
	}
}

func chirperCleaner(input string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}

	words := strings.Split(input, " ")
	res := make([]string, len(words))
	for idx, w := range words {
		if slices.Contains(badWords, strings.ToLower(w)) {
			w = "****"
		}
		res[idx] = w

	}
	return strings.Join(res, " ")
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type requestParameters struct {
		Email string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	params := requestParameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Something broke in json")
		return
	}
	currTime := time.Now()
	db_params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: currTime,
		UpdatedAt: currTime,
		Email:     params.Email,
	}
	usr, err := cfg.db.CreateUser(r.Context(), db_params)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("something broke in db: %s", err))
		return
	}

	type respParams struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}
	resp := respParams{
		ID:        usr.ID,
		CreatedAt: usr.CreatedAt,
		UpdatedAt: usr.UpdatedAt,
		Email:     usr.Email,
	}
	respondWithJSON(w, 201, resp)

}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("db failed to open: %s", err)
	}
	const filePathRoot = "."
	const port = "8080"
	apiCfg := apiConfig{}
	apiCfg.db = database.New(db)
	apiCfg.platform = platform
	serve_mux := http.NewServeMux()
	serve_mux.HandleFunc("GET /api/healthz", handlerReadiness)
	serve_mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot)))))
	serve_mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	serve_mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	serve_mux.HandleFunc("POST /api/validate_chirp", apiCfg.handlerChirp)
	serve_mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)

	server := http.Server{
		Handler: serve_mux,
		Addr:    ":" + port,
	}

	log.Fatal(server.ListenAndServe())
}
