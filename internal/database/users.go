package database

import (
	"encoding/json"
	"errors"
	"os"
)

var ErrUserAlreadyExists = errors.New("User already exists")

func (db *DB) GetUserByEmail(email string) (User, error) {
	file, err := db.ReadFile()
	if err != nil {
		return User{}, err
	}

	user, ok := file.Users[email]
	if !ok {
		return User{}, errors.New("User does not exist")
	}

	return user, nil
}

func (db *DB) CreateUser(email string, password string) (User, error) {
	file, err := db.ReadFile()
	if err != nil {
		return User{}, err
	}

	_, err = db.GetUserByEmail(email)
	if err == nil {
		return User{}, ErrUserAlreadyExists
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	id := len(file.Users) + 1

	user := User{
		ID:       id,
		Email:    email,
		Password: password,
	}

	file.Users[email] = user

	data, err := json.Marshal(file)
	if err != nil {
		return User{}, errors.New("Failed to marshal user object")
	}

	os.WriteFile(db.path, data, 0666)
	return user, nil
}
