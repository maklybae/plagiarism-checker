-- +goose Up
-- +goose StatementBegin
CREATE TABLE files IF NOT EXISTS (
    id   UUID PRIMARY KEY,
    path TEXT NOT NULL,
    hash TEXT NOT NULL
);
CREATE INDEX idx_files_hash ON files(hash);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_files_hash;
DROP TABLE IF EXISTS files;
-- +goose StatementEnd
