package repository

import (
	"context"
	"database/sql"
	"livoir-blog/internal/domain"
	"livoir-blog/pkg/common"
	"livoir-blog/pkg/logger"

	"go.uber.org/zap"
)

type AdministratorRepositoryImpl struct {
	db *sql.DB
}

func NewAdministratorRepository(db *sql.DB) (domain.AdministratorRepository, error) {
	if db == nil {
		logger.Log.Error("Database connection is nil")
		return nil, common.ErrInternalServerError
	}

	return &AdministratorRepositoryImpl{db: db}, nil
}

func (r *AdministratorRepositoryImpl) FindByEmail(ctx context.Context, email string) (*domain.Administrator, error) {
	query := `SELECT id, full_name, email, password_hash, created_at, updated_at FROM administrators WHERE email = $1`
	row := r.db.QueryRow(query, email)

	admin := domain.Administrator{}
	err := row.Scan(&admin.ID, &admin.FullName, &admin.Email, &admin.PasswordHash, &admin.CreatedAt, &admin.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.ErrUserNotFound
		}
		logger.Log.Error("failed to get administrator by email", zap.Error(err))
		return nil, err
	}

	return &admin, nil
}
