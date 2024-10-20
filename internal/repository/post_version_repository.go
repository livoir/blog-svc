package repository

import (
	"database/sql"
	"fmt"
	"livoir-blog/internal/domain"
)

type postVersionRepository struct {
	db *sql.DB
}

func NewPostVersionRepository(db *sql.DB) domain.PostVersionRepository {
	return &postVersionRepository{db: db}
}

func (r *postVersionRepository) Create(tx domain.Transaction, postVersion *domain.PostVersion) error {
	sqlTx := tx.GetTx()
	query := `INSERT INTO post_versions (version_number, post_id, created_at, title, content) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := sqlTx.QueryRow(query, postVersion.VersionNumber, postVersion.PostID, postVersion.CreatedAt, postVersion.Title, postVersion.Content).Scan(&postVersion.ID)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (r *postVersionRepository) GetByID(id int64) (*domain.PostVersion, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *postVersionRepository) Update(tx domain.Transaction, postVersion *domain.PostVersion) error {
	return fmt.Errorf("not implemented")
}
