package repository

import (
	"context"
	"database/sql"
	"errors"
	"livoir-blog/internal/domain"
	"livoir-blog/pkg/common"
	"livoir-blog/pkg/logger"
	"livoir-blog/pkg/ulid"
	"net/http"

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
		return common.NewCustomError(http.StatusInternalServerError, "error while creating post version")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error("Failed to get rows affected", zap.Error(err))
		return common.NewCustomError(http.StatusInternalServerError, "failed to verify post version creation")
	}
	if rowsAffected == 0 {
		return common.NewCustomError(http.StatusBadRequest, "post version creation failed due to constraint violation")
	}
	return nil
}

func (r *postVersionRepository) Update(ctx context.Context, tx domain.Transaction, postVersion *domain.PostVersion) error {
	sqlTx := tx.GetTx()
	query := `UPDATE post_versions SET title = $2, content = $3, published_at = $4, version_number = $5 WHERE id = $1`
	result, err := sqlTx.ExecContext(ctx, query, postVersion.ID, postVersion.Title, postVersion.Content, postVersion.PublishedAt, postVersion.VersionNumber)
	if err != nil {
		logger.Log.Error("Failed to update post version", zap.Error(err))
		return common.NewCustomError(http.StatusBadRequest, "error while updating post version")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error("Failed to get rows affected", zap.Error(err))
		return common.NewCustomError(http.StatusInternalServerError, "failed to verify post version update")
	}
	if rowsAffected == 0 {
		return common.NewCustomError(http.StatusInternalServerError, "failed to update post version due to constraint violation")
	}
	return nil
}

func (r *postVersionRepository) GetLatestByPostIDForUpdate(ctx context.Context, tx domain.Transaction, postID string) (*domain.PostVersion, error) {
	sqlTx := tx.GetTx()
	postVersion := &domain.PostVersion{}
	err := sqlTx.QueryRowContext(ctx, "SELECT id, version_number, post_id, created_at, title, content, published_at FROM post_versions WHERE post_id = $1 ORDER BY version_number DESC LIMIT 1 FOR UPDATE", postID).
		Scan(&postVersion.ID, &postVersion.VersionNumber, &postVersion.PostID, &postVersion.CreatedAt, &postVersion.Title, &postVersion.Content, &postVersion.PublishedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Error("No post versions found for post id", zap.String("postID", postID))
			return nil, common.NewCustomError(http.StatusNotFound, "no post versions found for post id")
		}
		logger.Log.Error("Failed to get latest post version by post id for update", zap.Error(err))
		return nil, common.NewCustomError(http.StatusInternalServerError, "error while trying to get latest post version by post id for update")
	}
	return postVersion, nil
}

func (r *postVersionRepository) Delete(ctx context.Context, tx domain.Transaction, id string) error {
	sqlTx := tx.GetTx()
	result, err := sqlTx.ExecContext(ctx, "DELETE FROM post_versions WHERE id = $1", id)
	if err != nil {
		logger.Log.Error("Failed to delete post version", zap.Error(err))
		return common.NewCustomError(http.StatusBadRequest, "error while deleting post version")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error("Failed to get rows affected", zap.Error(err))
		return common.NewCustomError(http.StatusInternalServerError, "failed to verify post version deletion")
	}
	if rowsAffected == 0 {
		logger.Log.Error("Post version not found or already deleted", zap.String("id", id))
		return common.NewCustomError(http.StatusInternalServerError, "failed to delete post version due to constraint violation")
	}
	return nil
}

func (r *postVersionRepository) GetByIDForUpdate(ctx context.Context, tx domain.Transaction, id string) (*domain.PostVersion, error) {
	sqlTx := tx.GetTx()
	postVersion := &domain.PostVersion{}
	err := sqlTx.QueryRowContext(ctx, "SELECT id, version_number, post_id, created_at, title, content, published_at FROM post_versions WHERE id = $1 FOR UPDATE", id).
		Scan(&postVersion.ID, &postVersion.VersionNumber, &postVersion.PostID, &postVersion.CreatedAt, &postVersion.Title, &postVersion.Content, &postVersion.PublishedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Error("No post versions found for id", zap.String("id", id))
			return nil, common.NewCustomError(http.StatusNotFound, "no post versions found for id")
		}
		logger.Log.Error("Failed to get post version by id for update", zap.Error(err))
		return nil, common.NewCustomError(http.StatusInternalServerError, "error while trying to get post version by id for update")
	}
	return postVersion, nil
}

func (r *postVersionRepository) GetByID(ctx context.Context, id string) (*domain.PostVersion, error) {
	query := `SELECT id, version_number, post_id, created_at, title, content, published_at FROM post_versions WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	var postVersion domain.PostVersion
	err := row.Scan(&postVersion.ID, &postVersion.VersionNumber, &postVersion.PostID, &postVersion.CreatedAt, &postVersion.Title, &postVersion.Content, &postVersion.PublishedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.ErrPostVersionNotFound
		}
		logger.Log.Error("Failed to get post version by id", zap.Error(err))
		return nil, common.NewCustomError(http.StatusInternalServerError, "error while trying to get post version by id")
	}
	return &postVersion, nil
}
