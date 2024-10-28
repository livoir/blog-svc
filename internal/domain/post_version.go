package domain

import (
	"context"
	"time"
)

type PostVersion struct {
	ID            string     `json:"id"`
	VersionNumber int64      `json:"version_number"`
	PostID        string     `json:"post_id"`
	Title         string     `json:"title"`
	Content       string     `json:"content"`
	PublishedAt   *time.Time `json:"published_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

type PostVersionRepository interface {
	Create(ctx context.Context, tx Transaction, postVersion *PostVersion) error
	Update(ctx context.Context, tx Transaction, postVersion *PostVersion) error
	GetLatestByPostIDForUpdate(ctx context.Context, tx Transaction, postID string) (*PostVersion, error)
	Delete(ctx context.Context, tx Transaction, id string) error
	GetByID(ctx context.Context, id string) (*PostVersion, error)
}
