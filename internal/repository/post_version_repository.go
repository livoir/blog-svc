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

type postVersionRepository struct {
	db *sql.DB
}

func NewPostVersionRepository(db *sql.DB) (domain.PostVersionRepository, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}
	return &postVersionRepository{db: db}, nil
}

func (r *postVersionRepository) Create(ctx context.Context, tx domain.Transaction, postVersion *domain.PostVersion) error {
	postVersion.ID = ulid.New()
	sqlTx := tx.GetTx()
	query := `INSERT INTO post_versions (id, version_number, post_id, created_at, title, content) VALUES ($1, $2, $3, $4, $5, $6)`
	result, err := sqlTx.ExecContext(ctx, query, postVersion.ID, postVersion.VersionNumber, postVersion.PostID, postVersion.CreatedAt, postVersion.Title, postVersion.Content)
	if err != nil {
		logger.Log.Error("Failed to create post version", zap.Error(err))
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error("Failed to get rows affected", zap.Error(err))
		return err
	}
	if rowsAffected == 0 {
		return errors.New("failed to create post version")
	}
	return nil
}

func (r *postVersionRepository) Update(ctx context.Context, tx domain.Transaction, postVersion *domain.PostVersion) error {
	sqlTx := tx.GetTx()
	query := `UPDATE post_versions SET title = $2, content = $3, published_at = $4, version_number = $5 WHERE id = $1`
	result, err := sqlTx.ExecContext(ctx, query, postVersion.ID, postVersion.Title, postVersion.Content, postVersion.PublishedAt, postVersion.VersionNumber)
	if err != nil {
		logger.Log.Error("Failed to update post version", zap.Error(err))
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error("Failed to get rows affected", zap.Error(err))
		return err
	}
	if rowsAffected == 0 {
		return errors.New("failed to update post version")
	}
	return nil
}

func (r *postVersionRepository) GetLatestByPostIDForUpdate(ctx context.Context, tx domain.Transaction, postID string) (*domain.PostVersion, error) {
	sqlTx := tx.GetTx()
	postVersion := &domain.PostVersion{}
	err := sqlTx.QueryRowContext(ctx, "SELECT id, version_number, post_id, created_at, title, content, published_at FROM post_versions WHERE post_id = $1 ORDER BY version_number DESC FOR UPDATE", postID).
		Scan(&postVersion.ID, &postVersion.VersionNumber, &postVersion.PostID, &postVersion.CreatedAt, &postVersion.Title, &postVersion.Content, &postVersion.PublishedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Error("No post versions found for post id", zap.String("postID", postID))
			return nil, errors.New("post version not found")
		}
		logger.Log.Error("Failed to get latest post version by post id for update", zap.Error(err))
		return nil, err
	}
	return postVersion, nil
}
