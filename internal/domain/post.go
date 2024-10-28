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
type PostDetail struct {
	Post
	Title         string     `json:"title"`
	Content       string     `json:"content"`
	VersionNumber int64      `json:"version_number"`
	Categories    []Category `json:"categories"`
}

type PostDetailDTO struct {
	PostDetail
}

type PostRepository interface {
	GetByID(ctx context.Context, id string) (*PostDetail, error)
	Create(ctx context.Context, tx Transaction, post *Post) error
	Update(ctx context.Context, tx Transaction, post *Post) error
	GetByIDForUpdate(ctx context.Context, tx Transaction, id string) (*Post, error)
}

type PostUsecase interface {
	GetByID(ctx context.Context, id string) (*PostDetailDTO, error)
	Create(ctx context.Context, post *CreatePostDTO) (*PostResponseDTO, error)
	Update(ctx context.Context, id string, post *UpdatePostDTO) (*PostResponseDTO, error)
	Publish(ctx context.Context, id string) (*PublishResponseDTO, error)
	DeletePostVersionByPostID(ctx context.Context, id string) error
}

type PostResponseDTO struct {
	PostID        string `json:"post_id"`
	PostVersionID string `json:"post_version_id"`
	Title         string `json:"title"`
	Content       string `json:"content"`
}

type PublishResponseDTO struct {
	PostID      string     `json:"post_id"`
	PublishedAt *time.Time `json:"published_at"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
}
