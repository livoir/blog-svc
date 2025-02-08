-- +goose Up
-- +goose StatementBegin
CREATE TABLE administrator_sessions (
    id VARCHAR(26) PRIMARY KEY,
    administrator_id VARCHAR(26) NOT NULL,
    hashed_token TEXT NOT NULL,
    revoked_at TIMESTAMP,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (administrator_id) REFERENCES administrators(id)
);

CREATE INDEX idx_administrator_sessions_user_id_hashed_token ON administrator_sessions (administrator_id, hashed_token);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE administrator_sessions;
-- +goose StatementEnd
