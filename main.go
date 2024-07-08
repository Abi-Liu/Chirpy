package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/abi-liu/chirpy/internal/database"
	"github.com/joho/godotenv"
)

type apiConfig struct {
	fileserverHits int
	UID            int
	db             *database.DB
	jwt            string
}

func createUIDClosure() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load env")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	mux := http.NewServeMux()
	appConfig := &apiConfig{}
	appConfig.jwt = jwtSecret

	db, err := database.CreateDB("database.json")
	if err != nil {
		log.Printf("Failed to create database: %s", err.Error())
	}
	appConfig.db = db

	mux.Handle("/app/", appConfig.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", getHealthCheck)
	mux.HandleFunc("GET /admin/metrics", appConfig.getMetrics)
	mux.HandleFunc("GET /api/reset", appConfig.resetMetrics)
	mux.HandleFunc("POST /api/chirps", appConfig.postChirp)
	mux.HandleFunc("GET /api/chirps", appConfig.getChirps)
	mux.HandleFunc("GET /api/chirps/{id}", appConfig.getChirpById)
	mux.HandleFunc("POST /api/users", appConfig.createUser)
	mux.HandleFunc("POST /api/login", appConfig.login)
	mux.HandleFunc("PUT /api/users", appConfig.updateUserCredentials)

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
