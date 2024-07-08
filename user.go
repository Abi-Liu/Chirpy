package main

import (
	"encoding/json"
	"net/http"

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

	if req.Email == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Please provide valid credentials")
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
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type Res struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	req := &Req{}
	err := decoder.Decode(req)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unable to decode request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Please provide valid credentials")
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

	respondWithJSON(w, http.StatusOK, Res{
		ID:    user.ID,
		Email: user.Email,
	})
}
