package main

import (
	"encoding/json"
	"net/http"
	"strings"
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
	chirp := Chirp{
		ID:   c.UID,
		Body: strings.Join(chirpSlice, " "),
	}
	err = c.db.addChirp(chirp)
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
	file, err := c.db.readChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	chirps := []Chirp{}
	for _, v := range file.Chirps {
		chirps = append(chirps, v)
	}

	respondWithJSON(w, http.StatusOK, chirps)
}
