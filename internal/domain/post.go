package domain

import "time"

type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type PostRepository interface {
	GetByID(id int64) (*Post, error)
	Create(post *Post) error
}

type PostUsecase interface {
	GetByID(id int64) (*Post, error)
	Create(post *Post) error
}
