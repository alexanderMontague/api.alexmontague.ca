package repository

import (
	"database/sql"
	"errors"
	"slices"
	"time"

	"api.alexmontague.ca/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var ALLOWED_USERS = []string{
	"me@alexmontague.ca",
}

func CreateUser(email, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	if !slices.Contains(ALLOWED_USERS, email) {
		return nil, errors.New("unauthorized user " + email)
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	result, err := database.DB.Exec(`
		INSERT INTO users (email, password_hash, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`, email, string(hashedPassword), now, now)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &User{
		ID:           int(id),
		Email:        email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func GetUserByEmail(email string) (*User, error) {
	var user User
	err := database.DB.QueryRow(`
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = ?
	`, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func GetUserByID(id int) (*User, error) {
	var user User
	err := database.DB.QueryRow(`
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = ?
	`, id).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func ValidatePassword(user *User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}
