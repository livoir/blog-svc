package repository

import (
	"database/sql"
	"fmt"
	"livoir-blog/internal/domain"
	"livoir-blog/pkg/logger"
	"livoir-blog/pkg/ulid"

	"go.uber.org/zap"
)

type postVersionRepository struct {
	db *sql.DB
}

func NewPostVersionRepository(db *sql.DB) domain.PostVersionRepository {
	return &postVersionRepository{db: db}
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
	return fmt.Errorf("not implemented")
}
