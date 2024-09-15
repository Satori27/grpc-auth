-- +goose Up
-- +goose StatementBegin
	ALTER TABLE apps DROP CONSTRAINT apps_secret_key;
	ALTER TABLE apps DROP CONSTRAINT apps_secret_key1;

-- goose -dir migrations postgres "postgresql://grpc:grpc@localhost:5432/grpc_auth?sslmode=disable" up
-- postgresql://grpc:grpc@localhost:5432/grpc_auth
-- +goose StatementEnd



-- +goose Down
-- +goose StatementBegin
ALTER TABLE apps ADD CONSTRAINT apps_secret_key UNIQUE (secret);
ALTER TABLE apps ADD CONSTRAINT apps_secret_key1 UNIQUE (secret);

-- +goose StatementEnd
