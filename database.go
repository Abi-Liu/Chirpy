package main

import (
	"encoding/json"
	"os"
	"sync"
)

type DB struct {
	path string
	mu   *sync.RWMutex
}

type File struct {
	Chirps map[int]Chirp `json:"chirps"`
}

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

func (db *DB) readChirps() (File, error) {
	db.mu.RLock()
	data, err := os.ReadFile("database.json")
	if err != nil {
		return File{}, err
	}
	db.mu.RUnlock()

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

func (db *DB) addChirp(chirp Chirp) error {
	file, err := db.readChirps()

	file.Chirps[chirp.ID] = chirp

	db.mu.Lock()
	defer db.mu.Unlock()
	data, err := json.Marshal(file)
	if err != nil {
		return err
	}

	os.WriteFile("database.json", data, 0666)

	return nil
}

func createDB(path string) (*DB, error) {
	_, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	db := &DB{path: path, mu: &sync.RWMutex{}}
	return db, nil
}
