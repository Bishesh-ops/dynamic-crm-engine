package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
}

func (db *DB) CreateUser(ctx context.Context, email, passwordHash string) (string, error) {
	var id string
	query := `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`

	err := db.Pool.QueryRow(ctx, query, email, passwordHash).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}
	return id, nil
}

func (db *DB) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	query := `SELECT id, email, password_hash, role FROM users WHERE email = $1`

	err := db.Pool.QueryRow(ctx, query, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &u, nil
}
