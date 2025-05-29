package postgres

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maklybae/plagiarism-checker/analysis/internal/application"
	"github.com/maklybae/plagiarism-checker/analysis/internal/domain"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db}
}

func (r *Repository) CreateTextInfo(ctx context.Context, text *domain.TextInfo) error {
	query, args, err := sq.Insert("texts").
		Columns("id", "lines", "words", "chars", "wordcloud_path").
		Values(text.ID, text.Lines, text.Words, text.Chars, text.WordcloudPath).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	_, execErr := r.db.Exec(ctx, query, args...)
	if execErr != nil {
		return fmt.Errorf("failed to execute insert query: %w", execErr)
	}

	return nil
}

func (r *Repository) GetTextInfoByID(ctx context.Context, id uuid.UUID) (*domain.TextInfo, error) {
	query, args, err := sq.Select("id", "lines", "words", "chars", "wordcloud_path").
		From("texts").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	row := r.db.QueryRow(ctx, query, args...)

	text := &domain.TextInfo{}
	if err := row.Scan(&text.ID, &text.Lines, &text.Words, &text.Chars, &text.WordcloudPath); err != nil {
		if err == pgx.ErrNoRows {
			return nil, &application.FileInfoNotFoundError{
				ID:  id,
				Err: err,
			}
		}

		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	return text, nil
}

func (r *Repository) SetWordcloudPath(ctx context.Context, id uuid.UUID, path string) error {
	query, args, err := sq.Update("texts").
		Set("wordcloud_path", path).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	_, execErr := r.db.Exec(ctx, query, args...)
	if execErr != nil {
		return fmt.Errorf("failed to execute update query: %w", execErr)
	}

	return nil
}
