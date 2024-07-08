package main

import (
	"encoding/json"
	"os"
)

type File struct {
	Chirps map[int]Chirp `json:"chirps"`
}

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

func readChirps() (File, error) {
	data, err := os.ReadFile("database.json")
	if err != nil {
		return File{}, err
	}

	if len(data) == 0 {
		file := File{Chirps: map[int]Chirp{}}
		return file, nil
	}
	file := File{}
	err = json.Unmarshal(data, &file)
	if err != nil {
		return File{}, err
	}

	return file, nil
}

func addChirp(chirp Chirp) error {
	file, err := readChirps()

	file.Chirps[chirp.ID] = chirp

	data, err := json.Marshal(file)
	if err != nil {
		return err
	}

	os.WriteFile("database.json", data, 0666)

	return nil
}
