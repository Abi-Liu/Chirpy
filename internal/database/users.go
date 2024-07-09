package database

import (
	"encoding/json"
	"errors"
	"os"
	"time"
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

func (db *DB) UpdateCredentials(id int, email, password string) (User, error) {
	file, err := db.ReadFile()
	if err != nil {
		return User{}, err
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		return User{}, err
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	updatedUser := User{
		ID:       id,
		Email:    email,
		Password: hashedPassword,
	}

	for k, v := range file.Users {
		if v.ID == id {
			delete(file.Users, k)
			file.Users[email] = updatedUser
		}
	}

	data, err := json.Marshal(file)
	if err != nil {
		return User{}, err
	}

	os.WriteFile(db.path, data, 0666)

	return updatedUser, nil
}

func (db *DB) UpdateRefreshToken(id int, token string) error {
	file, err := db.ReadFile()
	if err != nil {
		return err
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	tokenStruct := Token{
		Token:     token,
		ID:        id,
		ExpiresAt: time.Now().UTC().Add(60 * 24 * time.Hour),
	}

	file.Tokens[token] = tokenStruct

	data, err := json.Marshal(file)
	if err != nil {
		return err
	}

	os.WriteFile(db.path, data, 0666)

	return nil
}

func (db *DB) LookupToken(tokenStr string) (Token, error) {
	file, err := db.ReadFile()
	if err != nil {
		return Token{}, err
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	token, ok := file.Tokens[tokenStr]

	if !ok {
		return Token{}, errors.New("Token not found")
	}

	return token, nil
}

func CheckTokenExpiration(token Token) error {
	if token.ExpiresAt.Before(time.Now().UTC()) {
		return errors.New("Token is expired")
	}

	return nil
}
