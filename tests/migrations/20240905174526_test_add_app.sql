-- +goose Up
-- +goose StatementBegin
INSERT INTO apps (id, name, secret)
VALUES (2, 'test', 'sercet-test')
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
