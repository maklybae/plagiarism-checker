-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS texts (
    id UUID PRIMARY KEY,
    lines INTEGER NOT NULL,
    words INTEGER NOT NULL,
    chars INTEGER NOT NULL,
    wordcloud_path TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS texts;
-- +goose StatementEnd
