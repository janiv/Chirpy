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

type ChirpResp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
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
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
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
		return
	}
	usr, err := cfg.db.GetUserByID(r.Context(), params.UserID)
	if err != nil {
		respondWithError(w, 403, "scram")
		log.Printf("Error occurred: %s", err)
		return
	}
	curr_time := time.Now()
	chirp_params := database.CreateChirpParams{
		ID:        uuid.New(),
		CreatedAt: curr_time,
		UpdatedAt: curr_time,
		Body:      chirperCleaner(params.Body),
		UserID:    usr.ID,
	}
	chrp, err := cfg.db.CreateChirp(r.Context(), chirp_params)
	if err != nil {
		respondWithError(w, 500, "oops")
		log.Fatal(err)
		return
	}
	type returnVals struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}
	respBody := returnVals{
		ID:        chrp.ID,
		CreatedAt: chrp.CreatedAt,
		UpdatedAt: chrp.UpdatedAt,
		Body:      chrp.Body,
		UserID:    chrp.UserID,
	}
	respondWithJSON(w, 201, respBody)

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

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, 500, "chirps bottled")
		return
	}
	var resp []ChirpResp
	for _, c := range chirps {
		curr := ChirpResp{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserID:    c.UserID,
		}
		resp = append(resp, curr)
	}
	respondWithJSON(w, 200, resp)
}

func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	path_param := r.PathValue("chirpID")
	chirp_id, err := uuid.Parse(path_param)
	if err != nil {
		respondWithError(w, 500, "could not parse id")
		return
	}
	chirp, err := cfg.db.GetChirpByID(r.Context(), chirp_id)
	if err != nil {
		respondWithError(w, 404, "chirp no here")
		return
	}
	resp := ChirpResp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	respondWithJSON(w, 200, resp)
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
	serve_mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirp)
	serve_mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	serve_mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	serve_mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirpByID)

	server := http.Server{
		Handler: serve_mux,
		Addr:    ":" + port,
	}

	log.Fatal(server.ListenAndServe())
}
