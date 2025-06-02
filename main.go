package main

import (
	"log"
	"net/http"
)

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func main() {
	const filePathRoot = "."
	const port = "8080"

	serve_mux := http.NewServeMux()
	serve_mux.HandleFunc("/healthz", handlerReadiness)
	serve_mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot))))

	server := http.Server{
		Handler: serve_mux,
		Addr:    ":" + port,
	}

	log.Fatal(server.ListenAndServe())
}
