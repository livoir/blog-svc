package repository

import (
	"context"
	"database/sql"
	"livoir-blog/internal/domain"
	"livoir-blog/pkg/common"
	"livoir-blog/pkg/logger"
	"net/http"

	"github.com/lib/pq"
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
	if email == "" {
		return nil, common.NewCustomError(http.StatusBadRequest, "email is required")
	}
	query := `SELECT id, full_name, email, password_hash, created_at, updated_at FROM administrators WHERE email = $1`
	row := r.db.QueryRowContext(ctx, query, email)

	admin := domain.Administrator{}
	err := row.Scan(&admin.ID, &admin.FullName, &admin.Email, &admin.PasswordHash, &admin.CreatedAt, &admin.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.ErrUserNotFound
		}
		logger.Log.Error("failed to get administrator by email", zap.Error(err), zap.String("email", email))
		return nil, err
	}

	return &admin, nil
}

func (r *AdministratorRepositoryImpl) Insert(ctx context.Context, administrator *domain.Administrator) error {
	query := `INSERT INTO administrators (id, full_name, email, password_hash, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
	res, err := r.db.ExecContext(ctx, query, administrator.ID, administrator.FullName, administrator.Email, administrator.PasswordHash, administrator.CreatedAt, administrator.UpdatedAt)
	if err != nil {
		logger.Log.Error("failed to insert administrator", zap.Error(err), zap.String("email", administrator.Email))
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" { // PostgreSQL unique violation code
			return common.NewCustomError(http.StatusConflict, "administrator with this email already exists")
		}
		logger.Log.Error("failed to get rows affected", zap.Error(err))
		return err
	}

	if rowsAffected == 0 {
		return common.NewCustomError(http.StatusInternalServerError, "no rows affected")
	}

	return nil
}
