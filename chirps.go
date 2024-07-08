package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/abi-liu/chirpy/internal/database"
)

func (c *apiConfig) postChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		ID   int    `json:"id"`
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	chirpSlice := strings.Fields(params.Body)
	for i, word := range chirpSlice {
		if strings.ToLower(word) == "kerfuffle" || strings.ToLower(word) == "sharbert" || strings.ToLower(word) == "fornax" {
			chirpSlice[i] = "****"
		}
	}

	c.UID++
	chirp := database.Chirp{
		ID:   c.UID,
		Body: strings.Join(chirpSlice, " "),
	}
	err = c.db.AddChirp(chirp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, returnVals{
		ID:   c.UID,
		Body: strings.Join(chirpSlice, " "),
	})
}

func (c *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	file, err := c.db.ReadFile()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	chirps := []database.Chirp{}
	for _, v := range file.Chirps {
		chirps = append(chirps, v)
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (c *apiConfig) getChirpById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		log.Printf("Cannot convert %s to type int", r.PathValue("id"))
		respondWithError(w, http.StatusInternalServerError, "Please enter a valid id")
		return
	}

	chirp, err := c.db.GetChirpById(id)
	if err != nil {
		log.Printf("Cannot find Chirp with id %d", id)
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("Cannot find chirp with id %d", id))
		return
	}

	respondWithJSON(w, http.StatusOK, chirp)
}
