package repository

import (
	"context"
	"database/sql"
	"livoir-blog/internal/domain"
	"livoir-blog/pkg/common"
	"livoir-blog/pkg/logger"
	"livoir-blog/pkg/ulid"
	"net/http"

	"go.uber.org/zap"
)

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) (*CategoryRepository, error) {
	if db == nil {
		return nil, common.NewCustomError(http.StatusInternalServerError, "db is nil")
	}
	return &CategoryRepository{db}, nil
}

func (r *CategoryRepository) Create(ctx context.Context, tx domain.Transaction, category *domain.Category) error {
	sqlTx := tx.GetTx()
	category.ID = ulid.New()
	query := `INSERT INTO categories (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`
	result, err := sqlTx.ExecContext(ctx, query, category.ID, category.Name, category.CreatedAt, category.UpdatedAt)
	if err != nil {
		logger.Log.Error("Failed to create category", zap.Error(err))
		return common.ErrInternalServerError
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error("Failed to get rows affected", zap.Error(err))
		return common.ErrInternalServerError
	}
	if rowsAffected == 0 {
		return common.NewCustomError(http.StatusInternalServerError, "failed to create category")
	}
	return nil
}
