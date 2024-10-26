package domain

import (
	"context"
	"time"
)

type Category struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PostVersionCategory struct {
	PostVersionID string    `json:"post_version_id"`
	CategoryID    string    `json:"category_id"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateCategoryDTO struct {
	Name string `json:"name"`
}

type CategoryResponseDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CategoryRepository interface {
	Create(ctx context.Context, tx Transaction, category *Category) error
}

type CategoryUsecase interface {
	Create(ctx context.Context, request *CreateCategoryDTO) (*CategoryResponseDTO, error)
}
