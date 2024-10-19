package repository

import (
	"database/sql"
	"livoir-blog/internal/domain"
)

type postRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) domain.PostRepository {
	return &postRepository{db}
}

func (r *postRepository) GetByID(id int64) (*domain.Post, error) {
	post := &domain.Post{}
	err := r.db.QueryRow("SELECT id, title, content FROM posts WHERE id = $1", id).Scan(&post.ID, &post.Title, &post.Content)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return post, nil
}
