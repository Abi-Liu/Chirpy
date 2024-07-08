package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/abi-liu/chirpy/internal/auth"
	"github.com/abi-liu/chirpy/internal/database"
)

func (c *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type Res struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	req := Req{}
	err := decoder.Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ok := validateRequestCredentials(w, req.Email, req.Password)
	if !ok {
		return
	}

	_, err = c.db.GetUserByEmail(req.Email)
	if err == nil {
		respondWithError(w, http.StatusConflict, "User already exists")
		return
	}

	hashedPassword, err := database.HashPassword(req.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	user, err := c.db.CreateUser(req.Email, hashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithJSON(w, http.StatusCreated, Res{
		ID:    user.ID,
		Email: user.Email,
	})
}

func (c *apiConfig) login(w http.ResponseWriter, r *http.Request) {
	type Req struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	type Res struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
		Token string `json:"token"`
	}

	decoder := json.NewDecoder(r.Body)
	req := &Req{}
	err := decoder.Decode(req)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unable to decode request body")
		return
	}

	log.Print(req.ExpiresInSeconds)

	ok := validateRequestCredentials(w, req.Email, req.Password)
	if !ok {
		return
	}

	user, err := c.db.GetUserByEmail(req.Email)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User does not exist")
		return
	}

	err = database.ComparePassword(req.Password, user.Password)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Passwords do not match")
		return
	}

	token, err := auth.GenerateToken(c.jwt, user.ID, req.ExpiresInSeconds)
	if err != nil {
		log.Printf("failed to generate JWT: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to generate JWT")
		return
	}

	respondWithJSON(w, http.StatusOK, Res{
		ID:    user.ID,
		Email: user.Email,
		Token: token,
	})
}

func (c *apiConfig) updateUserCredentials(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	log.Print(token)
	arr := strings.Split(token, "Bearer ")
	if len(arr) < 2 {
		respondWithError(w, http.StatusUnauthorized, "Token not present")
		return
	}
	token = arr[1]
	id, err := auth.ParseToken(token, c.jwt)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	log.Print(id)

	intId, err := strconv.Atoi(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to convert id to int")
		return
	}

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type Res struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	req := &Req{}
	err = decoder.Decode(req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not decode request")
	}

	if ok := validateRequestCredentials(w, req.Email, req.Password); !ok {
		return
	}

	user, err := c.db.UpdateCredentials(intId, req.Email, req.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Res{
		ID:    user.ID,
		Email: user.Email,
	})
}

func validateRequestCredentials(w http.ResponseWriter, email, password string) bool {
	if email == "" || password == "" {
		respondWithError(w, http.StatusBadRequest, "Please provide valid credentials")
		return false
	}
	return true
}
