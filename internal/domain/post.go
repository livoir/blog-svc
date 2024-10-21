package domain

import (
	"time"
)

type CreatePostDTO struct {
	PostId  string `json:"post_id"`
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
	ID               string     `json:"id"`
	CurrentVersionID string     `json:"current_version_id"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at"`
}
type PostWithVersion struct {
	Post
	Title   string `json:"title"`
	Content string `json:"content"`
}
type PostRepository interface {
	GetByID(id string) (*PostWithVersion, error)
	Create(tx Transaction, post *Post) error
	Update(tx Transaction, post *Post) error
	GetByIDForUpdate(tx Transaction, id string) (*Post, error)
}

type PostUsecase interface {
	GetByID(id string) (*PostWithVersion, error)
	Create(post *CreatePostDTO) error
	Update(id string, post *UpdatePostDTO) error
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
	GetByID(id string) (*PostVersion, error)
	Create(tx Transaction, postVersion *PostVersion) error
	Update(tx Transaction, postVersion *PostVersion) error
	GetLatestByPostIDForUpdate(tx Transaction, postID string) (*PostVersion, error)
}
