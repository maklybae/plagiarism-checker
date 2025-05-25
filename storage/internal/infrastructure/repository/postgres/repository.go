package postgres

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maklybae/plagiarism-checker/storage/internal/domain"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db}
}

func (r *Repository) CreateFile(ctx context.Context, file *domain.FileInfo) error {
	query, args, err := sq.Insert("files").
		Columns("id", "path", "hash").
		Values(file.ID, file.Path, file.Hash).
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

func (r *Repository) GetFileByID(ctx context.Context, id uuid.UUID) (file *domain.FileInfo, err error) {
	query, args, err := sq.Select("id", "path", "hash").
		From("files").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	row := r.db.QueryRow(ctx, query, args...)

	file = &domain.FileInfo{}
	if err := row.Scan(&file.ID, &file.Path, &file.Hash); err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	return file, nil
}

func (r *Repository) GetFilesByHash(ctx context.Context, hash string) ([]*domain.FileInfo, error) {
	query, args, err := sq.Select("id", "path", "hash").
		From("files").
		Where(sq.Eq{"hash": hash}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute select query: %w", err)
	}
	defer rows.Close()

	var files []*domain.FileInfo

	for rows.Next() {
		file := &domain.FileInfo{}
		if err := rows.Scan(&file.ID, &file.Path, &file.Hash); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return files, nil
}
