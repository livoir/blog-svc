package repository

import (
	"context"
	"database/sql"
	"fmt"
	"livoir-blog/internal/domain"
	"livoir-blog/pkg/common"
	"livoir-blog/pkg/logger"
	"livoir-blog/pkg/ulid"
	"net/http"
	"strings"

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

func (r *CategoryRepository) AttachToPostVersion(ctx context.Context, tx domain.Transaction, postVersionCategories []domain.PostVersionCategory) error {
	if len(postVersionCategories) == 0 {
		return nil
	}
	sqlTx := tx.GetTx()
	valueStrings := make([]string, 0, len(postVersionCategories))
	valueArgs := make([]interface{}, 0, len(postVersionCategories)*2)
	for i, postVersionCategory := range postVersionCategories {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, postVersionCategory.PostVersionID, postVersionCategory.CategoryID)
	}
	query := fmt.Sprintf(`INSERT INTO post_version_categories (post_version_id, category_id) VALUES %s ON CONFLICT DO NOTHING`, strings.Join(valueStrings, ","))
	result, err := sqlTx.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		logger.Log.Error("Failed to attach categories to post version", zap.Error(err))
		return common.ErrInternalServerError
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error("Failed to get rows affected", zap.Error(err))
		return common.ErrInternalServerError
	}
	if rowsAffected == 0 {
		return common.NewCustomError(http.StatusInternalServerError, "failed to attach categories to post version")
	}
	return nil
}

func (r *CategoryRepository) GetByIDs(ctx context.Context, ids []string) ([]*domain.Category, error) {
	valueStrings := make([]string, 0, len(ids))
	valueArgs := make([]interface{}, 0, len(ids))
	for i, id := range ids {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d)", i+1))
		valueArgs = append(valueArgs, id)
	}
	query := fmt.Sprintf(`SELECT id, name, created_at, updated_at FROM categories WHERE id IN (%s)`, strings.Join(valueStrings, ","))
	rows, err := r.db.QueryContext(ctx, query, valueArgs...)
	fmt.Println(err)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.ErrCategoryNotFound
		}
		logger.Log.Error("Failed to get category by id", zap.Error(err))
		return nil, common.ErrInternalServerError
	}
	var categories []*domain.Category
	defer rows.Close()
	for rows.Next() {
		var category domain.Category
		err := rows.Scan(&category.ID, &category.Name, &category.CreatedAt, &category.UpdatedAt)
		if err != nil {
			logger.Log.Error("Failed to scan category", zap.Error(err))
			return nil, common.ErrInternalServerError
		}
		categories = append(categories, &category)
	}
	fmt.Println(categories)
	return categories, nil
}
