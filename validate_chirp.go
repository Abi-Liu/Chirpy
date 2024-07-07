package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func validateChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	type req struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := req{}

	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error when decoding request body: %s\n", err)
		w.WriteHeader(500)
		return
	}

	if len(params.Body) > 140 {
		type errorRes struct {
			Error string `json:"error"`
		}
		errResponse := errorRes{Error: "Chirp is too long"}
		data, err := json.Marshal(errResponse)
		if err != nil {
			log.Printf("Error when marshalling json response: %s\n", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(400)
		w.Write(data)
		return
	}

	type success struct {
		Valid bool `json:"valid"`
	}
	res := success{Valid: true}
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("Error when marshalling json response: %s\n", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Write(data)

}
