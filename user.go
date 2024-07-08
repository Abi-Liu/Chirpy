package main

import (
	"encoding/json"
	"net/http"
)

func (c *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	req := Req{}
	err := decoder.Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	user, err := c.db.createUser(req.Email, req.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error when saving to database")
		return
	}

	respondWithJSON(w, http.StatusCreated, user)
}
