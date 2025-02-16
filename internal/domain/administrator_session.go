package domain

import (
	"context"
	"database/sql"
)

type AdministratorSession struct {
	ID              string
	AdministratorID string
	EncryptedToken  string
	RevokedAt       sql.NullTime
	IpAddress       string
	UserAgent       string
}

type AdministratorSessionRepository interface {
	Insert(ctx context.Context, tx Transaction, session *AdministratorSession) error
	Revoke(ctx context.Context, tx Transaction, sessionID string) error
}
