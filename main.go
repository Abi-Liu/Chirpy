package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

type apiConfig struct {
	fileserverHits int
	UID            int
	file           *os.File
}

func createUIDClosure() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

func main() {
	mux := http.NewServeMux()
	appConfig := &apiConfig{}

	// create in memory json file to store chirps
	file, err := os.Create("database.json")
	if err != nil {
		log.Println("Failed to create database.json\nShutting down...")
	}
	appConfig.file = file

	mux.Handle("/app/", appConfig.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", getHealthCheck)
	mux.HandleFunc("GET /admin/metrics", appConfig.getMetrics)
	mux.HandleFunc("GET /api/reset", appConfig.resetMetrics)
	mux.HandleFunc("POST /api/chirps", appConfig.postChirp)
	mux.HandleFunc("GET /api/chirps", getChirps)

	server := &http.Server{Addr: ":8080", Handler: mux}

	fmt.Println("Server starting on port 8080")
	log.Fatal(server.ListenAndServe())
}

func getHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (c *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.fileserverHits++
		fmt.Println(c.fileserverHits)
		next.ServeHTTP(w, r)
	})
}

func (c *apiConfig) getMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`<html>

<body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
</body>

</html>`, c.fileserverHits)))
}

func (c *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	c.fileserverHits = 0
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
