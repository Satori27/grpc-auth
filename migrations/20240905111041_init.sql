-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users(
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    pass_hash TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_email ON users (email);

CREATE TABLE IF NOT EXISTS apps(
    id SERIAl PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    secret TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS roles(
    user_id INTEGER,
    user_role TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
-- goose -dir migrations postgres "postgresql://grpc:grpc@localhost:5432/grpc_auth?sslmode=disable" up
-- postgresql://grpc:grpc@localhost:5432/grpc_auth
-- +goose StatementEnd



-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS apps;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
