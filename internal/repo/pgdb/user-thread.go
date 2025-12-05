package pgdb

import (
	"context"
	"fmt"
	"sync"

	"feedback-bot/internal/entity"
	"feedback-bot/internal/repo/repoerrors"
	"feedback-bot/pkg/postgres"
)

type RepoUserThread struct {
	*postgres.Postgres
	mu *sync.RWMutex
}

func NewRepoUserThread(pg *postgres.Postgres) *RepoUserThread {
	return &RepoUserThread{pg, &sync.RWMutex{}}
}

func (r *RepoUserThread) Add(ctx context.Context, entity entity.UserThread) error {
	sql, args, _ := r.Builder.
		Insert("user_threads").
		Columns("user_id", "thread_id").
		Values(entity.UserId, entity.ThreadId).
		ToSql()

	result, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	if result.RowsAffected() != 1 {
		return repoerrors.ErrNotInserted
	}

	return nil
}

// func (r *RepoUserThread) Edit(ctx context.Context, postId int64, fwrdId int64) error {
// 	sql, args, _ := r.Builder.
// 		Update("user_threads").
// 		Set("thread_id", fwrdId).
// 		Where("post_id = ?", postId).
// 		ToSql()

// 	result, err := r.Pool.Exec(ctx, sql, args...)
// 	if err != nil {
// 		return err
// 	}

// 	if result.RowsAffected() != 1 {
// 		return repoerrors.ErrNotInserted
// 	}

// 	return nil
// }

func (r *RepoUserThread) GetList(ctx context.Context) (map[int64]int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var list = make(map[int64]int64)

	sql, args, _ := r.Builder.
		Select("user_id, thread_id").
		From("user_threads").
		ToSql()

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return list, err
	}
	defer rows.Close()

	for rows.Next() {
		var record entity.UserThread
		err = rows.Scan(&record.UserId, &record.ThreadId)
		if err != nil {
			return list, fmt.Errorf("RepoUserThread.GetList - rows.Scan: %v", err)
		}
		list[record.UserId] = record.ThreadId
	}

	return list, nil
}
