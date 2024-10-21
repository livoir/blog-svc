package repository

import (
	"database/sql"
	"errors"
	"fmt"
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

func (r *postVersionRepository) Create(tx domain.Transaction, postVersion *domain.PostVersion) error {
	postVersion.ID = ulid.New()
	sqlTx := tx.GetTx()
	query := `INSERT INTO post_versions (id, version_number, post_id, created_at, title, content) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	err := sqlTx.QueryRow(query, postVersion.ID, postVersion.VersionNumber, postVersion.PostID, postVersion.CreatedAt, postVersion.Title, postVersion.Content).Scan(&postVersion.ID)
	if err != nil {
		logger.Log.Error("Failed to create post version", zap.Error(err))
		return err
	}
	return nil
}

func (r *postVersionRepository) GetByID(id string) (*domain.PostVersion, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *postVersionRepository) Update(tx domain.Transaction, postVersion *domain.PostVersion) error {
	sqlTx := tx.GetTx()
	query := `UPDATE post_versions SET title = $2, content = $3, published_at = $4 WHERE id = $1`
	_, err := sqlTx.Exec(query, postVersion.ID, postVersion.Title, postVersion.Content, postVersion.PublishedAt)
	if err != nil {
		logger.Log.Error("Failed to update post version", zap.Error(err))
		return err
	}
	return nil
}

// GetLatestByPostIDForUpdate implements domain.PostVersionRepository.
func (r *postVersionRepository) GetLatestByPostIDForUpdate(tx domain.Transaction, postID string) (*domain.PostVersion, error) {
	sqlTx := tx.GetTx()
	postVersion := &domain.PostVersion{}
	err := sqlTx.QueryRow("SELECT id, version_number, post_id, created_at, title, content, published_at FROM post_versions WHERE post_id = $1 ORDER BY version_number DESC FOR UPDATE", postID).
		Scan(&postVersion.ID, &postVersion.VersionNumber, &postVersion.PostID, &postVersion.CreatedAt, &postVersion.Title, &postVersion.Content, &postVersion.PublishedAt)
	if err != nil {
		logger.Log.Error("Failed to get latest post version by post id for update", zap.Error(err))
		return nil, err
	}
	return postVersion, nil
}
