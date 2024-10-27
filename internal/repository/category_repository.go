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

func NewCategoryRepository(db *sql.DB) (domain.CategoryRepository, error) {
	if db == nil {
		return nil, common.NewCustomError(http.StatusInternalServerError, "db is nil")
	}
	return &CategoryRepository{db: db}, nil
}

func (r *CategoryRepository) GetByName(ctx context.Context, name string) (*domain.Category, error) {
	query := `SELECT id, name, created_at, updated_at FROM categories WHERE name = $1`
	row := r.db.QueryRowContext(ctx, query, name)
	var category domain.Category
	err := row.Scan(&category.ID, &category.Name, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.ErrCategoryNotFound
		}
		logger.Log.Error("Failed to get category by id", zap.Error(err))
		return nil, common.ErrInternalServerError
	}
	return &category, nil
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

func (r *CategoryRepository) Update(ctx context.Context, tx domain.Transaction, category *domain.Category) error {
	sqlTx := tx.GetTx()
	query := `UPDATE categories SET name = $1, updated_at = $2 WHERE id = $3`
	result, err := sqlTx.ExecContext(ctx, query, category.Name, category.UpdatedAt, category.ID)
	if err != nil {
		logger.Log.Error("Failed to update category", zap.Error(err))
		return common.ErrInternalServerError
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error("Failed to get rows affected", zap.Error(err))
		return common.ErrInternalServerError
	}
	if rowsAffected == 0 {
		return common.NewCustomError(http.StatusInternalServerError, "failed to update category")
	}
	return nil
}

func (r *CategoryRepository) GetByIDForUpdate(ctx context.Context, tx domain.Transaction, id string) (*domain.Category, error) {
	sqlTx := tx.GetTx()
	query := `SELECT id, name, created_at, updated_at FROM categories WHERE id = $1 FOR UPDATE`
	row := sqlTx.QueryRowContext(ctx, query, id)
	var category domain.Category
	err := row.Scan(&category.ID, &category.Name, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.ErrCategoryNotFound
		}
		logger.Log.Error("Failed to get category by id", zap.Error(err))
		return nil, common.ErrInternalServerError
	}
	return &category, nil
}
