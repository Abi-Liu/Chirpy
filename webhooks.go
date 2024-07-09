package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func (c *apiConfig) receiveWebhook(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	arr := strings.Split(authHeader, "ApiKey ")
	if len(arr) < 2 {
		log.Print("missing key")
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	token := arr[1]
	if token != c.polka {
		log.Printf("Not equal - token %s\n received %s", c.polka, token)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	type Req struct {
		Event string `json:"event"`
		Data  struct {
			UserId int `json:"user_id"`
		} `json:"data"`
	}

	type Res struct{}

	decoder := json.NewDecoder(r.Body)
	req := Req{}
	err := decoder.Decode(&req)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to decode params")
		return
	}

	switch req.Event {
	case "user.upgraded":
		// upgrade user
		user, err := c.db.FindUserById(req.Data.UserId)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "User not found")
			return
		}

		err = c.upgradeUser(user.Email)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error updating user")
			return
		}

		respondWithJSON(w, http.StatusNoContent, Res{})
		return
	default:
		respondWithJSON(w, http.StatusNoContent, Res{})
		return
	}
}

func (c *apiConfig) upgradeUser(email string) error {
	err := c.db.UpgradeUser(email)
	if err != nil {
		return err
	}
	return nil
}
