package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/janiv/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	secret         string
	polka_key      string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	jwtSecret := os.Getenv("JWT_SECRET")
	polkaKey := os.Getenv("POLKA_KEY")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("db failed to open: %s", err)
	}
	const filePathRoot = "."
	const port = "8080"
	apiCfg := apiConfig{}
	apiCfg.db = database.New(db)
	apiCfg.platform = platform
	apiCfg.secret = jwtSecret
	apiCfg.polka_key = polkaKey
	serve_mux := http.NewServeMux()
	serve_mux.HandleFunc("GET /api/healthz", handlerReadiness)
	serve_mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot)))))
	serve_mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	serve_mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	serve_mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirp)
	serve_mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	serve_mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	serve_mux.HandleFunc("GET /api/chirps/", apiCfg.handlerGetChirps)
	serve_mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirpByID)
	serve_mux.HandleFunc("POST /api/refresh", apiCfg.handlerRefresh)
	serve_mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevoke)
	serve_mux.HandleFunc("PUT /api/users", apiCfg.handlerUpdateEmailPassword)
	serve_mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChirp)
	serve_mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerSetChirpyRed)

	server := http.Server{
		Handler: serve_mux,
		Addr:    ":" + port,
	}

	log.Fatal(server.ListenAndServe())
}
