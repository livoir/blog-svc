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

type AdministratorSessionRepository struct {
	db *sql.DB
}

func NewAdministratorSessionRepository(db *sql.DB) (domain.AdministratorSessionRepository, error) {
	return &AdministratorSessionRepository{db: db}, nil
}

func (a *AdministratorSessionRepository) Insert(ctx context.Context, tx domain.Transaction, session *domain.AdministratorSession) error {
	sqlTx := tx.GetTx()
	session.ID = ulid.New()
	query := `INSERT INTO administrator_sessions (id, administrator_id, encrypted_token, ip_address, user_agent) VALUES ($1, $2, $3, $4, $5)`
	res, err := sqlTx.ExecContext(ctx, query, session.ID, session.AdministratorID, session.EncryptedToken, session.IpAddress, session.UserAgent)
	if err != nil {
		logger.Log.Error("Failed to save administrator session", zap.Error(err))
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		logger.Log.Error("Failed to get rows affected", zap.Error(err))
		return err
	}
	if rowsAffected == 0 {
		return common.NewCustomError(http.StatusInternalServerError, "Failed to save administrator session")
	}
	return nil
}
