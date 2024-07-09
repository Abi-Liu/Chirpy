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
		ID          int    `json:"id"`
		Email       string `json:"email"`
		IsChirpyRed bool   `json:"is_chirpy_red"`
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
		ID:          user.ID,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}

func (c *apiConfig) login(w http.ResponseWriter, r *http.Request) {
	type Req struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	type Res struct {
		ID           int    `json:"id"`
		Email        string `json:"email"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		IsChirpyRed  bool   `json:"is_chirpy_red"`
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

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	err = c.db.UpdateRefreshToken(user.ID, refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to save refresh token to db")
		return
	}

	respondWithJSON(w, http.StatusOK, Res{
		ID:           user.ID,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
		IsChirpyRed:  user.IsChirpyRed,
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
		ID          int    `json:"id"`
		Email       string `json:"email"`
		IsChirpyRed bool   `json:"is_chirpy_red"`
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
		ID:          user.ID,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}

func (c *apiConfig) refreshToken(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	arr := strings.Split(authHeader, "Bearer ")
	if len(arr) < 2 {
		respondWithError(w, http.StatusUnauthorized, "Token not provided")
		return
	}
	tokenStr := arr[1]

	token, err := c.db.LookupToken(tokenStr)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token does not exist")
		return
	}

	err = database.CheckTokenExpiration(token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token expired")
		return
	}

	jwt, err := auth.GenerateToken(c.jwt, token.ID, 60*60)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to generate access token")
		return
	}

	type Res struct {
		Token string `json:"token"`
	}

	respondWithJSON(w, http.StatusOK, Res{Token: jwt})
}

func (c *apiConfig) revokeToken(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	arr := strings.Split(authHeader, "Bearer ")
	if len(arr) < 2 {
		respondWithError(w, http.StatusUnauthorized, "Token not provided")
		return
	}
	tokenStr := arr[1]

	c.db.DeleteToken(tokenStr)

	respondWithJSON(w, http.StatusNoContent, "")
}

func validateRequestCredentials(w http.ResponseWriter, email, password string) bool {
	if email == "" || password == "" {
		respondWithError(w, http.StatusBadRequest, "Please provide valid credentials")
		return false
	}
	return true
}
