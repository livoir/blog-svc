package repository

import (
	"context"
	"database/sql"
	"errors"
	"livoir-blog/internal/domain"
	"livoir-blog/pkg/logger"
	"livoir-blog/pkg/ulid"

	"go.uber.org/zap"
)

type postRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) (domain.PostRepository, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}
	return &postRepository{db}, nil
}

func (r *postRepository) GetByID(ctx context.Context, id string) (*domain.PostWithVersion, error) {
	post := &domain.PostWithVersion{}
	err := r.db.QueryRowContext(ctx, "SELECT p.id, p.current_version_id, p.created_at, p.updated_at, pv.title, pv.content, pv.version_number FROM posts p JOIN post_versions pv ON p.current_version_id = pv.id WHERE p.id = $1", id).
		Scan(&post.ID, &post.CurrentVersionID, &post.CreatedAt, &post.UpdatedAt, &post.Title, &post.Content, &post.VersionNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		logger.Log.Error("Failed to get post by id", zap.Error(err))
		return nil, err
	}
	return post, nil
}

func (r *postRepository) Create(ctx context.Context, tx domain.Transaction, post *domain.Post) error {
	post.ID = ulid.New()
	sqlTx := tx.GetTx()
	query := `INSERT INTO posts (id, created_at, updated_at) VALUES ($1, $2, $3)`
	result, err := sqlTx.ExecContext(ctx, query, post.ID, post.CreatedAt, post.UpdatedAt)
	if err != nil {
		logger.Log.Error("Failed to create post", zap.Error(err))
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error("Failed to get rows affected", zap.Error(err))
		return err
	}
	if rowsAffected == 0 {
		return errors.New("failed to create post")
	}
	return nil
}

func (r *postRepository) Update(ctx context.Context, tx domain.Transaction, post *domain.Post) error {
	sqlTx := tx.GetTx()
	query := `UPDATE posts SET current_version_id = $1, updated_at = $2 WHERE id = $3`
	result, err := sqlTx.ExecContext(ctx, query, post.CurrentVersionID, post.UpdatedAt, post.ID)
	if err != nil {
		logger.Log.Error("Failed to update post", zap.Error(err))
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error("Failed to get rows affected", zap.Error(err))
		return err
	}
	if rowsAffected == 0 {
		return errors.New("failed to update post")
	}
	return nil
}

func (r *postRepository) GetByIDForUpdate(ctx context.Context, tx domain.Transaction, id string) (*domain.Post, error) {
	sqlTx := tx.GetTx()
	post := &domain.Post{}
	err := sqlTx.QueryRowContext(ctx, "SELECT id, current_version_id, created_at, updated_at, deleted_at FROM posts WHERE id = $1 FOR UPDATE", id).
		Scan(&post.ID, &post.CurrentVersionID, &post.CreatedAt, &post.UpdatedAt, &post.DeletedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Error("No post versions found for post id", zap.String("id", id))
			return nil, errors.New("post not found")
		}
		logger.Log.Error("Failed to get latest post by id for update", zap.Error(err))
		return nil, err
	}
	return post, nil
}
