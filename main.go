package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type apiConfig struct {
	fileserverHits int
}

func main() {
	mux := http.NewServeMux()
	appConfig := &apiConfig{fileserverHits: 0}
	mux.Handle("/app/", appConfig.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("/healthz", getHealthCheck)
	mux.HandleFunc("/metrics", appConfig.getMetrics)
	mux.HandleFunc("/reset", appConfig.resetMetrics)

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
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	res := strconv.Itoa(c.fileserverHits)
	w.Write([]byte("Hits: " + res))
}

func (c *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	c.fileserverHits = 0
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
