-- +goose Up
-- +goose StatementBegin
CREATE TABLE administrator (
    id VARCHAR(26) PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP DEFAULT NULL
);
CREATE UNIQUE INDEX idx_administrator_email ON administrator(email);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS administrator;
-- +goose StatementEnd
