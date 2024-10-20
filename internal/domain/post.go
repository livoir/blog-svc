package domain

import (
	"time"
)

type CreatePostDTO struct {
	PostId  int64  `json:"post_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type UpdatePostDTO struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}
type PostDTO struct {
	ID               int64  `json:"id"`
	Title            string `json:"title"`
	Content          string `json:"content"`
	CurrentVersionID int64  `json:"current_version_id"`
}

type Post struct {
	ID               int64      `json:"id"`
	CurrentVersionID int64      `json:"current_version_id"`
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
	GetByID(id int64) (*PostWithVersion, error)
	Create(post *Post) error
	Update(post *Post) error
}

type PostUsecase interface {
	GetByID(id int64) (*PostWithVersion, error)
	Create(post *CreatePostDTO) error
}

type PostVersion struct {
	ID            int64     `json:"id"`
	VersionNumber int64     `json:"version_number"`
	PostID        int64     `json:"post_id"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	PublishedAt   time.Time `json:"published_at"`
	CreatedAt     time.Time `json:"created_at"`
}

type PostVersionRepository interface {
	GetByID(id int64) (*PostVersion, error)
	Create(postVersion *PostVersion) error
	Update(postVersion *PostVersion) error
}
