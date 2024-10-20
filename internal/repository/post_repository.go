package repository

import (
	"database/sql"
	"fmt"
	"livoir-blog/internal/domain"
	"livoir-blog/pkg/ulid"
)

type postRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) domain.PostRepository {
	return &postRepository{db}
}

func (r *postRepository) GetByID(id string) (*domain.PostWithVersion, error) {
	post := &domain.PostWithVersion{}
	err := r.db.QueryRow("SELECT p.id, p.current_version_id, p.created_at, p.updated_at, p.deleted_at, pv.title, pv.content FROM posts p JOIN post_versions pv ON p.current_version_id = pv.id WHERE p.id = $1", id).
		Scan(&post.ID, &post.CurrentVersionID, &post.CreatedAt, &post.UpdatedAt, &post.DeletedAt, &post.Title, &post.Content)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		fmt.Println(err)
		return nil, err
	}
	return post, nil
}

func (r *postRepository) Create(tx domain.Transaction, post *domain.Post) error {
	post.ID = ulid.New()
	sqlTx := tx.GetTx()
	query := `INSERT INTO posts (id, created_at) VALUES ($1, $2) RETURNING id`
	err := sqlTx.QueryRow(query, post.ID, post.CreatedAt).Scan(&post.ID)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return err
}

func (r *postRepository) Update(tx domain.Transaction, post *domain.Post) error {
	sqlTx := tx.GetTx()
	query := `UPDATE posts SET current_version_id = $1, updated_at = $2 WHERE id = $3 RETURNING id`
	err := sqlTx.QueryRow(query, post.CurrentVersionID, post.UpdatedAt, post.ID).Scan(&post.ID)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
