package pgdb

import (
	"context"
	"errors"
	"fmt"

	"feedback-bot/internal/entity"
	"feedback-bot/internal/repo/repoerrors"
	"feedback-bot/pkg/postgres"

	"github.com/jackc/pgx"
)

type RepoMessageUpdate struct {
	*postgres.Postgres
}

func NewRepoMessageUpdate(pg *postgres.Postgres) *RepoMessageUpdate {
	return &RepoMessageUpdate{pg}
}

func (r *RepoMessageUpdate) Create(ctx context.Context, entity entity.MessageUpdate) (int, error) {
	sql, args, _ := r.Builder.
		Insert("message_updates").
		Columns("id", "message").
		Values(entity.Id, entity.Message).
		ToSql()

	result, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return 0, err
	}

	if result.RowsAffected() != 1 {
		return 0, repoerrors.ErrNotInserted
	}

	return entity.Id, nil
}

func (r *RepoMessageUpdate) GetById(ctx context.Context, id int) (entity.MessageUpdate, error) {
	sql, args, _ := r.Builder.
		Select("id, message, is_processed, created_at").
		From("message_updates").
		Where("id = ?", id).
		ToSql()

	var record entity.MessageUpdate
	err := r.Pool.QueryRow(ctx, sql, args...).Scan(
		&record.Id,
		&record.Message,
		&record.Processed,
		&record.CreatedAt,
	)

	if err == nil {
		return record, nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return entity.MessageUpdate{}, repoerrors.ErrNotFound
	}

	return entity.MessageUpdate{}, fmt.Errorf("RepoMessageUpdate.GetById - r.Pool.QueryRow: %v", err)
}

func (r *RepoMessageUpdate) GetList(ctx context.Context) ([]entity.MessageUpdate, error) {
	sql, args, _ := r.Builder.
		Select("id, message, is_processed, created_at").
		From("message_updates").
		OrderBy("id ASC").
		ToSql()

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("RepoMessageUpdate.GetList - r.Pool.Query: %v", err)
	}
	defer rows.Close()

	var records []entity.MessageUpdate
	for rows.Next() {
		var record entity.MessageUpdate
		err = rows.Scan(&record.Id, &record.Message, &record.Processed, &record.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("RepoMessageUpdate.GetList - rows.Scan: %v", err)
		}
		records = append(records, record)
	}

	return records, nil
}
