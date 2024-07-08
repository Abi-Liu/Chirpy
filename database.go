package main

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
	Users  map[int]User
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

func (db *DB) readFile() (File, error) {
	db.mu.RLock()
	data, err := os.ReadFile("database.json")
	if err != nil {
		return File{}, err
	}
	db.mu.RUnlock()

	if len(data) == 0 {
		file := File{
			Chirps: map[int]Chirp{},
			Users:  map[int]User{},
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

func (db *DB) addChirp(chirp Chirp) error {
	file, err := db.readFile()

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

func (db *DB) getChirpById(id int) (Chirp, error) {
	file, err := db.readFile()
	if err != nil {
		return Chirp{}, err
	}

	chirp, ok := file.Chirps[id]
	if !ok {
		return Chirp{}, errors.New("Not found")
	}

	return chirp, nil
}

func (db *DB) createUser(email string, password string) (User, error) {
	file, err := db.readFile()
	if err != nil {
		return User{}, errors.New("Could not unmarshal file")
	}
	db.mu.Lock()
	defer db.mu.Unlock()

	id := len(file.Users) + 1
	user := User{
		ID:       id,
		Email:    email,
		Password: password,
	}
	file.Users[id] = user

	data, err := json.Marshal(file)
	if err != nil {
		return User{}, errors.New("Failed to marshal user object")
	}

	os.WriteFile(db.path, data, 0666)
	return user, nil
}

func createDB(path string) (*DB, error) {
	_, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	db := &DB{path: path, mu: &sync.RWMutex{}}
	return db, nil
}
