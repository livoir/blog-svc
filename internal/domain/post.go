package domain

import (
	"context"
	"time"
)

type CreatePostDTO struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type UpdatePostDTO struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type PostDTO struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Content          string `json:"content"`
	CurrentVersionID string `json:"current_version_id"`
}

type Post struct {
	ID               string    `json:"id"`
	CurrentVersionID string    `json:"current_version_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
type PostWithVersion struct {
	Post
	Title         string `json:"title"`
	Content       string `json:"content"`
	VersionNumber int64  `json:"version_number"`
}
type PostRepository interface {
	GetByID(ctx context.Context, id string) (*PostWithVersion, error)
	Create(ctx context.Context, tx Transaction, post *Post) error
	Update(ctx context.Context, tx Transaction, post *Post) error
	GetByIDForUpdate(ctx context.Context, tx Transaction, id string) (*Post, error)
}

type PostUsecase interface {
	GetByID(ctx context.Context, id string) (*PostWithVersion, error)
	Create(ctx context.Context, post *CreatePostDTO) (*PostResponseDTO, error)
	Update(ctx context.Context, id string, post *UpdatePostDTO) (*PostResponseDTO, error)
}

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
}

type PostResponseDTO struct {
	PostID  string `json:"post_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}
