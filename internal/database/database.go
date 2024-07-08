package database

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

type DB struct {
	path string
	mu   *sync.RWMutex
}

type File struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[string]User
	Tokens map[int]string
}

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

func CreateDB(path string) (*DB, error) {
	_, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	db := &DB{path: path, mu: &sync.RWMutex{}}
	return db, nil
}

func (db *DB) ReadFile() (File, error) {
	db.mu.RLock()
	data, err := os.ReadFile("database.json")
	if err != nil {
		return File{}, err
	}
	db.mu.RUnlock()

	if len(data) == 0 {
		file := File{
			Chirps: map[int]Chirp{},
			Users:  map[string]User{},
			Tokens: map[int]string{},
		}
		return file, nil
	}
	file := File{}
	err = json.Unmarshal(data, &file)
	if err != nil {
		return File{}, err
	}

	return file, nil
}

func (db *DB) AddChirp(chirp Chirp) error {
	file, err := db.ReadFile()

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

func (db *DB) GetChirpById(id int) (Chirp, error) {
	file, err := db.ReadFile()
	if err != nil {
		return Chirp{}, err
	}

	chirp, ok := file.Chirps[id]
	if !ok {
		return Chirp{}, errors.New("Not found")
	}

	return chirp, nil
}
