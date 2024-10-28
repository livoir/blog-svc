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

type CategoryRequestDTO struct {
	Name string `json:"name"`
}

type CategoryResponseDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AttachCategoryToPostVersionRequestDTO struct {
	PostVersionID string   `json:"post_version_id"`
	CategoryIDs   []string `json:"category_ids"`
}

type CategoryRepository interface {
	Create(ctx context.Context, tx Transaction, category *Category) error
	Update(ctx context.Context, tx Transaction, category *Category) error
	GetByIDForUpdate(ctx context.Context, tx Transaction, id string) (*Category, error)
	GetByName(ctx context.Context, name string) (*Category, error)
	AttachToPostVersion(ctx context.Context, tx Transaction, postVersionCategories []PostVersionCategory) error
	GetByIDs(ctx context.Context, ids []string) ([]*Category, error)
}

type CategoryUsecase interface {
	Create(ctx context.Context, request *CategoryRequestDTO) (*CategoryResponseDTO, error)
	Update(ctx context.Context, id string, request *CategoryRequestDTO) (*CategoryResponseDTO, error)
	AttachToPostVersion(ctx context.Context, request *AttachCategoryToPostVersionRequestDTO) error
}
