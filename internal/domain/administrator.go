package domain

import (
	"context"
	"time"
)

type Administrator struct {
	ID           string    `json:"id"`
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AdministratorRepository interface {
	FindByEmail(ctx context.Context, email string) (*Administrator, error)
	Insert(ctx context.Context, administrator *Administrator) error
}
