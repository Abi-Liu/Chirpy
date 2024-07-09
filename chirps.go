package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/abi-liu/chirpy/internal/auth"
	"github.com/abi-liu/chirpy/internal/database"
)

func (c *apiConfig) postChirp(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	arr := strings.Split(authHeader, "Bearer ")
	if len(arr) < 2 {
		respondWithError(w, http.StatusUnauthorized, "Token not provided")
		return
	}
	tokenStr := arr[1]

	stringId, err := auth.ParseToken(tokenStr, c.jwt)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token expired")
		return
	}

	id, err := strconv.Atoi(stringId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot convert to int")
		return
	}

	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		ID       int    `json:"id"`
		Body     string `json:"body"`
		AuthorId int    `json:"author_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
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
		ID:       c.UID,
		Body:     strings.Join(chirpSlice, " "),
		AuthorId: id,
	}
	err = c.db.AddChirp(chirp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, returnVals{
		ID:       c.UID,
		Body:     strings.Join(chirpSlice, " "),
		AuthorId: id,
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

func (c *apiConfig) deleteChirpById(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	arr := strings.Split(authHeader, "Bearer ")
	if len(arr) < 2 {
		respondWithError(w, http.StatusUnauthorized, "Token missing")
		return
	}
	token := arr[1]
	userId, err := auth.ParseToken(token, c.jwt)

	intUser, _ := strconv.Atoi(userId)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	chirpId := r.PathValue("id")
	intId, err := strconv.Atoi(chirpId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Please enter a valid chirp id")
		return
	}

	chirp, err := c.db.GetChirpById(intId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Chirp not found")
		return
	}

	if chirp.AuthorId != intUser {
		respondWithError(w, http.StatusForbidden, "You are not allowed to delete this chirp")
		return
	}

	err = c.db.DeleteChirpById(intId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete chirp")
		return
	}

	respondWithJSON(w, http.StatusNoContent, ``)
}
